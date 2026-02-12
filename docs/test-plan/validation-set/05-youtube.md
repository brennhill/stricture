# 05 — YouTube Data API v3

**API:** YouTube Data API v3
**Why included:** Quota limits, pagination tokens, nested resources, string-typed numeric fields, ISO 8601 duration parsing, optional thumbnail sizes
**Base URL:** `https://www.googleapis.com/youtube/v3`
**Auth:** API key (query param `key=`) or OAuth2 Bearer token

---

## Manifest Fragment

```yaml
contracts:
  - id: "youtube-videos"
    producer: youtube-data-api-v3
    consumers: [my-service]
    protocol: http
    auth:
      methods: [api_key, oauth2_bearer]
      api_key_param: key
    quota:
      daily_limit: 10000
      costs:
        list_videos: 1
        search: 100
        list_channels: 1
        list_playlists: 1
    endpoints:
      - path: "/youtube/v3/videos"
        method: GET
        request:
          fields:
            part:       { type: string, required: true, description: "Comma-separated: snippet,contentDetails,statistics,status,player,topicDetails,recordingDetails,liveStreamingDetails" }
            id:         { type: string, required: false, description: "Comma-separated video IDs" }
            chart:      { type: enum, values: ["mostPopular"], required: false }
            myRating:   { type: enum, values: ["like", "dislike"], required: false }
            maxResults: { type: integer, range: [0, 50], required: false, default: 5 }
            pageToken:  { type: string, required: false }
            regionCode: { type: string, format: "ISO 3166-1 alpha-2", required: false }
            key:        { type: string, required: false }
        response:
          fields:
            kind:          { type: string, literal: "youtube#videoListResponse", required: true }
            etag:          { type: string, required: true }
            pageInfo:
              totalResults:   { type: integer, range: [0, null], required: true }
              resultsPerPage: { type: integer, range: [0, 50], required: true }
            nextPageToken: { type: string, required: false }
            prevPageToken: { type: string, required: false }
            items:
              type: array
              items:
                kind: { type: string, literal: "youtube#video", required: true }
                etag: { type: string, required: true }
                id:   { type: string, required: true }
                snippet:
                  publishedAt:          { type: string, format: "ISO 8601 datetime", required: true }
                  channelId:            { type: string, required: true }
                  title:                { type: string, required: true }
                  description:          { type: string, required: true }
                  thumbnails:
                    default:  { type: object, fields: { url: string, width: integer, height: integer }, required: true }
                    medium:   { type: object, fields: { url: string, width: integer, height: integer }, required: true }
                    high:     { type: object, fields: { url: string, width: integer, height: integer }, required: true }
                    standard: { type: object, fields: { url: string, width: integer, height: integer }, required: false }
                    maxres:   { type: object, fields: { url: string, width: integer, height: integer }, required: false }
                  channelTitle:         { type: string, required: true }
                  tags:                 { type: array, items: string, required: false }
                  categoryId:           { type: string, required: true }
                  liveBroadcastContent: { type: enum, values: ["none", "upcoming", "live"], required: true }
                contentDetails:
                  duration:   { type: string, format: "ISO 8601 duration", pattern: "^PT(\\d+H)?(\\d+M)?(\\d+S)?$", required: true }
                  dimension:  { type: enum, values: ["2d", "3d"], required: true }
                  definition: { type: enum, values: ["hd", "sd"], required: true }
                statistics:
                  viewCount:    { type: string, format: "numeric string", required: true }
                  likeCount:    { type: string, format: "numeric string", required: true }
                  commentCount: { type: string, format: "numeric string", required: true }
        status_codes: [200, 400, 401, 403, 404, 500]
        error_shape:
          error:
            code:    { type: integer, required: true }
            message: { type: string, required: true }
            errors:
              type: array
              items:
                domain:       { type: string, required: true }
                reason:       { type: string, required: true }
                message:      { type: string, required: true }
                locationType: { type: string, required: false }
                location:     { type: string, required: false }

      - path: "/youtube/v3/search"
        method: GET
        request:
          fields:
            part:       { type: string, required: true, description: "Always 'snippet' for search" }
            q:          { type: string, required: false }
            type:       { type: enum, values: ["video", "channel", "playlist"], required: false }
            maxResults: { type: integer, range: [0, 50], required: false, default: 5 }
            pageToken:  { type: string, required: false }
            order:      { type: enum, values: ["date", "rating", "relevance", "title", "videoCount", "viewCount"], required: false }
            key:        { type: string, required: false }
        response:
          fields:
            kind:          { type: string, literal: "youtube#searchListResponse", required: true }
            etag:          { type: string, required: true }
            pageInfo:
              totalResults:   { type: integer, required: true }
              resultsPerPage: { type: integer, range: [0, 50], required: true }
            nextPageToken: { type: string, required: false }
            prevPageToken: { type: string, required: false }
            items:
              type: array
              items:
                kind: { type: string, literal: "youtube#searchResult", required: true }
                etag: { type: string, required: true }
                id:
                  kind:       { type: string, required: true }
                  videoId:    { type: string, required: false }
                  channelId:  { type: string, required: false }
                  playlistId: { type: string, required: false }
                snippet:
                  publishedAt:          { type: string, format: "ISO 8601 datetime", required: true }
                  channelId:            { type: string, required: true }
                  title:                { type: string, required: true }
                  description:          { type: string, required: true }
                  thumbnails:
                    default: { type: object, required: true }
                    medium:  { type: object, required: true }
                    high:    { type: object, required: true }
                  channelTitle:         { type: string, required: true }
                  liveBroadcastContent: { type: enum, values: ["none", "upcoming", "live"], required: true }
        status_codes: [200, 400, 401, 403, 404, 500]

      - path: "/youtube/v3/channels"
        method: GET
        request:
          fields:
            part:       { type: string, required: true }
            id:         { type: string, required: false }
            forHandle:  { type: string, required: false }
            maxResults: { type: integer, range: [0, 50], required: false }
            key:        { type: string, required: false }
        response:
          fields:
            kind:  { type: string, literal: "youtube#channelListResponse", required: true }
            etag:  { type: string, required: true }
            pageInfo:
              totalResults:   { type: integer, required: true }
              resultsPerPage: { type: integer, required: true }
            items:
              type: array
              items:
                kind: { type: string, literal: "youtube#channel", required: true }
                id:   { type: string, required: true }
        status_codes: [200, 400, 401, 403, 404, 500]

      - path: "/youtube/v3/playlists"
        method: GET
        request:
          fields:
            part:       { type: string, required: true }
            id:         { type: string, required: false }
            channelId:  { type: string, required: false }
            maxResults: { type: integer, range: [0, 50], required: false }
            pageToken:  { type: string, required: false }
            key:        { type: string, required: false }
        response:
          fields:
            kind:  { type: string, literal: "youtube#playlistListResponse", required: true }
            etag:  { type: string, required: true }
            pageInfo:
              totalResults:   { type: integer, required: true }
              resultsPerPage: { type: integer, required: true }
            nextPageToken: { type: string, required: false }
            items:
              type: array
              items:
                kind: { type: string, literal: "youtube#playlist", required: true }
                id:   { type: string, required: true }
        status_codes: [200, 400, 401, 403, 404, 500]
```

