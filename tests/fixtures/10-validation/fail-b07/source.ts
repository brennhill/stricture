// src/services/google-maps-client.ts (B07 — wrong field types)

// BUG: value fields are string instead of number (integer)
interface DistanceDuration {
  text: string;
  value: string;  // WRONG: manifest says integer, API returns number
}

// BUG: lat/lng stored as strings
interface LatLng {
  lat: string;  // WRONG: manifest says number
  lng: string;  // WRONG: manifest says number
}

interface DirectionStep {
  polyline: { points: string };
  distance: DistanceDuration;
  duration: DistanceDuration;
}

interface DirectionLeg {
  distance: DistanceDuration;
  duration: DistanceDuration;
  steps: DirectionStep[];
}

export class GoogleMapsClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;

  constructor(apiKey: string, baseUrl = "https://maps.googleapis.com") {
    this.apiKey = apiKey;
    this.baseUrl = baseUrl;
  }

  async getDirections(origin: string, destination: string): Promise<DirectionRoute[]> {
    const url = new URL(`${this.baseUrl}/maps/api/directions/json`);
    url.searchParams.set("origin", origin);
    url.searchParams.set("destination", destination);
    url.searchParams.set("key", this.apiKey);

    try {
      const response = await fetch(url.toString());
      const body = await response.json();

      // BUG: Coerces numeric values to strings during processing
      for (const route of body.routes) {
        for (const leg of route.legs) {
          leg.distance.value = String(leg.distance.value);
          leg.duration.value = String(leg.duration.value);
        }
      }

      return body.routes;
    } catch (err) {
      throw new Error(`Directions failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }

  async geocode(address: string): Promise<GeoResult[]> {
    const url = new URL(`${this.baseUrl}/maps/api/geocode/json`);
    url.searchParams.set("address", address);
    url.searchParams.set("key", this.apiKey);

    try {
      const response = await fetch(url.toString());
      const body = await response.json();

      // BUG: Coerces lat/lng to strings
      for (const result of body.results) {
        result.geometry.location.lat = String(result.geometry.location.lat);
        result.geometry.location.lng = String(result.geometry.location.lng);
      }

      return body.results;
    } catch (err) {
      throw new Error(`Geocode failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }
}
