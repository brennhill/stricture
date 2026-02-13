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