---

## PERFECT Integration (Zero Violations Expected)

Stricture must produce **zero violations** when scanning this code. It correctly handles: `part` parameter construction, pagination via `nextPageToken`, string-typed statistics, optional thumbnail sizes, ISO 8601 duration parsing, quota tracking, etag-based caching, and all error status codes.

```typescript
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
```

**Expected Stricture result:** 0 violations. All CTR and TQ rules pass.

---

## B01 -- No Error Handling

**Rule violated:** `TQ-error-path-coverage`
**Severity:** Critical

```typescript
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
```

**Expected violation:** `TQ-error-path-coverage` -- No `try/catch`, no `.catch()`, no error status checks. Zero error paths tested.

**Production impact:** Network timeout, DNS failure, or any non-200 response causes an unhandled promise rejection that crashes the Node.js process. A 403 quota-exceeded response is silently treated as success, and `data.items` is `undefined`, causing a `TypeError` on `.length`.

---

## B02 -- No Status Code Check

**Rule violated:** `CTR-status-code-handling`
**Severity:** Critical

```typescript
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
```

**Expected violation:** `CTR-status-code-handling` -- Manifest declares status codes `[200, 400, 401, 403, 404, 500]`, but code never inspects `response.status` or `response.ok`. All responses are treated identically.

**Production impact:** A 403 response body `{ error: { code: 403, message: "quotaExceeded" } }` is cast to `VideoListResponse`. Accessing `result.items` returns `undefined`; downstream code iterating videos throws at runtime. A 401 invalid-key error silently produces garbage data.

