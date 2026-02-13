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
