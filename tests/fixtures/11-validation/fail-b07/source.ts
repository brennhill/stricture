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