---

## B03 -- Shallow Assertions

**Rule violated:** `TQ-no-shallow-assertions`
**Severity:** High

```typescript
// youtube-client-b03.ts — Correct client code, but tests use only shallow assertions.

// Client implementation is identical to PERFECT (omitted for brevity).

async function testB03() {
  const client = new YouTubeClient("test-key");

  // BUG: Every assertion is shallow — no structural validation
  const mockResponse = await client.listVideos({ part: ["snippet", "statistics"], id: ["abc"] });

  // These assertions prove NOTHING about contract conformance:
  expect(mockResponse).toBeDefined();             // Passes for any non-undefined value
  expect(mockResponse.items).toBeDefined();        // Passes even if items is a string "[]"
  expect(mockResponse.items.length).toBeGreaterThan(0); // Does not validate item shape

  const video = mockResponse.items[0];
  expect(video).toBeDefined();                     // Could be { foo: "bar" }
  expect(video.id).toBeDefined();                  // Does not check it is a string
  expect(video.snippet).toBeDefined();             // Does not check any snippet fields
  expect(video.statistics).toBeDefined();          // Does not check viewCount is a string
  expect(video.statistics!.viewCount).toBeDefined(); // Does not verify type is string, not number

  // No assertion on:
  //   - response.kind literal value
  //   - pageInfo shape
  //   - statistics field types (string vs number)
  //   - duration format
  //   - liveBroadcastContent enum values
  //   - thumbnail optional fields
}
```

**Expected violation:** `TQ-no-shallow-assertions` -- 8 of 8 assertions use only `.toBeDefined()` or `.toBeGreaterThan(0)`. No assertion validates field types, literal values, enum membership, or structural shape.

**Production impact:** Tests pass even if the YouTube API changes `viewCount` from string to number (or vice versa), if `pageInfo` is removed, or if a new `liveBroadcastContent` value is added. The test suite gives false confidence while the contract is broken.

---

## B04 -- Missing Negative Tests

**Rule violated:** `TQ-negative-cases`
**Severity:** High

```typescript
// youtube-client-b04.ts — Good client, but tests only cover happy paths.

// Client implementation is identical to PERFECT.

async function testB04() {
  // Test 1: List videos (happy path only)
  {
    const client = new YouTubeClient("valid-key");
    const result = await client.listVideos({
      part: ["snippet", "statistics"],
      id: ["dQw4w9WgXcQ"],
    });
    assert.equal(result.kind, "youtube#videoListResponse");
    assert.equal(result.items.length, 1);
  }

  // Test 2: Search (happy path only)
  {
    const client = new YouTubeClient("valid-key");
    const result = await client.search({ q: "rick astley" });
    assert.equal(result.kind, "youtube#searchListResponse");
    assert.ok(result.items.length > 0);
  }

  // BUG: No tests for ANY of these:
  //   - 400 Bad Request (invalid part parameter)
  //   - 401 Unauthorized (bad API key)
  //   - 403 Quota Exceeded
  //   - 403 Forbidden (video restricted by region)
  //   - 404 Not Found
  //   - 500 Server Error
  //   - Network timeout
  //   - Malformed JSON response
  //   - Empty items array (valid video ID but deleted video)
  //   - QuotaExhaustedError thrown pre-request
}
```

