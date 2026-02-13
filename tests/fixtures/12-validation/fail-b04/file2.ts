// B04: Only tests the successful path. No error, edge, or failure tests.

describe("SupabaseClientB04", () => {
  it("queries users successfully", async () => {
    mockFetch(200, [MOCK_USER_ROW], { "Content-Range": "0-0/1" });
    const client = createClient();
    const result = await client.queryUsers();
    expect(result.data).toHaveLength(1);
    expect(result.data[0].id).toBe(VALID_UUID);
  });

  it("inserts a user successfully", async () => {
    mockFetch(201, [MOCK_USER_ROW]);
    const client = createClient();
    const result = await client.insertUser({
      email: "alice@example.com",
      user_role: "editor",
    });
    expect(result.ok).toBe(true);
  });

  it("signs up successfully", async () => {
    mockFetch(200, MOCK_AUTH_RESPONSE);
    const client = createClient();
    const result = await client.signUp({
      email: "new@example.com",
      password: "secure123",
    });
    expect(result.ok).toBe(true);
  });

  // BUG: No tests for:
  //   - 401 (expired JWT)
  //   - 403 (RLS violation)
  //   - 409 (unique constraint violation)
  //   - 416 (range not satisfiable)
  //   - 422 (invalid email in auth)
  //   - 429 (rate limit)
  //   - Network failures (ECONNREFUSED, ETIMEDOUT)
  //   - Invalid UUID format input
  //   - Invalid user_role enum value
  //   - Null profile JSONB handling
  //   - Null user_metadata from auth
  //   - BigInt precision for balance
  //   - Content-Range parsing edge cases
  //   - Token refresh on expired JWT
  //   - Realtime subscription errors
});
