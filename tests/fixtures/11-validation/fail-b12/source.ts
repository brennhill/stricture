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
