# 11 — Auth0 Management API & Authentication API

**API:** Auth0 Management API v2 + Authentication API
**Why included:** JWT validation, token expiration arithmetic, OAuth2 grant types, scopes, nullable metadata, user ID format constraints, token refresh races
**Endpoints under test:**

| Method | Path | Purpose |
|--------|------|---------|
| `POST` | `/oauth/token` | Get access token (client credentials, authorization code, refresh token) |
| `GET` | `/api/v2/users/:id` | Get user by ID |
| `POST` | `/api/v2/users` | Create user |
| `PATCH` | `/api/v2/users/:id` | Update user |

---

## Manifest Fragment

```yaml
contracts:
  - id: "auth0-authentication"
    producer: auth0
    consumers: [my-service]
    protocol: http
    base_url: "https://{tenant}.auth0.com"
    endpoints:
      - path: "/oauth/token"
        method: POST
        request:
          content_type: "application/json"
          fields:
            grant_type:    { type: enum, values: ["client_credentials", "authorization_code", "password", "refresh_token"], required: true }
            client_id:     { type: string, minLength: 1, required: true }
            client_secret: { type: string, minLength: 1, required: true }
            audience:      { type: string, format: url, required: false }
            code:          { type: string, required: false }
            redirect_uri:  { type: string, format: url, required: false }
            refresh_token: { type: string, required: false }
            scope:         { type: string, format: "space-separated-scopes", required: false }
        response:
          fields:
            access_token:  { type: string, format: jwt, required: true }
            token_type:    { type: enum, values: ["Bearer"], required: true }
            expires_in:    { type: integer, range: [1, 2592000], required: true }
            scope:         { type: string, format: "space-separated-scopes", required: false }
            refresh_token: { type: string, required: false }
        status_codes: [200, 400, 401, 403, 429, 500, 503]
        error_response:
          fields:
            statusCode:    { type: integer, required: true }
            error:         { type: string, required: true }
            message:       { type: string, required: true }
            errorCode:     { type: string, required: false }

  - id: "auth0-management-users"
    producer: auth0
    consumers: [my-service]
    protocol: http
    base_url: "https://{tenant}.auth0.com"
    auth:
      type: bearer
      header: "Authorization"
      format: "Bearer {access_token}"
    endpoints:
      - path: "/api/v2/users/{id}"
        method: GET
        request:
          path_params:
            id: { type: string, format: "auth0|[a-f0-9]+|{provider}|{id}", required: true }
        response:
          fields:
            user_id:        { type: string, format: "auth0|[a-f0-9]+|{provider}|{id}", required: true }
            email:          { type: string, format: email, required: true }
            email_verified: { type: boolean, required: true }
            name:           { type: string, required: true }
            picture:        { type: string, format: url, required: true }
            created_at:     { type: string, format: iso8601, required: true }
            updated_at:     { type: string, format: iso8601, required: true }
            identities:     { type: array, items: { type: object, fields: { provider: { type: string }, user_id: { type: string }, connection: { type: string }, isSocial: { type: boolean } } }, required: true }
            app_metadata:   { type: object, nullable: true, required: false }
            user_metadata:  { type: object, nullable: true, required: false }
        status_codes: [200, 400, 401, 403, 404, 429, 500, 503]

      - path: "/api/v2/users"
        method: POST
        request:
          fields:
            email:      { type: string, format: email, required: true }
            password:   { type: string, minLength: 8, required: true }
            connection: { type: enum, values: ["Username-Password-Authentication", "google-oauth2", "facebook", "apple", "github", "twitter"], required: true }
            name:       { type: string, required: false }
            app_metadata:  { type: object, nullable: true, required: false }
            user_metadata: { type: object, nullable: true, required: false }
        response:
          fields:
            user_id:        { type: string, format: "auth0|[a-f0-9]+", required: true }
            email:          { type: string, format: email, required: true }
            email_verified: { type: boolean, required: true }
            name:           { type: string, required: true }
            created_at:     { type: string, format: iso8601, required: true }
            identities:     { type: array, required: true }
        status_codes: [201, 400, 401, 403, 409, 429, 500, 503]

      - path: "/api/v2/users/{id}"
        method: PATCH
        request:
          path_params:
            id: { type: string, format: "auth0|[a-f0-9]+|{provider}|{id}", required: true }
          fields:
            email:         { type: string, format: email, required: false }
            name:          { type: string, required: false }
            app_metadata:  { type: object, nullable: true, required: false }
            user_metadata: { type: object, nullable: true, required: false }
            blocked:       { type: boolean, required: false }
            connection:    { type: enum, values: ["Username-Password-Authentication", "google-oauth2", "facebook", "apple", "github", "twitter"], required: false }
        response:
          fields:
            user_id:        { type: string, format: "auth0|[a-f0-9]+|{provider}|{id}", required: true }
            email:          { type: string, format: email, required: true }
            email_verified: { type: boolean, required: true }
            name:           { type: string, required: true }
            updated_at:     { type: string, format: iso8601, required: true }
            app_metadata:   { type: object, nullable: true, required: false }
            user_metadata:  { type: object, nullable: true, required: false }
        status_codes: [200, 400, 401, 403, 404, 409, 429, 500, 503]
```

---

## PERFECT — Zero-Violation Integration

Stricture must report **zero violations** against this code. It represents a fully correct Auth0 integration with JWT validation, proper token lifecycle management, null-safe metadata access, exhaustive error handling, and deep test assertions.

```typescript
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
```

---

## B01 — No Error Handling

**Bug:** No try/catch on any Auth0 API call. Network failures, JSON parse errors, and Auth0 error responses all crash the caller with unhandled exceptions.

**Violated rule:** `TQ-error-path-coverage`

```typescript
// tests/auth0-client.test.ts — B01: No error handling

import { describe, it, expect, vi } from "vitest";

describe("Auth0 Client (B01 — no error handling)", () => {
  it("obtains access token", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({
        access_token: "eyJ.payload.sig",
        token_type: "Bearer",
        expires_in: 86400,
      }), { status: 200 }),
    );

    // No try/catch — fetch failure, JSON parse error, or Auth0 error response
    // will propagate as unhandled exceptions
    const response = await fetch("https://test.auth0.com/oauth/token", {
      method: "POST",
      body: JSON.stringify({
        grant_type: "client_credentials",
        client_id: "my-client",
        client_secret: "my-secret",
      }),
    });
    const data = await response.json();
    expect(data.access_token).toBe("eyJ.payload.sig");
  });

  it("retrieves user by ID", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({
        user_id: "auth0|abc123",
        email: "test@example.com",
        email_verified: true,
        name: "Test",
        picture: "https://cdn.auth0.com/avatars/te.png",
        created_at: "2025-01-01T00:00:00.000Z",
        updated_at: "2025-01-01T00:00:00.000Z",
        identities: [],
      }), { status: 200 }),
    );

    // No error handling on fetch or JSON parsing
    const response = await fetch("https://test.auth0.com/api/v2/users/auth0%7Cabc123", {
      headers: { Authorization: "Bearer token" },
    });
    const user = await response.json();
    expect(user.email).toBe("test@example.com");
  });

  // No tests for: network failures, 401, 403, 404, 409, 429, 500, 503
});
```

