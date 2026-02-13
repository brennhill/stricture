// tests/google-maps-client.test.ts (B01 — no error tests)

describe("GoogleMapsClient.geocode", () => {
  it("returns geocode results", async () => {
    mockFetchResponse(makeGeocodeOkResponse());
    const client = new GoogleMapsClient("AIzaSyA1234567890abcdefghijklmnopqrstuvw");
    const results = await client.geocode("123 Main St");
    expect(results).toHaveLength(1);
    // BUG: No test for network failure, no test for error statuses
  });
});
