# 10 — Google Maps Geocoding & Places API

Validates Stricture detection for APIs that **always return HTTP 200** with status in the JSON body. This is the same pattern as Slack (see `08-slack.md`). Every response is `200 OK` — the real result is in `body.status`. Additionally tests coordinate precision, geographic range validation, pagination tokens, and deeply nested geometry objects.

**Key endpoints:**

| Endpoint | Purpose |
|----------|---------|
| `GET /maps/api/geocode/json?address=` | Forward geocode (address to coordinates) |
| `GET /maps/api/geocode/json?latlng=` | Reverse geocode (coordinates to address) |
| `GET /maps/api/place/nearbysearch/json` | Nearby place search |
| `GET /maps/api/directions/json` | Driving/transit directions |

**Why included:** HTTP-200-always pattern, coordinate precision, geographic range constraints, nested geometry objects, encoded polylines, pagination via `next_page_token`, distance/duration integer semantics.

---

## Manifest Fragment

```yaml
# .stricture-manifest.yml (fragment)
contracts:
  - id: "google-maps-geocode"
    producer: google-maps
    consumers: [location-service]
    protocol: http
    endpoints:
      - path: "/maps/api/geocode/json"
        method: GET
        request:
          fields:
            key:      { type: string, format: "AIza*", required: true }
            address:  { type: string, required: false }
            latlng:   { type: string, format: "^-?\\d+\\.\\d+,-?\\d+\\.\\d+$", required: false }
            # NOTE: at least one of address or latlng is required
        response:
          status_in_body: true
          http_status: 200   # Always 200
          fields:
            status:
              type: enum
              values: ["OK", "ZERO_RESULTS", "OVER_DAILY_LIMIT", "OVER_QUERY_LIMIT", "REQUEST_DENIED", "INVALID_REQUEST", "UNKNOWN_ERROR"]
              required: true
            results:
              type: array
              required: true
              items:
                fields:
                  place_id:          { type: string, format: "^ChIJ", required: true }
                  formatted_address: { type: string, required: true }
                  geometry:
                    type: object
                    required: true
                    fields:
                      location:
                        type: object
                        required: true
                        fields:
                          lat: { type: number, range: [-90, 90], precision: 6, required: true }
                          lng: { type: number, range: [-180, 180], precision: 6, required: true }
                      location_type:
                        type: enum
                        values: ["ROOFTOP", "RANGE_INTERPOLATED", "GEOMETRIC_CENTER", "APPROXIMATE"]
                        required: true
                      viewport:
                        type: object
                        required: true
                        fields:
                          northeast: { type: object, fields: { lat: { type: number }, lng: { type: number } }, required: true }
                          southwest: { type: object, fields: { lat: { type: number }, lng: { type: number } }, required: true }
                  address_components:
                    type: array
                    required: true
                    items:
                      fields:
                        long_name:  { type: string, required: true }
                        short_name: { type: string, required: true }
                        types:      { type: array, items: { type: string }, required: true }
                  types: { type: array, items: { type: string }, required: true }
            error_message: { type: string, required: false }
        status_codes: [200]   # Always 200

  - id: "google-maps-nearby"
    producer: google-maps
    consumers: [location-service]
    protocol: http
    endpoints:
      - path: "/maps/api/place/nearbysearch/json"
        method: GET
        request:
          fields:
            key:      { type: string, format: "AIza*", required: true }
            location: { type: string, format: "^-?\\d+\\.\\d+,-?\\d+\\.\\d+$", required: true }
            radius:   { type: integer, range: [1, 50000], required: true }
            type:     { type: string, required: false }
        response:
          status_in_body: true
          http_status: 200
          fields:
            status:           { type: enum, values: ["OK", "ZERO_RESULTS", "OVER_QUERY_LIMIT", "REQUEST_DENIED", "INVALID_REQUEST", "UNKNOWN_ERROR"], required: true }
            results:          { type: array, required: true }
            next_page_token:  { type: string, required: false }

  - id: "google-maps-directions"
    producer: google-maps
    consumers: [location-service]
    protocol: http
    endpoints:
      - path: "/maps/api/directions/json"
        method: GET
        request:
          fields:
            key:         { type: string, format: "AIza*", required: true }
            origin:      { type: string, required: true }
            destination: { type: string, required: true }
            mode:        { type: enum, values: ["driving", "walking", "bicycling", "transit"], required: false }
        response:
          status_in_body: true
          http_status: 200
          fields:
            status: { type: enum, values: ["OK", "NOT_FOUND", "ZERO_RESULTS", "MAX_WAYPOINTS_EXCEEDED", "INVALID_REQUEST", "OVER_DAILY_LIMIT", "OVER_QUERY_LIMIT", "REQUEST_DENIED", "UNKNOWN_ERROR"], required: true }
            routes:
              type: array
              required: true
              items:
                fields:
                  legs:
                    type: array
                    required: true
                    items:
                      fields:
                        distance: { type: object, fields: { text: { type: string }, value: { type: integer } }, required: true }
                        duration: { type: object, fields: { text: { type: string }, value: { type: integer } }, required: true }
                        steps:
                          type: array
                          required: true
                          items:
                            fields:
                              polyline: { type: object, fields: { points: { type: string } }, required: true }
                              distance: { type: object, fields: { text: { type: string }, value: { type: integer } }, required: true }
                              duration: { type: object, fields: { text: { type: string }, value: { type: integer } }, required: true }
            geocoded_waypoints: { type: array, required: true }
        status_codes: [200]
```

---

## PERFECT — Zero Violations

The canonical correct integration. Stricture must produce **zero violations** against this code.

