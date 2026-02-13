// src/services/google-maps-client.ts (B15 — stale cache race condition)

export class GoogleMapsClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;
  private readonly geocodeCache = new Map<string, GeoResult[]>();

  constructor(apiKey: string, baseUrl = "https://maps.googleapis.com") {
    this.apiKey = apiKey;
    this.baseUrl = baseUrl;
  }

  async geocode(address: string): Promise<GeoResult[]> {
    if (!address) throw new Error("Address required");

    // BUG: Cache keyed by address string with no expiration, no versioning.
    //
    // Race condition scenario:
    // 1. Service A geocodes "123 New Development, Austin TX" → gets coords for empty lot
    // 2. Google updates DB — new building is now at that address with refined coordinates
    // 3. Service B geocodes same address → gets updated coords
    // 4. Service A uses cached stale coords — delivers to empty lot, Service B delivers to building
    //
    // Worse: concurrent requests:
    // 1. Thread 1: cache miss, sends request
    // 2. Thread 2: cache miss (not yet populated), sends request
    // 3. Thread 1: receives response v1, stores in cache
    // 4. Thread 2: receives response v2 (microseconds later, different result due to API update mid-rollout)
    // 5. Thread 2: overwrites cache with v2
    // 6. Thread 1: downstream code uses v1 (already returned) while cache now has v2
    // Two threads processing the same address have different coordinates.

    const cached = this.geocodeCache.get(address);
    if (cached) {
      return cached;
    }

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

      // BUG: Cache never expires. Stored indefinitely.
      // Google Maps coordinates can change when:
      // - New buildings are constructed
      // - Roads are rerouted
      // - Address systems are updated
      // - Google improves geocoding accuracy
      // - Boundary changes (annexation, redistricting)
      this.geocodeCache.set(address, body.results);

      return body.results;
    } catch (err) {
      throw new Error(`Geocode failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }

  async batchGeocode(addresses: string[]): Promise<Map<string, GeoResult[]>> {
    // BUG: Concurrent geocoding with shared mutable cache.
    // Promise.all fires all requests simultaneously.
    // Multiple responses write to the same cache concurrently.
    // JavaScript is single-threaded, but the interleaving of async operations
    // means cache reads and writes are not atomic with the fetch.
    const results = new Map<string, GeoResult[]>();
    await Promise.all(
      addresses.map(async (addr) => {
        const geo = await this.geocode(addr);
        results.set(addr, geo);
      }),
    );
    return results;
  }
}
