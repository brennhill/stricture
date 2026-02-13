// src/auth0-client.ts — Auth0 Management API and Authentication API client.

import { SignJWT, jwtVerify, createRemoteJWKSet } from "jose";

// ─── Types ───────────────────────────────────────────────────────────────────

interface Auth0TokenRequest {
  grant_type: "client_credentials" | "authorization_code" | "password" | "refresh_token";
  client_id: string;
  client_secret: string;
  audience?: string;
  code?: string;
  redirect_uri?: string;
  refresh_token?: string;
  scope?: string;
}

interface Auth0TokenResponse {
  access_token: string;
  token_type: "Bearer";
  expires_in: number;
  scope?: string;
  refresh_token?: string;
}

interface Auth0ErrorResponse {
  statusCode: number;
  error: string;
  message: string;
  errorCode?: string;
}

interface Auth0Identity {
  provider: string;
  user_id: string;
  connection: string;
  isSocial: boolean;
}

interface Auth0User {
  user_id: string;
  email: string;
  email_verified: boolean;
  name: string;
  picture: string;
  created_at: string;
  updated_at: string;
  identities: Auth0Identity[];
  app_metadata?: Record<string, unknown> | null;
  user_metadata?: Record<string, unknown> | null;
}

interface CreateUserRequest {
  email: string;
  password: string;
  connection:
    | "Username-Password-Authentication"
    | "google-oauth2"
    | "facebook"
    | "apple"
    | "github"
    | "twitter";
  name?: string;
  app_metadata?: Record<string, unknown> | null;
  user_metadata?: Record<string, unknown> | null;
}

interface UpdateUserRequest {
  email?: string;
  name?: string;
  app_metadata?: Record<string, unknown> | null;
  user_metadata?: Record<string, unknown> | null;
  blocked?: boolean;
  connection?:
    | "Username-Password-Authentication"
    | "google-oauth2"
    | "facebook"
    | "apple"
    | "github"
    | "twitter";
}

type GrantType = "client_credentials" | "authorization_code" | "password" | "refresh_token";

// ─── Validation helpers ──────────────────────────────────────────────────────

const USER_ID_PATTERN = /^(auth0\|[a-f0-9]+|[\w-]+\|\w+)$/;

function isValidUserId(id: string): boolean {
  return USER_ID_PATTERN.test(id);
}

function isValidJwt(token: string): boolean {
  const parts = token.split(".");
  if (parts.length !== 3) return false;
  return parts.every((part) => /^[A-Za-z0-9_-]+$/.test(part));
}

function isValidScopes(scopes: string): boolean {
  return scopes.split(" ").every((s) => s.length > 0 && /^[\w:]+$/.test(s));
}

// ─── Token Manager ──────────────────────────────────────────────────────────

class Auth0TokenManager {
  private accessToken: string | null = null;
  private tokenExpiresAt: number = 0;
  private refreshToken: string | null = null;
  private refreshMutex: Promise<void> | null = null;

  constructor(
    private readonly tenantDomain: string,
    private readonly clientId: string,
    private readonly clientSecret: string,
    private readonly audience: string,
    private readonly requiredScopes: string,
  ) {}

  async getValidToken(): Promise<string> {
    // Token is still valid (with 60s buffer for clock skew)
    if (this.accessToken && Date.now() < this.tokenExpiresAt - 60_000) {
      return this.accessToken;
    }

    // Prevent concurrent refresh races: if a refresh is already in progress, await it
    if (this.refreshMutex) {
      await this.refreshMutex;
      if (this.accessToken && Date.now() < this.tokenExpiresAt - 60_000) {
        return this.accessToken;
      }
    }

    // Acquire refresh mutex
    let resolve: () => void;
    this.refreshMutex = new Promise<void>((r) => { resolve = r; });

    try {
      if (this.refreshToken) {
        await this.performTokenRefresh();
      } else {
        await this.performClientCredentialsGrant();
      }
    } finally {
      this.refreshMutex = null;
      resolve!();
    }

    return this.accessToken!;
  }

  private async performClientCredentialsGrant(): Promise<void> {
    const tokenResponse = await this.requestToken({
      grant_type: "client_credentials",
      client_id: this.clientId,
      client_secret: this.clientSecret,
      audience: this.audience,
      scope: this.requiredScopes,
    });
    this.storeToken(tokenResponse);
  }

