// youtube-client-b02.ts — Has try/catch but never checks response.ok or status.

class YouTubeClientB02 {
  private apiKey: string;
  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async listVideos(videoIds: string[]): Promise<VideoListResponse> {
    try {
      const response = await fetch(
        `https://www.googleapis.com/youtube/v3/videos?part=snippet,statistics&id=${videoIds.join(",")}&key=${this.apiKey}`
      );
      // BUG: Never checks response.ok or response.status
      // A 403 quota-exceeded response is parsed as if it were a valid VideoListResponse
      const data = await response.json();
      return data as VideoListResponse;
    } catch (err) {
      throw new Error(`Failed to fetch videos: ${(err as Error).message}`);
    }
  }
}

// Tests exist but never test non-200 status codes
async function testB02() {
  const client = new YouTubeClientB02("key");
  try {
    const result = await client.listVideos(["abc"]);
    // Assertions exist but only for happy path
    assert.equal(result.kind, "youtube#videoListResponse");
    assert.ok(Array.isArray(result.items));
  } catch (err) {
    // This only catches network errors, not HTTP 4xx/5xx
    console.error("Caught:", err);
  }
}
