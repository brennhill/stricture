// test/payment-processor.anti6.test.ts (BAD)
import { describe, it, expect } from 'vitest';
import { PaymentProcessor } from '../src/payment-processor';

describe('PaymentProcessor with lazy assertions', () => {
  it('processes payment with any values', async () => {
    const processor = new PaymentProcessor();
    const result = await processor.process({
      amount: 100.00,
      currency: 'USD',
      userId: 'user_123',
      paymentMethodId: 'pm_456'
    });

    expect(result).toEqual({
      transactionId: expect.any(String),
      amount: expect.any(Number),
      currency: expect.any(String),
      status: expect.any(String),
      fee: expect.any(Number),
      netAmount: expect.any(Number),
      processedAt: expect.any(String),
      metadata: expect.any(Object)
    });
  });

  it('refunds with any transaction ID', async () => {
    const processor = new PaymentProcessor();
    const result = await processor.refund(expect.any(String), 50.00);

    expect(result.transactionId).toEqual(expect.any(String));
    expect(result.amount).toEqual(expect.any(Number));
  });

  it('validates with any inputs', async () => {
    const processor = new PaymentProcessor();
    const result = await processor.validate({
      amount: expect.any(Number),
      currency: expect.any(String),
      userId: expect.any(String),
      paymentMethodId: expect.any(String)
    });

    expect(result).toEqual(expect.any(Boolean));
  });

  it('returns object with string properties', async () => {
    const processor = new PaymentProcessor();
    const result = await processor.process({
      amount: 75.50,
      currency: 'EUR',
      userId: 'user_456',
      paymentMethodId: 'pm_789'
    });

    expect(typeof result.transactionId).toBe('string');
    expect(typeof result.currency).toBe('string');
    expect(typeof result.status).toBe('string');
    // Could be any string - "succeeded", "failed", "banana"
  });
});
