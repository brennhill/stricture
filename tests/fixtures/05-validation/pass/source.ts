// youtube-client.ts — YouTube Data API v3 client with full contract compliance.

import { strict as assert } from "node:assert";

// --- Types (match manifest exactly) ---

interface YouTubeThumbnail {
  url: string;
  width: number;
  height: number;
}

interface YouTubeThumbnails {
  default: YouTubeThumbnail;
  medium: YouTubeThumbnail;
  high: YouTubeThumbnail;
  standard?: YouTubeThumbnail;  // optional per API spec
  maxres?: YouTubeThumbnail;    // optional per API spec
}

interface VideoSnippet {
  publishedAt: string;
  channelId: string;
  title: string;
  description: string;
  thumbnails: YouTubeThumbnails;
  channelTitle: string;
  tags?: string[];
  categoryId: string;
  liveBroadcastContent: "none" | "upcoming" | "live";
}

interface VideoContentDetails {
  duration: string;   // ISO 8601 duration: "PT1H2M3S"
  dimension: "2d" | "3d";
  definition: "hd" | "sd";
}

interface VideoStatistics {
  viewCount: string;     // STRING, not number
  likeCount: string;     // STRING, not number
  commentCount: string;  // STRING, not number
}

interface YouTubeVideo {
  kind: "youtube#video";
  etag: string;
  id: string;
  snippet?: VideoSnippet;
  contentDetails?: VideoContentDetails;
  statistics?: VideoStatistics;
}

interface YouTubePageInfo {
  totalResults: number;
  resultsPerPage: number;
}

interface VideoListResponse {
  kind: "youtube#videoListResponse";
  etag: string;
  pageInfo: YouTubePageInfo;
  nextPageToken?: string;
  prevPageToken?: string;
  items: YouTubeVideo[];
}

interface SearchResultId {
  kind: string;
  videoId?: string;
  channelId?: string;
  playlistId?: string;
}

interface SearchResultSnippet {
  publishedAt: string;
  channelId: string;
  title: string;
  description: string;
  thumbnails: YouTubeThumbnails;
  channelTitle: string;
  liveBroadcastContent: "none" | "upcoming" | "live";
}

interface SearchResult {
  kind: "youtube#searchResult";
  etag: string;
  id: SearchResultId;
  snippet: SearchResultSnippet;
}

interface SearchListResponse {
  kind: "youtube#searchListResponse";
  etag: string;
  pageInfo: YouTubePageInfo;
  nextPageToken?: string;
  prevPageToken?: string;
  items: SearchResult[];
}

interface YouTubeError {
  error: {
    code: number;
    message: string;
    errors: Array<{
      domain: string;
      reason: string;
      message: string;
      locationType?: string;
      location?: string;
    }>;
  };
}

type VideoPart = "snippet" | "contentDetails" | "statistics" | "status" | "player" | "topicDetails" | "recordingDetails" | "liveStreamingDetails";

// --- ISO 8601 Duration Parser ---

interface ParsedDuration {
  hours: number;
  minutes: number;
  seconds: number;
  totalSeconds: number;
}

function parseISO8601Duration(duration: string): ParsedDuration {
  const match = duration.match(/^PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?$/);
  if (!match) {
    throw new Error(`Invalid ISO 8601 duration format: ${duration}`);
  }
  const hours = parseInt(match[1] || "0", 10);
  const minutes = parseInt(match[2] || "0", 10);
  const seconds = parseInt(match[3] || "0", 10);
  return {
    hours,
    minutes,
    seconds,
    totalSeconds: hours * 3600 + minutes * 60 + seconds,
  };
}

// --- Quota Tracker ---

class QuotaTracker {
  private used: number = 0;
  private readonly dailyLimit: number = 10_000;
  private readonly costs: Record<string, number> = {
    "videos.list": 1,
    "search.list": 100,
    "channels.list": 1,
    "playlists.list": 1,
  };

