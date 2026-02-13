// tests/storage/cache.test.ts
it('should store value in cache', () => {
  const cache = new Cache();
  const result = cache.set('key1', 'value1');

  // TQ-return-type-verified: Only checks side effect, not return value
  expect(cache.get('key1')).toBe('value1');
  // Missing: result.stored, result.key
});
