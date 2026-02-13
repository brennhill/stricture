// src/services/google-maps-client.ts (B06 — incomplete response type)

// BUG: Missing place_id, address_components, types fields
interface GeoResult {
  formatted_address: string;
  geometry: Geometry;
  // Missing: place_id: string
  // Missing: address_components: AddressComponent[]
  // Missing: types: string[]
}

// BUG: Missing location_type and viewport
interface Geometry {
  location: LatLng;
  // Missing: location_type: LocationType
  // Missing: viewport: Viewport
}

interface LatLng {
  lat: number;
  lng: number;
}

// BUG: AddressComponent interface not defined at all — never used in code
// The API returns it, but this client ignores it entirely

interface GeocodeResponse {
  status: string;
  results: GeoResult[];
}

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

    try {
      const response = await fetch(url.toString());
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      const body = await response.json() as GeocodeResponse;
      // The API returns place_id, address_components, types, viewport, location_type
      // but our type drops them silently. Callers cannot access these fields.
      return body.results;
    } catch (err) {
      throw new Error(`Geocode failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }
}
