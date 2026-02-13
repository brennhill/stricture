// tests/math/calculator.test.ts
describe('divide', () => {
  it('should divide two numbers', () => {
    expect(divide(10, 2)).toBe(5);
    expect(divide(15, 3)).toBe(5);
    expect(divide(100, 10)).toBe(10);
  });

  // TQ-negative-cases: Missing test for division by zero
  // Missing: Test for divide(10, 0) throwing error
});