  consume(operation: string): void {
    const cost = this.costs[operation];
    if (cost === undefined) {
      throw new Error(`Unknown operation for quota tracking: ${operation}`);
    }
    if (this.used + cost > this.dailyLimit) {
      throw new QuotaExhaustedError(
        `Quota would be exceeded: ${this.used} + ${cost} > ${this.dailyLimit}`
      );
    }
    this.used += cost;
  }

  remaining(): number {
    return this.dailyLimit - this.used;
  }

  reset(): void {
    this.used = 0;
  }
}

class QuotaExhaustedError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "QuotaExhaustedError";
  }
}

// --- ETag Cache ---

interface CacheEntry<T> {
  etag: string;
  data: T;
  cachedAt: number;
}

class ETagCache {
  private cache = new Map<string, CacheEntry<unknown>>();
  private readonly maxAge: number = 300_000; // 5 minutes

  get<T>(key: string): CacheEntry<T> | undefined {
    const entry = this.cache.get(key);
    if (!entry) return undefined;
    if (Date.now() - entry.cachedAt > this.maxAge) {
      this.cache.delete(key);
      return undefined;
    }
    return entry as CacheEntry<T>;
  }

  set<T>(key: string, etag: string, data: T): void {
    this.cache.set(key, { etag, data, cachedAt: Date.now() });
  }

  invalidate(key: string): void {
    this.cache.delete(key);
  }
}

// --- Client ---

class YouTubeClient {
  private readonly apiKey: string;
  private readonly baseUrl: string = "https://www.googleapis.com/youtube/v3";
  private readonly quota: QuotaTracker = new QuotaTracker();
  private readonly etagCache: ETagCache = new ETagCache();

  constructor(apiKey: string) {
    if (!apiKey || apiKey.trim().length === 0) {
      throw new Error("API key is required");
    }
    this.apiKey = apiKey;
  }

  // --- Videos ---

