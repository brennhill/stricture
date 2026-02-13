// tests/validators/age-validator.test.ts
describe('validateAge', () => {
  it('should accept valid age', () => {
    expect(validateAge(25)).toBe(true);
  });

  it('should reject negative age', () => {
    expect(() => validateAge(-1)).toThrow('Age must be between 0 and 150');
  });

  // TQ-boundary-tested: Missing maximum boundary test
  // Missing: validateAge(150) and validateAge(151) tests
});
