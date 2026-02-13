// src/services/google-maps-client.ts (B01 — no error handling)

export class GoogleMapsClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;

  constructor(apiKey: string, baseUrl = "https://maps.googleapis.com") {
    this.apiKey = apiKey;
    this.baseUrl = baseUrl;
  }

  async geocode(address: string): Promise<GeoResult[]> {
    const url = `${this.baseUrl}/maps/api/geocode/json?address=${encodeURIComponent(address)}&key=${this.apiKey}`;

    // BUG: No try/catch. fetch() can throw on network failure, DNS failure, timeout.
    // No error handling at all — crash propagates to caller.
    const response = await fetch(url);
    const body: GeocodeResponse = await response.json();
    return body.results;
  }

  async reverseGeocode(lat: number, lng: number): Promise<GeoResult[]> {
    const url = `${this.baseUrl}/maps/api/geocode/json?latlng=${lat},${lng}&key=${this.apiKey}`;

    // BUG: Same — no error handling
    const response = await fetch(url);
    const body: GeocodeResponse = await response.json();
    return body.results;
  }
}
