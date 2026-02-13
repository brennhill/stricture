// tests/google-maps-client.test.ts (B02)

describe("GoogleMapsClient.geocode", () => {
  it("returns results on success", async () => {
    mockFetchResponse(makeGeocodeOkResponse());
    const client = new GoogleMapsClient("AIzaSyA1234567890abcdefghijklmnopqrstuvw");
    const results = await client.geocode("123 Main St");
    expect(results).toHaveLength(1);
  });

  it("handles HTTP error", async () => {
    // BUG: This test is useless — Google Maps never returns non-200 HTTP status.
    // The test passes but tests an impossible scenario.
    global.fetch = jest.fn().mockResolvedValue({ ok: false, status: 500, json: () => Promise.resolve({}) });
    const client = new GoogleMapsClient("AIzaSyA1234567890abcdefghijklmnopqrstuvw");
    await expect(client.geocode("123 Main St")).rejects.toThrow(/HTTP error/);
  });
});