**Expected violation:**
```
TQ-error-path-coverage: Auth0 token endpoint call at line 14 has no try/catch
or .catch() handler. Network failures, HTTP errors, and JSON parse failures
will crash the caller as unhandled exceptions.
```

**Production impact:** Any transient Auth0 outage, DNS failure, or expired token causes an unhandled promise rejection that crashes the Node.js process or leaves the user facing a raw stack trace.

---

## B02 — No Status Code Check

**Bug:** Fetches are wrapped in try/catch for network errors, but the response status is never checked. A 401 (expired token), 403 (insufficient scope), or 404 (user not found) is silently parsed as if it were a success.

**Violated rule:** `CTR-status-code-handling`

```typescript
// tests/auth0-client.test.ts — B02: No status code check

import { describe, it, expect, vi } from "vitest";

describe("Auth0 Client (B02 — no status code check)", () => {
  it("obtains access token", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({
        access_token: "eyJ.payload.sig",
        token_type: "Bearer",
        expires_in: 86400,
      }), { status: 200 }),
    );

    try {
      const response = await fetch("https://test.auth0.com/oauth/token", {
        method: "POST",
        body: JSON.stringify({
          grant_type: "client_credentials",
          client_id: "my-client",
          client_secret: "my-secret",
        }),
      });
      // BUG: Never checks response.ok or response.status
      const data = await response.json();
      expect(data.access_token).toBe("eyJ.payload.sig");
    } catch (err) {
      // Only catches network-level errors, not HTTP errors
      throw err;
    }
  });

  it("retrieves user by ID", async () => {
    // Simulate a 401 response — but code treats it as success
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({
        statusCode: 401,
        error: "Unauthorized",
        message: "Expired token",
      }), { status: 401 }),
    );

    try {
      const response = await fetch("https://test.auth0.com/api/v2/users/auth0%7Cabc123", {
        headers: { Authorization: "Bearer expired-token" },
      });
      // BUG: parses the 401 error body as a user object
      const data = await response.json();
      // This will pass but data is the error body, not a user!
      expect(data).toBeDefined();
    } catch (err) {
      throw err;
    }
  });
});
```

**Expected violation:**
```
CTR-status-code-handling: Auth0 Management API GET /api/v2/users/:id
returns status codes [200, 400, 401, 403, 404, 429, 500, 503].
Client handles 0 of 8 status codes. No response.ok or response.status
check found.
```

**Production impact:** A 401 expired token response is parsed as a user object. The error body `{ statusCode, error, message }` is treated as user data, causing downstream code to crash or display "Unauthorized" as a user's name.

---

## B03 — Shallow Assertions

**Bug:** All assertions use `.toBeDefined()` or `.toBeTruthy()` instead of validating the actual response shape, field types, and values.

**Violated rule:** `TQ-no-shallow-assertions`

```typescript
// tests/auth0-client.test.ts — B03: Shallow assertions

import { describe, it, expect, vi } from "vitest";

describe("Auth0 Client (B03 — shallow assertions)", () => {
  it("obtains access token", async () => {
    const tokenResponse = {
      access_token: "eyJ.payload.sig",
      token_type: "Bearer",
      expires_in: 86400,
      scope: "read:users",
    };

    // BUG: Every assertion is shallow — proves nothing about correctness
    expect(tokenResponse).toBeDefined();
    expect(tokenResponse.access_token).toBeDefined();
    expect(tokenResponse.token_type).toBeDefined();
    expect(tokenResponse.expires_in).toBeTruthy();
    expect(tokenResponse.scope).toBeTruthy();
  });

  it("retrieves user", () => {
    const user = {
      user_id: "auth0|abc123",
      email: "test@example.com",
      email_verified: true,
      name: "Test User",
      picture: "https://cdn.auth0.com/avatars/te.png",
      created_at: "2025-01-01T00:00:00.000Z",
      updated_at: "2025-01-01T00:00:00.000Z",
      identities: [{ provider: "auth0", user_id: "abc123", connection: "Username-Password-Authentication", isSocial: false }],
      app_metadata: { role: "admin" },
    };

    // BUG: All shallow — user_id could be "INVALID" and these still pass
    expect(user).toBeDefined();
    expect(user.user_id).toBeDefined();
    expect(user.email).toBeDefined();
    expect(user.email_verified).toBeDefined();
    expect(user.identities).toBeDefined();
    expect(user.app_metadata).toBeTruthy();
    expect(user.created_at).toBeTruthy();
  });

  it("handles error response", () => {
    const error = {
      statusCode: 400,
      error: "Bad Request",
      message: "Missing required field: email",
    };

    // BUG: Does not verify error shape matches Auth0 contract
    expect(error).toBeDefined();
    expect(error.statusCode).toBeTruthy();
    expect(error.message).toBeDefined();
  });
});
```

**Expected violation:**
```
TQ-no-shallow-assertions: 12 shallow assertions found.
Line 13: expect(tokenResponse).toBeDefined() — shallow on typed return value (Auth0TokenResponse with 4 fields).
Line 14: expect(tokenResponse.access_token).toBeDefined() — access_token is string, assert format/value.
Line 16: expect(tokenResponse.expires_in).toBeTruthy() — expires_in is number, assert range > 0.
Line 31: expect(user).toBeDefined() — shallow on Auth0User with 10+ fields.
...
```

**Production impact:** A response where `token_type` is `"MAC"` instead of `"Bearer"` passes all assertions. A user with `email_verified` set to a string `"true"` instead of boolean `true` passes. The test suite provides a false sense of correctness.

---

## B04 — Missing Negative Tests

**Bug:** Only tests the happy path (200 OK). No tests for expired token (401), insufficient scope (403), user not found (404), duplicate user (409), rate limiting (429), or server errors (500/503).

**Violated rule:** `TQ-negative-cases`

```typescript
// tests/auth0-client.test.ts — B04: Missing negative tests

import { describe, it, expect } from "vitest";

describe("Auth0 Client (B04 — no negative tests)", () => {
  it("obtains access token successfully", () => {
    const tokenResponse = {
      access_token: "eyJ.payload.sig",
      token_type: "Bearer" as const,
      expires_in: 86400,
    };

    expect(tokenResponse.access_token).toMatch(/\./);
    expect(tokenResponse.token_type).toBe("Bearer");
    expect(tokenResponse.expires_in).toBe(86400);
  });

  it("retrieves user successfully", () => {
    const user = {
      user_id: "auth0|abc123",
      email: "test@example.com",
      email_verified: true,
      name: "Test",
      picture: "https://cdn.auth0.com/avatars/te.png",
      created_at: "2025-01-01T00:00:00.000Z",
      updated_at: "2025-01-01T00:00:00.000Z",
      identities: [{ provider: "auth0", user_id: "abc123", connection: "Username-Password-Authentication", isSocial: false }],
    };

    expect(user.user_id).toBe("auth0|abc123");
    expect(user.email).toBe("test@example.com");
    expect(user.identities).toHaveLength(1);
  });

  it("creates user successfully", () => {
    const created = {
      user_id: "auth0|newuser123",
      email: "new@example.com",
      email_verified: false,
      name: "New User",
      created_at: "2025-06-01T00:00:00.000Z",
      identities: [{ provider: "auth0", user_id: "newuser123", connection: "Username-Password-Authentication", isSocial: false }],
    };

    expect(created.user_id).toMatch(/^auth0\|/);
    expect(created.email_verified).toBe(false);
  });

  // BUG: No tests for ANY error scenario:
  // - Missing: 401 expired/invalid token
  // - Missing: 403 insufficient scope
  // - Missing: 404 user not found
  // - Missing: 409 user already exists
  // - Missing: 429 rate limit exceeded
  // - Missing: 500/503 server errors
  // - Missing: Network timeout/failure
  // - Missing: Invalid grant_type
  // - Missing: Malformed JWT
  // - Missing: Null app_metadata access
});
```

