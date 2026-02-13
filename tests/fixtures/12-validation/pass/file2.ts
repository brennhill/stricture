// supabase-client.test.ts -- Comprehensive tests for Supabase integration.

import { describe, it, expect, beforeEach, vi, afterEach } from "vitest";
import {
  SupabaseClient,
  parseContentRange,
  parseUserRow,
  isTokenExpired,
  validateUUIDv4,
  validateEmail,
  validateJWT,
  validateISO8601,
  validateRole,
  validatePageSize,
  validatePassword,
} from "./supabase-client";
import type {
  UserRow,
  AuthSession,
  SupabaseError,
  AuthError,
  RealtimeMessage,
} from "./supabase-client";

// ── Fixtures ────────────────────────────────────────────────

const VALID_UUID = "550e8400-e29b-41d4-a716-446655440000";

const MOCK_USER_ROW: Record<string, unknown> = {
  id: VALID_UUID,
  email: "alice@example.com",
  user_role: "editor",
  profile: { avatar_url: "https://example.com/avatar.jpg", bio: "Hello" },
  balance: "1000",
  metadata: { theme: "dark" },
  created_at: "2026-01-15T10:30:00.000Z",
  updated_at: "2026-02-01T14:00:00.000Z",
};

const MOCK_USER_NULL_PROFILE: Record<string, unknown> = {
  ...MOCK_USER_ROW,
  id: "660e8400-e29b-41d4-a716-446655440001",
  profile: null,
  metadata: null,
};

const MOCK_USER_BIGINT: Record<string, unknown> = {
  ...MOCK_USER_ROW,
  id: "770e8400-e29b-41d4-a716-446655440002",
  balance: "9007199254740993", // MAX_SAFE_INTEGER + 2
};

const MOCK_AUTH_RESPONSE = {
  id: VALID_UUID,
  aud: "authenticated",
  role: "authenticated",
  email: "new@example.com",
  phone: null,
  confirmation_sent_at: "2026-02-12T10:00:00.000Z",
  email_confirmed_at: null,
  created_at: "2026-02-12T10:00:00.000Z",
  updated_at: "2026-02-12T10:00:00.000Z",
  user_metadata: { display_name: "New User" },
  app_metadata: { provider: "email" },
  identities: [{ id: VALID_UUID, provider: "email" }],
  access_token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
  token_type: "bearer",
  expires_in: 3600,
  expires_at: Math.floor(Date.now() / 1000) + 3600,
  refresh_token: "refresh_abc123",
};

const MOCK_POSTGREST_ERROR_401: SupabaseError = {
  message: "JWT expired",
  details: null,
  hint: "Provide a valid JWT",
  code: "PGRST301",
};

const MOCK_POSTGREST_ERROR_403: SupabaseError = {
  message: "permission denied for table users",
  details: null,
  hint: null,
  code: "42501",
};

const MOCK_POSTGREST_ERROR_409: SupabaseError = {
  message: "duplicate key value violates unique constraint",
  details: "Key (email)=(alice@example.com) already exists.",
  hint: null,
  code: "23505",
};

const MOCK_POSTGREST_ERROR_416: SupabaseError = {
  message: "Requested range not satisfiable",
  details: null,
  hint: null,
  code: "PGRST103",
};

const MOCK_AUTH_ERROR_422: AuthError = {
  error: "invalid_grant",
  error_description: "Email address is not valid",
  msg: "Unable to validate email address: invalid format",
  code: 422,
};

const MOCK_AUTH_ERROR_429: AuthError = {
  msg: "For security purposes, you can only request this after 60 seconds.",
  code: 429,
};

// ── Helpers ─────────────────────────────────────────────────

function mockFetch(status: number, body: unknown, headers?: Record<string, string>): void {
  global.fetch = vi.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    json: vi.fn().mockResolvedValue(body),
    headers: new Map(Object.entries(headers ?? {})),
  } as unknown as Response);
}