```typescript
// src/services/google-maps-client.ts
// stricture-contract: consumer=location-service producer=google-maps

// --- Types ---

interface LatLng {
  lat: number;
  lng: number;
}

interface Viewport {
  northeast: LatLng;
  southwest: LatLng;
}

type LocationType = "ROOFTOP" | "RANGE_INTERPOLATED" | "GEOMETRIC_CENTER" | "APPROXIMATE";

interface Geometry {
  location: LatLng;
  location_type: LocationType;
  viewport: Viewport;
}

interface AddressComponent {
  long_name: string;
  short_name: string;
  types: string[];
}

interface GeoResult {
  place_id: string;
  formatted_address: string;
  geometry: Geometry;
  address_components: AddressComponent[];
  types: string[];
}

type GeoStatus =
  | "OK"
  | "ZERO_RESULTS"
  | "OVER_DAILY_LIMIT"
  | "OVER_QUERY_LIMIT"
  | "REQUEST_DENIED"
  | "INVALID_REQUEST"
  | "UNKNOWN_ERROR";

interface GeocodeResponse {
  status: GeoStatus;
  results: GeoResult[];
  error_message?: string;
}

interface NearbyResult {
  place_id: string;
  name: string;
  geometry: Geometry;
  types: string[];
  vicinity: string;
}

interface NearbyResponse {
  status: GeoStatus;
  results: NearbyResult[];
  next_page_token?: string;
}

interface DistanceDuration {
  text: string;
  value: number;  // integer — meters for distance, seconds for duration
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

interface DirectionRoute {
  legs: DirectionLeg[];
}

type DirectionStatus =
  | "OK"
  | "NOT_FOUND"
  | "ZERO_RESULTS"
  | "MAX_WAYPOINTS_EXCEEDED"
  | "INVALID_REQUEST"
  | "OVER_DAILY_LIMIT"
  | "OVER_QUERY_LIMIT"
  | "REQUEST_DENIED"
  | "UNKNOWN_ERROR";

interface DirectionsResponse {
  status: DirectionStatus;
  routes: DirectionRoute[];
  geocoded_waypoints: Array<{ geocoder_status: string; place_id: string }>;
}

// --- Errors ---

class GoogleMapsError extends Error {
  constructor(
    public readonly status: string,
    public readonly errorMessage?: string,
  ) {
    super(`Google Maps API error: ${status}${errorMessage ? ` — ${errorMessage}` : ""}`);
    this.name = "GoogleMapsError";
  }
}

class RateLimitError extends GoogleMapsError {
  constructor(errorMessage?: string) {
    super("OVER_QUERY_LIMIT", errorMessage);
    this.name = "RateLimitError";
  }
}

// --- Validation helpers ---

const API_KEY_PATTERN = /^AIza[0-9A-Za-z_-]{35}$/;
const PLACE_ID_PATTERN = /^ChIJ/;
const LAT_RANGE: [number, number] = [-90, 90];
const LNG_RANGE: [number, number] = [-180, 180];
const COORDINATE_PRECISION = 6;

function validateApiKey(key: string): void {
  if (!API_KEY_PATTERN.test(key)) {
    throw new Error(`Invalid API key format. Expected AIza* pattern, got: ${key.substring(0, 8)}...`);
  }
}

function validateLatitude(lat: number): void {
  if (typeof lat !== "number" || Number.isNaN(lat)) {
    throw new Error(`Invalid latitude: must be a number, got ${typeof lat}`);
  }
  if (lat < LAT_RANGE[0] || lat > LAT_RANGE[1]) {
    throw new Error(`Latitude out of range: ${lat}. Valid range: [${LAT_RANGE[0]}, ${LAT_RANGE[1]}]`);
  }
}

function validateLongitude(lng: number): void {
  if (typeof lng !== "number" || Number.isNaN(lng)) {
    throw new Error(`Invalid longitude: must be a number, got ${typeof lng}`);
  }
  if (lng < LNG_RANGE[0] || lng > LNG_RANGE[1]) {
    throw new Error(`Longitude out of range: ${lng}. Valid range: [${LNG_RANGE[0]}, ${LNG_RANGE[1]}]`);
  }
}

function validateCoordinatePrecision(value: number): number {
  // Preserve at least 6 decimal places (~0.11m precision)
  const str = value.toString();
  const decimalPart = str.split(".")[1];
  if (decimalPart && decimalPart.length >= COORDINATE_PRECISION) {
    return value;
  }
  // Return as-is — the API provides sufficient precision; we do not truncate
  return value;
}

function validatePlaceId(placeId: string): void {
  if (!PLACE_ID_PATTERN.test(placeId)) {
    throw new Error(`Invalid place_id format. Expected ChIJ* prefix, got: ${placeId.substring(0, 10)}`);
  }
}

// --- Response status handler (CRITICAL: Google Maps always returns HTTP 200) ---

function handleGeoStatus(status: GeoStatus, errorMessage?: string): void {
  switch (status) {
    case "OK":
    case "ZERO_RESULTS":
      return; // These are non-error statuses
    case "OVER_DAILY_LIMIT":
    case "OVER_QUERY_LIMIT":
      throw new RateLimitError(errorMessage);
    case "REQUEST_DENIED":
      throw new GoogleMapsError(status, errorMessage ?? "API key may be invalid or restricted");
    case "INVALID_REQUEST":
      throw new GoogleMapsError(status, errorMessage ?? "Missing or invalid parameters");
    case "UNKNOWN_ERROR":
      throw new GoogleMapsError(status, errorMessage ?? "Server error — retry may succeed");
    default: {
      // Exhaustive check: if Google adds a new status, TypeScript compilation fails
      const _exhaustive: never = status;
      throw new GoogleMapsError(String(_exhaustive), "Unknown status from API");
    }
  }
}

// --- Client ---

export class GoogleMapsClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;

  constructor(apiKey: string, baseUrl = "https://maps.googleapis.com") {
    validateApiKey(apiKey);
    this.apiKey = apiKey;
    this.baseUrl = baseUrl;
  }

  async geocode(address: string): Promise<GeoResult[]> {
    if (!address || address.trim().length === 0) {
      throw new Error("Address is required for geocoding");
    }

    const url = new URL(`${this.baseUrl}/maps/api/geocode/json`);
    url.searchParams.set("address", address);
    url.searchParams.set("key", this.apiKey);

    let response: Response;
    try {
      response = await fetch(url.toString());
    } catch (err) {
      throw new Error(`Network error during geocode: ${err instanceof Error ? err.message : String(err)}`);
    }

    if (!response.ok) {
      // Should never happen with Google Maps, but handle defensively
      throw new Error(`Unexpected HTTP ${response.status} from Geocoding API`);
    }

    let body: GeocodeResponse;
    try {
      body = await response.json() as GeocodeResponse;
    } catch {
      throw new Error("Failed to parse Geocoding API response as JSON");
    }

    // CRITICAL: Check body.status, not HTTP status (always 200)
    handleGeoStatus(body.status, body.error_message);

    if (body.status === "ZERO_RESULTS") {
      return []; // Empty array, not crash
    }

    // Validate each result
    for (const result of body.results) {
      validatePlaceId(result.place_id);
      validateLatitude(result.geometry.location.lat);
      validateLongitude(result.geometry.location.lng);
      validateCoordinatePrecision(result.geometry.location.lat);
      validateCoordinatePrecision(result.geometry.location.lng);
    }

    return body.results;
  }

  async reverseGeocode(lat: number, lng: number): Promise<GeoResult[]> {
    validateLatitude(lat);
    validateLongitude(lng);

    const url = new URL(`${this.baseUrl}/maps/api/geocode/json`);
    url.searchParams.set("latlng", `${lat},${lng}`);
    url.searchParams.set("key", this.apiKey);

    let response: Response;
    try {
      response = await fetch(url.toString());
    } catch (err) {
      throw new Error(`Network error during reverse geocode: ${err instanceof Error ? err.message : String(err)}`);
    }

    if (!response.ok) {
      throw new Error(`Unexpected HTTP ${response.status} from Geocoding API`);
    }

    let body: GeocodeResponse;
    try {
      body = await response.json() as GeocodeResponse;
    } catch {
      throw new Error("Failed to parse Geocoding API response as JSON");
    }

    handleGeoStatus(body.status, body.error_message);

    if (body.status === "ZERO_RESULTS") {
      return [];
    }

    for (const result of body.results) {
      validatePlaceId(result.place_id);
      validateLatitude(result.geometry.location.lat);
      validateLongitude(result.geometry.location.lng);
    }

    return body.results;
  }

  async nearbySearch(
    lat: number,
    lng: number,
    radiusMeters: number,
    type?: string,
  ): Promise<NearbyResult[]> {
    validateLatitude(lat);
    validateLongitude(lng);
    if (radiusMeters < 1 || radiusMeters > 50000) {
      throw new Error(`Radius must be between 1 and 50000 meters, got: ${radiusMeters}`);
    }

    const allResults: NearbyResult[] = [];
    let pageToken: string | undefined;

    // Pagination loop: Google returns max 20 per page, up to 60 total via next_page_token
    do {
      const url = new URL(`${this.baseUrl}/maps/api/place/nearbysearch/json`);
      url.searchParams.set("location", `${lat},${lng}`);
      url.searchParams.set("radius", String(radiusMeters));
      url.searchParams.set("key", this.apiKey);
      if (type) url.searchParams.set("type", type);
      if (pageToken) url.searchParams.set("pagetoken", pageToken);

      let response: Response;
      try {
        response = await fetch(url.toString());
      } catch (err) {
        throw new Error(`Network error during nearby search: ${err instanceof Error ? err.message : String(err)}`);
      }

      if (!response.ok) {
        throw new Error(`Unexpected HTTP ${response.status} from Places API`);
      }

      let body: NearbyResponse;
      try {
        body = await response.json() as NearbyResponse;
      } catch {
        throw new Error("Failed to parse Places API response as JSON");
      }

      handleGeoStatus(body.status, undefined);

      if (body.status === "ZERO_RESULTS") {
        break;
      }

      for (const result of body.results) {
        validatePlaceId(result.place_id);
        validateLatitude(result.geometry.location.lat);
        validateLongitude(result.geometry.location.lng);
      }

      allResults.push(...body.results);
      pageToken = body.next_page_token;

      // Google requires a short delay before using next_page_token
      if (pageToken) {
        await new Promise((resolve) => setTimeout(resolve, 2000));
      }
    } while (pageToken);

    return allResults;
  }

  async getDirections(
    origin: string,
    destination: string,
    mode: "driving" | "walking" | "bicycling" | "transit" = "driving",
  ): Promise<DirectionRoute[]> {
    if (!origin || origin.trim().length === 0) {
      throw new Error("Origin is required for directions");
    }
    if (!destination || destination.trim().length === 0) {
      throw new Error("Destination is required for directions");
    }

    const url = new URL(`${this.baseUrl}/maps/api/directions/json`);
    url.searchParams.set("origin", origin);
    url.searchParams.set("destination", destination);
    url.searchParams.set("mode", mode);
    url.searchParams.set("key", this.apiKey);

    let response: Response;
    try {
      response = await fetch(url.toString());
    } catch (err) {
      throw new Error(`Network error during directions: ${err instanceof Error ? err.message : String(err)}`);
    }

    if (!response.ok) {
      throw new Error(`Unexpected HTTP ${response.status} from Directions API`);
    }

    let body: DirectionsResponse;
    try {
      body = await response.json() as DirectionsResponse;
    } catch {
      throw new Error("Failed to parse Directions API response as JSON");
    }

    // Directions has additional statuses beyond the geocode set
    switch (body.status) {
      case "OK":
        break;
      case "NOT_FOUND":
        throw new GoogleMapsError(body.status, "Origin or destination not found");
      case "ZERO_RESULTS":
        return [];
      case "MAX_WAYPOINTS_EXCEEDED":
        throw new GoogleMapsError(body.status, "Too many waypoints (max 25)");
      case "OVER_DAILY_LIMIT":
      case "OVER_QUERY_LIMIT":
        throw new RateLimitError();
      case "REQUEST_DENIED":
        throw new GoogleMapsError(body.status, "Directions API not enabled or key restricted");
      case "INVALID_REQUEST":
        throw new GoogleMapsError(body.status, "Missing origin or destination");
      case "UNKNOWN_ERROR":
        throw new GoogleMapsError(body.status, "Server error — retry may succeed");
      default: {
        const _exhaustive: never = body.status;
        throw new GoogleMapsError(String(_exhaustive), "Unknown directions status");
      }
    }

    // Validate distance/duration values are integers (meters/seconds)
    for (const route of body.routes) {
      for (const leg of route.legs) {
        if (!Number.isInteger(leg.distance.value)) {
          throw new Error(`Distance value must be integer (meters), got: ${leg.distance.value}`);
        }
        if (!Number.isInteger(leg.duration.value)) {
          throw new Error(`Duration value must be integer (seconds), got: ${leg.duration.value}`);
        }
        for (const step of leg.steps) {
          if (!Number.isInteger(step.distance.value)) {
            throw new Error(`Step distance value must be integer, got: ${step.distance.value}`);
          }
          if (!Number.isInteger(step.duration.value)) {
            throw new Error(`Step duration value must be integer, got: ${step.duration.value}`);
          }
        }
      }
    }

    return body.routes;
  }
}
```

