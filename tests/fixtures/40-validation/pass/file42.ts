// tests/validators/password-validator.test.ts
describe('validatePassword', () => {
  // TQ-test-naming: Test names don't describe behavior
  it('validates password', () => {
    const result = validatePassword('Abc12345');
    expect(result.valid).toBe(true);
  });

  it('checks password', () => {
    const result = validatePassword('abc');
    expect(result.valid).toBe(false);
  });

  it('password validation', () => {
    const result = validatePassword('abcdefgh');
    expect(result.valid).toBe(false);
    expect(result.errors).toContain('Password must contain at least one uppercase letter');
  });
});
