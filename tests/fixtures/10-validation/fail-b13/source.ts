// src/services/google-maps-client.ts (B13 — no API key validation)

export class GoogleMapsClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;

  constructor(apiKey: string, baseUrl = "https://maps.googleapis.com") {
    // BUG: No API key format validation
    // Google Maps API keys always start with "AIza" followed by 35 alphanumeric chars
    // Accepts:
    //   "" (empty string)
    //   "sk_test_abc123" (Stripe key)
    //   "xoxb-123-456-abc" (Slack token)
    //   "null"
    //   undefined (runtime error later when building URL)
    this.apiKey = apiKey;
    this.baseUrl = baseUrl;
  }

  async geocode(address: string): Promise<GeoResult[]> {
    if (!address) throw new Error("Address required");

    const url = new URL(`${this.baseUrl}/maps/api/geocode/json`);
    url.searchParams.set("address", address);
    // BUG: Sends potentially invalid key to API
    // Every request with a bad key costs nothing (Google rejects it) but:
    // 1. Delays error detection until runtime
    // 2. Wastes a network round-trip
    // 3. Returns REQUEST_DENIED which may be silently swallowed (see B02, B08)
    url.searchParams.set("key", this.apiKey);

    try {
      const response = await fetch(url.toString());
      const body = await response.json() as GeocodeResponse;

      if (body.status === "OK") return body.results;
      if (body.status === "ZERO_RESULTS") return [];

      // This error message will say "REQUEST_DENIED" but the developer
      // has to figure out that it's because of a bad key format
      throw new Error(`Geocode error: ${body.status} — ${body.error_message}`);
    } catch (err) {
      throw new Error(`Geocode failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }
}
