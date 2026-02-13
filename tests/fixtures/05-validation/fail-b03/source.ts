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
