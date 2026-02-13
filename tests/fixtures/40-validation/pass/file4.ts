// tests/api/fetch-user.test.ts
it('should fetch user', async () => {
  const result = await fetchUser('user-123');

  // TQ-no-shallow-assertions: Shallow assertion doesn't verify API contract
  expect(result).toBeDefined();
  // Missing: result.id, result.name, result.email, etc.
});