function mockFetchSequence(
  ...responses: Array<{ status: number; body: unknown; headers?: Record<string, string> }>
): void {
  const fn = vi.fn();
  for (const [i, resp] of responses.entries()) {
    fn.mockResolvedValueOnce({
      ok: resp.status >= 200 && resp.status < 300,
      status: resp.status,
      json: vi.fn().mockResolvedValue(resp.body),
      headers: new Map(Object.entries(resp.headers ?? {})),
    } as unknown as Response);
  }
  global.fetch = fn;
}

function mockFetchNetworkError(message: string): void {
  global.fetch = vi.fn().mockRejectedValue(new Error(message));
}

function createClient(): SupabaseClient {
  return new SupabaseClient(
    "https://test-project.supabase.co",
    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSJ9.fake",
    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
    "refresh_token_abc",
    Math.floor(Date.now() / 1000) + 3600 // expires 1 hour from now
  );
}

// ── Tests ───────────────────────────────────────────────────

describe("SupabaseClient", () => {
  let client: SupabaseClient;

  beforeEach(() => {
    client = createClient();
    vi.restoreAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  // ── Validation Helpers ──────────────────────────────

  describe("validateUUIDv4", () => {
    it("accepts valid UUID v4", () => {
      expect(() => validateUUIDv4(VALID_UUID)).not.toThrow();
    });

    it("rejects UUID v1", () => {
      expect(() => validateUUIDv4("6ba7b810-9dad-11d1-80b4-00c04fd430c8")).toThrow("Invalid UUID v4");
    });

    it("rejects non-UUID strings", () => {
      expect(() => validateUUIDv4("not-a-uuid")).toThrow("Invalid UUID v4");
      expect(() => validateUUIDv4("")).toThrow("Invalid UUID v4");
      expect(() => validateUUIDv4("123")).toThrow("Invalid UUID v4");
    });
  });

  describe("validateEmail", () => {
    it("accepts valid emails", () => {
      expect(() => validateEmail("user@example.com")).not.toThrow();
    });

    it("rejects invalid emails", () => {
      expect(() => validateEmail("not-an-email")).toThrow("Invalid email");
      expect(() => validateEmail("")).toThrow("Invalid email");
    });
  });

  describe("validateRole", () => {
    it("accepts all valid roles", () => {
      for (const role of ["admin", "editor", "viewer", "guest"]) {
        expect(() => validateRole(role)).not.toThrow();
      }
    });

    it("rejects invalid roles", () => {
      expect(() => validateRole("superadmin")).toThrow("Invalid user_role");
      expect(() => validateRole("")).toThrow("Invalid user_role");
    });
  });

  describe("validatePageSize", () => {
    it("accepts valid page sizes", () => {
      expect(() => validatePageSize(1)).not.toThrow();
      expect(() => validatePageSize(25)).not.toThrow();
      expect(() => validatePageSize(1000)).not.toThrow();
    });

    it("rejects out-of-range sizes", () => {
      expect(() => validatePageSize(0)).toThrow("Page size must be between");
      expect(() => validatePageSize(1001)).toThrow("Page size must be between");
      expect(() => validatePageSize(-1)).toThrow("Page size must be between");
    });
  });

  describe("validatePassword", () => {
    it("accepts valid passwords", () => {
      expect(() => validatePassword("secure123")).not.toThrow();
      expect(() => validatePassword("123456")).not.toThrow();
    });

    it("rejects short passwords", () => {
      expect(() => validatePassword("12345")).toThrow("at least 6 characters");
      expect(() => validatePassword("")).toThrow("at least 6 characters");
    });
  });

  // ── Content-Range Parser ────────────────────────────

  describe("parseContentRange", () => {
    it("parses valid Content-Range header", () => {
      const result = parseContentRange("0-9/100");
      expect(result).toEqual({ start: 0, end: 9, total: 100 });
    });

    it("handles unknown total (*)", () => {
      const result = parseContentRange("0-24/*");
      expect(result).toEqual({ start: 0, end: 24, total: -1 });
    });

    it("returns null for missing header", () => {
      expect(parseContentRange(null)).toBeNull();
    });

    it("returns null for malformed header", () => {
      expect(parseContentRange("invalid")).toBeNull();
      expect(parseContentRange("abc-def/ghi")).toBeNull();
    });
  });

  // ── Row Parser ──────────────────────────────────────

  describe("parseUserRow", () => {
    it("parses a complete row with all fields", () => {
      const user = parseUserRow(MOCK_USER_ROW);
      expect(user.id).toBe(VALID_UUID);
      expect(user.email).toBe("alice@example.com");
      expect(user.user_role).toBe("editor");
      expect(user.profile).not.toBeNull();
      expect(user.profile!.avatar_url).toBe("https://example.com/avatar.jpg");
      expect(user.profile!.bio).toBe("Hello");
      expect(typeof user.balance).toBe("bigint");
      expect(user.balance).toBe(BigInt(1000));
      expect(user.metadata).toEqual({ theme: "dark" });
      expect(user.created_at).toBe("2026-01-15T10:30:00.000Z");
      expect(user.updated_at).toBe("2026-02-01T14:00:00.000Z");
    });

    it("handles null profile and metadata", () => {
      const user = parseUserRow(MOCK_USER_NULL_PROFILE);
      expect(user.profile).toBeNull();
      expect(user.metadata).toBeNull();
    });

    it("preserves bigint precision beyond MAX_SAFE_INTEGER", () => {
      const user = parseUserRow(MOCK_USER_BIGINT);
      expect(typeof user.balance).toBe("bigint");
      expect(user.balance).toBe(BigInt("9007199254740993"));
    });

    it("rejects invalid UUID", () => {
      expect(() => parseUserRow({ ...MOCK_USER_ROW, id: "not-uuid" })).toThrow(
        "Invalid UUID v4"
      );
    });

    it("rejects invalid user_role", () => {
      expect(() =>
        parseUserRow({ ...MOCK_USER_ROW, user_role: "superadmin" })
      ).toThrow("Invalid user_role");
    });

    it("rejects invalid timestamp", () => {
      expect(() =>
        parseUserRow({ ...MOCK_USER_ROW, created_at: "not-a-date" })
      ).toThrow("Invalid ISO 8601");
    });
  });

  // ── queryUsers (PostgREST GET) ──────────────────────

  describe("queryUsers", () => {
    it("returns parsed users on 200", async () => {
      mockFetch(200, [MOCK_USER_ROW], { "Content-Range": "0-0/1" });
      const result = await client.queryUsers();

      expect(result.data).toHaveLength(1);
      expect(result.data[0].id).toBe(VALID_UUID);
      expect(result.data[0].email).toBe("alice@example.com");
      expect(result.data[0].user_role).toBe("editor");
      expect(typeof result.data[0].balance).toBe("bigint");
      expect(result.totalCount).toBe(1);
    });

    it("handles 206 Partial Content for paginated responses", async () => {
      mockFetchSequence(
        {
          status: 206,
          body: [MOCK_USER_ROW],
          headers: { "Content-Range": "0-0/2" },
        },
        {
          status: 200,
          body: [MOCK_USER_NULL_PROFILE],
          headers: { "Content-Range": "1-1/2" },
        }
      );

      const result = await client.queryUsers(undefined, 1);
      expect(result.data).toHaveLength(2);
      expect(result.totalCount).toBe(2);
    });

    it("throws on 401 unauthorized", async () => {
      mockFetch(401, MOCK_POSTGREST_ERROR_401);
      await expect(client.queryUsers()).rejects.toThrow("Unauthorized");
    });

    it("throws on 403 RLS violation", async () => {
      mockFetch(403, MOCK_POSTGREST_ERROR_403);
      await expect(client.queryUsers()).rejects.toThrow("Forbidden (RLS)");
    });

    it("throws on 416 range not satisfiable", async () => {
      mockFetch(416, MOCK_POSTGREST_ERROR_416);
      await expect(client.queryUsers()).rejects.toThrow("Range not satisfiable");
    });

    it("throws on network error", async () => {
      mockFetchNetworkError("ECONNREFUSED");
      await expect(client.queryUsers()).rejects.toThrow(
        "Network error querying users: ECONNREFUSED"
      );
    });

    it("validates page size before calling API", async () => {
      await expect(client.queryUsers(undefined, 0)).rejects.toThrow(
        "Page size must be between"
      );
      await expect(client.queryUsers(undefined, 1001)).rejects.toThrow(
        "Page size must be between"
      );
    });

    it("validates role filter", async () => {
      await expect(
        client.queryUsers({ role: "superadmin" as never })
      ).rejects.toThrow("Invalid user_role");
    });

    it("handles users with null profile without crashing", async () => {
      mockFetch(200, [MOCK_USER_NULL_PROFILE], { "Content-Range": "0-0/1" });
      const result = await client.queryUsers();

      expect(result.data[0].profile).toBeNull();
      // Accessing nested fields does NOT crash
    });
  });

  // ── insertUser (PostgREST POST) ─────────────────────

  describe("insertUser", () => {
    it("inserts user and returns parsed row on 201", async () => {
      mockFetch(201, [MOCK_USER_ROW]);
      const result = await client.insertUser({
        email: "alice@example.com",
        user_role: "editor",
      });

      expect(result.ok).toBe(true);
      if (result.ok) {
        expect(result.data.id).toBe(VALID_UUID);
        expect(result.data.email).toBe("alice@example.com");
        expect(result.data.user_role).toBe("editor");
      }
    });

    it("returns error on 409 unique constraint", async () => {
      mockFetch(409, MOCK_POSTGREST_ERROR_409);
      const result = await client.insertUser({
        email: "alice@example.com",
        user_role: "editor",
      });

      expect(result.ok).toBe(false);
      if (!result.ok) {
        expect(result.status).toBe(409);
        expect(result.error.code).toBe("23505");
        expect(result.error.message).toContain("Unique constraint");
      }
    });

    it("returns error on 401", async () => {
      mockFetch(401, MOCK_POSTGREST_ERROR_401);
      const result = await client.insertUser({
        email: "test@example.com",
        user_role: "viewer",
      });

      expect(result.ok).toBe(false);
      if (!result.ok) {
        expect(result.status).toBe(401);
      }
    });

    it("returns error on 403 RLS violation", async () => {
      mockFetch(403, MOCK_POSTGREST_ERROR_403);
      const result = await client.insertUser({
        email: "test@example.com",
        user_role: "viewer",
      });

      expect(result.ok).toBe(false);
      if (!result.ok) {
        expect(result.status).toBe(403);
      }
    });

    it("validates email before calling API", async () => {
      await expect(
        client.insertUser({ email: "not-email", user_role: "viewer" })
      ).rejects.toThrow("Invalid email");
    });

    it("validates role before calling API", async () => {
      await expect(
        client.insertUser({
          email: "test@example.com",
          user_role: "superadmin" as never,
        })
      ).rejects.toThrow("Invalid user_role");
    });

    it("throws on network failure", async () => {
      mockFetchNetworkError("ETIMEDOUT");
      await expect(
        client.insertUser({ email: "test@example.com", user_role: "viewer" })
      ).rejects.toThrow("Network error inserting user: ETIMEDOUT");
    });
  });

  // ── signUp (Auth) ───────────────────────────────────

  describe("signUp", () => {
    it("returns session on successful sign-up", async () => {
      mockFetch(200, MOCK_AUTH_RESPONSE);
      const result = await client.signUp({
        email: "new@example.com",
        password: "secure123",
      });

      expect(result.ok).toBe(true);
      if (result.ok) {
        expect(result.session.user.id).toBe(VALID_UUID);
        expect(result.session.user.email).toBe("new@example.com");
        expect(result.session.token_type).toBe("bearer");
        expect(typeof result.session.access_token).toBe("string");
        expect(typeof result.session.expires_in).toBe("number");
        expect(typeof result.session.expires_at).toBe("number");
        expect(typeof result.session.refresh_token).toBe("string");
        expect(result.session.user.user_metadata).toEqual({ display_name: "New User" });
        expect(result.session.user.app_metadata).toEqual({ provider: "email" });
        expect(Array.isArray(result.session.user.identities)).toBe(true);
      }
    });

    it("handles null user_metadata", async () => {
      mockFetch(200, { ...MOCK_AUTH_RESPONSE, user_metadata: null });
      const result = await client.signUp({
        email: "new@example.com",
        password: "secure123",
      });

      expect(result.ok).toBe(true);
      if (result.ok) {
        expect(result.session.user.user_metadata).toBeNull();
      }
    });

    it("returns error on 422 (invalid email)", async () => {
      mockFetch(422, MOCK_AUTH_ERROR_422);
      const result = await client.signUp({
        email: "bad-email@",
        password: "secure123",
      });

      // Note: local validation may also catch this, but server can reject too
      expect(result.ok).toBe(false);
      if (!result.ok) {
        expect(result.status).toBe(422);
      }
    });

    it("returns error on 429 (rate limit)", async () => {
      mockFetch(429, MOCK_AUTH_ERROR_429);
      const result = await client.signUp({
        email: "new@example.com",
        password: "secure123",
      });

      expect(result.ok).toBe(false);
      if (!result.ok) {
        expect(result.status).toBe(429);
        expect(result.error.msg).toContain("Rate limit");
      }
    });

    it("validates email before calling API", async () => {
      await expect(
        client.signUp({ email: "invalid", password: "secure123" })
      ).rejects.toThrow("Invalid email");
    });

    it("validates password length before calling API", async () => {
      await expect(
        client.signUp({ email: "test@example.com", password: "12345" })
      ).rejects.toThrow("at least 6 characters");
    });

    it("throws on network failure", async () => {
      mockFetchNetworkError("DNS_RESOLUTION_FAILED");
      await expect(
        client.signUp({ email: "test@example.com", password: "secure123" })
      ).rejects.toThrow("Network error during sign-up");
    });
  });

  // ── Token Refresh ───────────────────────────────────

  describe("token refresh", () => {
    it("refreshes token before API call when expired", async () => {
      // Create client with expired token
      const expiredClient = new SupabaseClient(
        "https://test-project.supabase.co",
        "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSJ9.fake",
        "expired-access-token",
        "valid-refresh-token",
        Math.floor(Date.now() / 1000) - 100 // expired 100 seconds ago
      );

      mockFetchSequence(
        {
          status: 200,
          body: {
            access_token: "new-access-token",
            refresh_token: "new-refresh-token",
            expires_at: Math.floor(Date.now() / 1000) + 3600,
          },
        },
        {
          status: 200,
          body: [MOCK_USER_ROW],
          headers: { "Content-Range": "0-0/1" },
        }
      );

      const result = await expiredClient.queryUsers();
      expect(result.data).toHaveLength(1);

      // Verify refresh was called first
      const calls = (global.fetch as ReturnType<typeof vi.fn>).mock.calls;
      expect(calls[0][0]).toContain("/auth/v1/token?grant_type=refresh_token");
      expect(calls[1][0]).toContain("/rest/v1/users");
    });
  });

  // ── isTokenExpired ──────────────────────────────────

  describe("isTokenExpired", () => {
    it("returns false for future expiry", () => {
      const future = Math.floor(Date.now() / 1000) + 3600;
      expect(isTokenExpired(future)).toBe(false);
    });

    it("returns true for past expiry", () => {
      const past = Math.floor(Date.now() / 1000) - 100;
      expect(isTokenExpired(past)).toBe(true);
    });

    it("returns true within 60-second grace period", () => {
      const almostExpired = Math.floor(Date.now() / 1000) + 30;
      expect(isTokenExpired(almostExpired)).toBe(true);
    });
  });
});
