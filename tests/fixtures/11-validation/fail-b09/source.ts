// tests/auth0-client.test.ts — B09: Missing range validation

import { describe, it, expect } from "vitest";

describe("Auth0 Client (B09 — no range validation)", () => {
  function storeTokenExpiry(expiresIn: number): number {
    // BUG: No validation that expires_in is positive or within reasonable bounds
    // Manifest says range: [1, 2592000] (1 second to 30 days)
    return Date.now() + expiresIn * 1000;
  }

  it("stores token with valid expiry", () => {
    const expiresAt = storeTokenExpiry(86400);
    expect(expiresAt).toBeGreaterThan(Date.now());
  });

  it("accepts negative expires_in without error", () => {
    // BUG: Negative expiry means the token is already expired at storage time
    const expiresAt = storeTokenExpiry(-3600);
    // This should throw or reject, but it silently stores a past timestamp
    expect(expiresAt).toBeLessThan(Date.now()); // Token is already "expired"
    // No assertion that this is invalid — test proves the bug exists
  });

  it("accepts zero expires_in without error", () => {
    // BUG: Zero expiry means the token expires immediately
    const expiresAt = storeTokenExpiry(0);
    // Difference should be ~0ms but no validation catches this
    expect(Math.abs(expiresAt - Date.now())).toBeLessThan(100);
  });

  it("accepts absurdly large expires_in without error", () => {
    // BUG: 10 years in seconds — far exceeds Auth0's max of 30 days
    const expiresAt = storeTokenExpiry(315_360_000);
    // No upper bound check — token "never" expires
    expect(expiresAt).toBeGreaterThan(Date.now());
  });
});