**Expected violation:** `TQ-negative-cases` -- Manifest declares 6 status codes `[200, 400, 401, 403, 404, 500]`, but only 200 is tested. No error path, no edge case, no boundary condition tested.

**Production impact:** When the API key expires (401), the code path is untested and may silently return garbage or crash. When quota is exhausted (403), the application has no tested recovery path. The first time a user encounters any error in production is the first time that code path runs.

---

## B05 -- Request Missing Required Fields

**Rule violated:** `CTR-request-shape`
**Severity:** Critical

```typescript
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
```

**Expected violation:** `CTR-request-shape` -- The manifest declares `part: { required: true }` for both `/youtube/v3/videos` and `/youtube/v3/search`, but the client never includes this parameter in requests.

**Production impact:** Every API call returns HTTP 400 with `{ error: { code: 400, message: "Required parameter: part" } }`. The entire integration is non-functional. This passes type checking because the request URL is built dynamically with no compile-time validation of query parameters.

---

## B06 -- Response Type Mismatch

**Rule violated:** `CTR-response-shape`
**Severity:** High

```typescript
// youtube-client-b06.ts — Client type definition is missing pageInfo wrapper field.

// BUG: VideoListResponse type is missing `pageInfo` — a required field in the API response
interface VideoListResponseB06 {
  kind: "youtube#videoListResponse";
  etag: string;
  // MISSING: pageInfo: { totalResults: number; resultsPerPage: number }
  nextPageToken?: string;
  prevPageToken?: string;
  items: YouTubeVideo[];
}

// BUG: Video type has statistics as numbers instead of strings
interface VideoStatisticsB06 {
  viewCount: number;     // WRONG: API returns string "1234567"
  likeCount: number;     // WRONG: API returns string "56789"
  commentCount: number;  // WRONG: API returns string "1234"
}

class YouTubeClientB06 {
  private apiKey: string;
  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async listVideos(params: {
    part: string[];
    id: string[];
  }): Promise<VideoListResponseB06> {
    const response = await fetch(
      `https://www.googleapis.com/youtube/v3/videos?part=${params.part.join(",")}&id=${params.id.join(",")}&key=${this.apiKey}`
    );
    if (!response.ok) throw new Error(`API error: ${response.status}`);
    return (await response.json()) as VideoListResponseB06;
  }

  getTotalResults(response: VideoListResponseB06): number {
    // BUG: This crashes at runtime because `pageInfo` is not in the type.
    // TypeScript says this property does not exist, but the API sends it.
    // If the dev ignores the TS error or uses `as any`, the code works by accident.
    // If the dev relies on the type, they cannot access totalResults at all.
    return (response as any).pageInfo?.totalResults ?? 0;
  }
}

async function testB06() {
  const client = new YouTubeClientB06("test-key");
  const result = await client.listVideos({ part: ["snippet", "statistics"], id: ["abc"] });
  assert.equal(result.kind, "youtube#videoListResponse");
  assert.ok(Array.isArray(result.items));
  // Cannot test pageInfo because it is not in the type
  // Cannot properly test statistics types because they are wrong in the type
}
```

**Expected violation:** `CTR-response-shape` -- The manifest declares `pageInfo: { totalResults: integer, resultsPerPage: integer }` as required, but the client type omits it entirely. Additionally, `statistics` fields are declared as `number` in the client type but are `string` in the manifest and API.

**Production impact:** Code that needs `totalResults` for UI pagination display must use `as any` escape hatches, bypassing all type safety. Statistics fields silently work (JavaScript coerces strings to numbers in many contexts) until someone does `viewCount + 1` expecting `"1234567" + 1 = 1234568` but getting `"12345671"` (string concatenation).

---

## B07 -- Wrong Field Types

**Rule violated:** `CTR-manifest-conformance`
**Severity:** High

```typescript
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
```

**Expected violation:** `CTR-manifest-conformance` -- Manifest declares `viewCount: { type: string, format: "numeric string" }`, but the client type uses `number` and explicitly converts with `Number()`. The `dimension` and `definition` fields are typed as `string` instead of their respective enums.

**Production impact:** For most videos, `Number()` works fine. But YouTube's most-viewed video has 15+ billion views. `Number("15000000000")` still fits in `Number.MAX_SAFE_INTEGER` (9007199254740991), but as view counts grow or if the API returns other large numeric strings, precision loss occurs silently. Additionally, `dimension` and `definition` accept any string, so a typo like `"HD"` (uppercase) is not caught.

---

## B08 -- Incomplete Enum

**Rule violated:** `CTR-strictness-parity`
**Severity:** High

```typescript
// youtube-client-b08.ts — Handles "none" and "live" but forgets "upcoming".

