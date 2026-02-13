// src/services/google-maps-client.ts (B11 — coordinate precision loss)

interface LatLng {
  lat: number;
  lng: number;
}

export class GoogleMapsClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;
  private readonly cache = new Map<string, GeoResult[]>();

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

      if (body.status !== "OK") {
        if (body.status === "ZERO_RESULTS") return [];
        throw new Error(`Geocode error: ${body.status}`);
      }

      // BUG: Rounds coordinates to 2 decimal places
      // API returns: 39.781721 (6 decimals = ~0.11m precision)
      // After rounding: 39.78 (2 decimals = ~1,110m precision)
      // Manifest requires precision: 6
      for (const result of body.results) {
        result.geometry.location.lat = Math.round(result.geometry.location.lat * 100) / 100;
        result.geometry.location.lng = Math.round(result.geometry.location.lng * 100) / 100;
      }

      return body.results;
    } catch (err) {
      throw new Error(`Geocode failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }

  async reverseGeocode(lat: number, lng: number): Promise<GeoResult[]> {
    // BUG: Truncates input coordinates too — sending imprecise request
    const roundedLat = Math.round(lat * 100) / 100;
    const roundedLng = Math.round(lng * 100) / 100;

    const url = new URL(`${this.baseUrl}/maps/api/geocode/json`);
    url.searchParams.set("latlng", `${roundedLat},${roundedLng}`);
    url.searchParams.set("key", this.apiKey);

    try {
      const response = await fetch(url.toString());
      const body = await response.json() as GeocodeResponse;
      if (body.status !== "OK") {
        if (body.status === "ZERO_RESULTS") return [];
        throw new Error(`Reverse geocode error: ${body.status}`);
      }
      return body.results;
    } catch (err) {
      throw new Error(`Reverse geocode failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }
}