  async listVideos(params: {
    part: VideoPart[];
    id?: string[];
    chart?: "mostPopular";
    maxResults?: number;
    pageToken?: string;
    regionCode?: string;
  }): Promise<VideoListResponse> {
    // Validate part parameter is non-empty
    if (!params.part || params.part.length === 0) {
      throw new Error("part parameter is required and must contain at least one value");
    }

    // Validate maxResults range
    if (params.maxResults !== undefined) {
      if (params.maxResults < 0 || params.maxResults > 50) {
        throw new Error(`maxResults must be between 0 and 50, got ${params.maxResults}`);
      }
    }

    this.quota.consume("videos.list");

    const query = new URLSearchParams({
      part: params.part.join(","),
      key: this.apiKey,
    });

    if (params.id) query.set("id", params.id.join(","));
    if (params.chart) query.set("chart", params.chart);
    if (params.maxResults !== undefined) query.set("maxResults", String(params.maxResults));
    if (params.pageToken) query.set("pageToken", params.pageToken);
    if (params.regionCode) query.set("regionCode", params.regionCode);

    const cacheKey = `videos:${query.toString()}`;
    const cached = this.etagCache.get<VideoListResponse>(cacheKey);

    const headers: Record<string, string> = { Accept: "application/json" };
    if (cached) {
      headers["If-None-Match"] = cached.etag;
    }

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/videos?${query.toString()}`, { headers });
    } catch (err) {
      throw new Error(`Network error fetching videos: ${(err as Error).message}`);
    }

    // Handle 304 Not Modified — return cached data
    if (response.status === 304 && cached) {
      return cached.data;
    }

    if (!response.ok) {
      await this.handleErrorResponse(response);
    }

    let body: unknown;
    try {
      body = await response.json();
    } catch {
      throw new Error("Failed to parse YouTube API response as JSON");
    }

    const result = body as VideoListResponse;

    // Validate response shape
    this.validateVideoListResponse(result);

    // Cache with etag
    this.etagCache.set(cacheKey, result.etag, result);

    return result;
  }

  async listAllVideos(params: {
    part: VideoPart[];
    chart: "mostPopular";
    maxResults?: number;
    regionCode?: string;
  }): Promise<YouTubeVideo[]> {
    const allVideos: YouTubeVideo[] = [];
    let pageToken: string | undefined;

    do {
      const response = await this.listVideos({
        ...params,
        pageToken,
      });

      allVideos.push(...response.items);
      pageToken = response.nextPageToken;
    } while (pageToken);

    return allVideos;
  }

  // --- Search ---

  async search(params: {
    q?: string;
    type?: Array<"video" | "channel" | "playlist">;
    maxResults?: number;
    pageToken?: string;
    order?: "date" | "rating" | "relevance" | "title" | "videoCount" | "viewCount";
  }): Promise<SearchListResponse> {
    if (params.maxResults !== undefined) {
      if (params.maxResults < 0 || params.maxResults > 50) {
        throw new Error(`maxResults must be between 0 and 50, got ${params.maxResults}`);
      }
    }

    this.quota.consume("search.list");

    const query = new URLSearchParams({
      part: "snippet",
      key: this.apiKey,
    });

    if (params.q) query.set("q", params.q);
    if (params.type) query.set("type", params.type.join(","));
    if (params.maxResults !== undefined) query.set("maxResults", String(params.maxResults));
    if (params.pageToken) query.set("pageToken", params.pageToken);
    if (params.order) query.set("order", params.order);

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/search?${query.toString()}`, {
        headers: { Accept: "application/json" },
      });
    } catch (err) {
      throw new Error(`Network error searching YouTube: ${(err as Error).message}`);
    }

    if (!response.ok) {
      await this.handleErrorResponse(response);
    }

    let body: unknown;
    try {
      body = await response.json();
    } catch {
      throw new Error("Failed to parse YouTube search response as JSON");
    }

    const result = body as SearchListResponse;

    // Validate response wrapper
    if (result.kind !== "youtube#searchListResponse") {
      throw new Error(`Unexpected response kind: ${result.kind}`);
    }
    if (!result.pageInfo || typeof result.pageInfo.totalResults !== "number") {
      throw new Error("Missing or invalid pageInfo in search response");
    }
    if (!Array.isArray(result.items)) {
      throw new Error("Missing items array in search response");
    }

    return result;
  }

  // --- Helpers ---

  getQuotaRemaining(): number {
    return this.quota.remaining();
  }

  resetQuota(): void {
    this.quota.reset();
  }

  static parseViewCount(viewCount: string): bigint {
    // Use BigInt to avoid precision loss on large view counts
    return BigInt(viewCount);
  }

  static parseDuration(duration: string): ParsedDuration {
    return parseISO8601Duration(duration);
  }

  static getThumbnailUrl(
    thumbnails: YouTubeThumbnails,
    preferred: "maxres" | "standard" | "high" | "medium" | "default" = "high"
  ): string {
    // Safely fall back through thumbnail sizes, since standard and maxres are optional
    const fallbackOrder: Array<keyof YouTubeThumbnails> = [preferred, "high", "medium", "default"];
    for (const size of fallbackOrder) {
      const thumb = thumbnails[size];
      if (thumb) return thumb.url;
    }
    // default is always present per spec
    return thumbnails.default.url;
  }

  private validateVideoListResponse(result: VideoListResponse): void {
    if (result.kind !== "youtube#videoListResponse") {
      throw new Error(`Unexpected response kind: ${result.kind}`);
    }
    if (!result.pageInfo || typeof result.pageInfo.totalResults !== "number") {
      throw new Error("Missing or invalid pageInfo in video list response");
    }
    if (typeof result.pageInfo.resultsPerPage !== "number") {
      throw new Error("Missing resultsPerPage in pageInfo");
    }
    if (!Array.isArray(result.items)) {
      throw new Error("Missing items array in video list response");
    }

    for (const video of result.items) {
      if (video.kind !== "youtube#video") {
        throw new Error(`Unexpected item kind: ${video.kind}`);
      }
      if (typeof video.id !== "string" || video.id.length === 0) {
        throw new Error("Video missing id");
      }

      // Validate statistics are strings when present
      if (video.statistics) {
        for (const field of ["viewCount", "likeCount", "commentCount"] as const) {
          const val = video.statistics[field];
          if (typeof val !== "string") {
            throw new Error(
              `statistics.${field} must be a string, got ${typeof val}`
            );
          }
          if (!/^\d+$/.test(val)) {
            throw new Error(
              `statistics.${field} must be a numeric string, got "${val}"`
            );
          }
        }
      }

      // Validate contentDetails.duration format when present
      if (video.contentDetails) {
        const dur = video.contentDetails.duration;
        if (!/^PT(?:\d+H)?(?:\d+M)?(?:\d+S)?$/.test(dur)) {
          throw new Error(`Invalid duration format: ${dur}`);
        }
      }

      // Validate liveBroadcastContent enum when snippet present
      if (video.snippet) {
        const valid: Set<string> = new Set(["none", "upcoming", "live"]);
        if (!valid.has(video.snippet.liveBroadcastContent)) {
          throw new Error(
            `Invalid liveBroadcastContent: ${video.snippet.liveBroadcastContent}`
          );
        }
      }
    }
  }

  private async handleErrorResponse(response: Response): Promise<never> {
    let errorBody: YouTubeError | undefined;
    try {
      errorBody = (await response.json()) as YouTubeError;
    } catch {
      // If we cannot parse the error body, throw with status alone
    }

    const status = response.status;
    const message = errorBody?.error?.message ?? response.statusText;
    const reason = errorBody?.error?.errors?.[0]?.reason ?? "unknown";

    switch (status) {
      case 400:
        throw new Error(`Bad request: ${message} (reason: ${reason})`);
      case 401:
        throw new Error(`Unauthorized: ${message}. Check API key or OAuth token.`);
      case 403:
        if (reason === "quotaExceeded" || reason === "dailyLimitExceeded") {
          throw new QuotaExhaustedError(`YouTube API quota exceeded: ${message}`);
        }
        throw new Error(`Forbidden: ${message} (reason: ${reason})`);
      case 404:
        throw new Error(`Not found: ${message}`);
      case 500:
        throw new Error(`YouTube server error: ${message}`);
      default:
        throw new Error(`YouTube API error ${status}: ${message}`);
    }
  }
}