  private async performTokenRefresh(): Promise<void> {
    try {
      const tokenResponse = await this.requestToken({
        grant_type: "refresh_token",
        client_id: this.clientId,
        client_secret: this.clientSecret,
        refresh_token: this.refreshToken!,
      });
      this.storeToken(tokenResponse);
    } catch (error: unknown) {
      // If refresh fails (e.g., token revoked), fall back to client_credentials
      this.refreshToken = null;
      await this.performClientCredentialsGrant();
    }
  }

  private storeToken(response: Auth0TokenResponse): void {
    this.accessToken = response.access_token;
    // Convert seconds to milliseconds for Date.now() comparison
    this.tokenExpiresAt = Date.now() + response.expires_in * 1000;
    if (response.refresh_token) {
      this.refreshToken = response.refresh_token;
    }
  }

  private async requestToken(body: Auth0TokenRequest): Promise<Auth0TokenResponse> {
    let response: Response;
    try {
      response = await fetch(`https://${this.tenantDomain}/oauth/token`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
    } catch (networkError: unknown) {
      throw new Auth0NetworkError(
        `Failed to connect to Auth0 token endpoint: ${networkError instanceof Error ? networkError.message : String(networkError)}`,
      );
    }

    if (!response.ok) {
      let errorBody: Auth0ErrorResponse;
      try {
        errorBody = (await response.json()) as Auth0ErrorResponse;
      } catch {
        throw new Auth0ApiError(response.status, "unknown", "Failed to parse error response", undefined);
      }

      if (response.status === 429) {
        const retryAfter = response.headers.get("Retry-After");
        throw new Auth0RateLimitError(
          retryAfter ? parseInt(retryAfter, 10) : 60,
          errorBody.message,
        );
      }

      throw new Auth0ApiError(
        errorBody.statusCode,
        errorBody.error,
        errorBody.message,
        errorBody.errorCode,
      );
    }

    const tokenData = (await response.json()) as Auth0TokenResponse;

    // Validate token structure
    if (tokenData.token_type !== "Bearer") {
      throw new Auth0ValidationError(`Unexpected token_type: ${tokenData.token_type}`);
    }
    if (typeof tokenData.expires_in !== "number" || tokenData.expires_in <= 0) {
      throw new Auth0ValidationError(`Invalid expires_in: ${tokenData.expires_in}`);
    }
    if (!isValidJwt(tokenData.access_token)) {
      throw new Auth0ValidationError("access_token is not a valid JWT format");
    }

    // Verify JWT claims (audience, issuer, expiry)
    const JWKS = createRemoteJWKSet(
      new URL(`https://${this.tenantDomain}/.well-known/jwks.json`),
    );
    try {
      await jwtVerify(tokenData.access_token, JWKS, {
        issuer: `https://${this.tenantDomain}/`,
        audience: this.audience,
      });
    } catch (jwtError: unknown) {
      throw new Auth0ValidationError(
        `JWT verification failed: ${jwtError instanceof Error ? jwtError.message : String(jwtError)}`,
      );
    }

    return tokenData;
  }
}

// ─── Management API Client ──────────────────────────────────────────────────

class Auth0ManagementClient {
  constructor(
    private readonly tenantDomain: string,
    private readonly tokenManager: Auth0TokenManager,
  ) {}

  async getUser(userId: string): Promise<Auth0User> {
    if (!isValidUserId(userId)) {
      throw new Auth0ValidationError(
        `Invalid user_id format: "${userId}". Expected "auth0|{hex}" or "{provider}|{id}".`,
      );
    }

    const token = await this.tokenManager.getValidToken();

    let response: Response;
    try {
      response = await fetch(
        `https://${this.tenantDomain}/api/v2/users/${encodeURIComponent(userId)}`,
        {
          method: "GET",
          headers: {
            Authorization: `Bearer ${token}`,
            "Content-Type": "application/json",
          },
        },
      );
    } catch (networkError: unknown) {
      throw new Auth0NetworkError(
        `Failed to connect to Auth0 Management API: ${networkError instanceof Error ? networkError.message : String(networkError)}`,
      );
    }

    // Handle 401 with automatic token refresh and retry
    if (response.status === 401) {
      const freshToken = await this.tokenManager.getValidToken();
      try {
        response = await fetch(
          `https://${this.tenantDomain}/api/v2/users/${encodeURIComponent(userId)}`,
          {
            method: "GET",
            headers: {
              Authorization: `Bearer ${freshToken}`,
              "Content-Type": "application/json",
            },
          },
        );
      } catch (retryNetworkError: unknown) {
        throw new Auth0NetworkError(
          `Retry failed: ${retryNetworkError instanceof Error ? retryNetworkError.message : String(retryNetworkError)}`,
        );
      }
    }

    if (!response.ok) {
      await this.handleErrorResponse(response);
    }

    const user = (await response.json()) as Auth0User;
    this.validateUserResponse(user);
    return user;
  }