```typescript
// tests/google-maps-client.test.ts
// PERFECT test suite — covers happy path, all error statuses, edge cases

import { GoogleMapsClient, GoogleMapsError, RateLimitError } from "../src/services/google-maps-client";

// --- Mock setup ---

function mockFetchResponse(body: unknown): void {
  global.fetch = jest.fn().mockResolvedValue({
    ok: true,
    status: 200,
    json: () => Promise.resolve(body),
  });
}

function mockFetchNetworkError(message: string): void {
  global.fetch = jest.fn().mockRejectedValue(new Error(message));
}

const VALID_API_KEY = "AIzaSyA1234567890abcdefghijklmnopqrstuvw";

function makeGeocodeOkResponse(overrides?: Partial<{ lat: number; lng: number; placeId: string }>): object {
  return {
    status: "OK",
    results: [
      {
        place_id: overrides?.placeId ?? "ChIJN1t_tDeuEmsRUsoyG83frY4",
        formatted_address: "123 Main St, Springfield, IL 62701, USA",
        geometry: {
          location: {
            lat: overrides?.lat ?? 39.781721,
            lng: overrides?.lng ?? -89.650148,
          },
          location_type: "ROOFTOP",
          viewport: {
            northeast: { lat: 39.7830699, lng: -89.6487990 },
            southwest: { lat: 39.7803719, lng: -89.6514970 },
          },
        },
        address_components: [
          { long_name: "123", short_name: "123", types: ["street_number"] },
          { long_name: "Main Street", short_name: "Main St", types: ["route"] },
          { long_name: "Springfield", short_name: "Springfield", types: ["locality", "political"] },
          { long_name: "Illinois", short_name: "IL", types: ["administrative_area_level_1", "political"] },
          { long_name: "United States", short_name: "US", types: ["country", "political"] },
          { long_name: "62701", short_name: "62701", types: ["postal_code"] },
        ],
        types: ["street_address"],
      },
    ],
  };
}

// --- Geocode tests ---

describe("GoogleMapsClient.geocode", () => {
  const client = new GoogleMapsClient(VALID_API_KEY);

  it("returns geocode results for valid address", async () => {
    mockFetchResponse(makeGeocodeOkResponse());
    const results = await client.geocode("123 Main St, Springfield, IL");

    expect(results).toHaveLength(1);
    expect(results[0].place_id).toMatch(/^ChIJ/);
    expect(results[0].formatted_address).toContain("Springfield");
    expect(results[0].geometry.location.lat).toBeCloseTo(39.781721, 6);
    expect(results[0].geometry.location.lng).toBeCloseTo(-89.650148, 6);
    expect(results[0].geometry.location_type).toBe("ROOFTOP");
    expect(results[0].geometry.viewport.northeast.lat).toBeGreaterThan(results[0].geometry.viewport.southwest.lat);
    expect(results[0].address_components.length).toBeGreaterThan(0);
    expect(results[0].address_components[0]).toEqual(
      expect.objectContaining({ long_name: expect.any(String), short_name: expect.any(String), types: expect.any(Array) }),
    );
    expect(results[0].types).toContain("street_address");
  });

  it("returns empty array for ZERO_RESULTS", async () => {
    mockFetchResponse({ status: "ZERO_RESULTS", results: [] });
    const results = await client.geocode("xyznonexistentaddress999");
    expect(results).toEqual([]);
    expect(results).toHaveLength(0);
  });

  it("throws RateLimitError for OVER_QUERY_LIMIT", async () => {
    mockFetchResponse({ status: "OVER_QUERY_LIMIT", results: [], error_message: "You have exceeded your rate limit." });
    await expect(client.geocode("123 Main St")).rejects.toThrow(RateLimitError);
    await expect(client.geocode("123 Main St")).rejects.toThrow(/OVER_QUERY_LIMIT/);
  });

  it("throws RateLimitError for OVER_DAILY_LIMIT", async () => {
    mockFetchResponse({ status: "OVER_DAILY_LIMIT", results: [], error_message: "Daily limit exceeded." });
    await expect(client.geocode("123 Main St")).rejects.toThrow(RateLimitError);
  });

  it("throws GoogleMapsError for REQUEST_DENIED", async () => {
    mockFetchResponse({ status: "REQUEST_DENIED", results: [], error_message: "API key is invalid." });
    await expect(client.geocode("123 Main St")).rejects.toThrow(GoogleMapsError);
    await expect(client.geocode("123 Main St")).rejects.toThrow(/REQUEST_DENIED/);
  });

  it("throws GoogleMapsError for INVALID_REQUEST", async () => {
    mockFetchResponse({ status: "INVALID_REQUEST", results: [], error_message: "Missing address parameter." });
    await expect(client.geocode("")).rejects.toThrow(/Address is required/);
  });

  it("throws GoogleMapsError for UNKNOWN_ERROR", async () => {
    mockFetchResponse({ status: "UNKNOWN_ERROR", results: [] });
    await expect(client.geocode("123 Main St")).rejects.toThrow(GoogleMapsError);
  });

  it("throws on network failure", async () => {
    mockFetchNetworkError("ECONNREFUSED");
    await expect(client.geocode("123 Main St")).rejects.toThrow(/Network error/);
  });

  it("throws on empty address", async () => {
    await expect(client.geocode("")).rejects.toThrow(/Address is required/);
    await expect(client.geocode("   ")).rejects.toThrow(/Address is required/);
  });

  it("validates latitude range on response", async () => {
    mockFetchResponse(makeGeocodeOkResponse({ lat: 91.0 }));
    await expect(client.geocode("bad coords")).rejects.toThrow(/Latitude out of range/);
  });

  it("validates longitude range on response", async () => {
    mockFetchResponse(makeGeocodeOkResponse({ lng: 181.0 }));
    await expect(client.geocode("bad coords")).rejects.toThrow(/Longitude out of range/);
  });

  it("validates place_id format on response", async () => {
    mockFetchResponse(makeGeocodeOkResponse({ placeId: "INVALID_ID_123" }));
    await expect(client.geocode("bad place")).rejects.toThrow(/Invalid place_id format/);
  });
});

// --- Reverse geocode tests ---

describe("GoogleMapsClient.reverseGeocode", () => {
  const client = new GoogleMapsClient(VALID_API_KEY);

  it("returns results for valid coordinates", async () => {
    mockFetchResponse(makeGeocodeOkResponse());
    const results = await client.reverseGeocode(39.781721, -89.650148);
    expect(results).toHaveLength(1);
    expect(results[0].geometry.location.lat).toBeCloseTo(39.781721, 6);
  });

  it("throws on latitude out of range", async () => {
    await expect(client.reverseGeocode(91.0, 0)).rejects.toThrow(/Latitude out of range/);
    await expect(client.reverseGeocode(-91.0, 0)).rejects.toThrow(/Latitude out of range/);
  });

  it("throws on longitude out of range", async () => {
    await expect(client.reverseGeocode(0, 181.0)).rejects.toThrow(/Longitude out of range/);
    await expect(client.reverseGeocode(0, -181.0)).rejects.toThrow(/Longitude out of range/);
  });
});

// --- Nearby search tests ---

describe("GoogleMapsClient.nearbySearch", () => {
  const client = new GoogleMapsClient(VALID_API_KEY);

  it("returns all pages of results using next_page_token", async () => {
    let callCount = 0;
    global.fetch = jest.fn().mockImplementation(() => {
      callCount++;
      if (callCount === 1) {
        return Promise.resolve({
          ok: true, status: 200,
          json: () => Promise.resolve({
            status: "OK",
            results: Array.from({ length: 20 }, (_, i) => ({
              place_id: `ChIJ${String(i).padStart(10, "0")}page1`,
              name: `Place ${i}`,
              geometry: { location: { lat: 40.0 + i * 0.001, lng: -74.0 + i * 0.001 }, location_type: "APPROXIMATE", viewport: { northeast: { lat: 40.1, lng: -73.9 }, southwest: { lat: 39.9, lng: -74.1 } } },
              types: ["restaurant"],
              vicinity: "Nearby",
            })),
            next_page_token: "token_page_2",
          }),
        });
      }
      return Promise.resolve({
        ok: true, status: 200,
        json: () => Promise.resolve({
          status: "OK",
          results: Array.from({ length: 5 }, (_, i) => ({
            place_id: `ChIJ${String(i).padStart(10, "0")}page2`,
            name: `Place ${20 + i}`,
            geometry: { location: { lat: 40.02 + i * 0.001, lng: -74.02 + i * 0.001 }, location_type: "APPROXIMATE", viewport: { northeast: { lat: 40.1, lng: -73.9 }, southwest: { lat: 39.9, lng: -74.1 } } },
            types: ["restaurant"],
            vicinity: "Nearby",
          })),
          // No next_page_token — last page
        }),
      });
    });

    const results = await client.nearbySearch(40.0, -74.0, 5000, "restaurant");
    expect(results).toHaveLength(25); // 20 + 5 across 2 pages
    expect(global.fetch).toHaveBeenCalledTimes(2);
  });

  it("throws on invalid radius", async () => {
    await expect(client.nearbySearch(40.0, -74.0, 0)).rejects.toThrow(/Radius must be between/);
    await expect(client.nearbySearch(40.0, -74.0, 50001)).rejects.toThrow(/Radius must be between/);
  });
});

// --- Directions tests ---

describe("GoogleMapsClient.getDirections", () => {
  const client = new GoogleMapsClient(VALID_API_KEY);

  it("returns routes with integer distance/duration values", async () => {
    mockFetchResponse({
      status: "OK",
      routes: [{
        legs: [{
          distance: { text: "5.2 km", value: 5200 },
          duration: { text: "12 mins", value: 720 },
          steps: [
            {
              polyline: { points: "a~l~Fjk~uOwHJy@P" },
              distance: { text: "0.3 km", value: 300 },
              duration: { text: "1 min", value: 60 },
            },
          ],
        }],
      }],
      geocoded_waypoints: [
        { geocoder_status: "OK", place_id: "ChIJN1t_tDeuEmsRUsoyG83frY4" },
        { geocoder_status: "OK", place_id: "ChIJP3Sa8ziYEmsRUKgyFmh9AQM" },
      ],
    });

    const routes = await client.getDirections("Sydney", "Melbourne");
    expect(routes).toHaveLength(1);
    expect(routes[0].legs[0].distance.value).toBe(5200);
    expect(Number.isInteger(routes[0].legs[0].distance.value)).toBe(true);
    expect(routes[0].legs[0].duration.value).toBe(720);
    expect(Number.isInteger(routes[0].legs[0].duration.value)).toBe(true);
    expect(routes[0].legs[0].steps[0].polyline.points).toBe("a~l~Fjk~uOwHJy@P");
    expect(typeof routes[0].legs[0].distance.text).toBe("string");
  });

  it("returns empty array for ZERO_RESULTS", async () => {
    mockFetchResponse({ status: "ZERO_RESULTS", routes: [], geocoded_waypoints: [] });
    const routes = await client.getDirections("North Pole", "South Pole", "walking");
    expect(routes).toEqual([]);
  });

  it("throws for NOT_FOUND", async () => {
    mockFetchResponse({ status: "NOT_FOUND", routes: [], geocoded_waypoints: [] });
    await expect(client.getDirections("xyznotaplace", "abc")).rejects.toThrow(/NOT_FOUND/);
  });

  it("throws on empty origin or destination", async () => {
    await expect(client.getDirections("", "Melbourne")).rejects.toThrow(/Origin is required/);
    await expect(client.getDirections("Sydney", "")).rejects.toThrow(/Destination is required/);
  });
});

// --- API key validation ---

describe("GoogleMapsClient constructor", () => {
  it("throws on invalid API key format", () => {
    expect(() => new GoogleMapsClient("invalid_key")).toThrow(/Invalid API key format/);
    expect(() => new GoogleMapsClient("")).toThrow(/Invalid API key format/);
    expect(() => new GoogleMapsClient("sk_test_123")).toThrow(/Invalid API key format/);
  });

  it("accepts valid API key", () => {
    expect(() => new GoogleMapsClient(VALID_API_KEY)).not.toThrow();
  });
});
```

