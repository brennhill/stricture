// tests/auth0-client.test.ts — B15: Read-modify-write race condition

import { describe, it, expect, vi } from "vitest";

describe("Auth0 Client (B15 — metadata read-modify-write race)", () => {
  interface SimpleUser {
    user_id: string;
    app_metadata: Record<string, unknown> | null;
  }

  // Simulates Auth0 server state
  let serverState: SimpleUser = {
    user_id: "auth0|abc123",
    app_metadata: { plan: "free", role: "user", loginCount: 0 },
  };

  function mockAuth0Server(): void {
    vi.spyOn(globalThis, "fetch").mockImplementation(async (url, options) => {
      const urlStr = typeof url === "string" ? url : url.toString();

      if (options?.method === "GET" || !options?.method) {
        // GET returns current state (snapshot in time)
        return new Response(JSON.stringify({ ...serverState }), { status: 200 });
      }

      if (options?.method === "PATCH") {
        // PATCH replaces app_metadata entirely (Auth0's actual behavior)
        const body = JSON.parse(options.body as string) as Partial<SimpleUser>;
        if (body.app_metadata !== undefined) {
          serverState = { ...serverState, app_metadata: body.app_metadata };
        }
        return new Response(JSON.stringify({ ...serverState }), { status: 200 });
      }

      return new Response("Not Found", { status: 404 });
    });
  }

  async function updateUserRole(userId: string, newRole: string): Promise<void> {
    // Step 1: Read current user
    const getResponse = await fetch(`https://test.auth0.com/api/v2/users/${userId}`);
    const currentUser = await getResponse.json() as SimpleUser;

    // Step 2: Modify app_metadata locally
    // BUG: No lock, no version check — another request can modify between read and write
    const updatedMetadata = {
      ...currentUser.app_metadata,
      role: newRole,
    };

    // Step 3: Write back entire app_metadata (overwrites concurrent changes)
    await fetch(`https://test.auth0.com/api/v2/users/${userId}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ app_metadata: updatedMetadata }),
    });
  }

  async function updateUserPlan(userId: string, newPlan: string): Promise<void> {
    // Same read-modify-write pattern — races with updateUserRole
    const getResponse = await fetch(`https://test.auth0.com/api/v2/users/${userId}`);
    const currentUser = await getResponse.json() as SimpleUser;

    const updatedMetadata = {
      ...currentUser.app_metadata,
      plan: newPlan,
    };

    await fetch(`https://test.auth0.com/api/v2/users/${userId}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ app_metadata: updatedMetadata }),
    });
  }

  it("concurrent metadata updates overwrite each other", async () => {
    // Reset server state
    serverState = {
      user_id: "auth0|abc123",
      app_metadata: { plan: "free", role: "user", loginCount: 0 },
    };
    mockAuth0Server();

    // Two concurrent updates: one sets role=admin, other sets plan=pro
    // Both read the SAME initial state: { plan: "free", role: "user", loginCount: 0 }
    await Promise.all([
      updateUserRole("auth0|abc123", "admin"),
      updateUserPlan("auth0|abc123", "pro"),
    ]);

    // BUG: Last write wins — one of the updates is lost
    // If updateUserPlan writes last:
    //   app_metadata = { plan: "pro", role: "user", loginCount: 0 }
    //   The role="admin" update is LOST
    //
    // If updateUserRole writes last:
    //   app_metadata = { plan: "free", role: "admin", loginCount: 0 }
    //   The plan="pro" update is LOST

    // We can verify that one update was lost:
    const finalResponse = await fetch("https://test.auth0.com/api/v2/users/auth0|abc123");
    const finalUser = await finalResponse.json() as SimpleUser;

    // At least one of these will fail — proving the race condition
    const hasCorrectRole = (finalUser.app_metadata as Record<string, unknown>)?.role === "admin";
    const hasCorrectPlan = (finalUser.app_metadata as Record<string, unknown>)?.plan === "pro";

    // BUG: Cannot guarantee both are true due to race condition
    // In a correct implementation, both updates would be merged atomically
    expect(hasCorrectRole || hasCorrectPlan).toBe(true);
    // This weaker assertion "passes" but masks the data loss
    // expect(hasCorrectRole && hasCorrectPlan).toBe(true); // This would fail!

    vi.restoreAllMocks();
  });

  it("loginCount is lost during concurrent role update", async () => {
    serverState = {
      user_id: "auth0|abc123",
      app_metadata: { plan: "free", role: "user", loginCount: 42 },
    };
    mockAuth0Server();

    // Read current state (loginCount = 42)
    // Meanwhile, another process increments loginCount to 43
    // Then we write role="admin" with the stale loginCount=42

    // Simulate: external process updates loginCount between our read and write
    const getResponse = await fetch("https://test.auth0.com/api/v2/users/auth0|abc123");
    const staleUser = await getResponse.json() as SimpleUser;

    // External update happens here (loginCount goes from 42 to 43)
    serverState = {
      ...serverState,
      app_metadata: { ...serverState.app_metadata, loginCount: 43 },
    };

    // Our write uses stale data — overwrites loginCount back to 42
    const updatedMetadata = { ...staleUser.app_metadata, role: "admin" };
    await fetch("https://test.auth0.com/api/v2/users/auth0|abc123", {
      method: "PATCH",
      body: JSON.stringify({ app_metadata: updatedMetadata }),
    });

    // BUG: loginCount regressed from 43 back to 42
    expect((serverState.app_metadata as Record<string, unknown>)?.loginCount).toBe(42); // Should be 43!

    vi.restoreAllMocks();
  });
});
