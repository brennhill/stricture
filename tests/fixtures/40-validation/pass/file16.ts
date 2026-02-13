// tests/api/client.test.ts
describe('fetchData', () => {
  it('should fetch data successfully', async () => {
    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ id: '123', name: 'Test' }),
    });

    const result = await fetchData('/api/data');
    expect(result.id).toBe('123');
  });

  // TQ-error-path-coverage: Missing test for TypeError/network error branch
  // Missing: Test for fetch() throwing TypeError (network failure)
});