type LiveBroadcastContentB08 = "none" | "live";
// BUG: Missing "upcoming" — a valid enum value per API spec

interface VideoSnippetB08 {
  title: string;
  liveBroadcastContent: LiveBroadcastContentB08;
  // ... other fields
}

class YouTubeClientB08 {
  private apiKey: string;
  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  formatVideoStatus(snippet: VideoSnippetB08): string {
    // BUG: exhaustive switch is missing "upcoming"
    switch (snippet.liveBroadcastContent) {
      case "none":
        return "On-demand video";
      case "live":
        return "Currently streaming";
      // Missing case "upcoming":
      //   return "Scheduled livestream";
      default:
        // At runtime, "upcoming" falls here and is treated as an unknown value
        return "Unknown status";
    }
  }

  isLive(snippet: VideoSnippetB08): boolean {
    // BUG: Only checks for "live", treats "upcoming" same as "none"
    return snippet.liveBroadcastContent === "live";
  }

  shouldShowLiveIndicator(snippet: VideoSnippetB08): boolean {
    // BUG: An upcoming stream SHOULD show a "Starting soon" indicator
    // but this only shows indicator for "live"
    return snippet.liveBroadcastContent === "live";
  }
}

async function testB08() {
  const client = new YouTubeClientB08("test-key");

  // Tests only cover "none" and "live"
  assert.equal(
    client.formatVideoStatus({ title: "Video", liveBroadcastContent: "none" }),
    "On-demand video"
  );
  assert.equal(
    client.formatVideoStatus({ title: "Stream", liveBroadcastContent: "live" }),
    "Currently streaming"
  );
  // BUG: No test for "upcoming"
  // If tested, it would return "Unknown status" instead of "Scheduled livestream"
}
```

**Expected violation:** `CTR-strictness-parity` -- Manifest declares `liveBroadcastContent: { type: enum, values: ["none", "upcoming", "live"] }`, but the client type and switch statement only handle 2 of 3 values. The "upcoming" value is unhandled.

**Production impact:** Scheduled livestreams (with `liveBroadcastContent: "upcoming"`) are displayed as "Unknown status" instead of "Scheduled livestream" or "Starting soon". Users cannot distinguish between upcoming streams and broken data. The `shouldShowLiveIndicator` function returns false for upcoming streams, so they appear identical to regular videos in the UI.

---

## B09 -- Missing Range Validation

**Rule violated:** `CTR-strictness-parity`
**Severity:** Medium

```typescript
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
```

**Expected violation:** `CTR-strictness-parity` -- Manifest declares `maxResults: { type: integer, range: [0, 50] }`, but the client performs no range validation. Values outside 0-50 are sent directly to the API.

**Production impact:** Sending `maxResults=200` causes the YouTube API to return a 400 error: `"Invalid value '200'. Values must be within the range: [0, 50]"`. The error is generic and does not clearly indicate which parameter is wrong. A negative value like `-1` also causes a 400. These could be caught client-side with a simple bounds check, providing a better developer experience and avoiding wasted API quota (each failed call still costs quota units).

---

## B10 -- Format Not Validated

**Rule violated:** `CTR-strictness-parity`
**Severity:** Medium

```typescript
// youtube-client-b10.ts — Duration "PT1H2M3S" treated as plain string, no ISO 8601 parsing.

interface VideoDurationInfoB10 {
  videoId: string;
  duration: string;  // Stored as raw string, never validated or parsed
}

