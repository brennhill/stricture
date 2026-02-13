// tests/auth0-client.test.ts — B14: Token refresh race condition

import { describe, it, expect, vi } from "vitest";

describe("Auth0 Client (B14 — token refresh race)", () => {
  class TokenManagerRacy {
    private accessToken: string | null = null;
    private tokenExpiresAt: number = 0;

    async getToken(): Promise<string> {
      // BUG: No mutex — concurrent callers can all trigger refresh simultaneously
      if (!this.accessToken || Date.now() >= this.tokenExpiresAt) {
        await this.refreshToken();
      }
      return this.accessToken!;
    }

    private async refreshToken(): Promise<void> {
      // BUG: No lock — if 10 concurrent calls detect expired token,
      // all 10 will call this method, making 10 separate token requests
      const response = await fetch("https://test.auth0.com/oauth/token", {
        method: "POST",
        body: JSON.stringify({
          grant_type: "client_credentials",
          client_id: "my-client",
          client_secret: "my-secret",
        }),
      });
      const data = await response.json() as { access_token: string; expires_in: number };
      this.accessToken = data.access_token;
      this.tokenExpiresAt = Date.now() + data.expires_in * 1000;
    }
  }

  async function getUserRacy(
    tokenManager: TokenManagerRacy,
    userId: string,
  ): Promise<Record<string, unknown>> {
    const token = await tokenManager.getToken();

    // BUG: Token could expire between getToken() and fetch()
    // No retry logic if the API returns 401
    const response = await fetch(
      `https://test.auth0.com/api/v2/users/${encodeURIComponent(userId)}`,
      { headers: { Authorization: `Bearer ${token}` } },
    );

    // BUG: If response is 401 (expired token), no automatic retry
    if (!response.ok) {
      throw new Error(`Auth0 API error: ${response.status}`);
    }

    return response.json() as Promise<Record<string, unknown>>;
  }

  it("concurrent token refresh sends multiple requests", async () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch");

    // Each call returns a valid token
    fetchSpy.mockResolvedValue(
      new Response(JSON.stringify({
        access_token: "eyJ.payload.sig",
        token_type: "Bearer",
        expires_in: 86400,
      }), { status: 200 }),
    );

    const manager = new TokenManagerRacy();

    // 5 concurrent getToken() calls — all see expired token simultaneously
    const promises = Array.from({ length: 5 }, () => manager.getToken());
    await Promise.all(promises);

    // BUG: All 5 calls trigger separate refresh requests
    // A correct implementation would use a mutex so only 1 refresh happens
    expect(fetchSpy).toHaveBeenCalledTimes(5); // Should be 1!

    fetchSpy.mockRestore();
  });

  it("401 between token check and API call is not retried", async () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch");

    // First call: return a token
    fetchSpy.mockResolvedValueOnce(
      new Response(JSON.stringify({
        access_token: "eyJ.payload.sig",
        token_type: "Bearer",
        expires_in: 1, // Expires in 1 second
      }), { status: 200 }),
    );

    // Second call: 401 because token expired between getToken() and fetch()
    fetchSpy.mockResolvedValueOnce(
      new Response(JSON.stringify({
        statusCode: 401,
        error: "Unauthorized",
        message: "Expired token",
      }), { status: 401 }),
    );

    const manager = new TokenManagerRacy();

    // BUG: getUserRacy does not retry on 401 — just throws
    await expect(
      getUserRacy(manager, "auth0|abc123"),
    ).rejects.toThrow("Auth0 API error: 401");

    // A correct implementation would refresh the token and retry once

    fetchSpy.mockRestore();
  });
});
