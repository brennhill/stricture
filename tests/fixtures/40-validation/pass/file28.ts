// tests/utils/time.test.ts
describe('getCurrentTime', () => {
  it('should return current timestamp', () => {
    // TQ-mock-scope: Global Date.now() mocked but never cleaned up
    const mockNow = jest.fn(() => 1234567890);
    global.Date.now = mockNow;

    expect(getCurrentTime()).toBe(1234567890);
    // Missing: Restore Date.now() in afterEach
  });

  it('should return different timestamp', () => {
    // This test will FAIL because Date.now is still mocked from previous test
    expect(getCurrentTime()).toBeGreaterThan(0);
  });
});
