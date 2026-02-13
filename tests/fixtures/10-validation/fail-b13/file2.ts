// tests/google-maps-client.test.ts (B13)

describe("GoogleMapsClient constructor", () => {
  // BUG: No tests for invalid API key formats
  it("creates client", () => {
    // Accepts any string — no validation
    const client = new GoogleMapsClient("anything");
    expect(client).toBeDefined();
  });

  // Missing tests:
  // - empty string key
  // - wrong provider key format (sk_test_*, xoxb-*)
  // - key too short
  // - key too long
  // - key with special characters
});
