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
