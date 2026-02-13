// tests/auth/login.test.ts
it('should login successfully', async () => {
  const result = await login('user@example.com', 'password123');

  // TQ-no-shallow-assertions: Doesn't verify side effects or state changes
  expect(result).toBeTruthy();
  // Missing: session state, auth token, user context, etc.
});
