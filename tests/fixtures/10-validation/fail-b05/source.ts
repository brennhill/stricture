// src/services/google-maps-client.ts (B05 — missing required fields)

export class GoogleMapsClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;

  constructor(apiKey: string, baseUrl = "https://maps.googleapis.com") {
    this.apiKey = apiKey;
    this.baseUrl = baseUrl;
  }

  async geocode(address: string): Promise<GeoResult[]> {
    // BUG: No validation that address is non-empty.
    // Sends request with address="" which Google rejects as INVALID_REQUEST
    const url = new URL(`${this.baseUrl}/maps/api/geocode/json`);
    url.searchParams.set("address", address);  // Could be empty string
    url.searchParams.set("key", this.apiKey);

    const response = await fetch(url.toString());
    const body = await response.json() as GeocodeResponse;
    return body.results;
  }

  async reverseGeocode(lat: number, lng: number): Promise<GeoResult[]> {
    // BUG: Sends request without the latlng parameter — constructs URL wrong
    const url = new URL(`${this.baseUrl}/maps/api/geocode/json`);
    // Missing: url.searchParams.set("latlng", `${lat},${lng}`);
    url.searchParams.set("key", this.apiKey);

    const response = await fetch(url.toString());
    const body = await response.json() as GeocodeResponse;
    return body.results;
  }

  async getDirections(origin: string, destination: string): Promise<DirectionRoute[]> {
    const url = new URL(`${this.baseUrl}/maps/api/directions/json`);
    url.searchParams.set("origin", origin);
    // BUG: destination parameter not set
    // url.searchParams.set("destination", destination);
    url.searchParams.set("key", this.apiKey);

    const response = await fetch(url.toString());
    const body = await response.json() as DirectionsResponse;
    return body.routes;
  }
}