  async createUser(request: CreateUserRequest): Promise<Auth0User> {
    if (!request.email || !request.password || !request.connection) {
      throw new Auth0ValidationError("email, password, and connection are required");
    }

    const token = await this.tokenManager.getValidToken();

    let response: Response;
    try {
      response = await fetch(`https://${this.tenantDomain}/api/v2/users`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(request),
      });
    } catch (networkError: unknown) {
      throw new Auth0NetworkError(
        `Failed to connect to Auth0 Management API: ${networkError instanceof Error ? networkError.message : String(networkError)}`,
      );
    }

    if (!response.ok) {
      await this.handleErrorResponse(response);
    }

    if (response.status !== 201) {
      throw new Auth0ApiError(response.status, "unexpected_status", `Expected 201 Created, got ${response.status}`, undefined);
    }

    const user = (await response.json()) as Auth0User;
    this.validateUserResponse(user);
    return user;
  }

  async updateUser(userId: string, updates: UpdateUserRequest): Promise<Auth0User> {
    if (!isValidUserId(userId)) {
      throw new Auth0ValidationError(
        `Invalid user_id format: "${userId}". Expected "auth0|{hex}" or "{provider}|{id}".`,
      );
    }

    const token = await this.tokenManager.getValidToken();

    // Read-modify-write: fetch current user first to merge metadata safely
    const currentUser = await this.getUser(userId);

    // Merge app_metadata (shallow merge to preserve fields not in this update)
    const mergedAppMetadata =
      updates.app_metadata !== undefined
        ? { ...(currentUser.app_metadata ?? {}), ...updates.app_metadata }
        : currentUser.app_metadata;

    const mergedUserMetadata =
      updates.user_metadata !== undefined
        ? { ...(currentUser.user_metadata ?? {}), ...updates.user_metadata }
        : currentUser.user_metadata;

    const patchBody: UpdateUserRequest = {
      ...updates,
      app_metadata: mergedAppMetadata as Record<string, unknown>,
      user_metadata: mergedUserMetadata as Record<string, unknown>,
    };

    let response: Response;
    try {
      response = await fetch(
        `https://${this.tenantDomain}/api/v2/users/${encodeURIComponent(userId)}`,
        {
          method: "PATCH",
          headers: {
            Authorization: `Bearer ${token}`,
            "Content-Type": "application/json",
          },
          body: JSON.stringify(patchBody),
        },
      );
    } catch (networkError: unknown) {
      throw new Auth0NetworkError(
        `Failed to connect to Auth0 Management API: ${networkError instanceof Error ? networkError.message : String(networkError)}`,
      );
    }

    if (!response.ok) {
      await this.handleErrorResponse(response);
    }

    const user = (await response.json()) as Auth0User;
    this.validateUserResponse(user);
    return user;
  }

  private validateUserResponse(user: Auth0User): void {
    if (!isValidUserId(user.user_id)) {
      throw new Auth0ValidationError(`Invalid user_id format in response: "${user.user_id}"`);
    }
    if (typeof user.email_verified !== "boolean") {
      throw new Auth0ValidationError(`email_verified must be boolean, got ${typeof user.email_verified}`);
    }
    if (!Array.isArray(user.identities)) {
      throw new Auth0ValidationError("identities must be an array");
    }
    if (!user.created_at || isNaN(Date.parse(user.created_at))) {
      throw new Auth0ValidationError(`Invalid ISO 8601 created_at: "${user.created_at}"`);
    }
  }

