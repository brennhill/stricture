// youtube-client-b01.ts — No error handling anywhere.

class YouTubeClientB01 {
  private apiKey: string;
  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async listVideos(videoIds: string[]): Promise<any> {
    // BUG: No try/catch around fetch — network errors crash the caller
    // BUG: No error handling for non-200 responses
    // BUG: No error handling for JSON parse failures
    const response = await fetch(
      `https://www.googleapis.com/youtube/v3/videos?part=snippet&id=${videoIds.join(",")}&key=${this.apiKey}`
    );
    const data = await response.json();
    return data;
  }

  async search(query: string): Promise<any> {
    // BUG: Same pattern — bare fetch with no error handling
    const response = await fetch(
      `https://www.googleapis.com/youtube/v3/search?part=snippet&q=${query}&key=${this.apiKey}`
    );
    return response.json();
  }
}

// Test has no error-path tests at all
async function testB01() {
  const client = new YouTubeClientB01("key");
  const videos = await client.listVideos(["dQw4w9WgXcQ"]);
  console.log("Got videos:", videos.items.length);
}
