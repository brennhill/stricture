// youtube-client-b05.ts — Omits required `part` parameter.

class YouTubeClientB05 {
  private apiKey: string;
  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async listVideos(videoIds: string[]): Promise<VideoListResponse> {
    // BUG: `part` parameter is REQUIRED by the YouTube API but is completely omitted.
    // The API will return a 400: "No filter selected. Expected one of: chart, id, myRating"
    // or "Required parameter: part" depending on the exact request.
    const url = new URL("https://www.googleapis.com/youtube/v3/videos");
    url.searchParams.set("id", videoIds.join(","));
    url.searchParams.set("key", this.apiKey);
    // Missing: url.searchParams.set("part", "snippet,contentDetails,statistics");

    const response = await fetch(url.toString());
    if (!response.ok) {
      throw new Error(`YouTube API error: ${response.status}`);
    }
    return (await response.json()) as VideoListResponse;
  }

  async search(query: string): Promise<SearchListResponse> {
    const url = new URL("https://www.googleapis.com/youtube/v3/search");
    url.searchParams.set("q", query);
    url.searchParams.set("key", this.apiKey);
    // BUG: `part` is required for search too — always "snippet" for search results
    // Missing: url.searchParams.set("part", "snippet");

    const response = await fetch(url.toString());
    if (!response.ok) {
      throw new Error(`YouTube search error: ${response.status}`);
    }
    return (await response.json()) as SearchListResponse;
  }
}

async function testB05() {
  const client = new YouTubeClientB05("test-key");
  // Tests would fail at runtime because every call returns 400
  const result = await client.listVideos(["dQw4w9WgXcQ"]);
  assert.equal(result.kind, "youtube#videoListResponse");
}