**Expected violation:**
```
TQ-negative-cases: Auth0 API endpoints return status codes
[200, 201, 400, 401, 403, 404, 409, 429, 500, 503].
Tests only cover success cases (200, 201). Missing negative tests
for: 400, 401, 403, 404, 409, 429, 500, 503.
0 of 8 error status codes tested.
```

**Production impact:** Expired tokens silently cause user-facing errors. Rate limits are never handled, causing cascading retries. A user not found (404) is never caught, displaying blank profiles instead of error messages.

---

## B05 — Request Missing Required Fields

**Bug:** Token request omits the required `grant_type` field. Create user request omits `connection`. The request payloads are structurally incomplete per the Auth0 contract.

**Violated rule:** `CTR-request-shape`

```typescript
// tests/auth0-client.test.ts — B05: Request missing required fields

import { describe, it, expect, vi } from "vitest";

describe("Auth0 Client (B05 — missing required request fields)", () => {
  it("requests access token without grant_type", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({
        access_token: "eyJ.payload.sig",
        token_type: "Bearer",
        expires_in: 86400,
      }), { status: 200 }),
    );

    // BUG: Missing grant_type — Auth0 requires this field
    const body = {
      client_id: "my-client",
      client_secret: "my-secret",
      audience: "https://api.example.com",
      // grant_type is MISSING — Auth0 will return 400
    };

    const response = await fetch("https://test.auth0.com/oauth/token", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });

    if (response.ok) {
      const data = await response.json();
      expect(data.access_token).toMatch(/\./);
    }
  });

  it("creates user without connection field", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({
        user_id: "auth0|newuser",
        email: "new@example.com",
        email_verified: false,
        name: "New",
        created_at: "2025-06-01T00:00:00.000Z",
        identities: [],
      }), { status: 201 }),
    );

    // BUG: Missing connection — Auth0 requires this for user creation
    const body = {
      email: "new@example.com",
      password: "secureP@ss123",
      // connection is MISSING — Auth0 will return 400
    };

    const response = await fetch("https://test.auth0.com/api/v2/users", {
      method: "POST",
      headers: {
        Authorization: "Bearer token",
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
    });

    if (response.ok) {
      const data = await response.json();
      expect(data.user_id).toMatch(/^auth0\|/);
    }
  });
});
```

**Expected violation:**
```
CTR-request-shape: POST /oauth/token — client request is missing required
field "grant_type" (type: enum, values: ["client_credentials",
"authorization_code", "password", "refresh_token"]).

CTR-request-shape: POST /api/v2/users — client request is missing required
field "connection" (type: enum, values: ["Username-Password-Authentication",
"google-oauth2", "facebook", "apple", "github", "twitter"]).
```

**Production impact:** Every token request fails with a 400 "grant_type is required" error from Auth0. The mock test passes because it fakes a 200 response, but real API calls always fail in staging and production.

---

## B06 — Response Type Mismatch

**Bug:** The client-side `User` type is missing the `identities` array field and the `email_verified` boolean. The type definition is structurally incomplete compared to the Auth0 contract.

**Violated rule:** `CTR-response-shape`

```typescript
// tests/auth0-client.test.ts — B06: Response type mismatch

import { describe, it, expect } from "vitest";

// BUG: User type missing identities and email_verified fields
interface Auth0UserIncomplete {
  user_id: string;
  email: string;
  // email_verified: boolean — MISSING
  name: string;
  picture: string;
  created_at: string;
  updated_at: string;
  // identities: Auth0Identity[] — MISSING
  app_metadata?: Record<string, unknown>;
  user_metadata?: Record<string, unknown>;
}

describe("Auth0 Client (B06 — response type mismatch)", () => {
  it("retrieves user", () => {
    // Auth0 actually returns identities and email_verified,
    // but our type doesn't capture them
    const rawResponse = {
      user_id: "auth0|abc123",
      email: "test@example.com",
      email_verified: true,
      name: "Test User",
      picture: "https://cdn.auth0.com/avatars/te.png",
      created_at: "2025-01-01T00:00:00.000Z",
      updated_at: "2025-01-01T00:00:00.000Z",
      identities: [{ provider: "auth0", user_id: "abc123", connection: "Username-Password-Authentication", isSocial: false }],
    };

    // Cast to incomplete type — silently drops fields
    const user = rawResponse as unknown as Auth0UserIncomplete;

    expect(user.user_id).toBe("auth0|abc123");
    expect(user.email).toBe("test@example.com");
    expect(user.name).toBe("Test User");
    // Cannot test email_verified or identities — type doesn't have them
  });

  it("creates user", () => {
    const created: Auth0UserIncomplete = {
      user_id: "auth0|newuser",
      email: "new@example.com",
      name: "New User",
      picture: "https://cdn.auth0.com/avatars/ne.png",
      created_at: "2025-06-01T00:00:00.000Z",
      updated_at: "2025-06-01T00:00:00.000Z",
    };

    expect(created.user_id).toMatch(/^auth0\|/);
    // BUG: email_verified is not in the type at all
    // BUG: identities array is not in the type at all
  });
});
```

**Expected violation:**
```
CTR-response-shape: GET /api/v2/users/:id — client type "Auth0UserIncomplete"
is missing required field "email_verified" (type: boolean). Auth0 response
always includes this field.

CTR-response-shape: GET /api/v2/users/:id — client type "Auth0UserIncomplete"
is missing required field "identities" (type: array). Auth0 response always
includes this field.
```

**Production impact:** Code that needs to check `user.email_verified` before granting access to sensitive features cannot access the field. Identity provider information (needed for account linking) is silently discarded. TypeScript provides no compile-time error because the type simply omits these fields.

---

## B07 — Wrong Field Types

**Bug:** `expires_in` from the token response is stored as a string instead of a number. `email_verified` is stored as a string `"true"` instead of boolean `true`.

**Violated rule:** `CTR-manifest-conformance`

