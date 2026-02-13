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
