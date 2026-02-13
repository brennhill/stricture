// src/services/google-maps-client.ts (B12 — crash on empty results)

export class GoogleMapsClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;

  constructor(apiKey: string, baseUrl = "https://maps.googleapis.com") {
    this.apiKey = apiKey;
    this.baseUrl = baseUrl;
  }

  async geocode(address: string): Promise<{ lat: number; lng: number; address: string }> {
    if (!address) throw new Error("Address required");

    const url = new URL(`${this.baseUrl}/maps/api/geocode/json`);
    url.searchParams.set("address", address);
    url.searchParams.set("key", this.apiKey);

    try {
      const response = await fetch(url.toString());
      const body = await response.json() as GeocodeResponse;

      if (body.status !== "OK" && body.status !== "ZERO_RESULTS") {
        throw new Error(`Geocode error: ${body.status}`);
      }

      // BUG: Accesses results[0] without checking array length.
      // When status is "ZERO_RESULTS", results is [] (empty array).
      // results[0] is undefined.
      // results[0].geometry throws: "Cannot read properties of undefined (reading 'geometry')"
      const firstResult = body.results[0];
      return {
        lat: firstResult.geometry.location.lat,   // TypeError when results is empty
        lng: firstResult.geometry.location.lng,   // Never reached
        address: firstResult.formatted_address,    // Never reached
      };
    } catch (err) {
      throw new Error(`Geocode failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }

  async findNearestPlace(lat: number, lng: number): Promise<string> {
    const url = new URL(`${this.baseUrl}/maps/api/place/nearbysearch/json`);
    url.searchParams.set("location", `${lat},${lng}`);
    url.searchParams.set("radius", "1000");
    url.searchParams.set("key", this.apiKey);

    try {
      const response = await fetch(url.toString());
      const body = await response.json() as NearbyResponse;

      // BUG: Same pattern — crashes when results is empty
      // Also crashes when geometry or location is null (some place results
      // have null geometry for certain types)
      return body.results[0].name;  // TypeError when no nearby places found
    } catch (err) {
      throw new Error(`Nearby search failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }
}
