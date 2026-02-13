// youtube-client-b11.ts — parseInt on viewCount loses precision for very large values.

class YouTubeClientB11 {
  private apiKey: string;
  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async getVideoStats(videoId: string): Promise<{
    viewCount: number;
    likeCount: number;
    commentCount: number;
  }> {
    const response = await fetch(
      `https://www.googleapis.com/youtube/v3/videos?part=statistics&id=${videoId}&key=${this.apiKey}`
    );
    if (!response.ok) throw new Error(`API error: ${response.status}`);
    const data = (await response.json()) as VideoListResponse;
    const stats = data.items[0].statistics!;

    // BUG: parseInt loses precision for values > Number.MAX_SAFE_INTEGER (9007199254740991)
    // YouTube's most-viewed video has 15+ billion views — fits in MAX_SAFE_INTEGER today,
    // but cumulative view counts across channels or aggregated stats can exceed it.
    return {
      viewCount: parseInt(stats.viewCount, 10),
      likeCount: parseInt(stats.likeCount, 10),
      commentCount: parseInt(stats.commentCount, 10),
    };
  }

  // BUG: Aggregation across videos compounds precision loss
  async getTotalViews(videoIds: string[]): Promise<number> {
    let total = 0;
    const response = await fetch(
      `https://www.googleapis.com/youtube/v3/videos?part=statistics&id=${videoIds.join(",")}&key=${this.apiKey}`
    );
    if (!response.ok) throw new Error(`API error: ${response.status}`);
    const data = (await response.json()) as VideoListResponse;

    for (const video of data.items) {
      // BUG: Each parseInt can lose precision, and summing makes it worse
      total += parseInt(video.statistics!.viewCount, 10);
    }
    return total;
  }
}

async function testB11() {
  const client = new YouTubeClientB11("test-key");

  // Test with small numbers — passes, masking the bug
  const mockSmall = { viewCount: "1234567", likeCount: "8901", commentCount: "234" };
  assert.equal(parseInt(mockSmall.viewCount, 10), 1234567); // Correct

  // BUG: This test SHOULD exist but does not:
  const hugeCount = "10000000000000001"; // 10 quadrillion + 1
  const parsed = parseInt(hugeCount, 10);
  // parsed === 10000000000000000 (lost the trailing "1")
  // assert.equal(parsed.toString(), hugeCount); // WOULD FAIL — proves precision loss

  // PERFECT approach uses BigInt:
  // const safe = BigInt(hugeCount);
  // assert.equal(safe.toString(), "10000000000000001"); // No precision loss
}
