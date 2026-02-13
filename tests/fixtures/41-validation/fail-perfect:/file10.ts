// test/payment-processor.anti5.test.ts (BAD)
import { describe, it, expect } from 'vitest';
import { PaymentProcessor } from '../src/payment-processor';

describe('PaymentProcessor execution tests', () => {
  it('calls process method', async () => {
    const processor = new PaymentProcessor();
    await processor.process({
      amount: 100.00,
      currency: 'USD',
      userId: 'user_123',
      paymentMethodId: 'pm_456'
    });
    // No assertions - just checking it doesn't throw
  });

  it('executes refund', async () => {
    const processor = new PaymentProcessor();
    const result = await processor.refund('txn_123', 50.00);
    console.log('Refund result:', result); // Logging instead of asserting
  });

  it('runs validation', async () => {
    const processor = new PaymentProcessor();
    processor.validate({
      amount: 100.00,
      currency: 'USD',
      userId: 'user_123',
      paymentMethodId: 'pm_456'
    });
    // Missing await, missing assertion
  });

  it('processes multiple payments', async () => {
    const processor = new PaymentProcessor();
    for (let i = 0; i < 5; i++) {
      await processor.process({
        amount: 100.00 * i,
        currency: 'USD',
        userId: `user_${i}`,
        paymentMethodId: 'pm_456'
      });
    }
    // No assertions on any iteration
  });

  it('handles various currencies', async () => {
    const processor = new PaymentProcessor();
    const currencies = ['USD', 'EUR', 'GBP'];

    for (const currency of currencies) {
      const result = await processor.process({
        amount: 100.00,
        currency,
        userId: 'user_123',
        paymentMethodId: 'pm_456'
      });
    }
    // Loop executes but verifies nothing
  });
});