**Expected Stricture result:** 0 violations. All rules pass.

**Why this is correct:**
- `body.status` checked for all 7 geocode statuses + 9 direction statuses (exhaustive switch)
- Empty results handled safely (no `results[0]` access on empty array)
- Latitude validated `[-90, 90]`, longitude `[-180, 180]`
- Coordinates preserved with 6+ decimal precision
- `place_id` format validated (`ChIJ*` prefix)
- API key format validated (`AIza*`)
- `next_page_token` pagination followed in `nearbySearch`
- `distance.value` / `duration.value` validated as integers
- All fetch calls wrapped in try/catch
- Tests cover happy path AND all error statuses AND edge cases

---

## B01 — No Error Handling

**Rule violated:** `TQ-error-path-coverage`

**What changed:** Removed all try/catch around fetch calls. No error handling whatsoever.

```typescript
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
```

```typescript
// tests/google-maps-client.test.ts (B01 — no error tests)

describe("GoogleMapsClient.geocode", () => {
  it("returns geocode results", async () => {
    mockFetchResponse(makeGeocodeOkResponse());
    const client = new GoogleMapsClient("AIzaSyA1234567890abcdefghijklmnopqrstuvw");
    const results = await client.geocode("123 Main St");
    expect(results).toHaveLength(1);
    // BUG: No test for network failure, no test for error statuses
  });
});
```

