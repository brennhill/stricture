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
