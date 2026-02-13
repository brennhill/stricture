// tests/api/client.test.ts
it('should handle API response', async () => {
  // TQ-schema-conformance: Mock missing required 'meta' field
  const mockResponse: ApiResponse = {
    status: 200,
    data: {
      userId: 'user-123',
      userName: 'John Doe',
      userEmail: 'john@example.com',
      createdAt: '2026-01-01T00:00:00Z',
    },
    // Missing: meta field (required by ApiResponse interface)
  } as ApiResponse; // Type assertion bypasses TypeScript checking

  const result = processResponse(mockResponse);
  expect(result.userId).toBe('user-123');
});
