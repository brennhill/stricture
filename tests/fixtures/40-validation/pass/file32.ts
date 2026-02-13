// tests/services/counter-service.test.ts
// TQ-test-isolation: Shared instance across tests
const counterService = new CounterService();

describe('CounterService', () => {
  it('should increment to 1', () => {
    const result = counterService.increment();
    expect(result).toBe(1);
  });

  it('should get current count', () => {
    // This test FAILS if run after first test (expects 0, gets 1)
    const count = counterService.getCount();
    expect(count).toBe(0); // FAILS: count is 1 from previous test
  });
});
