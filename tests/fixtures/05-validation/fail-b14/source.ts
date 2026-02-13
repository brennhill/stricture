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