**Expected violation:** `TQ-error-path-coverage` — No try/catch or `.catch()` on fetch calls. Zero error paths tested.

**Production impact:** Any network failure (DNS timeout, connection refused, TLS error) causes an unhandled promise rejection that crashes the Node.js process. The caller has no opportunity to retry or degrade gracefully.

---

## B02 — No Status Field Check

**Rule violated:** `CTR-status-code-handling`

**What changed:** Checks HTTP status (which is always 200) but never examines `body.status`. This is the **most common Google Maps integration bug** — the API always returns HTTP 200.

```typescript
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
```

```typescript
// tests/google-maps-client.test.ts (B02)

describe("GoogleMapsClient.geocode", () => {
  it("returns results on success", async () => {
    mockFetchResponse(makeGeocodeOkResponse());
    const client = new GoogleMapsClient("AIzaSyA1234567890abcdefghijklmnopqrstuvw");
    const results = await client.geocode("123 Main St");
    expect(results).toHaveLength(1);
  });

  it("handles HTTP error", async () => {
    // BUG: This test is useless — Google Maps never returns non-200 HTTP status.
    // The test passes but tests an impossible scenario.
    global.fetch = jest.fn().mockResolvedValue({ ok: false, status: 500, json: () => Promise.resolve({}) });
    const client = new GoogleMapsClient("AIzaSyA1234567890abcdefghijklmnopqrstuvw");
    await expect(client.geocode("123 Main St")).rejects.toThrow(/HTTP error/);
  });
});
```

**Expected violation:** `CTR-status-code-handling` — Response status is conveyed in `body.status` (as declared in manifest: `status_in_body: true`), but client never reads `body.status`. Handles 0 of 7 body-level status values.

**Production impact:** When the API key is invalid (`REQUEST_DENIED`), when quota is exceeded (`OVER_QUERY_LIMIT`), or when the request is malformed (`INVALID_REQUEST`), the client silently returns an empty array instead of throwing an error. The calling code interprets this as "no results found" — a silent, undetectable failure. Users see no locations on the map with no error message.

---

## B03 — Shallow Assertions

**Rule violated:** `TQ-no-shallow-assertions`

**What changed:** Tests exist but only use `.toBeDefined()`, `.toBeTruthy()`, and `.toHaveLength()` without checking field values or types.

```typescript
// tests/google-maps-client.test.ts (B03 — shallow assertions)

describe("GoogleMapsClient.geocode", () => {
  const client = new GoogleMapsClient("AIzaSyA1234567890abcdefghijklmnopqrstuvw");

  it("returns geocode results", async () => {
    mockFetchResponse(makeGeocodeOkResponse());
    const results = await client.geocode("123 Main St");

    // BUG: All assertions are shallow — they prove nothing about correctness
    expect(results).toBeDefined();                          // Shallow: any non-undefined passes
    expect(results).toBeTruthy();                           // Shallow: any non-falsy passes
    expect(results.length).toBeGreaterThan(0);              // Shallow: knows count, not content
  });

  it("returns geometry", async () => {
    mockFetchResponse(makeGeocodeOkResponse());
    const results = await client.geocode("123 Main St");

    // BUG: Checks existence but not values
    expect(results[0].geometry).toBeDefined();              // Shallow
    expect(results[0].geometry.location).toBeDefined();     // Shallow
    expect(results[0].geometry.location.lat).toBeDefined(); // Shallow — could be 999.0
    expect(results[0].geometry.location.lng).toBeDefined(); // Shallow — could be "not a number"
    expect(results[0].place_id).toBeTruthy();               // Shallow — could be "INVALID"
  });

  it("returns address components", async () => {
    mockFetchResponse(makeGeocodeOkResponse());
    const results = await client.geocode("123 Main St");

    // BUG: Array length check without element validation
    expect(results[0].address_components).toBeDefined();    // Shallow
    expect(results[0].address_components.length).toBeGreaterThan(0); // Shallow
  });
});
```

**Expected violation:** `TQ-no-shallow-assertions` at 8 locations. Assertions use `toBeDefined()`, `toBeTruthy()`, and `toBeGreaterThan(0)` on typed return values with known shapes (`GeoResult`, `Geometry`, `LatLng`, `AddressComponent`). None verify actual field values, ranges, formats, or types.

**Production impact:** Tests pass even if coordinates are out of range (lat: 999.0), place_id has wrong format ("INVALID"), or geometry structure is malformed. A regression that returns `{ lat: "NaN", lng: null }` would pass all tests. Zero protection against shape-breaking changes.

---

## B04 — Missing Negative Tests

**Rule violated:** `TQ-negative-cases`

**What changed:** Only the happy path is tested. No tests for `ZERO_RESULTS`, `OVER_QUERY_LIMIT`, `REQUEST_DENIED`, `INVALID_REQUEST`, `UNKNOWN_ERROR`, network failures, or invalid inputs.