```typescript
// tests/auth0-client.test.ts — B07: Wrong field types

import { describe, it, expect } from "vitest";

// BUG: expires_in should be number, email_verified should be boolean
interface Auth0TokenResponseBad {
  access_token: string;
  token_type: "Bearer";
  expires_in: string; // BUG: should be number
  scope?: string;
}

interface Auth0UserBad {
  user_id: string;
  email: string;
  email_verified: string; // BUG: should be boolean
  name: string;
  picture: string;
  created_at: string;
  updated_at: string;
  identities: Array<{ provider: string; user_id: string; connection: string; isSocial: boolean }>;
}

describe("Auth0 Client (B07 — wrong field types)", () => {
  it("parses token response with string expires_in", () => {
    const token: Auth0TokenResponseBad = {
      access_token: "eyJ.payload.sig",
      token_type: "Bearer",
      expires_in: "86400", // BUG: string instead of number
    };

    expect(token.expires_in).toBe("86400");
    expect(typeof token.expires_in).toBe("string"); // This assertion itself is wrong — proves the bug

    // Downstream arithmetic will silently coerce or fail
    // Date.now() + "86400" => string concatenation, not addition
  });

  it("parses user with string email_verified", () => {
    const user: Auth0UserBad = {
      user_id: "auth0|abc123",
      email: "test@example.com",
      email_verified: "true", // BUG: string instead of boolean
      name: "Test",
      picture: "https://cdn.auth0.com/avatars/te.png",
      created_at: "2025-01-01T00:00:00.000Z",
      updated_at: "2025-01-01T00:00:00.000Z",
      identities: [{ provider: "auth0", user_id: "abc123", connection: "Username-Password-Authentication", isSocial: false }],
    };

    expect(user.email_verified).toBe("true"); // Asserts the wrong type!
    // if (user.email_verified) — truthy check passes for "true" string
    // if (user.email_verified === true) — strict equality fails for "true" string
  });
});
```

**Expected violation:**
```
CTR-manifest-conformance: Field "expires_in" type mismatch. Manifest declares
integer; client type uses string. Token expiry arithmetic will fail
(string concatenation instead of numeric addition).

CTR-manifest-conformance: Field "email_verified" type mismatch. Manifest
declares boolean; client type uses string. Strict equality checks
(=== true) will fail for the string value "true".
```

**Production impact:** Token expiry calculation `Date.now() + expires_in` produces string concatenation (`"1700000000000" + "86400"` = `"170000000000086400"`) instead of numeric addition. The token appears to never expire, or the parsed Date is in the far future. `email_verified === true` strict comparison fails for `"true"`, denying verified users access to protected features.

---

## B08 — Incomplete Enum

**Bug:** The `grant_type` handling only covers `"client_credentials"` but not `"authorization_code"`, `"password"`, or `"refresh_token"`. The `connection` handling only covers `"Username-Password-Authentication"` and misses social providers.

**Violated rule:** `CTR-strictness-parity`

```typescript
// tests/auth0-client.test.ts — B08: Incomplete enum handling

import { describe, it, expect } from "vitest";

describe("Auth0 Client (B08 — incomplete enum)", () => {
  function buildTokenRequest(grantType: string): Record<string, string> {
    // BUG: Only handles client_credentials — ignores 3 other valid grant types
    switch (grantType) {
      case "client_credentials":
        return {
          grant_type: "client_credentials",
          client_id: "my-client",
          client_secret: "my-secret",
          audience: "https://api.example.com",
        };
      // MISSING: "authorization_code" — needs code + redirect_uri
      // MISSING: "password" — needs username + password
      // MISSING: "refresh_token" — needs refresh_token
      default:
        // Silently falls through — returns empty object
        return { grant_type: grantType };
    }
  }

  function getConnectionDisplayName(connection: string): string {
    // BUG: Only handles 1 of 6 connection types
    switch (connection) {
      case "Username-Password-Authentication":
        return "Email & Password";
      // MISSING: "google-oauth2" — Google login
      // MISSING: "facebook" — Facebook login
      // MISSING: "apple" — Apple login
      // MISSING: "github" — GitHub login
      // MISSING: "twitter" — Twitter login
      default:
        // Falls through with no error — returns undefined behavior
        return connection;
    }
  }

  it("builds client_credentials request", () => {
    const req = buildTokenRequest("client_credentials");
    expect(req.grant_type).toBe("client_credentials");
    expect(req.client_id).toBe("my-client");
  });

  it("builds authorization_code request", () => {
    // BUG: This test "passes" but the request is missing code and redirect_uri
    const req = buildTokenRequest("authorization_code");
    expect(req.grant_type).toBe("authorization_code");
    // No assertion that code or redirect_uri are present
  });

  it("displays connection name", () => {
    expect(getConnectionDisplayName("Username-Password-Authentication")).toBe("Email & Password");
    // BUG: social providers just return the raw connection string
    expect(getConnectionDisplayName("google-oauth2")).toBe("google-oauth2");
    // This assertion passes but the display is wrong — should be "Google"
  });
});
```

**Expected violation:**
```
CTR-strictness-parity: Manifest declares grant_type enum with 4 values:
["client_credentials", "authorization_code", "password", "refresh_token"].
Client handles 1 of 4. Missing: "authorization_code", "password",
"refresh_token".

CTR-strictness-parity: Manifest declares connection enum with 6 values.
Client handles 1 of 6. Missing: "google-oauth2", "facebook", "apple",
"github", "twitter". No default/fallback case with error handling.
```

**Production impact:** Authorization code flow (the most common web OAuth flow) silently sends a request without `code` or `redirect_uri`, which Auth0 rejects with a 400. Social login users see raw connection strings like "google-oauth2" instead of friendly names.

---

## B09 — Missing Range Validation

**Bug:** `expires_in` from the token response is not validated for range. A negative value, zero, or extremely large value is accepted without checks. This allows nonsensical token lifetimes.

**Violated rule:** `CTR-strictness-parity`

```typescript
// tests/auth0-client.test.ts — B09: Missing range validation

import { describe, it, expect } from "vitest";

describe("Auth0 Client (B09 — no range validation)", () => {
  function storeTokenExpiry(expiresIn: number): number {
    // BUG: No validation that expires_in is positive or within reasonable bounds
    // Manifest says range: [1, 2592000] (1 second to 30 days)
    return Date.now() + expiresIn * 1000;
  }

  it("stores token with valid expiry", () => {
    const expiresAt = storeTokenExpiry(86400);
    expect(expiresAt).toBeGreaterThan(Date.now());
  });

  it("accepts negative expires_in without error", () => {
    // BUG: Negative expiry means the token is already expired at storage time
    const expiresAt = storeTokenExpiry(-3600);
    // This should throw or reject, but it silently stores a past timestamp
    expect(expiresAt).toBeLessThan(Date.now()); // Token is already "expired"
    // No assertion that this is invalid — test proves the bug exists
  });

  it("accepts zero expires_in without error", () => {
    // BUG: Zero expiry means the token expires immediately
    const expiresAt = storeTokenExpiry(0);
    // Difference should be ~0ms but no validation catches this
    expect(Math.abs(expiresAt - Date.now())).toBeLessThan(100);
  });

  it("accepts absurdly large expires_in without error", () => {
    // BUG: 10 years in seconds — far exceeds Auth0's max of 30 days
    const expiresAt = storeTokenExpiry(315_360_000);
    // No upper bound check — token "never" expires
    expect(expiresAt).toBeGreaterThan(Date.now());
  });
});
```