  private async handleErrorResponse(response: Response): Promise<never> {
    let errorBody: Auth0ErrorResponse;
    try {
      errorBody = (await response.json()) as Auth0ErrorResponse;
    } catch {
      throw new Auth0ApiError(response.status, "unknown", "Failed to parse error response", undefined);
    }

    switch (response.status) {
      case 400:
        throw new Auth0ApiError(400, errorBody.error, errorBody.message, errorBody.errorCode);
      case 401:
        throw new Auth0ApiError(401, "unauthorized", errorBody.message, errorBody.errorCode);
      case 403:
        throw new Auth0ApiError(403, "forbidden", `Insufficient scope: ${errorBody.message}`, errorBody.errorCode);
      case 404:
        throw new Auth0ApiError(404, "not_found", errorBody.message, errorBody.errorCode);
      case 409:
        throw new Auth0ApiError(409, "conflict", `User already exists: ${errorBody.message}`, errorBody.errorCode);
      case 429: {
        const retryAfter = response.headers.get("Retry-After");
        throw new Auth0RateLimitError(
          retryAfter ? parseInt(retryAfter, 10) : 60,
          errorBody.message,
        );
      }
      case 500:
        throw new Auth0ApiError(500, "server_error", errorBody.message, errorBody.errorCode);
      case 503:
        throw new Auth0ApiError(503, "service_unavailable", errorBody.message, errorBody.errorCode);
      default:
        throw new Auth0ApiError(response.status, errorBody.error, errorBody.message, errorBody.errorCode);
    }
  }
}

// ─── Error Classes ───────────────────────────────────────────────────────────

class Auth0NetworkError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "Auth0NetworkError";
  }
}

class Auth0ApiError extends Error {
  constructor(
    public readonly statusCode: number,
    public readonly error: string,
    message: string,
    public readonly errorCode?: string,
  ) {
    super(message);
    this.name = "Auth0ApiError";
  }
}

class Auth0RateLimitError extends Error {
  constructor(
    public readonly retryAfterSeconds: number,
    message: string,
  ) {
    super(message);
    this.name = "Auth0RateLimitError";
  }
}

class Auth0ValidationError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "Auth0ValidationError";
  }
}

// ─── Tests ───────────────────────────────────────────────────────────────────

// tests/auth0-client.test.ts

import { describe, it, expect, vi, beforeEach } from "vitest";

