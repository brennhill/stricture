// src/services/google-maps-client.ts (B09 — no coordinate range validation)

export class GoogleMapsClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;

  constructor(apiKey: string, baseUrl = "https://maps.googleapis.com") {
    this.apiKey = apiKey;
    this.baseUrl = baseUrl;
  }

  async reverseGeocode(lat: number, lng: number): Promise<GeoResult[]> {
    // BUG: No validation on lat/lng ranges
    // Valid: lat in [-90, 90], lng in [-180, 180]
    // Accepts: lat: 91.0, lat: -999.0, lng: 500.0
    // Google will return INVALID_REQUEST, but we don't check body.status either
    const url = new URL(`${this.baseUrl}/maps/api/geocode/json`);
    url.searchParams.set("latlng", `${lat},${lng}`);
    url.searchParams.set("key", this.apiKey);

    try {
      const response = await fetch(url.toString());
      const body = await response.json() as GeocodeResponse;
      return body.results;
    } catch (err) {
      throw new Error(`Reverse geocode failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }

  async nearbySearch(lat: number, lng: number, radiusMeters: number, type?: string): Promise<NearbyResult[]> {
    // BUG: No validation on lat, lng, or radius
    // radius valid range: [1, 50000]
    // Accepts radius: 0, radius: 100000, radius: -5
    const url = new URL(`${this.baseUrl}/maps/api/place/nearbysearch/json`);
    url.searchParams.set("location", `${lat},${lng}`);
    url.searchParams.set("radius", String(radiusMeters));
    url.searchParams.set("key", this.apiKey);
    if (type) url.searchParams.set("type", type);

    try {
      const response = await fetch(url.toString());
      const body = await response.json() as NearbyResponse;
      return body.results;
    } catch (err) {
      throw new Error(`Nearby search failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }
}
