// test/payment-processor.test.ts (PERFECT)
import { describe, it, expect } from 'vitest';
import { PaymentProcessor } from '../src/payment-processor';

describe('PaymentProcessor', () => {
  describe('process', () => {
    it('processes valid USD payment with correct fee calculation', async () => {
      const processor = new PaymentProcessor();
      const result = await processor.process({
        amount: 100.00,
        currency: 'USD',
        userId: 'user_123',
        paymentMethodId: 'pm_456'
      });

      expect(result.amount).toBe(100.00);
      expect(result.currency).toBe('USD');
      expect(result.status).toBe('succeeded');
      expect(result.fee).toBeCloseTo(3.20, 2); // 100 * 0.029 + 0.30
      expect(result.netAmount).toBeCloseTo(96.80, 2);
      expect(result.transactionId).toMatch(/^txn_\d+$/);
      expect(result.processedAt).toMatch(/^\d{4}-\d{2}-\d{2}T/);
      expect(result.metadata).toEqual({});
    });

    it('processes payment with custom metadata', async () => {
      const processor = new PaymentProcessor();
      const metadata = { orderId: 'ord_789', source: 'web' };
      const result = await processor.process({
        amount: 50.00,
        currency: 'EUR',
        userId: 'user_123',
        paymentMethodId: 'pm_456',
        metadata
      });

      expect(result.metadata).toEqual(metadata);
      expect(Object.keys(result.metadata)).toHaveLength(2);
    });

    it('throws error for negative amount', async () => {
      const processor = new PaymentProcessor();
      await expect(processor.process({
        amount: -10.00,
        currency: 'USD',
        userId: 'user_123',
        paymentMethodId: 'pm_456'
      })).rejects.toThrow('Amount must be positive');
    });

    it('throws error for zero amount', async () => {
      const processor = new PaymentProcessor();
      await expect(processor.process({
        amount: 0,
        currency: 'USD',
        userId: 'user_123',
        paymentMethodId: 'pm_456'
      })).rejects.toThrow('Amount must be positive');
    });

    it('throws error for unsupported currency', async () => {
      const processor = new PaymentProcessor();
      await expect(processor.process({
        amount: 100.00,
        currency: 'JPY',
        userId: 'user_123',
        paymentMethodId: 'pm_456'
      })).rejects.toThrow('Unsupported currency');
    });

    it('handles boundary at minimum valid amount', async () => {
      const processor = new PaymentProcessor();
      const result = await processor.process({
        amount: 0.01,
        currency: 'USD',
        userId: 'user_123',
        paymentMethodId: 'pm_456'
      });

      expect(result.status).toBe('succeeded');
      expect(result.amount).toBe(0.01);
    });
  });

  describe('refund', () => {
    it('refunds full amount with correct metadata', async () => {
      const processor = new PaymentProcessor();
      const result = await processor.refund('txn_1234567890', 100.00);

      expect(result.transactionId).toMatch(/^rfnd_\d+$/);
      expect(result.amount).toBe(100.00);
      expect(result.status).toBe('succeeded');
      expect(result.fee).toBe(0);
      expect(result.netAmount).toBe(100.00);
      expect(result.metadata).toEqual({ originalTransaction: 'txn_1234567890' });
    });

    it('throws error for invalid transaction ID format', async () => {
      const processor = new PaymentProcessor();
      await expect(processor.refund('invalid_123', 50.00))
        .rejects.toThrow('Invalid transaction ID');
    });
  });
});
