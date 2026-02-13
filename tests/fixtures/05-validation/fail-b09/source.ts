// youtube-client-b09.ts — No validation on maxResults range (0-50).

class YouTubeClientB09 {
  private apiKey: string;
  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async listVideos(params: {
    part: string[];
    id?: string[];
    maxResults?: number;  // BUG: No validation that this is 0-50
  }): Promise<VideoListResponse> {
    const url = new URL("https://www.googleapis.com/youtube/v3/videos");
    url.searchParams.set("part", params.part.join(","));
    url.searchParams.set("key", this.apiKey);
    if (params.id) url.searchParams.set("id", params.id.join(","));
    // BUG: Passes maxResults=200 straight to the API without validation
    if (params.maxResults !== undefined) {
      url.searchParams.set("maxResults", String(params.maxResults));
    }

    const response = await fetch(url.toString());
    if (!response.ok) throw new Error(`API error: ${response.status}`);
    return (await response.json()) as VideoListResponse;
  }

  async search(params: {
    q: string;
    maxResults?: number;  // BUG: No validation here either
  }): Promise<SearchListResponse> {
    const url = new URL("https://www.googleapis.com/youtube/v3/search");
    url.searchParams.set("part", "snippet");
    url.searchParams.set("q", params.q);
    url.searchParams.set("key", this.apiKey);
    // BUG: maxResults=-1 or maxResults=1000 sent without validation
    if (params.maxResults !== undefined) {
      url.searchParams.set("maxResults", String(params.maxResults));
    }

    const response = await fetch(url.toString());
    if (!response.ok) throw new Error(`Search error: ${response.status}`);
    return (await response.json()) as SearchListResponse;
  }
}

async function testB09() {
  const client = new YouTubeClientB09("test-key");

  // Test with valid maxResults — passes
  await client.listVideos({ part: ["snippet"], maxResults: 25 });

  // BUG: Never tests boundary conditions:
  //   maxResults = 0   (minimum)
  //   maxResults = 50  (maximum)
  //   maxResults = 51  (should be rejected)
  //   maxResults = -1  (should be rejected)
  //   maxResults = 200 (should be rejected)
}
