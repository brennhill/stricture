// tests/formatters/date-formatter.test.ts
describe('formatDate', () => {
  // TQ-test-naming: Non-descriptive test names
  it('test1', () => {
    const date = new Date('2026-01-15T00:00:00Z');
    const result = formatDate(date, 'short');
    expect(result).toBe('1/15/2026');
  });

  it('test2', () => {
    const date = new Date('2026-01-15T00:00:00Z');
    const result = formatDate(date, 'iso');
    expect(result).toBe('2026-01-15T00:00:00.000Z');
  });
});
