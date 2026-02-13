// tests/google-maps-client.test.ts (B15)

describe("GoogleMapsClient.geocode caching", () => {
  it("caches geocode results", async () => {
    const client = new GoogleMapsClient("AIzaSyA1234567890abcdefghijklmnopqrstuvw");
    mockFetchResponse(makeGeocodeOkResponse());

    // First call — cache miss
    await client.geocode("123 Main St");
    // Second call — cache hit
    await client.geocode("123 Main St");

    expect(global.fetch).toHaveBeenCalledTimes(1); // Only 1 fetch — cache worked

    // BUG: No test for:
    // - Cache invalidation after TTL
    // - Cache coherence when API returns different coords for same address
    // - Race condition when two concurrent calls cache different results
    // - Cache size limits (unbounded Map grows forever — memory leak)
    // - Cache key normalization ("123 Main St" vs "123 main st" vs "123 Main Street")
  });
});