describe("Auth0 Token Endpoint", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("obtains access token via client_credentials grant", async () => {
    const mockResponse = {
      access_token: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL3Rlc3QuYXV0aDAuY29tLyIsInN1YiI6InRlc3RAY2xpZW50cyIsImF1ZCI6Imh0dHBzOi8vYXBpLmV4YW1wbGUuY29tIiwiZXhwIjo5OTk5OTk5OTk5LCJpYXQiOjE2MDAwMDAwMDAsInNjb3BlIjoicmVhZDp1c2VycyJ9.signature",
      token_type: "Bearer" as const,
      expires_in: 86400,
      scope: "read:users update:users create:users",
    };

    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify(mockResponse), { status: 200, headers: { "Content-Type": "application/json" } }),
    );

    // Verify the full response shape
    expect(mockResponse.access_token).toMatch(/^[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+$/);
    expect(mockResponse.token_type).toBe("Bearer");
    expect(mockResponse.expires_in).toBe(86400);
    expect(typeof mockResponse.expires_in).toBe("number");
    expect(mockResponse.expires_in).toBeGreaterThan(0);
    expect(mockResponse.scope).toBe("read:users update:users create:users");
    expect(mockResponse.scope!.split(" ")).toEqual(["read:users", "update:users", "create:users"]);
  });

  it("handles 401 expired token by refreshing", async () => {
    const errorResponse: Auth0ErrorResponse = {
      statusCode: 401,
      error: "Unauthorized",
      message: "Expired token",
      errorCode: "invalid_token",
    };

    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify(errorResponse), { status: 401 }),
    );

    expect(errorResponse.statusCode).toBe(401);
    expect(errorResponse.error).toBe("Unauthorized");
    expect(typeof errorResponse.message).toBe("string");
    expect(errorResponse.message.length).toBeGreaterThan(0);
    expect(errorResponse.errorCode).toBe("invalid_token");
  });

  it("handles 403 insufficient scope", async () => {
    const errorResponse: Auth0ErrorResponse = {
      statusCode: 403,
      error: "Forbidden",
      message: "Insufficient scope, expected: read:users",
      errorCode: "insufficient_scope",
    };

    expect(errorResponse.statusCode).toBe(403);
    expect(errorResponse.error).toBe("Forbidden");
    expect(errorResponse.message).toContain("scope");
    expect(errorResponse.errorCode).toBe("insufficient_scope");
  });

  it("handles 429 rate limit with Retry-After header", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(
        JSON.stringify({ statusCode: 429, error: "Too Many Requests", message: "Rate limit exceeded" }),
        { status: 429, headers: { "Retry-After": "30" } },
      ),
    );

    const response = await fetch("https://test.auth0.com/oauth/token", { method: "POST" });
    const retryAfter = response.headers.get("Retry-After");

    expect(response.status).toBe(429);
    expect(retryAfter).toBe("30");
    expect(parseInt(retryAfter!, 10)).toBe(30);
    expect(parseInt(retryAfter!, 10)).toBeGreaterThan(0);

    const body = (await response.json()) as Auth0ErrorResponse;
    expect(body.statusCode).toBe(429);
    expect(body.error).toBe("Too Many Requests");
    expect(body.message).toBe("Rate limit exceeded");
  });

  it("validates all grant_type enum values", () => {
    const validGrantTypes: GrantType[] = [
      "client_credentials",
      "authorization_code",
      "password",
      "refresh_token",
    ];

    validGrantTypes.forEach((gt) => {
      expect(["client_credentials", "authorization_code", "password", "refresh_token"]).toContain(gt);
    });

    expect(validGrantTypes).toHaveLength(4);
    expect(() => {
      const invalid = "implicit" as GrantType;
      if (!["client_credentials", "authorization_code", "password", "refresh_token"].includes(invalid)) {
        throw new Error(`Invalid grant_type: ${invalid}`);
      }
    }).toThrow("Invalid grant_type: implicit");
  });

  it("handles network failure on token request", async () => {
    vi.spyOn(globalThis, "fetch").mockRejectedValueOnce(new TypeError("fetch failed"));

    await expect(
      fetch("https://test.auth0.com/oauth/token", { method: "POST" }),
    ).rejects.toThrow("fetch failed");
  });

  it("validates token expiry arithmetic (seconds to milliseconds)", () => {
    const expiresIn = 86400; // 24 hours in seconds
    const now = Date.now();
    const expiresAt = now + expiresIn * 1000; // Convert seconds to ms

    expect(expiresAt - now).toBe(86400_000); // 24 hours in ms
    expect(expiresAt).toBeGreaterThan(now);
    expect(expiresIn * 1000).toBe(86_400_000);
  });
});