```typescript
// tests/google-maps-client.test.ts (B04 — missing negative tests)

describe("GoogleMapsClient.geocode", () => {
  const client = new GoogleMapsClient("AIzaSyA1234567890abcdefghijklmnopqrstuvw");

  it("geocodes an address", async () => {
    mockFetchResponse(makeGeocodeOkResponse());
    const results = await client.geocode("123 Main St, Springfield, IL");
    expect(results).toHaveLength(1);
    expect(results[0].geometry.location.lat).toBeCloseTo(39.781721, 6);
  });

  it("reverse geocodes coordinates", async () => {
    mockFetchResponse(makeGeocodeOkResponse());
    const results = await client.reverseGeocode(39.781721, -89.650148);
    expect(results).toHaveLength(1);
  });

  it("gets directions", async () => {
    mockFetchResponse({
      status: "OK",
      routes: [{ legs: [{ distance: { text: "5.2 km", value: 5200 }, duration: { text: "12 mins", value: 720 }, steps: [] }] }],
      geocoded_waypoints: [],
    });
    const routes = await client.getDirections("Sydney", "Melbourne");
    expect(routes).toHaveLength(1);
  });

  // BUG: No negative tests at all. Missing:
  // - ZERO_RESULTS status
  // - OVER_QUERY_LIMIT status
  // - OVER_DAILY_LIMIT status
  // - REQUEST_DENIED status (invalid API key)
  // - INVALID_REQUEST status (missing params)
  // - UNKNOWN_ERROR status
  // - Network failure (fetch throws)
  // - Empty address input
  // - Out-of-range coordinates
  // - Invalid place_id format
  // - NOT_FOUND for directions
});
```

**Expected violation:** `TQ-negative-cases` — Test file has 3 test cases, all for `status: "OK"`. The manifest declares 7 status values for geocode and 9 for directions. Zero negative/error paths tested. Missing tests for: ZERO_RESULTS, OVER_QUERY_LIMIT, OVER_DAILY_LIMIT, REQUEST_DENIED, INVALID_REQUEST, UNKNOWN_ERROR, NOT_FOUND, MAX_WAYPOINTS_EXCEEDED, network failure, invalid inputs.

**Production impact:** If the error-handling code has a bug (e.g., throws wrong error type, silently swallows errors, crashes on empty results array), it will ship to production untested. Rate limiting and quota exhaustion will crash the application instead of being handled gracefully.

---

## B05 — Request Missing Required Fields

**Rule violated:** `CTR-request-shape`

**What changed:** Geocode request sent without `address` or `latlng` parameter. Directions request sent without `destination`. API key not included in some calls.

```typescript
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
```

**Expected violation:** `CTR-request-shape` — Three violations:
1. `geocode()` does not validate that `address` is non-empty before sending. Manifest declares `address` as conditionally required (at least one of `address` or `latlng`).
2. `reverseGeocode()` never sets the `latlng` query parameter. Manifest declares it as conditionally required.
3. `getDirections()` never sets the `destination` query parameter. Manifest declares `destination` as required.

**Production impact:** Every call to `reverseGeocode()` returns `INVALID_REQUEST` because the `latlng` param is missing. Directions always fail because Google requires both origin and destination. These are guaranteed failures masked as empty results (since B02's lack of status checking compounds the problem).

---

## B06 — Response Type Mismatch

**Rule violated:** `CTR-response-shape`

**What changed:** The `GeoResult` type definition is missing `place_id`, `address_components`, and `types` fields. The `Geometry` type is missing `viewport` and `location_type`.

```typescript
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
```

**Expected violation:** `CTR-response-shape` — Client type `GeoResult` is missing fields that the manifest declares as required:
- Missing `place_id` (type: string, format: "^ChIJ", required: true)
- Missing `address_components` (type: array, required: true)
- Missing `types` (type: array, required: true)
- `Geometry` missing `location_type` (type: enum, required: true)
- `Geometry` missing `viewport` (type: object, required: true)

**Production impact:** Downstream code that needs `place_id` for Place Details lookups, `address_components` for address parsing, or `location_type` for precision assessment will get `undefined`. TypeScript will not flag the access because the field is absent from the type, making it a silent `undefined` at runtime. The `viewport` is critical for map display — without it, the map cannot zoom to show the result.

---

## B07 — Wrong Field Types

**Rule violated:** `CTR-manifest-conformance`

**What changed:** `distance.value` and `duration.value` stored as strings instead of numbers. Latitude/longitude stored as strings.

```typescript
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
```

**Expected violation:** `CTR-manifest-conformance` — Field type mismatches:
- `distance.value`: manifest declares `integer`, code type is `string`
- `duration.value`: manifest declares `integer`, code type is `string`
- `location.lat`: manifest declares `number`, code type is `string`
- `location.lng`: manifest declares `number`, code type is `string`

**Production impact:** Arithmetic on distance/duration produces string concatenation instead of addition: `"5200" + "3100"` yields `"52003100"` instead of `8300`. Coordinate comparisons fail: `"39.78" < "40.0"` uses lexicographic ordering, which gives wrong results for negative coordinates (`"-89.6" > "-74.0"` is lexicographically true but mathematically false). Map rendering breaks because mapping libraries expect numeric coordinates.

---

## B08 — Incomplete Enum

**Rule violated:** `CTR-strictness-parity`

**What changed:** Status handler only covers "OK" and "ZERO_RESULTS", ignoring "OVER_QUERY_LIMIT", "OVER_DAILY_LIMIT", "REQUEST_DENIED", "INVALID_REQUEST", "UNKNOWN_ERROR".

```typescript
// src/services/google-maps-client.ts (B08 — incomplete status enum handling)

type GeoStatus = "OK" | "ZERO_RESULTS" | "OVER_DAILY_LIMIT" | "OVER_QUERY_LIMIT" | "REQUEST_DENIED" | "INVALID_REQUEST" | "UNKNOWN_ERROR";

// BUG: Only handles 2 of 7 status values
function handleGeoStatus(status: GeoStatus): void {
  switch (status) {
    case "OK":
      return;
    case "ZERO_RESULTS":
      return;
    // BUG: Missing cases:
    // case "OVER_DAILY_LIMIT":
    // case "OVER_QUERY_LIMIT":
    // case "REQUEST_DENIED":
    // case "INVALID_REQUEST":
    // case "UNKNOWN_ERROR":
    default:
      // Silently falls through — treats rate limiting, denied access, and
      // invalid requests the same as success
      return;
  }
}

export class GoogleMapsClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;

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

      handleGeoStatus(body.status);

      // BUG: When status is "REQUEST_DENIED" or "OVER_QUERY_LIMIT",
      // handleGeoStatus silently returns, and we proceed to return
      // body.results which is []. Caller thinks "no results found"
      // instead of "your key is invalid" or "you're rate limited".
      return body.results;
    } catch (err) {
      throw new Error(`Geocode failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }
}
```

**Expected violation:** `CTR-strictness-parity` — Consumer handles 2 of 7 enum values for field `status`. Missing explicit handling for: `OVER_DAILY_LIMIT`, `OVER_QUERY_LIMIT`, `REQUEST_DENIED`, `INVALID_REQUEST`, `UNKNOWN_ERROR`. The `default` case silently returns instead of throwing, making unhandled statuses indistinguishable from success.

**Production impact:** When the API key expires, is restricted, or quota is exceeded, the client silently returns empty results. Operations teams cannot distinguish "genuinely no results" from "service is broken." Rate limiting goes undetected, and the application continues sending requests that will all fail, worsening the quota situation.

---

## B09 — Missing Range Validation

**Rule violated:** `CTR-strictness-parity`

**What changed:** No validation on latitude or longitude ranges. Accepts `lat: 91.0` or `lng: -999.0` without error.

```typescript
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
```

**Expected violation:** `CTR-strictness-parity` — Manifest declares ranges:
- `lat`: range `[-90, 90]` — consumer has no range validation
- `lng`: range `[-180, 180]` — consumer has no range validation
- `radius`: range `[1, 50000]` — consumer has no range validation

Producer (Google Maps API) validates these ranges and returns `INVALID_REQUEST`, but consumer does not pre-validate, sending known-bad requests over the network.

**Production impact:** Invalid coordinates are sent to the API, consuming quota on requests guaranteed to fail. A user-input bug that produces `lat: 91.0` (e.g., swapping lat/lng for a point at longitude 91) causes silent failures. The application cannot distinguish "coordinate slightly off" from "coordinate completely invalid" because there is no client-side validation.

---

## B10 — Format Not Validated

**Rule violated:** `CTR-strictness-parity`

**What changed:** `place_id` accepted as any string instead of requiring `ChIJ*` prefix format. API key accepted without format check.

```typescript
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
```

**Expected violation:** `CTR-strictness-parity` — Format constraints declared in manifest but not enforced by consumer:
- `key`: manifest format `AIza*` — constructor has no format validation
- `place_id`: manifest format `^ChIJ` — geocode results and placeDetails input not validated

**Production impact:** An accidental use of a Stripe API key (`sk_test_*`) as the Google Maps key silently fails every request (REQUEST_DENIED), but the error is only discovered after deployment. Corrupted or spoofed `place_id` values from untrusted sources are passed to Place Details without validation, potentially triggering unexpected behavior. In a microservice architecture, a producer service sending malformed `place_id` values would not be caught at the consumer boundary.

---

## B11 — Precision Loss

**Rule violated:** `CTR-strictness-parity`

**What changed:** Latitude and longitude stored with only 2 decimal places, losing approximately 1.1 km of precision. Full precision requires 6+ decimal places.

```typescript
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
```

**Expected violation:** `CTR-strictness-parity` — Manifest declares `precision: 6` on `location.lat` and `location.lng`. Consumer truncates to 2 decimal places via `Math.round(value * 100) / 100`.

Precision table for reference:
| Decimals | Precision | Example |
|----------|-----------|---------|
| 0 | 111 km | Country |
| 1 | 11.1 km | City |
| 2 | 1.11 km | Neighborhood |
| 6 | 0.11 m | Individual building |
| 8 | 1.1 mm | Survey-grade |

**Production impact:** A ride-hailing app places pickup pins 1.1 km from the actual location. A delivery app routes drivers to the wrong block. Navigation starts from a point over a kilometer away from the user. Two addresses on the same street (39.78172 and 39.78234) map to the same rounded coordinate (39.78), making them indistinguishable. This bug is particularly insidious because the application "mostly works" — results look approximately right on a zoomed-out map, but are useless at street level.

---

## B12 — Nullable Field Crash

**Rule violated:** `CTR-response-shape`

**What changed:** Accesses `results[0].geometry.location` without checking if `results` is empty. When status is `ZERO_RESULTS`, `results` is `[]` and `results[0]` is `undefined`.

```typescript
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
```

**Expected violation:** `CTR-response-shape` — Accesses `results[0]` without null/empty check. The manifest declares `results` as `type: array` which can be empty (specifically when `status` is `ZERO_RESULTS`). Accessing index `[0]` on an empty array returns `undefined`, and then accessing `.geometry.location` on `undefined` throws `TypeError`.

**Production impact:** Searching for an address in a remote area (ocean, uninhabited desert) or using a misspelling returns `ZERO_RESULTS` with an empty array. The application crashes with an unhandled `TypeError: Cannot read properties of undefined (reading 'geometry')` instead of showing a "no results found" message. In a server context, this crashes the request handler, returning an HTTP 500 to the user. Error monitoring shows a cryptic TypeError with no indication that the root cause is simply "no geocode results."

---

## B13 — Missing API Key Validation

**Rule violated:** `CTR-request-shape`

**What changed:** No validation on API key format. Accepts empty strings, wrong provider keys, and malformed strings.

```typescript
// src/services/google-maps-client.ts (B13 — no API key validation)

