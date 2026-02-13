// youtube-client-b12.ts — Accesses optional thumbnail.maxres.url without null check.

class YouTubeClientB12 {
  private apiKey: string;
  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async getVideoThumbnail(videoId: string): Promise<string> {
    const response = await fetch(
      `https://www.googleapis.com/youtube/v3/videos?part=snippet&id=${videoId}&key=${this.apiKey}`
    );
    if (!response.ok) throw new Error(`API error: ${response.status}`);
    const data = (await response.json()) as VideoListResponse;
    const video = data.items[0];

    // BUG: `maxres` is OPTIONAL per the API spec (manifest: required: false).
    // Many videos do not have a maxres thumbnail.
    // This crashes with: TypeError: Cannot read properties of undefined (reading 'url')
    return video.snippet!.thumbnails.maxres!.url;
  }

  async getVideoThumbnails(videoIds: string[]): Promise<Map<string, string>> {
    const response = await fetch(
      `https://www.googleapis.com/youtube/v3/videos?part=snippet&id=${videoIds.join(",")}&key=${this.apiKey}`
    );
    if (!response.ok) throw new Error(`API error: ${response.status}`);
    const data = (await response.json()) as VideoListResponse;
    const thumbnails = new Map<string, string>();

    for (const video of data.items) {
      // BUG: Same crash — standard is also optional
      const bestThumb = video.snippet!.thumbnails.standard!.url
        ?? video.snippet!.thumbnails.high.url;
      thumbnails.set(video.id, bestThumb);
    }
    return thumbnails;
  }

  async buildOGMetaTags(videoId: string): Promise<string> {
    const response = await fetch(
      `https://www.googleapis.com/youtube/v3/videos?part=snippet&id=${videoId}&key=${this.apiKey}`
    );
    if (!response.ok) throw new Error(`API error: ${response.status}`);
    const data = (await response.json()) as VideoListResponse;
    const video = data.items[0];

    // BUG: Uses non-null assertion on optional fields
    return `<meta property="og:image" content="${video.snippet!.thumbnails.maxres!.url}" />`;
  }
}

async function testB12() {
  const client = new YouTubeClientB12("test-key");

  // Test only uses videos that happen to HAVE maxres — masks the bug
  // A video uploaded at low resolution will NOT have maxres or standard thumbnails
  // This test passes but does not cover the crash scenario
}
