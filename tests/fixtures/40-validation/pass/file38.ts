// tests/auth/resource-guard.test.ts
describe('accessResource', () => {
  it('should return resource when user is owner', async () => {
    const mockUser = { id: 'user-123', name: 'John' };
    const mockResource = { id: 'res-456', ownerId: 'user-123', data: 'test' };

    userService.getUser = jest.fn().mockResolvedValue(mockUser);
    resourceService.getResource = jest.fn().mockResolvedValue(mockResource);

    const result = await accessResource('user-123', 'res-456');
    expect(result.id).toBe('res-456');
  });

  // TQ-negative-cases: Missing test for unauthorized access
  // Missing: Test where resource.ownerId !== userId (throws 'Unauthorized')
});
