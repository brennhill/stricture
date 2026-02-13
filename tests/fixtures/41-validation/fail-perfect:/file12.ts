// test/auth-middleware.anti7.test.ts (BAD)
import { describe, it, expect } from 'vitest';
import { AuthMiddleware } from '../src/auth-middleware';

describe('AuthMiddleware with shared state', () => {
  // SHARED MUTABLE STATE - disaster waiting to happen
  const auth = new AuthMiddleware();
  let testToken: string;
  let testUser = { userId: 'user_123', roles: ['admin'] };

  it('creates initial token', () => {
    testToken = 'tkn_initial';
    auth['tokens'].set(testToken, {
      ...testUser,
      expiresAt: Date.now() + 3600000
    });
    expect(auth.verify(testToken).userId).toBe('user_123');
  });

  it('verifies existing token', () => {
    // Depends on previous test setting testToken
    const result = auth.verify(testToken);
    expect(result.userId).toBe('user_123');
  });

  it('refreshes token', () => {
    // Depends on previous tests
    const newToken = auth.refresh(testToken);
    testToken = newToken; // Mutates shared state
    expect(newToken).toMatch(/^tkn_/);
  });

  it('checks role on refreshed token', () => {
    // Depends on previous test mutating testToken
    const hasRole = auth.hasRole(testToken, 'admin');
    expect(hasRole).toBe(true);
  });

  it('modifies user roles', () => {
    // Mutates shared testUser object
    testUser.roles.push('moderator');
    auth['tokens'].set(testToken, {
      ...testUser,
      expiresAt: Date.now() + 3600000
    });
    expect(auth.hasRole(testToken, 'moderator')).toBe(true);
  });

  it('still has admin role', () => {
    // Depends on previous test's mutation
    expect(auth.hasRole(testToken, 'admin')).toBe(true);
  });
});
