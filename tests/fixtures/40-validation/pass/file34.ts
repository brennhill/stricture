// tests/storage/store.test.ts
describe('Store', () => {
  const store = new Store();

  it('should set value', () => {
    store.set('key1', 'value1');
    expect(store.get('key1')).toBe('value1');
  });

  it('should get previously set value', () => {
    // TQ-test-isolation: Depends on previous test running first
    expect(store.get('key1')).toBe('value1'); // FAILS if run in isolation
  });

  it('should handle multiple keys', () => {
    store.set('key2', 'value2');
    // Assumes key1 still exists from first test
    expect(store.get('key1')).toBe('value1');
    expect(store.get('key2')).toBe('value2');
  });
});
