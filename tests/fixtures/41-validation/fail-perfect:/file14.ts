// test/payment-processor.anti9.test.ts (BAD)
import { describe, it, expect } from 'vitest';
import { PaymentProcessor } from '../src/payment-processor';

describe('PaymentProcessor partial assertions', () => {
  it('processes payment', async () => {
    const processor = new PaymentProcessor();
    const result = await processor.process({
      amount: 100.00,
      currency: 'USD',
      userId: 'user_123',
      paymentMethodId: 'pm_456'
    });

    // Only checks 2 of 8 fields
    expect(result.status).toBe('succeeded');
    expect(result.amount).toBe(100.00);
    // Missing: transactionId, currency, fee, netAmount, processedAt, metadata
  });

  it('refunds payment', async () => {
    const processor = new PaymentProcessor();
    const result = await processor.refund('txn_123', 50.00);

    // Only checks 1 of 8 fields
    expect(result.amount).toBe(50.00);
    // Missing: transactionId, currency, status, fee, netAmount, processedAt, metadata
  });

  it('validates payment request', async () => {
    const processor = new PaymentProcessor();
    const result = await processor.validate({
      amount: 100.00,
      currency: 'USD',
      userId: 'user_123',
      paymentMethodId: 'pm_456'
    });

    // Checks boolean but not why
    expect(typeof result).toBe('boolean');
    // Should verify: what makes it true vs false
  });

  it('transforms data', () => {
    const transformer = new DataTransformer();
    const result = transformer.transform([{ id: 1 }, null, { id: 2 }]);

    // Only checks records array
    expect(result.records).toHaveLength(2);
    // Missing: errors array, stats.total, stats.successful, stats.failed, stats.duration
  });

  it('transforms empty array', () => {
    const transformer = new DataTransformer();
    const result = transformer.transform([]);

    // Only checks length
    expect(result.records).toHaveLength(0);
    // Missing: errors should be empty, stats should be zero
  });
});