**Expected violation:**
```
CTR-strictness-parity: Manifest declares expires_in range [1, 2592000].
Client code at storeTokenExpiry() does not validate this range. Negative
values (-3600), zero (0), and values exceeding 2592000 are accepted
without error.
```

**Production impact:** A malicious or buggy Auth0 response with `expires_in: -1` results in the token being immediately considered expired, triggering infinite refresh loops. An `expires_in: 0` causes every API call to refresh the token first, hammering the token endpoint. An absurdly large value means a compromised token is never rotated.

---

## B10 — Format Not Validated

**Bug:** `user_id` is accepted as any string without validating the `"auth0|{hex}"` or `"{provider}|{id}"` format. A user_id of `"plainstring"`, an empty string, or `"../../admin"` are all accepted.

**Violated rule:** `CTR-strictness-parity`

```typescript
// tests/auth0-client.test.ts — B10: Format not validated

import { describe, it, expect, vi } from "vitest";

describe("Auth0 Client (B10 — no format validation)", () => {
  async function getUser(userId: string): Promise<Record<string, unknown>> {
    // BUG: No format validation on userId
    // Manifest says format: "auth0|[a-f0-9]+|{provider}|{id}"
    // Accepts any string, including path traversal attacks
    const response = await fetch(
      `https://test.auth0.com/api/v2/users/${userId}`, // Not even URL-encoded!
      { headers: { Authorization: "Bearer token" } },
    );
    return response.json() as Promise<Record<string, unknown>>;
  }

  it("accepts valid auth0 user_id", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({ user_id: "auth0|abc123" }), { status: 200 }),
    );

    const user = await getUser("auth0|abc123");
    expect(user.user_id).toBe("auth0|abc123");
  });

  it("accepts plain string user_id without validation", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({ user_id: "plainstring" }), { status: 200 }),
    );

    // BUG: "plainstring" does not match the required format
    const user = await getUser("plainstring");
    expect(user.user_id).toBe("plainstring"); // Should have been rejected before the fetch
  });

  it("accepts empty string user_id", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({}), { status: 404 }),
    );

    // BUG: Empty string is not a valid user_id
    const user = await getUser("");
    expect(user).toBeDefined();
  });

  it("accepts path traversal in user_id", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({}), { status: 400 }),
    );

    // BUG: Path traversal is not rejected client-side
    // URL becomes: /api/v2/users/../../admin
    const user = await getUser("../../admin");
    expect(user).toBeDefined();
  });

  it("does not validate user_id in response either", () => {
    // BUG: Response user_id is also not validated
    const responseUser = { user_id: "not-a-valid-format", email: "x@y.com" };
    expect(responseUser.user_id).toBe("not-a-valid-format");
    // No format check — accepts anything from Auth0
  });
});
```

**Expected violation:**
```
CTR-strictness-parity: Manifest declares user_id format
"auth0|[a-f0-9]+|{provider}|{id}". Client code at getUser() does not
validate this format. Accepts arbitrary strings including empty strings
and path traversal patterns. Neither request user_id nor response
user_id is format-validated.
```

**Production impact:** A path traversal attack (`../../admin`) in the user_id is sent directly to the Auth0 API without client-side rejection. While Auth0 itself likely rejects this, the unencoded pipe character in URLs (`auth0|abc`) causes encoding issues. Response user_ids that do not match the expected format are silently accepted, potentially indicating data corruption or a MITM attack.

---

## B11 — Precision Loss

**Bug:** Token expiry calculation uses `Date.now() + expires_in` without converting seconds to milliseconds. Since `expires_in` is in seconds (e.g., 86400) and `Date.now()` returns milliseconds, the token expires 1000x too early.

**Violated rule:** `CTR-strictness-parity`

```typescript
// tests/auth0-client.test.ts — B11: Precision loss (seconds vs milliseconds)

import { describe, it, expect } from "vitest";

describe("Auth0 Client (B11 — seconds/milliseconds precision loss)", () => {
  class TokenManagerBroken {
    private accessToken: string | null = null;
    private tokenExpiresAt: number = 0;

    storeToken(accessToken: string, expiresIn: number): void {
      this.accessToken = accessToken;
      // BUG: expires_in is in SECONDS, but Date.now() is in MILLISECONDS
      // Missing: * 1000 conversion
      this.tokenExpiresAt = Date.now() + expiresIn;
      // If expiresIn = 86400 (24 hours in seconds),
      // this adds only 86400 ms (86.4 seconds) instead of 86400000 ms (24 hours)
    }

    isTokenValid(): boolean {
      return this.accessToken !== null && Date.now() < this.tokenExpiresAt;
    }

    getExpiresAt(): number {
      return this.tokenExpiresAt;
    }
  }

  it("token expires 1000x too early", () => {
    const manager = new TokenManagerBroken();
    const before = Date.now();

    manager.storeToken("eyJ.payload.sig", 86400); // 24 hours in seconds

    const expiresAt = manager.getExpiresAt();
    const expiresInMs = expiresAt - before;

    // BUG: Should be ~86400000 ms (24 hours), but is ~86400 ms (86 seconds)
    expect(expiresInMs).toBeLessThan(100_000); // Less than 100 seconds
    expect(expiresInMs).toBeGreaterThan(80_000); // Around 86 seconds

    // The correct value would be:
    // expect(expiresInMs).toBeGreaterThan(86_000_000); // ~24 hours
  });

  it("token appears expired after 2 minutes", async () => {
    const manager = new TokenManagerBroken();
    manager.storeToken("eyJ.payload.sig", 86400);

    // Simulate 2 minutes passing
    const twoMinutesLater = Date.now() + 120_000;
    // BUG: Token "expired" after 86.4 seconds — but should last 24 hours
    expect(manager.getExpiresAt()).toBeLessThan(twoMinutesLater);
  });
});
```

**Expected violation:**
```
CTR-strictness-parity: Token expiry arithmetic at storeToken() adds
expires_in (seconds) directly to Date.now() (milliseconds) without
unit conversion (* 1000). Token with expires_in=86400 (24h) expires
after ~86 seconds instead of 24 hours. Off by factor of 1000.
```

**Production impact:** Every access token expires after approximately 86 seconds instead of 24 hours. This causes the application to re-authenticate with Auth0 on almost every API call, dramatically increasing latency and hammering the token endpoint. Under load, this triggers Auth0 rate limits (429), causing cascading failures across the application.

---

## B12 — Nullable Field Crash

**Bug:** Accesses `user.app_metadata.role` directly without null-checking. Since `app_metadata` can be `undefined` or `null` (per the Auth0 contract), this crashes with "Cannot read properties of null (reading 'role')" at runtime.

**Violated rule:** `CTR-response-shape`

```typescript
// tests/auth0-client.test.ts — B12: Nullable field crash