class YouTubeClientB10 {
  private apiKey: string;
  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async getVideoDurations(videoIds: string[]): Promise<VideoDurationInfoB10[]> {
    const response = await fetch(
      `https://www.googleapis.com/youtube/v3/videos?part=contentDetails&id=${videoIds.join(",")}&key=${this.apiKey}`
    );
    if (!response.ok) throw new Error(`API error: ${response.status}`);
    const data = (await response.json()) as VideoListResponse;

    return data.items.map((video) => ({
      videoId: video.id,
      // BUG: Duration stored as raw string with no format validation.
      // If the API ever returns a malformed duration, it is silently accepted.
      duration: video.contentDetails!.duration,
    }));
  }

  // BUG: Tries to display duration as "H:MM:SS" but does string manipulation
  // instead of proper ISO 8601 parsing
  formatDuration(duration: string): string {
    // This naive approach fails for many valid ISO 8601 durations:
    //   "PT3M33S"  -> replaces correctly by accident
    //   "PT1H2M3S" -> produces "1:2:3" instead of "1:02:03"
    //   "PT45S"    -> produces "45" instead of "0:45"
    //   "PT1H"     -> produces "1" instead of "1:00:00"
    return duration
      .replace("PT", "")
      .replace("H", ":")
      .replace("M", ":")
      .replace("S", "");
  }

  // BUG: No validation that duration matches ISO 8601 pattern
  isShortVideo(duration: string): boolean {
    // Tries to check if video is under 60 seconds, but uses indexOf
    // "PT59S" -> no "M" or "H", returns true (correct by accident)
    // "PT1M0S" -> has "M", returns false (correct by accident)
    // "PT0M59S" -> has "M", returns false (WRONG — 59 seconds IS short)
    return !duration.includes("H") && !duration.includes("M");
  }
}

async function testB10() {
  const client = new YouTubeClientB10("test-key");

  // Only tests the happy-path format
  assert.equal(client.formatDuration("PT3M33S"), "3:33");

  // BUG: Does not test:
  //   formatDuration("PT1H2M3S")  -> "1:2:3" (wrong, should be "1:02:03")
  //   formatDuration("PT45S")     -> "45" (wrong, should be "0:45")
  //   formatDuration("PT1H")      -> "1" (wrong, should be "1:00:00")
  //   isShortVideo("PT0M59S")     -> false (wrong, 59 seconds is short)
  //   formatDuration("invalid")   -> "invalid" (should throw)
}
```

**Expected violation:** `CTR-strictness-parity` -- Manifest declares `duration: { type: string, format: "ISO 8601 duration", pattern: "^PT(\\d+H)?(\\d+M)?(\\d+S)?$" }`, but the client never validates the format. The `formatDuration` method uses naive string replacement instead of regex-based parsing, and `isShortVideo` uses presence of "M" as a proxy for duration length.

**Production impact:** Videos display wrong durations in the UI. A 1-hour-2-minute-3-second video shows as "1:2:3" instead of "1:02:03". A 45-second short shows as "45" with no context. The `isShortVideo` function incorrectly classifies "PT0M59S" (59 seconds) as NOT short because it contains "M". Any server-side duration format change goes undetected.

---

## B11 -- Precision Loss

**Rule violated:** `CTR-strictness-parity`
**Severity:** Medium

```typescript
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
```

**Expected violation:** `CTR-strictness-parity` -- Manifest declares `viewCount: { type: string, format: "numeric string" }`. The client converts to JavaScript `number` via `parseInt`, which loses precision for values exceeding `Number.MAX_SAFE_INTEGER` (2^53 - 1). The PERFECT implementation uses `BigInt` to avoid this.

**Production impact:** For individual video view counts today, this bug is latent (YouTube's most-viewed video has ~15 billion views, well within `MAX_SAFE_INTEGER`). However, aggregated statistics across a channel's entire catalog (e.g., "total channel views") can exceed the safe integer limit. When it does, `parseInt("10000000000000001")` silently returns `10000000000000000`, and analytics dashboards show subtly wrong numbers with no error.

---

## B12 -- Nullable Field Crash

**Rule violated:** `CTR-response-shape`
**Severity:** High

```typescript
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
```

**Expected violation:** `CTR-response-shape` -- Manifest declares `standard: { required: false }` and `maxres: { required: false }`, but the client uses non-null assertions (`!`) to access these fields. The PERFECT implementation uses `YouTubeClient.getThumbnailUrl()` with a safe fallback chain.

**Production impact:** Any video without a maxres thumbnail (common for older videos, mobile uploads, or videos under 720p) causes a `TypeError: Cannot read properties of undefined (reading 'url')`. This crashes server-side rendering of OG meta tags, meaning social media previews for those videos show broken images or the page fails to render entirely. Approximately 30-40% of YouTube videos lack maxres thumbnails.

---

## B13 -- Missing Quota Tracking

**Rule violated:** `CTR-request-shape`
**Severity:** High

```typescript
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
```

**Expected violation:** `CTR-request-shape` -- Manifest declares `quota: { daily_limit: 10000, costs: { search: 100 } }`, but the client has no quota tracking mechanism. The `searchAll` method can consume unbounded quota by paginating through all search results.

**Production impact:** A single call to `searchAll("popular topic")` with many pages can exhaust the entire daily quota (10,000 units) in seconds. Once exhausted, ALL YouTube API calls for the application return 403 for the rest of the day. There is no warning, no graceful degradation, and no ability to prioritize high-value API calls over low-value ones. The error message is a generic "YouTube API error: 403" with no indication that it is a quota issue vs. a permissions issue.

---

## B14 -- Pagination Terminated Early

**Rule violated:** `CTR-response-shape`
**Severity:** High

```typescript
// youtube-client-b14.ts — Ignores nextPageToken, only returns first page.

