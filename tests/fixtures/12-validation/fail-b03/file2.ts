// B03: Tests only check existence, never shape or value.

describe("SupabaseClientB03", () => {
  it("queries users", async () => {
    mockFetch(200, [MOCK_USER_ROW], { "Content-Range": "0-0/1" });
    const client = createClient();
    const result = await client.queryUsers();

    // BUG: These assertions prove nothing about correctness.
    // They pass even if result is { data: [{ id: 999, email: null }] }.
    expect(result).toBeDefined();
    expect(result.data).toBeDefined();
    expect(result.data.length).toBeTruthy();
    if (result.data.length > 0) {
      expect(result.data[0]).toBeDefined();
      expect(result.data[0].id).toBeDefined();
      expect(result.data[0].email).toBeDefined();
      expect(result.data[0].balance).toBeDefined();
      expect(result.data[0].user_role).toBeDefined();
    }
  });

  it("signs up user", async () => {
    mockFetch(200, MOCK_AUTH_RESPONSE);
    const client = createClient();
    const result = await client.signUp({
      email: "new@example.com",
      password: "secure123",
    });

    // BUG: Only checks that session exists, not its shape or values.
    expect(result).toBeDefined();
    expect(result.ok).toBeTruthy();
    if (result.ok) {
      expect(result.session).toBeDefined();
      expect(result.session.user).toBeDefined();
      expect(result.session.access_token).toBeDefined();
    }
  });

  it("handles errors", async () => {
    mockFetch(401, MOCK_POSTGREST_ERROR_401);
    const client = createClient();
    // BUG: Only checks that it throws, not what it throws.
    await expect(client.queryUsers()).rejects.toBeDefined();
  });
});