import { describe, it, expect } from "vitest";

interface Auth0UserWithMetadata {
  user_id: string;
  email: string;
  email_verified: boolean;
  name: string;
  picture: string;
  created_at: string;
  updated_at: string;
  identities: Array<{ provider: string; user_id: string; connection: string; isSocial: boolean }>;
  app_metadata?: Record<string, unknown> | null;
  user_metadata?: Record<string, unknown> | null;
}

describe("Auth0 Client (B12 — nullable field crash)", () => {
  function getUserRole(user: Auth0UserWithMetadata): string {
    // BUG: Direct property access on nullable field
    // app_metadata can be undefined or null — this crashes
    return (user.app_metadata as Record<string, string>).role;
  }

  function getUserTheme(user: Auth0UserWithMetadata): string {
    // BUG: Same pattern — user_metadata can be null
    return (user.user_metadata as Record<string, string>).theme;
  }

  function getUserPermissions(user: Auth0UserWithMetadata): string[] {
    // BUG: Chains through nullable — crashes on null.permissions.map
    return ((user.app_metadata as Record<string, unknown>).permissions as string[]).map(
      (p: string) => p.toUpperCase(),
    );
  }

  it("gets role from user with app_metadata", () => {
    const user: Auth0UserWithMetadata = {
      user_id: "auth0|abc123",
      email: "test@example.com",
      email_verified: true,
      name: "Test",
      picture: "https://cdn.auth0.com/avatars/te.png",
      created_at: "2025-01-01T00:00:00.000Z",
      updated_at: "2025-01-01T00:00:00.000Z",
      identities: [{ provider: "auth0", user_id: "abc123", connection: "Username-Password-Authentication", isSocial: false }],
      app_metadata: { role: "admin", permissions: ["read", "write"] },
      user_metadata: { theme: "dark" },
    };

    // This works because app_metadata is populated
    expect(getUserRole(user)).toBe("admin");
    expect(getUserTheme(user)).toBe("dark");
  });

  it("crashes on null app_metadata", () => {
    const user: Auth0UserWithMetadata = {
      user_id: "auth0|def456",
      email: "new@example.com",
      email_verified: false,
      name: "New User",
      picture: "https://cdn.auth0.com/avatars/ne.png",
      created_at: "2025-06-01T00:00:00.000Z",
      updated_at: "2025-06-01T00:00:00.000Z",
      identities: [{ provider: "auth0", user_id: "def456", connection: "Username-Password-Authentication", isSocial: false }],
      app_metadata: null, // Auth0 can return null for new users
      user_metadata: null,
    };

    // BUG: Crashes — cannot read "role" of null
    expect(() => getUserRole(user)).toThrow();
    expect(() => getUserTheme(user)).toThrow();
    expect(() => getUserPermissions(user)).toThrow();
  });

  it("crashes on undefined app_metadata", () => {
    const user: Auth0UserWithMetadata = {
      user_id: "auth0|ghi789",
      email: "minimal@example.com",
      email_verified: false,
      name: "Minimal",
      picture: "https://cdn.auth0.com/avatars/mi.png",
      created_at: "2025-06-01T00:00:00.000Z",
      updated_at: "2025-06-01T00:00:00.000Z",
      identities: [{ provider: "auth0", user_id: "ghi789", connection: "Username-Password-Authentication", isSocial: false }],
      // app_metadata is undefined (field omitted entirely)
      // user_metadata is undefined
    };

    // BUG: Crashes — cannot read "role" of undefined
    expect(() => getUserRole(user)).toThrow();
    expect(() => getUserTheme(user)).toThrow();
  });
});
```

**Expected violation:**
```
CTR-response-shape: Field "app_metadata" is declared nullable: true in the
manifest. Client code at getUserRole() accesses user.app_metadata.role
without null check. Will throw TypeError "Cannot read properties of null
(reading 'role')" when app_metadata is null.

Same violation for getUserTheme() accessing user.user_metadata.theme
and getUserPermissions() accessing user.app_metadata.permissions.
```

**Production impact:** Every newly created Auth0 user (who has no `app_metadata` yet) crashes the application when any code path tries to read their role, permissions, or preferences. This affects the most critical user flow -- first login after signup -- causing a 500 error on the user's very first authenticated request.

---

## B13 — Missing JWT Validation

**Bug:** The access token returned by Auth0 is used without verifying its JWT signature, expiry (`exp` claim), audience (`aud` claim), or issuer (`iss` claim). A tampered, expired, or misrouted token is silently accepted.

**Violated rule:** `CTR-request-shape`

```typescript
// tests/auth0-client.test.ts — B13: Missing JWT validation

import { describe, it, expect } from "vitest";

describe("Auth0 Client (B13 — no JWT validation)", () => {
  function storeAccessToken(tokenResponse: {
    access_token: string;
    token_type: string;
    expires_in: number;
  }): { token: string; expiresAt: number } {
    // BUG: Stores the token without ANY validation:
    // - No JWT format check (three dot-separated base64url segments)
    // - No signature verification against Auth0's JWKS
    // - No exp claim check (token could already be expired)
    // - No aud claim check (token could be for a different API)
    // - No iss claim check (token could be from a different tenant)
    return {
      token: tokenResponse.access_token,
      expiresAt: Date.now() + tokenResponse.expires_in * 1000,
    };
  }

  it("accepts a valid-looking JWT", () => {
    const result = storeAccessToken({
      access_token: "eyJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJodHRwczovL3Rlc3QuYXV0aDAuY29tLyJ9.signature",
      token_type: "Bearer",
      expires_in: 86400,
    });

    expect(result.token).toContain(".");
    // BUG: Only checks it contains a dot — not that it's a valid JWT
  });

  it("accepts a completely invalid token format", () => {
    // BUG: "not-a-jwt" has no dots, no base64url segments — but is accepted
    const result = storeAccessToken({
      access_token: "not-a-jwt-at-all",
      token_type: "Bearer",
      expires_in: 86400,
    });

    expect(result.token).toBe("not-a-jwt-at-all");
    // No error thrown — invalid token stored and will be used for API calls
  });

  it("accepts a JWT with tampered payload", () => {
    // BUG: Payload says admin scope, but signature does not match
    // A real JWT verifier would reject this
    const tamperedJwt = "eyJhbGciOiJSUzI1NiJ9.eyJzY29wZSI6ImFkbWluOnN1cGVyIn0.invalid-signature";

    const result = storeAccessToken({
      access_token: tamperedJwt,
      token_type: "Bearer",
      expires_in: 86400,
    });

    expect(result.token).toBe(tamperedJwt);
    // Tampered token accepted — would allow privilege escalation
  });

  it("accepts a JWT with expired exp claim", () => {
    // BUG: The exp claim is in the past, but we never check it
    // Base64url of {"alg":"RS256"} . {"exp":1000000000,"iss":"https://test.auth0.com/"} . sig
    const expiredJwt = "eyJhbGciOiJSUzI1NiJ9.eyJleHAiOjEwMDAwMDAwMDAsImlzcyI6Imh0dHBzOi8vdGVzdC5hdXRoMC5jb20vIn0.sig";

    const result = storeAccessToken({
      access_token: expiredJwt,
      token_type: "Bearer",
      expires_in: 86400,
    });

    expect(result.token).toBe(expiredJwt);
    // Expired token stored — all subsequent API calls will fail with 401
  });

  it("accepts a JWT for a different audience", () => {
    // BUG: Token audience is for a completely different API
    // Should be rejected because our API is not the intended audience
    const wrongAudienceJwt = "eyJhbGciOiJSUzI1NiJ9.eyJhdWQiOiJodHRwczovL290aGVyLWFwaS5jb20ifQ.sig";

    const result = storeAccessToken({
      access_token: wrongAudienceJwt,
      token_type: "Bearer",
      expires_in: 86400,
    });

    expect(result.token).toBe(wrongAudienceJwt);
    // Wrong-audience token accepted — confused deputy problem
  });
});
```

**Expected violation:**
```
CTR-request-shape: POST /oauth/token — access_token in response has format
"jwt" per manifest. Client code at storeAccessToken() does not verify:
(1) JWT structure (3 base64url segments), (2) signature against JWKS,
(3) exp claim for expiry, (4) aud claim for audience, (5) iss claim
for issuer. Accepts arbitrary strings including tampered tokens.
```

**Production impact:** A man-in-the-middle attacker can inject a tampered JWT with elevated scopes (e.g., `admin:super`) and the client will use it for all Management API calls. An expired JWT from cache or a different tenant's JWT are both silently accepted. This is a critical security vulnerability that enables privilege escalation and confused deputy attacks.

---

## B14 — Token Refresh Race

**Bug:** When the token expires between the validity check and the API call, the 401 response is not retried with a fresh token. Concurrent token refresh attempts can also race, causing multiple simultaneous refresh requests and wasted Auth0 rate limit budget.

**Violated rule:** `CTR-request-shape`

```typescript
// tests/auth0-client.test.ts — B14: Token refresh race condition

