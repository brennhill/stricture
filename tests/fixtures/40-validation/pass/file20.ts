// tests/services/profile-service.test.ts
it('should get user profile', () => {
  const profile = getProfile('user-123');

  // TQ-assertion-depth: Only checks top-level fields
  expect(profile.userId).toBe('user-123');
  expect(profile.name).toBe('John Doe');
  expect(profile.contact).toBeDefined();
  // Missing: contact.email, contact.phone, contact.address.street, etc.
});
