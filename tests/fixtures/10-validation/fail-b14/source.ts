// src/services/google-maps-client.ts (B14 — pagination ignored)

export class GoogleMapsClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;

  constructor(apiKey: string, baseUrl = "https://maps.googleapis.com") {
    this.apiKey = apiKey;
    this.baseUrl = baseUrl;
  }

  async nearbySearch(lat: number, lng: number, radiusMeters: number, type?: string): Promise<NearbyResult[]> {
    const url = new URL(`${this.baseUrl}/maps/api/place/nearbysearch/json`);
    url.searchParams.set("location", `${lat},${lng}`);
    url.searchParams.set("radius", String(radiusMeters));
    url.searchParams.set("key", this.apiKey);
    if (type) url.searchParams.set("type", type);

    try {
      const response = await fetch(url.toString());
      const body = await response.json() as NearbyResponse;

      if (body.status !== "OK") {
        if (body.status === "ZERO_RESULTS") return [];
        throw new Error(`Nearby search error: ${body.status}`);
      }

      // BUG: Returns only the first page (max 20 results).
      // Google returns next_page_token when more results exist.
      // The token must be used in a subsequent request after a ~2 second delay.
      // Ignoring it means:
      //   - "Find all restaurants within 5km" returns max 20 out of potentially 60
      //   - Results are not random — they're ordered by "prominence"
      //   - So the code consistently returns only the most prominent places
      //   - Less popular but closer places may be on pages 2 or 3

      if (body.next_page_token) {
        // Token exists but is completely ignored
        console.log("More results available but not fetched");
      }

      return body.results;
    } catch (err) {
      throw new Error(`Nearby search failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }
}
