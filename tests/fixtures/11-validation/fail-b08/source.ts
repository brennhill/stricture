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
