// src/services/google-maps-client.ts (B08 — incomplete status enum handling)

type GeoStatus = "OK" | "ZERO_RESULTS" | "OVER_DAILY_LIMIT" | "OVER_QUERY_LIMIT" | "REQUEST_DENIED" | "INVALID_REQUEST" | "UNKNOWN_ERROR";

// BUG: Only handles 2 of 7 status values
function handleGeoStatus(status: GeoStatus): void {
  switch (status) {
    case "OK":
      return;
    case "ZERO_RESULTS":
      return;
    // BUG: Missing cases:
    // case "OVER_DAILY_LIMIT":
    // case "OVER_QUERY_LIMIT":
    // case "REQUEST_DENIED":
    // case "INVALID_REQUEST":
    // case "UNKNOWN_ERROR":
    default:
      // Silently falls through — treats rate limiting, denied access, and
      // invalid requests the same as success
      return;
  }
}

export class GoogleMapsClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;

  constructor(apiKey: string, baseUrl = "https://maps.googleapis.com") {
    this.apiKey = apiKey;
    this.baseUrl = baseUrl;
  }

  async geocode(address: string): Promise<GeoResult[]> {
    if (!address) throw new Error("Address required");

    const url = new URL(`${this.baseUrl}/maps/api/geocode/json`);
    url.searchParams.set("address", address);
    url.searchParams.set("key", this.apiKey);

    try {
      const response = await fetch(url.toString());
      const body = await response.json() as GeocodeResponse;

      handleGeoStatus(body.status);

      // BUG: When status is "REQUEST_DENIED" or "OVER_QUERY_LIMIT",
      // handleGeoStatus silently returns, and we proceed to return
      // body.results which is []. Caller thinks "no results found"
      // instead of "your key is invalid" or "you're rate limited".
      return body.results;
    } catch (err) {
      throw new Error(`Geocode failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }
}
