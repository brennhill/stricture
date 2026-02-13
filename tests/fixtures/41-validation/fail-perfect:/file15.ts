// test/payment-processor.anti10.test.ts (BAD - WORST CASE)
import { describe, it, expect, vi } from 'vitest';
import { PaymentProcessor } from '../src/payment-processor';
import { DataTransformer } from '../src/data-transformer';
import { AuthMiddleware } from '../src/auth-middleware';

// ANTI-PATTERN 3: Global mocks
vi.mock('../src/payment-gateway');
vi.mock('../src/logger');

describe('Full test suite', () => {
  // ANTI-PATTERN 7: Shared mutable state
  const processor = new PaymentProcessor();
  const transformer = new DataTransformer();
  const auth = new AuthMiddleware();
  let sharedToken: string;
  let sharedResult: any;

  // ANTI-PATTERN 4: Generic name
  it('test 1', async () => {
    // ANTI-PATTERN 5: No assertions
    await processor.process({
      amount: 100.00,
      currency: 'USD',
      userId: 'user_123',
      paymentMethodId: 'pm_456'
    });
  });

  // ANTI-PATTERN 4: Generic name
  it('should work', async () => {
    // ANTI-PATTERN 7: Mutates shared state
    sharedResult = await processor.process({
      amount: 200.00,
      currency: 'EUR',
      userId: 'user_456',
      paymentMethodId: 'pm_789'
    });

    // ANTI-PATTERN 1: toBeDefined carpet bomb
    expect(sharedResult).toBeDefined();
    expect(sharedResult.transactionId).toBeDefined();
    expect(sharedResult.status).toBeDefined();
  });

  // ANTI-PATTERN 4: Generic name
  it('handles data', () => {
    // ANTI-PATTERN 8: No boundary testing
    const result = transformer.transform([{ id: 5 }, { id: 10 }]);

    // ANTI-PATTERN 1: toBeDefined
    // ANTI-PATTERN 9: Partial field check
    expect(result).toBeDefined();
    expect(result.records).toBeDefined();
  });

  // ANTI-PATTERN 2: Only happy paths (3 success tests, 0 error tests)
  it('processes USD payment', async () => {
    const result = await processor.process({
      amount: 100.00,
      currency: 'USD',
      userId: 'user_123',
      paymentMethodId: 'pm_456'
    });
    // ANTI-PATTERN 6: expect.any()
    expect(result.status).toEqual(expect.any(String));
  });

  it('processes EUR payment', async () => {
    const result = await processor.process({
      amount: 50.00,
      currency: 'EUR',
      userId: 'user_456',
      paymentMethodId: 'pm_789'
    });
    expect(result.status).toEqual(expect.any(String));
  });

  it('processes GBP payment', async () => {
    const result = await processor.process({
      amount: 75.00,
      currency: 'GBP',
      userId: 'user_789',
      paymentMethodId: 'pm_012'
    });
    expect(result.status).toEqual(expect.any(String));
  });

  // ANTI-PATTERN 7: Depends on shared state from previous test
  it('checks previous result', () => {
    expect(sharedResult).toBeDefined();
    expect(sharedResult.amount).toBeGreaterThan(0);
  });

  // ANTI-PATTERN 4 + 5: Generic name, no assertions
  it('test auth', () => {
    sharedToken = 'tkn_test';
    auth['tokens'].set(sharedToken, {
      userId: 'user_123',
      roles: ['admin'],
      expiresAt: Date.now() + 3600000
    });
  });

  // ANTI-PATTERN 7: Depends on previous test
  it('verifies token', () => {
    const decoded = auth.verify(sharedToken);
    // ANTI-PATTERN 1: toBeDefined
    expect(decoded).toBeDefined();
  });
});
