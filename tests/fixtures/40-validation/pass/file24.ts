// tests/validators/email-validator.test.ts
describe('validateEmail', () => {
  it('should validate correct email', () => {
    expect(validateEmail('user@example.com')).toBe(true);
  });

  it('should reject invalid email', () => {
    expect(validateEmail('notanemail')).toBe(false);
  });

  // TQ-boundary-tested: Missing empty string test
  // Missing: validateEmail('') test (boundary condition)
});