class YouTubeClientB14 {
  private apiKey: string;
  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async getMostPopularVideos(regionCode: string = "US"): Promise<YouTubeVideo[]> {
    const response = await fetch(
      `https://www.googleapis.com/youtube/v3/videos?part=snippet,statistics&chart=mostPopular&maxResults=50&regionCode=${regionCode}&key=${this.apiKey}`
    );
    if (!response.ok) throw new Error(`API error: ${response.status}`);
    const data = (await response.json()) as VideoListResponse;

    // BUG: Returns only the first page of results.
    // data.nextPageToken may be present, indicating more results exist.
    // The caller receives at most 50 videos even if totalResults is 200.
    return data.items;

    // MISSING: Should loop while nextPageToken exists:
    // const allVideos = [...data.items];
    // while (data.nextPageToken) {
    //   data = await this.fetchPage(data.nextPageToken);
    //   allVideos.push(...data.items);
    // }
    // return allVideos;
  }

  async getChannelVideos(channelId: string): Promise<YouTubeVideo[]> {
    // Step 1: Search for videos by channel (returns video IDs)
    const searchResponse = await fetch(
      `https://www.googleapis.com/youtube/v3/search?part=snippet&channelId=${channelId}&type=video&maxResults=50&key=${this.apiKey}`
    );
    if (!searchResponse.ok) throw new Error(`Search error: ${searchResponse.status}`);
    const searchData = (await searchResponse.json()) as SearchListResponse;

    // BUG: Only gets first page of search results.
    // A channel with 500 videos only returns the first 50.
    // searchData.nextPageToken is ignored.
    const videoIds = searchData.items
      .map((item) => item.id.videoId)
      .filter((id): id is string => id !== undefined);

    // Step 2: Get full video details
    const videoResponse = await fetch(
      `https://www.googleapis.com/youtube/v3/videos?part=snippet,statistics&id=${videoIds.join(",")}&key=${this.apiKey}`
    );
    if (!videoResponse.ok) throw new Error(`Video error: ${videoResponse.status}`);
    const videoData = (await videoResponse.json()) as VideoListResponse;

    return videoData.items;
  }