export class GoogleMapsClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;

  constructor(apiKey: string, baseUrl = "https://maps.googleapis.com") {
    // BUG: No API key format validation
    // Google Maps API keys always start with "AIza" followed by 35 alphanumeric chars
    // Accepts:
    //   "" (empty string)
    //   "sk_test_abc123" (Stripe key)
    //   "xoxb-123-456-abc" (Slack token)
    //   "null"
    //   undefined (runtime error later when building URL)
    this.apiKey = apiKey;
    this.baseUrl = baseUrl;
  }

  async geocode(address: string): Promise<GeoResult[]> {
    if (!address) throw new Error("Address required");

    const url = new URL(`${this.baseUrl}/maps/api/geocode/json`);
    url.searchParams.set("address", address);
    // BUG: Sends potentially invalid key to API
    // Every request with a bad key costs nothing (Google rejects it) but:
    // 1. Delays error detection until runtime
    // 2. Wastes a network round-trip
    // 3. Returns REQUEST_DENIED which may be silently swallowed (see B02, B08)
    url.searchParams.set("key", this.apiKey);

    try {
      const response = await fetch(url.toString());
      const body = await response.json() as GeocodeResponse;

      if (body.status === "OK") return body.results;
      if (body.status === "ZERO_RESULTS") return [];

      // This error message will say "REQUEST_DENIED" but the developer
      // has to figure out that it's because of a bad key format
      throw new Error(`Geocode error: ${body.status} — ${body.error_message}`);
    } catch (err) {
      throw new Error(`Geocode failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }
}
```

```typescript
// tests/google-maps-client.test.ts (B13)

describe("GoogleMapsClient constructor", () => {
  // BUG: No tests for invalid API key formats
  it("creates client", () => {
    // Accepts any string — no validation
    const client = new GoogleMapsClient("anything");
    expect(client).toBeDefined();
  });

  // Missing tests:
  // - empty string key
  // - wrong provider key format (sk_test_*, xoxb-*)
  // - key too short
  // - key too long
  // - key with special characters
});
```

**Expected violation:** `CTR-request-shape` — Manifest declares `key` field with format `AIza*`. The consumer constructor accepts any string without format validation. No test verifies that invalid key formats are rejected.

**Production impact:** Configuration errors are not caught at application startup. Instead, they manifest as `REQUEST_DENIED` errors on the first API call, which might be minutes after startup in a lazy-initialization pattern. In a CI/CD pipeline, an environment variable typo (`GOGLE_MAPS_KEY` instead of `GOOGLE_MAPS_KEY`) results in an empty string key that passes construction but fails every runtime call. The error message "REQUEST_DENIED" does not hint at a key format issue, making debugging harder.

---

## B14 — Pagination Not Handled

**Rule violated:** `CTR-response-shape`

**What changed:** Nearby search ignores `next_page_token`, returning at most 20 results when up to 60 are available.

```typescript
// src/services/google-maps-client.ts (B14 — pagination ignored)

export class GoogleMapsClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;

  constructor(apiKey: string, baseUrl = "https://maps.googleapis.com") {
    this.apiKey = apiKey;
    this.baseUrl = baseUrl;
  }

  async nearbySearch(lat: number, lng: number, radiusMeters: number, type?: string): Promise<NearbyResult[]> {
    const url = new URL(`${this.baseUrl}/maps/api/place/nearbysearch/json`);
    url.searchParams.set("location", `${lat},${lng}`);
    url.searchParams.set("radius", String(radiusMeters));
    url.searchParams.set("key", this.apiKey);
    if (type) url.searchParams.set("type", type);

    try {
      const response = await fetch(url.toString());
      const body = await response.json() as NearbyResponse;

      if (body.status !== "OK") {
        if (body.status === "ZERO_RESULTS") return [];
        throw new Error(`Nearby search error: ${body.status}`);
      }

      // BUG: Returns only the first page (max 20 results).
      // Google returns next_page_token when more results exist.
      // The token must be used in a subsequent request after a ~2 second delay.
      // Ignoring it means:
      //   - "Find all restaurants within 5km" returns max 20 out of potentially 60
      //   - Results are not random — they're ordered by "prominence"
      //   - So the code consistently returns only the most prominent places
      //   - Less popular but closer places may be on pages 2 or 3

      if (body.next_page_token) {
        // Token exists but is completely ignored
        console.log("More results available but not fetched");
      }

      return body.results;
    } catch (err) {
      throw new Error(`Nearby search failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }
}
```

```typescript
// tests/google-maps-client.test.ts (B14)

describe("GoogleMapsClient.nearbySearch", () => {
  const client = new GoogleMapsClient("AIzaSyA1234567890abcdefghijklmnopqrstuvw");

  it("returns nearby places", async () => {
    mockFetchResponse({
      status: "OK",
      results: [
        { place_id: "ChIJ0000000001", name: "Restaurant A", geometry: { location: { lat: 40.0, lng: -74.0 } }, types: ["restaurant"], vicinity: "Nearby" },
      ],
      next_page_token: "CpQCAgEAAFxg8o-eU7_uKn7Yqjana-HQIx1hr5BrT4zBaEko29ANsXtp9mrqN0yrKWhf-y2PUpHRLQb1GT-mtxNcXou8TwkXhi1Jbk-RtEnfxlGKv-",
    });

    const results = await client.nearbySearch(40.0, -74.0, 5000, "restaurant");
    expect(results).toHaveLength(1);
    // BUG: Test does not verify that next_page_token is followed
    // Test passes with incomplete data — only first page tested
    // A correct test would mock multiple fetch calls and verify all pages are collected
  });
});
```

**Expected violation:** `CTR-response-shape` — Manifest declares `next_page_token` as a response field (type: string, required: false). The consumer reads it but does not use it to fetch subsequent pages. The response shape is technically read correctly, but the pagination contract is not honored.

Additionally: `TQ-negative-cases` — No test verifies multi-page behavior.

**Production impact:** A "find nearby restaurants" feature shows only the 20 most prominent restaurants, not all ~60 available. Users in dense areas (Manhattan, Tokyo) miss two-thirds of results. This creates a subtle bias: only chain restaurants and popular venues appear, while independent local businesses on pages 2-3 are invisible. Feature appears to work correctly until a user compares results with the Google Maps app and notices missing places.

---

## B15 — Race Condition

**Rule violated:** `CTR-request-shape`

**What changed:** Geocode results are cached by address string, but the same address can resolve to different coordinates over time as Google updates its database. No cache invalidation or versioning.

```typescript
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
```

```typescript
// tests/google-maps-client.test.ts (B15)

describe("GoogleMapsClient.geocode caching", () => {
  it("caches geocode results", async () => {
    const client = new GoogleMapsClient("AIzaSyA1234567890abcdefghijklmnopqrstuvw");
    mockFetchResponse(makeGeocodeOkResponse());

    // First call — cache miss
    await client.geocode("123 Main St");
    // Second call — cache hit
    await client.geocode("123 Main St");

    expect(global.fetch).toHaveBeenCalledTimes(1); // Only 1 fetch — cache worked

    // BUG: No test for:
    // - Cache invalidation after TTL
    // - Cache coherence when API returns different coords for same address
    // - Race condition when two concurrent calls cache different results
    // - Cache size limits (unbounded Map grows forever — memory leak)
    // - Cache key normalization ("123 Main St" vs "123 main st" vs "123 Main Street")
  });
});
```

**Expected violation:** `CTR-request-shape` — Cache implementation creates a read-modify-write race condition. The cache key (address string) is not normalized, and the cache has no TTL or versioning. Concurrent calls to `geocode()` with the same address can produce different results stored in different local variables while the cache is overwritten by the last responder. Additionally, the cache grows without bound (no eviction), creating a memory leak in long-running services.

**Production impact:**
1. **Stale data:** A logistics company caches geocode results for "100 New Construction Ave." The address initially resolves to the nearest road intersection. Six months later, the building exists in Google's database with precise coordinates. The application still uses the stale cache, delivering packages to the intersection instead of the building entrance.
2. **Inconsistency:** Two microservice instances geocode the same address. Instance A has a cached result from January (before a Google accuracy improvement). Instance B gets the fresh result. Two orders for the same address are routed to different locations.
3. **Memory leak:** In a busy service geocoding thousands of unique addresses per day, the cache Map grows indefinitely. After weeks of operation, the process runs out of memory and is OOM-killed.
4. **Cache key collision:** "123 MAIN ST" and "123 Main St" and "123 Main Street" are separate cache keys that all resolve to the same location, tripling API usage and cache memory.

---

## Summary Table

| Bug | Rule | Description | Severity |
|-----|------|-------------|----------|
| B01 | TQ-error-path-coverage | No try/catch on any fetch call | Critical |
| B02 | CTR-status-code-handling | Checks HTTP 200 but not body.status (always 200) | Critical |
| B03 | TQ-no-shallow-assertions | `toBeDefined()` / `toBeTruthy()` only, no value checks | High |
| B04 | TQ-negative-cases | Only happy-path tests, no error status tests | High |
| B05 | CTR-request-shape | Missing required params (address, latlng, destination) | Critical |
| B06 | CTR-response-shape | GeoResult type missing place_id, viewport, address_components | High |
| B07 | CTR-manifest-conformance | distance.value and lat/lng stored as strings not numbers | Critical |
| B08 | CTR-strictness-parity | Handles 2/7 status enum values, silent default fallthrough | Critical |
| B09 | CTR-strictness-parity | No lat/lng range validation (accepts 91.0, -999.0) | Medium |
| B10 | CTR-strictness-parity | No format validation on place_id or API key | Medium |
| B11 | CTR-strictness-parity | Coordinates rounded to 2 decimals, ~1.1km precision loss | High |
| B12 | CTR-response-shape | Crashes on results[0] when results is empty (ZERO_RESULTS) | Critical |
| B13 | CTR-request-shape | No API key format validation (accepts any string) | Medium |
| B14 | CTR-response-shape | Ignores next_page_token, returns max 20 of 60 results | Medium |
| B15 | CTR-request-shape | Unbounded cache with no TTL, no normalization, race on write | Low |

### Detection Expectations

- **B01-B04** (Level 1-2): Any quality linter should catch these. 100% detection rate expected.
- **B05-B07** (Level 3): Requires manifest awareness and type comparison. 100% detection rate expected.
- **B08-B10** (Level 4): Requires enum completeness analysis and format constraint enforcement. 100% detection rate expected.
- **B11-B15** (Level 5): Requires precision analysis, nullable traversal, pagination awareness, and cache coherence analysis. >90% detection rate expected. B15 is the hardest — detecting cache race conditions requires understanding async control flow and mutable shared state.