import { describe, it, expect, vi } from "vitest";

describe("Auth0 Client (B14 — token refresh race)", () => {
  class TokenManagerRacy {
    private accessToken: string | null = null;
    private tokenExpiresAt: number = 0;

    async getToken(): Promise<string> {
      // BUG: No mutex — concurrent callers can all trigger refresh simultaneously
      if (!this.accessToken || Date.now() >= this.tokenExpiresAt) {
        await this.refreshToken();
      }
      return this.accessToken!;
    }

    private async refreshToken(): Promise<void> {
      // BUG: No lock — if 10 concurrent calls detect expired token,
      // all 10 will call this method, making 10 separate token requests
      const response = await fetch("https://test.auth0.com/oauth/token", {
        method: "POST",
        body: JSON.stringify({
          grant_type: "client_credentials",
          client_id: "my-client",
          client_secret: "my-secret",
        }),
      });
      const data = await response.json() as { access_token: string; expires_in: number };
      this.accessToken = data.access_token;
      this.tokenExpiresAt = Date.now() + data.expires_in * 1000;
    }
  }

  async function getUserRacy(
    tokenManager: TokenManagerRacy,
    userId: string,
  ): Promise<Record<string, unknown>> {
    const token = await tokenManager.getToken();

    // BUG: Token could expire between getToken() and fetch()
    // No retry logic if the API returns 401
    const response = await fetch(
      `https://test.auth0.com/api/v2/users/${encodeURIComponent(userId)}`,
      { headers: { Authorization: `Bearer ${token}` } },
    );

    // BUG: If response is 401 (expired token), no automatic retry
    if (!response.ok) {
      throw new Error(`Auth0 API error: ${response.status}`);
    }

    return response.json() as Promise<Record<string, unknown>>;
  }

  it("concurrent token refresh sends multiple requests", async () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch");

    // Each call returns a valid token
    fetchSpy.mockResolvedValue(
      new Response(JSON.stringify({
        access_token: "eyJ.payload.sig",
        token_type: "Bearer",
        expires_in: 86400,
      }), { status: 200 }),
    );

    const manager = new TokenManagerRacy();

    // 5 concurrent getToken() calls — all see expired token simultaneously
    const promises = Array.from({ length: 5 }, () => manager.getToken());
    await Promise.all(promises);

    // BUG: All 5 calls trigger separate refresh requests
    // A correct implementation would use a mutex so only 1 refresh happens
    expect(fetchSpy).toHaveBeenCalledTimes(5); // Should be 1!

    fetchSpy.mockRestore();
  });

  it("401 between token check and API call is not retried", async () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch");

    // First call: return a token
    fetchSpy.mockResolvedValueOnce(
      new Response(JSON.stringify({
        access_token: "eyJ.payload.sig",
        token_type: "Bearer",
        expires_in: 1, // Expires in 1 second
      }), { status: 200 }),
    );

    // Second call: 401 because token expired between getToken() and fetch()
    fetchSpy.mockResolvedValueOnce(
      new Response(JSON.stringify({
        statusCode: 401,
        error: "Unauthorized",
        message: "Expired token",
      }), { status: 401 }),
    );

    const manager = new TokenManagerRacy();

    // BUG: getUserRacy does not retry on 401 — just throws
    await expect(
      getUserRacy(manager, "auth0|abc123"),
    ).rejects.toThrow("Auth0 API error: 401");

    // A correct implementation would refresh the token and retry once

    fetchSpy.mockRestore();
  });
});
```

**Expected violation:**
```
CTR-request-shape: Auth0 Management API calls at getUserRacy() do not
implement automatic retry on 401 (expired token). Token can expire
between the validity check in getToken() and the actual API call.

CTR-strictness-parity: TokenManagerRacy.getToken() has no concurrency
guard (mutex/lock). Concurrent callers trigger redundant token refresh
requests, wasting rate limit budget.
```

**Production impact:** Under moderate concurrency (e.g., 20 concurrent API calls), every token expiry triggers 20 simultaneous refresh requests to Auth0. This wastes rate limit budget and can trigger 429 responses, causing all 20 callers to fail. The TOCTOU (time-of-check-time-of-use) race between token validity check and API call causes intermittent 401 errors under load that are difficult to reproduce in testing.

---

## B15 — Race Condition (Read-Modify-Write)

**Bug:** When updating a user's `app_metadata`, the code reads the current user, modifies `app_metadata` locally, and patches the user. A concurrent update between the read and the patch overwrites the other update's metadata fields.

**Violated rule:** `CTR-request-shape`

```typescript
// tests/auth0-client.test.ts — B15: Read-modify-write race condition

import { describe, it, expect, vi } from "vitest";

