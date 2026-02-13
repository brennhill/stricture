// tests/google-maps-client.test.ts (B14)

describe("GoogleMapsClient.nearbySearch", () => {
  const client = new GoogleMapsClient("AIzaSyA1234567890abcdefghijklmnopqrstuvw");

  it("returns nearby places", async () => {
    mockFetchResponse({
      status: "OK",
      results: [
        { place_id: "ChIJ0000000001", name: "Restaurant A", geometry: { location: { lat: 40.0, lng: -74.0 } }, types: ["restaurant"], vicinity: "Nearby" },
      ],
      next_page_token: "CpQCAgEAAFxg8o-eU7_uKn7Yqjana-HQIx1hr5BrT4zBaEko29ANsXtp9mrqN0yrKWhf-y2PUpHRLQb1GT-mtxNcXou8TwkXhi1Jbk-RtEnfxlGKv-",
    });

    const results = await client.nearbySearch(40.0, -74.0, 5000, "restaurant");
    expect(results).toHaveLength(1);
    // BUG: Test does not verify that next_page_token is followed
    // Test passes with incomplete data — only first page tested
    // A correct test would mock multiple fetch calls and verify all pages are collected
  });
});
