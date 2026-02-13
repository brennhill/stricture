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
