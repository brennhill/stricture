// src/services/google-maps-client.ts (B02 — HTTP-200-always trap)

export class GoogleMapsClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;

  constructor(apiKey: string, baseUrl = "https://maps.googleapis.com") {
    this.apiKey = apiKey;
    this.baseUrl = baseUrl;
  }

  async geocode(address: string): Promise<GeoResult[]> {
    const url = new URL(`${this.baseUrl}/maps/api/geocode/json`);
    url.searchParams.set("address", address);
    url.searchParams.set("key", this.apiKey);

    let response: Response;
    try {
      response = await fetch(url.toString());
    } catch (err) {
      throw new Error(`Network error: ${err instanceof Error ? err.message : String(err)}`);
    }

    // BUG: Checks HTTP status — but Google Maps ALWAYS returns 200.
    // Even REQUEST_DENIED, OVER_QUERY_LIMIT, INVALID_REQUEST all come back as HTTP 200.
    // The real status is in body.status, which is never checked.
    if (!response.ok) {
      throw new Error(`HTTP error: ${response.status}`);
    }

    const body: GeocodeResponse = await response.json();

    // BUG: Goes straight to results without checking body.status.
    // If status is "REQUEST_DENIED", results is [] and caller gets empty array
    // instead of an error. Silent failure.
    return body.results;
  }
}