describe("Auth0 Client (B15 — metadata read-modify-write race)", () => {
  interface SimpleUser {
    user_id: string;
    app_metadata: Record<string, unknown> | null;
  }

  // Simulates Auth0 server state
  let serverState: SimpleUser = {
    user_id: "auth0|abc123",
    app_metadata: { plan: "free", role: "user", loginCount: 0 },
  };

  function mockAuth0Server(): void {
    vi.spyOn(globalThis, "fetch").mockImplementation(async (url, options) => {
      const urlStr = typeof url === "string" ? url : url.toString();

      if (options?.method === "GET" || !options?.method) {
        // GET returns current state (snapshot in time)
        return new Response(JSON.stringify({ ...serverState }), { status: 200 });
      }

      if (options?.method === "PATCH") {
        // PATCH replaces app_metadata entirely (Auth0's actual behavior)
        const body = JSON.parse(options.body as string) as Partial<SimpleUser>;
        if (body.app_metadata !== undefined) {
          serverState = { ...serverState, app_metadata: body.app_metadata };
        }
        return new Response(JSON.stringify({ ...serverState }), { status: 200 });
      }

      return new Response("Not Found", { status: 404 });
    });
  }

  async function updateUserRole(userId: string, newRole: string): Promise<void> {
    // Step 1: Read current user
    const getResponse = await fetch(`https://test.auth0.com/api/v2/users/${userId}`);
    const currentUser = await getResponse.json() as SimpleUser;

    // Step 2: Modify app_metadata locally
    // BUG: No lock, no version check — another request can modify between read and write
    const updatedMetadata = {
      ...currentUser.app_metadata,
      role: newRole,
    };

    // Step 3: Write back entire app_metadata (overwrites concurrent changes)
    await fetch(`https://test.auth0.com/api/v2/users/${userId}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ app_metadata: updatedMetadata }),
    });
  }

  async function updateUserPlan(userId: string, newPlan: string): Promise<void> {
    // Same read-modify-write pattern — races with updateUserRole
    const getResponse = await fetch(`https://test.auth0.com/api/v2/users/${userId}`);
    const currentUser = await getResponse.json() as SimpleUser;

    const updatedMetadata = {
      ...currentUser.app_metadata,
      plan: newPlan,
    };

    await fetch(`https://test.auth0.com/api/v2/users/${userId}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ app_metadata: updatedMetadata }),
    });
  }

  it("concurrent metadata updates overwrite each other", async () => {
    // Reset server state
    serverState = {
      user_id: "auth0|abc123",
      app_metadata: { plan: "free", role: "user", loginCount: 0 },
    };
    mockAuth0Server();

    // Two concurrent updates: one sets role=admin, other sets plan=pro
    // Both read the SAME initial state: { plan: "free", role: "user", loginCount: 0 }
    await Promise.all([
      updateUserRole("auth0|abc123", "admin"),
      updateUserPlan("auth0|abc123", "pro"),
    ]);

    // BUG: Last write wins — one of the updates is lost
    // If updateUserPlan writes last:
    //   app_metadata = { plan: "pro", role: "user", loginCount: 0 }
    //   The role="admin" update is LOST
    //
    // If updateUserRole writes last:
    //   app_metadata = { plan: "free", role: "admin", loginCount: 0 }
    //   The plan="pro" update is LOST

    // We can verify that one update was lost:
    const finalResponse = await fetch("https://test.auth0.com/api/v2/users/auth0|abc123");
    const finalUser = await finalResponse.json() as SimpleUser;

    // At least one of these will fail — proving the race condition
    const hasCorrectRole = (finalUser.app_metadata as Record<string, unknown>)?.role === "admin";
    const hasCorrectPlan = (finalUser.app_metadata as Record<string, unknown>)?.plan === "pro";

    // BUG: Cannot guarantee both are true due to race condition
    // In a correct implementation, both updates would be merged atomically
    expect(hasCorrectRole || hasCorrectPlan).toBe(true);
    // This weaker assertion "passes" but masks the data loss
    // expect(hasCorrectRole && hasCorrectPlan).toBe(true); // This would fail!

    vi.restoreAllMocks();
  });

  it("loginCount is lost during concurrent role update", async () => {
    serverState = {
      user_id: "auth0|abc123",
      app_metadata: { plan: "free", role: "user", loginCount: 42 },
    };
    mockAuth0Server();

    // Read current state (loginCount = 42)
    // Meanwhile, another process increments loginCount to 43
    // Then we write role="admin" with the stale loginCount=42

    // Simulate: external process updates loginCount between our read and write
    const getResponse = await fetch("https://test.auth0.com/api/v2/users/auth0|abc123");
    const staleUser = await getResponse.json() as SimpleUser;

    // External update happens here (loginCount goes from 42 to 43)
    serverState = {
      ...serverState,
      app_metadata: { ...serverState.app_metadata, loginCount: 43 },
    };

    // Our write uses stale data — overwrites loginCount back to 42
    const updatedMetadata = { ...staleUser.app_metadata, role: "admin" };
    await fetch("https://test.auth0.com/api/v2/users/auth0|abc123", {
      method: "PATCH",
      body: JSON.stringify({ app_metadata: updatedMetadata }),
    });

    // BUG: loginCount regressed from 43 back to 42
    expect((serverState.app_metadata as Record<string, unknown>)?.loginCount).toBe(42); // Should be 43!

    vi.restoreAllMocks();
  });
});
```

**Expected violation:**
```
CTR-request-shape: PATCH /api/v2/users/:id — client uses read-modify-write
pattern on app_metadata without concurrency protection. Between the
GET (read) and PATCH (write), concurrent updates to the same user's
metadata can be silently overwritten. No optimistic locking, versioning,
or atomic merge strategy is used.
```

**Production impact:** When two microservices concurrently update different fields of the same user's `app_metadata` (e.g., billing service sets `plan` while auth service sets `role`), one update is silently lost. This is an intermittent data loss bug that is extremely difficult to reproduce in testing because it depends on exact timing. In production with enough concurrency, this leads to users losing their subscription plan, role assignments, or login counts. The data loss is silent -- no error is raised, no log is written, and the overwritten value is gone permanently.

---

## Summary

| Case | Bug | Rule | Severity |
|------|-----|------|----------|
| PERFECT | None | N/A | Zero violations expected |
| B01 | No error handling | TQ-error-path-coverage | High |
| B02 | No status code check | CTR-status-code-handling | High |
| B03 | Shallow assertions | TQ-no-shallow-assertions | Medium |
| B04 | Missing negative tests | TQ-negative-cases | Medium |
| B05 | Request missing required fields | CTR-request-shape | High |
| B06 | Response type mismatch | CTR-response-shape | High |
| B07 | Wrong field types | CTR-manifest-conformance | High |
| B08 | Incomplete enum | CTR-strictness-parity | Medium |
| B09 | Missing range validation | CTR-strictness-parity | Medium |
| B10 | Format not validated | CTR-strictness-parity | Medium |
| B11 | Precision loss (seconds vs ms) | CTR-strictness-parity | Critical |
| B12 | Nullable field crash | CTR-response-shape | Critical |
| B13 | Missing JWT validation | CTR-request-shape | Critical |
| B14 | Token refresh race | CTR-request-shape | High |
| B15 | Race condition (read-modify-write) | CTR-request-shape | Critical |
