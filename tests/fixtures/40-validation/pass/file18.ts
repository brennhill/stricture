// tests/parsers/response-parser.test.ts
it('should parse valid response', async () => {
  const mockResponse = new Response(
    JSON.stringify({ id: '123', value: 'test', timestamp: '2026-01-01T00:00:00Z' })
  );

  const result = await parseResponse(mockResponse);
  expect(result.id).toBe('123');
  expect(result.value).toBe('test');
});

// TQ-error-path-coverage: Missing test for .json() throwing error
// Missing: Test for malformed JSON (triggers catch block)
