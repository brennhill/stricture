// tests/google-maps-client.test.ts (B04 — missing negative tests)

describe("GoogleMapsClient.geocode", () => {
  const client = new GoogleMapsClient("AIzaSyA1234567890abcdefghijklmnopqrstuvw");

  it("geocodes an address", async () => {
    mockFetchResponse(makeGeocodeOkResponse());
    const results = await client.geocode("123 Main St, Springfield, IL");
    expect(results).toHaveLength(1);
    expect(results[0].geometry.location.lat).toBeCloseTo(39.781721, 6);
  });

  it("reverse geocodes coordinates", async () => {
    mockFetchResponse(makeGeocodeOkResponse());
    const results = await client.reverseGeocode(39.781721, -89.650148);
    expect(results).toHaveLength(1);
  });

  it("gets directions", async () => {
    mockFetchResponse({
      status: "OK",
      routes: [{ legs: [{ distance: { text: "5.2 km", value: 5200 }, duration: { text: "12 mins", value: 720 }, steps: [] }] }],
      geocoded_waypoints: [],
    });
    const routes = await client.getDirections("Sydney", "Melbourne");
    expect(routes).toHaveLength(1);
  });

  // BUG: No negative tests at all. Missing:
  // - ZERO_RESULTS status
  // - OVER_QUERY_LIMIT status
  // - OVER_DAILY_LIMIT status
  // - REQUEST_DENIED status (invalid API key)
  // - INVALID_REQUEST status (missing params)
  // - UNKNOWN_ERROR status
  // - Network failure (fetch throws)
  // - Empty address input
  // - Out-of-range coordinates
  // - Invalid place_id format
  // - NOT_FOUND for directions
});
