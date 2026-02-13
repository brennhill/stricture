// tests/google-maps-client.test.ts (B03 — shallow assertions)

describe("GoogleMapsClient.geocode", () => {
  const client = new GoogleMapsClient("AIzaSyA1234567890abcdefghijklmnopqrstuvw");

  it("returns geocode results", async () => {
    mockFetchResponse(makeGeocodeOkResponse());
    const results = await client.geocode("123 Main St");

    // BUG: All assertions are shallow — they prove nothing about correctness
    expect(results).toBeDefined();                          // Shallow: any non-undefined passes
    expect(results).toBeTruthy();                           // Shallow: any non-falsy passes
    expect(results.length).toBeGreaterThan(0);              // Shallow: knows count, not content
  });

  it("returns geometry", async () => {
    mockFetchResponse(makeGeocodeOkResponse());
    const results = await client.geocode("123 Main St");

    // BUG: Checks existence but not values
    expect(results[0].geometry).toBeDefined();              // Shallow
    expect(results[0].geometry.location).toBeDefined();     // Shallow
    expect(results[0].geometry.location.lat).toBeDefined(); // Shallow — could be 999.0
    expect(results[0].geometry.location.lng).toBeDefined(); // Shallow — could be "not a number"
    expect(results[0].place_id).toBeTruthy();               // Shallow — could be "INVALID"
  });

  it("returns address components", async () => {
    mockFetchResponse(makeGeocodeOkResponse());
    const results = await client.geocode("123 Main St");

    // BUG: Array length check without element validation
    expect(results[0].address_components).toBeDefined();    // Shallow
    expect(results[0].address_components.length).toBeGreaterThan(0); // Shallow
  });
});
