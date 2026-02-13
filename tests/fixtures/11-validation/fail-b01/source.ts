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
