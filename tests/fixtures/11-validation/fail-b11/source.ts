// tests/auth0-client.test.ts — B11: Precision loss (seconds vs milliseconds)

import { describe, it, expect } from "vitest";

describe("Auth0 Client (B11 — seconds/milliseconds precision loss)", () => {
  class TokenManagerBroken {
    private accessToken: string | null = null;
    private tokenExpiresAt: number = 0;

    storeToken(accessToken: string, expiresIn: number): void {
      this.accessToken = accessToken;
      // BUG: expires_in is in SECONDS, but Date.now() is in MILLISECONDS
      // Missing: * 1000 conversion
      this.tokenExpiresAt = Date.now() + expiresIn;
      // If expiresIn = 86400 (24 hours in seconds),
      // this adds only 86400 ms (86.4 seconds) instead of 86400000 ms (24 hours)
    }

    isTokenValid(): boolean {
      return this.accessToken !== null && Date.now() < this.tokenExpiresAt;
    }

    getExpiresAt(): number {
      return this.tokenExpiresAt;
    }
  }

  it("token expires 1000x too early", () => {
    const manager = new TokenManagerBroken();
    const before = Date.now();

    manager.storeToken("eyJ.payload.sig", 86400); // 24 hours in seconds

    const expiresAt = manager.getExpiresAt();
    const expiresInMs = expiresAt - before;

    // BUG: Should be ~86400000 ms (24 hours), but is ~86400 ms (86 seconds)
    expect(expiresInMs).toBeLessThan(100_000); // Less than 100 seconds
    expect(expiresInMs).toBeGreaterThan(80_000); // Around 86 seconds

    // The correct value would be:
    // expect(expiresInMs).toBeGreaterThan(86_000_000); // ~24 hours
  });

  it("token appears expired after 2 minutes", async () => {
    const manager = new TokenManagerBroken();
    manager.storeToken("eyJ.payload.sig", 86400);

    // Simulate 2 minutes passing
    const twoMinutesLater = Date.now() + 120_000;
    // BUG: Token "expired" after 86.4 seconds — but should last 24 hours
    expect(manager.getExpiresAt()).toBeLessThan(twoMinutesLater);
  });
});