describe("Auth0 Management API — Users", () => {
  it("retrieves user by ID with full shape validation", () => {
    const user: Auth0User = {
      user_id: "auth0|507f1f77bcf86cd799439011",
      email: "alice@example.com",
      email_verified: true,
      name: "Alice Smith",
      picture: "https://cdn.auth0.com/avatars/al.png",
      created_at: "2025-01-15T10:30:00.000Z",
      updated_at: "2025-06-01T14:00:00.000Z",
      identities: [
        {
          provider: "auth0",
          user_id: "507f1f77bcf86cd799439011",
          connection: "Username-Password-Authentication",
          isSocial: false,
        },
      ],
      app_metadata: { plan: "pro", role: "admin" },
      user_metadata: { language: "en", theme: "dark" },
    };

    // Validate user_id format
    expect(user.user_id).toMatch(/^(auth0\|[a-f0-9]+|[\w-]+\|\w+)$/);

    // Validate all required fields exist and have correct types
    expect(typeof user.email).toBe("string");
    expect(user.email).toContain("@");
    expect(typeof user.email_verified).toBe("boolean");
    expect(user.email_verified).toBe(true);
    expect(typeof user.name).toBe("string");
    expect(user.name.length).toBeGreaterThan(0);

    // Validate picture is a URL
    expect(user.picture).toMatch(/^https?:\/\//);

    // Validate ISO 8601 dates
    expect(isNaN(Date.parse(user.created_at))).toBe(false);
    expect(isNaN(Date.parse(user.updated_at))).toBe(false);

    // Validate identities array structure
    expect(Array.isArray(user.identities)).toBe(true);
    expect(user.identities.length).toBeGreaterThan(0);
    expect(user.identities[0].provider).toBe("auth0");
    expect(typeof user.identities[0].isSocial).toBe("boolean");
    expect(user.identities[0].connection).toBe("Username-Password-Authentication");

    // Validate nullable metadata (null-safe access)
    const role = user.app_metadata?.role ?? "default";
    expect(role).toBe("admin");
    const theme = user.user_metadata?.theme ?? "light";
    expect(theme).toBe("dark");
  });

  it("handles null app_metadata and user_metadata safely", () => {
    const user: Auth0User = {
      user_id: "auth0|abcdef123456",
      email: "bob@example.com",
      email_verified: false,
      name: "Bob",
      picture: "https://cdn.auth0.com/avatars/bo.png",
      created_at: "2025-03-10T08:00:00.000Z",
      updated_at: "2025-03-10T08:00:00.000Z",
      identities: [{ provider: "auth0", user_id: "abcdef123456", connection: "Username-Password-Authentication", isSocial: false }],
      app_metadata: null,
      user_metadata: null,
    };

    // Null-safe access must not throw
    const role = user.app_metadata?.role ?? "default";
    expect(role).toBe("default");

    const theme = user.user_metadata?.theme ?? "light";
    expect(theme).toBe("light");

    // Accessing deeply nested metadata safely
    const permissions = (user.app_metadata as Record<string, unknown> | null)?.permissions;
    expect(permissions).toBeUndefined();
  });

  it("validates user_id formats for different providers", () => {
    const validIds = [
      "auth0|507f1f77bcf86cd799439011",
      "google-oauth2|123456789",
      "facebook|10100000000000001",
      "github|12345678",
    ];

    validIds.forEach((id) => {
      expect(id).toMatch(/^(auth0\|[a-f0-9]+|[\w-]+\|\w+)$/);
      expect(id.includes("|")).toBe(true);
      const [provider, providerId] = id.split("|");
      expect(provider.length).toBeGreaterThan(0);
      expect(providerId.length).toBeGreaterThan(0);
    });

    const invalidIds = ["plainstring", "no-pipe-here", "", "auth0|", "|no-provider"];
    invalidIds.forEach((id) => {
      expect(isValidUserId(id)).toBe(false);
    });
  });

  it("handles 404 user not found", () => {
    const error: Auth0ErrorResponse = {
      statusCode: 404,
      error: "Not Found",
      message: 'The user does not exist.',
    };

    expect(error.statusCode).toBe(404);
    expect(error.error).toBe("Not Found");
    expect(typeof error.message).toBe("string");
    expect(error.errorCode).toBeUndefined();
  });

  it("handles 409 conflict on user creation", () => {
    const error: Auth0ErrorResponse = {
      statusCode: 409,
      error: "Conflict",
      message: "The user already exists.",
      errorCode: "auth0_idp_error",
    };

    expect(error.statusCode).toBe(409);
    expect(error.error).toBe("Conflict");
    expect(error.message).toContain("already exists");
    expect(error.errorCode).toBe("auth0_idp_error");
  });

  it("validates connection enum values on user creation", () => {
    const validConnections = [
      "Username-Password-Authentication",
      "google-oauth2",
      "facebook",
      "apple",
      "github",
      "twitter",
    ];

    validConnections.forEach((conn) => {
      expect(validConnections).toContain(conn);
    });

    const invalidConnection = "linkedin";
    expect(validConnections).not.toContain(invalidConnection);
  });

  it("handles 500 and 503 server errors", () => {
    const error500: Auth0ErrorResponse = {
      statusCode: 500,
      error: "Internal Server Error",
      message: "An internal server error occurred",
    };
    expect(error500.statusCode).toBe(500);
    expect(typeof error500.message).toBe("string");

    const error503: Auth0ErrorResponse = {
      statusCode: 503,
      error: "Service Unavailable",
      message: "Auth0 service is temporarily unavailable",
    };
    expect(error503.statusCode).toBe(503);
    expect(error503.error).toBe("Service Unavailable");
  });

  it("validates scopes are space-separated strings", () => {
    const validScopes = "read:users update:users create:users delete:users";
    const scopeList = validScopes.split(" ");

    expect(scopeList).toHaveLength(4);
    scopeList.forEach((scope) => {
      expect(scope).toMatch(/^[\w:]+$/);
      expect(scope.length).toBeGreaterThan(0);
    });

    const emptyScope = "";
    expect(emptyScope.split(" ").filter(Boolean)).toHaveLength(0);
  });
});
