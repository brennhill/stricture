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
