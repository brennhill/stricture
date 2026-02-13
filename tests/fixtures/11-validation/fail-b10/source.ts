// tests/auth0-client.test.ts — B10: Format not validated

import { describe, it, expect, vi } from "vitest";

describe("Auth0 Client (B10 — no format validation)", () => {
  async function getUser(userId: string): Promise<Record<string, unknown>> {
    // BUG: No format validation on userId
    // Manifest says format: "auth0|[a-f0-9]+|{provider}|{id}"
    // Accepts any string, including path traversal attacks
    const response = await fetch(
      `https://test.auth0.com/api/v2/users/${userId}`, // Not even URL-encoded!
      { headers: { Authorization: "Bearer token" } },
    );
    return response.json() as Promise<Record<string, unknown>>;
  }

  it("accepts valid auth0 user_id", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({ user_id: "auth0|abc123" }), { status: 200 }),
    );

    const user = await getUser("auth0|abc123");
    expect(user.user_id).toBe("auth0|abc123");
  });

  it("accepts plain string user_id without validation", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({ user_id: "plainstring" }), { status: 200 }),
    );

    // BUG: "plainstring" does not match the required format
    const user = await getUser("plainstring");
    expect(user.user_id).toBe("plainstring"); // Should have been rejected before the fetch
  });

  it("accepts empty string user_id", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({}), { status: 404 }),
    );

    // BUG: Empty string is not a valid user_id
    const user = await getUser("");
    expect(user).toBeDefined();
  });

  it("accepts path traversal in user_id", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({}), { status: 400 }),
    );

    // BUG: Path traversal is not rejected client-side
    // URL becomes: /api/v2/users/../../admin
    const user = await getUser("../../admin");
    expect(user).toBeDefined();
  });

  it("does not validate user_id in response either", () => {
    // BUG: Response user_id is also not validated
    const responseUser = { user_id: "not-a-valid-format", email: "x@y.com" };
    expect(responseUser.user_id).toBe("not-a-valid-format");
    // No format check — accepts anything from Auth0
  });
});