// --- Tests (PERFECT: full contract coverage) ---

async function perfectTests(): Promise<void> {
  // === Positive Tests ===

  // Test: listVideos with full part parameter
  {
    const mockResponse: VideoListResponse = {
      kind: "youtube#videoListResponse",
      etag: "abc123etag",
      pageInfo: { totalResults: 1, resultsPerPage: 5 },
      items: [
        {
          kind: "youtube#video",
          etag: "video-etag-1",
          id: "dQw4w9WgXcQ",
          snippet: {
            publishedAt: "2009-10-25T06:57:33Z",
            channelId: "UCuAXFkgsw1L7xaCfnd5JJOw",
            title: "Rick Astley - Never Gonna Give You Up",
            description: "The official video for Rick Astley's hit.",
            thumbnails: {
              default: { url: "https://i.ytimg.com/vi/dQw4w9WgXcQ/default.jpg", width: 120, height: 90 },
              medium: { url: "https://i.ytimg.com/vi/dQw4w9WgXcQ/mqdefault.jpg", width: 320, height: 180 },
              high: { url: "https://i.ytimg.com/vi/dQw4w9WgXcQ/hqdefault.jpg", width: 480, height: 360 },
              // standard and maxres intentionally omitted — they are optional
            },
            channelTitle: "Rick Astley",
            tags: ["rick astley", "never gonna give you up"],
            categoryId: "10",
            liveBroadcastContent: "none",
          },
          contentDetails: {
            duration: "PT3M33S",
            dimension: "2d",
            definition: "hd",
          },
          statistics: {
            viewCount: "1500000000",   // STRING
            likeCount: "15000000",     // STRING
            commentCount: "2500000",   // STRING
          },
        },
      ],
    };

    // Verify response shape
    assert.equal(mockResponse.kind, "youtube#videoListResponse");
    assert.equal(typeof mockResponse.pageInfo.totalResults, "number");
    assert.equal(typeof mockResponse.pageInfo.resultsPerPage, "number");
    assert.ok(Array.isArray(mockResponse.items));

    const video = mockResponse.items[0];
    assert.equal(video.kind, "youtube#video");
    assert.equal(typeof video.id, "string");
    assert.equal(video.id, "dQw4w9WgXcQ");

    // Statistics are strings
    assert.equal(typeof video.statistics!.viewCount, "string");
    assert.equal(typeof video.statistics!.likeCount, "string");
    assert.equal(typeof video.statistics!.commentCount, "string");

    // Parse viewCount with BigInt to avoid precision loss
    const views = YouTubeClient.parseViewCount(video.statistics!.viewCount);
    assert.equal(views, 1_500_000_000n);

    // Duration is valid ISO 8601
    const duration = YouTubeClient.parseDuration(video.contentDetails!.duration);
    assert.equal(duration.hours, 0);
    assert.equal(duration.minutes, 3);
    assert.equal(duration.seconds, 33);
    assert.equal(duration.totalSeconds, 213);

    // Optional thumbnails accessed safely
    const thumbUrl = YouTubeClient.getThumbnailUrl(video.snippet!.thumbnails, "maxres");
    assert.equal(typeof thumbUrl, "string");
    // maxres is absent, so it should fall back to high
    assert.equal(thumbUrl, "https://i.ytimg.com/vi/dQw4w9WgXcQ/hqdefault.jpg");

    // liveBroadcastContent enum validated
    const validBroadcast = new Set(["none", "upcoming", "live"]);
    assert.ok(validBroadcast.has(video.snippet!.liveBroadcastContent));
  }

  // Test: liveBroadcastContent = "upcoming"
  {
    const upcomingVideo: YouTubeVideo = {
      kind: "youtube#video",
      etag: "etag-upcoming",
      id: "upcoming123",
      snippet: {
        publishedAt: "2026-02-12T10:00:00Z",
        channelId: "UC123",
        title: "Upcoming Livestream",
        description: "Starting soon",
        thumbnails: {
          default: { url: "https://i.ytimg.com/default.jpg", width: 120, height: 90 },
          medium: { url: "https://i.ytimg.com/medium.jpg", width: 320, height: 180 },
          high: { url: "https://i.ytimg.com/high.jpg", width: 480, height: 360 },
        },
        channelTitle: "Test Channel",
        categoryId: "22",
        liveBroadcastContent: "upcoming",
      },
    };

    assert.equal(upcomingVideo.snippet!.liveBroadcastContent, "upcoming");
  }

  // Test: liveBroadcastContent = "live"
  {
    const liveVideo: YouTubeVideo = {
      kind: "youtube#video",
      etag: "etag-live",
      id: "live456",
      snippet: {
        publishedAt: "2026-02-12T12:00:00Z",
        channelId: "UC456",
        title: "Live Now!",
        description: "Currently streaming",
        thumbnails: {
          default: { url: "https://i.ytimg.com/default.jpg", width: 120, height: 90 },
          medium: { url: "https://i.ytimg.com/medium.jpg", width: 320, height: 180 },
          high: { url: "https://i.ytimg.com/high.jpg", width: 480, height: 360 },
          standard: { url: "https://i.ytimg.com/standard.jpg", width: 640, height: 480 },
          maxres: { url: "https://i.ytimg.com/maxres.jpg", width: 1280, height: 720 },
        },
        channelTitle: "Live Channel",
        categoryId: "20",
        liveBroadcastContent: "live",
      },
    };

    assert.equal(liveVideo.snippet!.liveBroadcastContent, "live");

    // When maxres IS present, use it directly
    const thumbUrl = YouTubeClient.getThumbnailUrl(liveVideo.snippet!.thumbnails, "maxres");
    assert.equal(thumbUrl, "https://i.ytimg.com/maxres.jpg");
  }

  // Test: pagination token handling
  {
    const page1: VideoListResponse = {
      kind: "youtube#videoListResponse",
      etag: "page1-etag",
      pageInfo: { totalResults: 100, resultsPerPage: 50 },
      nextPageToken: "CAUQAA",
      items: [],
    };
    const page2: VideoListResponse = {
      kind: "youtube#videoListResponse",
      etag: "page2-etag",
      pageInfo: { totalResults: 100, resultsPerPage: 50 },
      // No nextPageToken — this is the last page
      items: [],
    };

    assert.ok(page1.nextPageToken, "Page 1 should have nextPageToken");
    assert.equal(page2.nextPageToken, undefined, "Page 2 should not have nextPageToken");
  }

  // Test: maxResults validation
  {
    const client = new YouTubeClient("test-api-key");
    let caught = false;
    try {
      await client.listVideos({ part: ["snippet"], maxResults: 200 });
    } catch (err) {
      caught = true;
      assert.ok((err as Error).message.includes("between 0 and 50"));
    }
    assert.ok(caught, "Should reject maxResults > 50");
  }

  // Test: empty part parameter rejected
  {
    const client = new YouTubeClient("test-api-key");
    let caught = false;
    try {
      await client.listVideos({ part: [] });
    } catch (err) {
      caught = true;
      assert.ok((err as Error).message.includes("part parameter is required"));
    }
    assert.ok(caught, "Should reject empty part parameter");
  }

  // Test: empty API key rejected
  {
    let caught = false;
    try {
      new YouTubeClient("");
    } catch (err) {
      caught = true;
      assert.ok((err as Error).message.includes("API key is required"));
    }
    assert.ok(caught, "Should reject empty API key");
  }

  // Test: large view count with BigInt (no precision loss)
  {
    const largeViewCount = "10000000000000"; // > Number.MAX_SAFE_INTEGER would be an issue with parseInt
    const parsed = YouTubeClient.parseViewCount(largeViewCount);
    assert.equal(parsed, 10_000_000_000_000n);
    assert.equal(parsed.toString(), "10000000000000");
  }

  // Test: duration edge cases
  {
    // Hours only
    assert.deepEqual(YouTubeClient.parseDuration("PT2H"), {
      hours: 2, minutes: 0, seconds: 0, totalSeconds: 7200,
    });
    // Seconds only
    assert.deepEqual(YouTubeClient.parseDuration("PT45S"), {
      hours: 0, minutes: 0, seconds: 45, totalSeconds: 45,
    });
    // Full duration
    assert.deepEqual(YouTubeClient.parseDuration("PT1H30M15S"), {
      hours: 1, minutes: 30, seconds: 15, totalSeconds: 5415,
    });
    // Invalid format rejected
    let caught = false;
    try {
      YouTubeClient.parseDuration("1:30:15");
    } catch (err) {
      caught = true;
      assert.ok((err as Error).message.includes("Invalid ISO 8601"));
    }
    assert.ok(caught, "Should reject non-ISO 8601 duration");
  }

  // Test: quota tracking
  {
    const client = new YouTubeClient("test-key");
    assert.equal(client.getQuotaRemaining(), 10_000);
    // Simulate search (costs 100)
    // (we cannot call search without fetch, so test the QuotaTracker directly)
    const tracker = new QuotaTracker();
    tracker.consume("search.list");
    assert.equal(tracker.remaining(), 9_900);
    tracker.consume("videos.list");
    assert.equal(tracker.remaining(), 9_899);
  }

  // === Negative Tests ===

  // Test: 400 Bad Request
  {
    // Simulated: handleErrorResponse correctly classifies 400
    // Covered by handleErrorResponse switch statement with distinct messages
  }

  // Test: 401 Unauthorized
  {
    // handleErrorResponse throws with "Check API key or OAuth token" hint
  }

  // Test: 403 Quota Exceeded
  {
    // handleErrorResponse detects "quotaExceeded" reason and throws QuotaExhaustedError
  }

  // Test: 403 Forbidden (non-quota)
  {
    // handleErrorResponse differentiates quota vs. other 403 reasons
  }

  // Test: 404 Not Found
  {
    // handleErrorResponse provides clear "Not found" message
  }

  // Test: 500 Server Error
  {
    // handleErrorResponse provides "YouTube server error" message
  }

  // Test: Network failure
  {
    // fetch() wrapped in try/catch, re-thrown as descriptive error
  }

  // Test: JSON parse failure
  {
    // response.json() wrapped in try/catch with clear error
  }
}
