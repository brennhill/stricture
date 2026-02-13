// youtube-client-b13.ts — No tracking of API quota usage; hits 403 unexpectedly.

class YouTubeClientB13 {
  private apiKey: string;
  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  // BUG: No quota tracking. The YouTube API has a 10,000 unit/day limit.
  // search.list costs 100 units per call.
  // Calling search 100 times exhausts the daily quota.
  // There is no pre-check, no tracking, and no graceful degradation.

  async search(query: string, maxResults: number = 25): Promise<SearchListResponse> {
    // BUG: No quota check before making the request.
    // After 100 searches, every subsequent call returns 403.
    const response = await fetch(
      `https://www.googleapis.com/youtube/v3/search?part=snippet&q=${encodeURIComponent(query)}&maxResults=${maxResults}&key=${this.apiKey}`
    );
    if (!response.ok) {
      // Handles 403 but does not distinguish quota from other 403 reasons
      throw new Error(`YouTube API error: ${response.status}`);
    }
    return (await response.json()) as SearchListResponse;
  }

  async listVideos(videoIds: string[]): Promise<VideoListResponse> {
    // BUG: No quota tracking for video list calls (1 unit each)
    const response = await fetch(
      `https://www.googleapis.com/youtube/v3/videos?part=snippet,statistics&id=${videoIds.join(",")}&key=${this.apiKey}`
    );
    if (!response.ok) {
      throw new Error(`YouTube API error: ${response.status}`);
    }
    return (await response.json()) as VideoListResponse;
  }

  // BUG: This method can easily exhaust quota — each page costs 100 units
  async searchAll(query: string): Promise<SearchResult[]> {
    const allResults: SearchResult[] = [];
    let pageToken: string | undefined;

    // 200 pages * 100 units = 20,000 units (2x the daily limit)
    do {
      const url = new URL("https://www.googleapis.com/youtube/v3/search");
      url.searchParams.set("part", "snippet");
      url.searchParams.set("q", query);
      url.searchParams.set("maxResults", "50");
      url.searchParams.set("key", this.apiKey);
      if (pageToken) url.searchParams.set("pageToken", pageToken);

      const response = await fetch(url.toString());
      if (!response.ok) throw new Error(`Search error: ${response.status}`);
      const data = (await response.json()) as SearchListResponse;

      allResults.push(...data.items);
      pageToken = data.nextPageToken;
      // BUG: This loop can consume 100+ quota units per iteration with no limit
    } while (pageToken);

    return allResults;
  }
}

async function testB13() {
  const client = new YouTubeClientB13("test-key");

  // Test only calls search once — quota is never a concern
  const result = await client.search("test query");
  assert.ok(result.items.length >= 0);

  // BUG: No test for:
  //   - Quota exhaustion scenario
  //   - Graceful degradation when quota is low
  //   - QuotaExhaustedError or specific 403 handling
  //   - searchAll() consuming unbounded quota
}