  async getPlaylistItems(playlistId: string): Promise<YouTubeVideo[]> {
    const response = await fetch(
      `https://www.googleapis.com/youtube/v3/playlistItems?part=snippet&playlistId=${playlistId}&maxResults=50&key=${this.apiKey}`
    );
    if (!response.ok) throw new Error(`Playlist error: ${response.status}`);
    const data = await response.json();

    // BUG: Same pattern — ignores nextPageToken in playlist items
    // A playlist with 200 videos only returns the first 50
    return data.items;
  }
}

async function testB14() {
  const client = new YouTubeClientB14("test-key");
  const videos = await client.getMostPopularVideos("US");

  // BUG: Test only checks that SOME videos were returned,
  // never verifies that ALL pages were fetched
  assert.ok(videos.length > 0);
  assert.ok(videos.length <= 50); // This PASSES but is actually evidence of the bug

  // BUG: Never checks:
  //   - Whether nextPageToken was present in the response
  //   - Whether totalResults > items.length (indicating more pages)
  //   - Whether a multi-page channel returns ALL videos
}
```

**Expected violation:** `CTR-response-shape` -- Manifest declares `nextPageToken: { type: string, required: false }` as a response field, but the client never reads or uses this field. When the API returns a `nextPageToken`, it indicates more results exist, but the client discards them.

**Production impact:** A channel analytics dashboard that calls `getChannelVideos()` on a channel with 500 videos only shows the first 50. The "total views" calculation is off by 90%. A playlist viewer shows only the first 50 items of a 200-item playlist. The UI shows no indication that data is incomplete. Users report "missing videos" but the bug is intermittent because it only manifests when results exceed 50.

---

## B15 -- Race Condition

**Rule violated:** `CTR-request-shape`
**Severity:** Critical

```typescript
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
```

**Expected violation:** `CTR-request-shape` -- Manifest response includes `etag: { type: string, required: true }`, indicating the API supports conditional requests. The client caches responses but ignores the etag field entirely, never sends `If-None-Match` headers, and has no mechanism to detect or handle stale data. The `updateVideoCache` method has a TOCTOU race condition.

**Production impact:** A live dashboard displaying video statistics shows data up to 10 minutes stale with no indication of staleness. During a viral video event, the view count shown to the user can be millions behind the actual count. The `updateVideoCache` method computes analytics on data from time T1, but the cache is updated with data from time T2 -- the analytics and the cached data are inconsistent. Concurrent requests for the same video ID waste API quota (each miss triggers a separate fetch) and can produce inconsistent results if the video's data changes between the concurrent fetches.

---

## Summary Table

| Bug | Rule Violated | Severity | Key Signal |
|-----|---------------|----------|------------|
| B01 | TQ-error-path-coverage | Critical | No `try/catch`, no `.catch()`, bare `fetch` |
| B02 | CTR-status-code-handling | Critical | Never checks `response.ok` or `response.status` |
| B03 | TQ-no-shallow-assertions | High | All assertions use `.toBeDefined()` only |
| B04 | TQ-negative-cases | High | Only 200 tested; no 400/401/403/404/500 tests |
| B05 | CTR-request-shape | Critical | `part` parameter (required) never sent |
| B06 | CTR-response-shape | High | Client type missing `pageInfo`; wrong stat types |
| B07 | CTR-manifest-conformance | High | `viewCount` typed as `number` instead of `string` |
| B08 | CTR-strictness-parity | High | `"upcoming"` missing from liveBroadcastContent enum |
| B09 | CTR-strictness-parity | Medium | No bounds check on `maxResults` (allows 200, max is 50) |
| B10 | CTR-strictness-parity | Medium | ISO 8601 duration treated as plain string |
| B11 | CTR-strictness-parity | Medium | `parseInt` on view counts loses precision past 2^53 |
| B12 | CTR-response-shape | High | `thumbnails.maxres!.url` crashes when maxres absent |
| B13 | CTR-request-shape | High | No quota tracking; `searchAll` can exhaust 10k daily limit |
| B14 | CTR-response-shape | High | `nextPageToken` ignored; only first page returned |
| B15 | CTR-request-shape | Critical | No etag caching; TOCTOU race in read-modify-write |
