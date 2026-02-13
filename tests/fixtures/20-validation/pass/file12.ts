// tests/client/api-client.test.ts — Missing negative tests
describe("UserApiClient", () => {
  it("should get user by ID", async () => {
    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ id: "123", name: "Alice", email: "alice@example.com", role: "editor", avatar: null, createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z" }),
    });

    const client = new UserApiClient("http://localhost:3000", "test-token");
    const user = await client.getUser("123");
    expect(user.id).toBe("123");
  });

  // BUG: No test for 404 case
  // Missing: it("should throw error for non-existent user", ...)
});
