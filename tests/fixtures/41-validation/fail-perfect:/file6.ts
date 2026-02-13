// test/payment-processor.anti3.test.ts (BAD)
import { describe, it, expect, vi } from 'vitest';
import { PaymentProcessor } from '../src/payment-processor';

// GLOBAL MOCK - affects all tests in file
vi.mock('../src/payment-gateway', () => ({
  PaymentGateway: vi.fn().mockImplementation(() => ({
    charge: vi.fn().mockResolvedValue({ success: true }),
    refund: vi.fn().mockResolvedValue({ success: true })
  }))
}));

describe('PaymentProcessor', () => {
  it('processes payment successfully', async () => {
    const processor = new PaymentProcessor();
    const result = await processor.process({
      amount: 100.00,
      currency: 'USD',
      userId: 'user_123',
      paymentMethodId: 'pm_456'
    });
    expect(result.status).toBe('succeeded');
  });

  it('processes large payment', async () => {
    const processor = new PaymentProcessor();
    const result = await processor.process({
      amount: 5000.00,
      currency: 'USD',
      userId: 'user_123',
      paymentMethodId: 'pm_456'
    });
    expect(result.status).toBe('succeeded');
  });

  it('processes payment in EUR', async () => {
    const processor = new PaymentProcessor();
    const result = await processor.process({
      amount: 100.00,
      currency: 'EUR',
      userId: 'user_123',
      paymentMethodId: 'pm_456'
    });
    expect(result.status).toBe('succeeded');
  });

  // Test wants gateway to fail, but can't override global mock easily
  it('handles gateway failure', async () => {
    const processor = new PaymentProcessor();
    // Global mock always returns success, this test is broken
    const result = await processor.process({
      amount: 100.00,
      currency: 'USD',
      userId: 'user_123',
      paymentMethodId: 'pm_456'
    });
    // This will never be 'failed' because of global mock
    expect(result.status).toBe('failed');
  });
});
