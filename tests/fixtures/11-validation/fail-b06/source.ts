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
