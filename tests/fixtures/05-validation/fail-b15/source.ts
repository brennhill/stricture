// youtube-client-b15.ts — Reads video stats then caches, but stats change between read and use.
// No etag-based cache invalidation.

interface CachedVideoData {
  video: YouTubeVideo;
  fetchedAt: number;
  // BUG: No etag stored — cannot detect staleness
}

class YouTubeClientB15 {
  private apiKey: string;
  private cache: Map<string, CachedVideoData> = new Map();
  private readonly cacheTTL = 600_000; // 10 minutes

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async getVideo(videoId: string): Promise<YouTubeVideo> {
    const cached = this.cache.get(videoId);
    if (cached && Date.now() - cached.fetchedAt < this.cacheTTL) {
      // BUG: Returns stale data without etag validation.
      // If another client updated the video (title change, new stats),
      // this returns outdated information for up to 10 minutes.
      return cached.video;
    }

    // BUG: No If-None-Match header sent — misses 304 optimization
    const response = await fetch(
      `https://www.googleapis.com/youtube/v3/videos?part=snippet,statistics&id=${videoId}&key=${this.apiKey}`
    );
    if (!response.ok) throw new Error(`API error: ${response.status}`);
    const data = (await response.json()) as VideoListResponse;
    const video = data.items[0];

    // BUG: Stores video data without its etag.
    // Cannot use conditional requests (If-None-Match) on next fetch.
    this.cache.set(videoId, {
      video,
      fetchedAt: Date.now(),
    });

    return video;
  }

  // BUG: Race condition in read-modify-write pattern
  async updateVideoCache(videoId: string): Promise<YouTubeVideo> {
    // Step 1: Read current data
    const current = await this.getVideo(videoId);

    // Step 2: Some business logic runs here (takes time)
    await this.processVideoAnalytics(current);

    // Step 3: Read again to "refresh" cache
    // BUG: Between step 1 and step 3, the video data may have changed.
    // The analytics from step 2 were based on step 1's data,
    // but step 3's data is now different. The cache now has step 3's data
    // but the analytics were computed from step 1's data.
    // This is a TOCTOU (time-of-check-to-time-of-use) race.
    const refreshed = await this.getVideo(videoId);
    return refreshed;
  }

  // BUG: Concurrent callers can cause redundant fetches and cache overwrites
  async getMultipleVideos(videoIds: string[]): Promise<YouTubeVideo[]> {
    // Multiple concurrent calls to getVideo for the same ID can:
    // 1. All miss the cache simultaneously
    // 2. All fetch from the API (wasting quota)
    // 3. Overwrite each other's cache entries
    // No deduplication or locking mechanism
    return Promise.all(videoIds.map((id) => this.getVideo(id)));
  }

  private async processVideoAnalytics(video: YouTubeVideo): Promise<void> {
    // Simulates time-consuming processing
    await new Promise((resolve) => setTimeout(resolve, 5000));
    // Analytics computed here are based on potentially stale data
  }
}

async function testB15() {
  const client = new YouTubeClientB15("test-key");

  // Test: basic caching works
  const video1 = await client.getVideo("abc");
  const video2 = await client.getVideo("abc");
  // BUG: Only tests that caching returns a result, not that it handles staleness

  // BUG: No tests for:
  //   - Etag-based conditional requests (If-None-Match / 304)
  //   - Cache invalidation when data changes
  //   - Concurrent access to the same cache key
  //   - TOCTOU race in updateVideoCache
  //   - Quota waste from duplicate concurrent fetches
}
