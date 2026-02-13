// src/services/google-maps-client.ts (B10 — no format validation)

interface GeoResult {
  place_id: string;      // No format constraint — any string accepted
  formatted_address: string;
  geometry: Geometry;
  address_components: AddressComponent[];
  types: string[];
}

export class GoogleMapsClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;

  constructor(apiKey: string, baseUrl = "https://maps.googleapis.com") {
    // BUG: No API key format validation
    // Accepts: "", "invalid", "sk_test_123" — anything
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

      if (body.status !== "OK") {
        if (body.status === "ZERO_RESULTS") return [];
        throw new Error(`Geocode error: ${body.status}`);
      }

      // BUG: No place_id format validation
      // Manifest says format: "^ChIJ" — must start with "ChIJ"
      // But we accept any string, including empty strings, numeric IDs, etc.
      // A corrupted response or API change could return place_ids in a different format
      // and we would silently accept them

      return body.results;
    } catch (err) {
      throw new Error(`Geocode failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }

  async placeDetails(placeId: string): Promise<unknown> {
    // BUG: No placeId format check — accepts "123", "", "null"
    // A wrong placeId wastes an API call and returns REQUEST_DENIED or INVALID_REQUEST
    const url = new URL(`${this.baseUrl}/maps/api/place/details/json`);
    url.searchParams.set("place_id", placeId);
    url.searchParams.set("key", this.apiKey);

    const response = await fetch(url.toString());
    return response.json();
  }
}
