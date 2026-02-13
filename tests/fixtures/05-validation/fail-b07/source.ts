// youtube-client-b07.ts — viewCount stored as number; it is a string in the API contract.

interface VideoStatsB07 {
  viewCount: number;     // BUG: YouTube API returns "1234567" (string), not 1234567 (number)
  likeCount: number;     // BUG: Same — string in API, number in client type
  commentCount: number;  // BUG: Same
}

interface VideoB07 {
  kind: "youtube#video";
  id: string;
  snippet?: VideoSnippet;
  contentDetails?: {
    duration: string;
    dimension: string;    // BUG: Should be enum "2d" | "3d"
    definition: string;   // BUG: Should be enum "hd" | "sd"
  };
  statistics?: VideoStatsB07;
}

class YouTubeClientB07 {
  private apiKey: string;
  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async getVideoStats(videoId: string): Promise<VideoStatsB07> {
    const response = await fetch(
      `https://www.googleapis.com/youtube/v3/videos?part=statistics&id=${videoId}&key=${this.apiKey}`
    );
    if (!response.ok) throw new Error(`API error: ${response.status}`);
    const data = await response.json();
    const video = data.items[0];

    // BUG: Converts string viewCount to number — loses precision for large values
    return {
      viewCount: Number(video.statistics.viewCount),
      likeCount: Number(video.statistics.likeCount),
      commentCount: Number(video.statistics.commentCount),
    };
  }
}

async function testB07() {
  const client = new YouTubeClientB07("test-key");
  const stats = await client.getVideoStats("dQw4w9WgXcQ");
  // These assertions enforce the WRONG type:
  assert.equal(typeof stats.viewCount, "number");     // Passes, but the contract says string
  assert.equal(typeof stats.likeCount, "number");     // Passes, but wrong
  assert.ok(stats.viewCount > 0);                     // Works for small numbers but masks the type mismatch
}
