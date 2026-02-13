// tests/client/api-client.test.ts — Shallow assertions
it("should create user", async () => {
  const client = new UserApiClient("http://localhost:3000", "test-token");
  const result = await client.createUser({
    name: "Alice",
    email: "alice@example.com",
    role: "editor",
  });

  // BUG: Only checks if result exists, not field values
  expect(result).toBeDefined();
  expect(result.id).toBeDefined();
  // Missing: name, email, role, avatar, createdAt, updatedAt checks
});
