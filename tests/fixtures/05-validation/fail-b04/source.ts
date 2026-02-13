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
