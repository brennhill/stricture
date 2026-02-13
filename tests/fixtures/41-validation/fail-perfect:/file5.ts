// test/payment-processor.anti2.test.ts (BAD)
import { describe, it, expect } from 'vitest';
import { PaymentProcessor } from '../src/payment-processor';

describe('PaymentProcessor - success scenarios', () => {
  it('processes USD payment', async () => {
    const processor = new PaymentProcessor();
    const result = await processor.process({
      amount: 100.00,
      currency: 'USD',
      userId: 'user_123',
      paymentMethodId: 'pm_456'
    });
    expect(result.status).toBe('succeeded');
  });

  it('processes EUR payment', async () => {
    const processor = new PaymentProcessor();
    const result = await processor.process({
      amount: 50.00,
      currency: 'EUR',
      userId: 'user_456',
      paymentMethodId: 'pm_789'
    });
    expect(result.status).toBe('succeeded');
  });

  it('processes GBP payment', async () => {
    const processor = new PaymentProcessor();
    const result = await processor.process({
      amount: 75.00,
      currency: 'GBP',
      userId: 'user_789',
      paymentMethodId: 'pm_012'
    });
    expect(result.status).toBe('succeeded');
  });

  it('processes payment with small amount', async () => {
    const processor = new PaymentProcessor();
    const result = await processor.process({
      amount: 1.00,
      currency: 'USD',
      userId: 'user_123',
      paymentMethodId: 'pm_456'
    });
    expect(result.status).toBe('succeeded');
  });

  it('processes payment with large amount', async () => {
    const processor = new PaymentProcessor();
    const result = await processor.process({
      amount: 10000.00,
      currency: 'USD',
      userId: 'user_123',
      paymentMethodId: 'pm_456'
    });
    expect(result.status).toBe('succeeded');
  });

  it('processes payment with metadata', async () => {
    const processor = new PaymentProcessor();
    const result = await processor.process({
      amount: 100.00,
      currency: 'USD',
      userId: 'user_123',
      paymentMethodId: 'pm_456',
      metadata: { orderId: 'ord_123' }
    });
    expect(result.status).toBe('succeeded');
  });

  it('processes payment without metadata', async () => {
    const processor = new PaymentProcessor();
    const result = await processor.process({
      amount: 100.00,
      currency: 'USD',
      userId: 'user_123',
      paymentMethodId: 'pm_456'
    });
    expect(result.status).toBe('succeeded');
  });

  it('refunds full amount', async () => {
    const processor = new PaymentProcessor();
    const result = await processor.refund('txn_123', 100.00);
    expect(result.status).toBe('succeeded');
  });

  it('refunds partial amount', async () => {
    const processor = new PaymentProcessor();
    const result = await processor.refund('txn_123', 50.00);
    expect(result.status).toBe('succeeded');
  });

  it('validates payment', async () => {
    const processor = new PaymentProcessor();
    const result = await processor.validate({
      amount: 100.00,
      currency: 'USD',
      userId: 'user_123',
      paymentMethodId: 'pm_456'
    });
    expect(result).toBe(true);
  });
});
