# 41 — AI-Generated Test Patterns

## Overview

This validation set captures Stricture's core value proposition: catching AI-generated tests that achieve 100% coverage while verifying nothing. Modern AI code assistants excel at generating syntactically correct tests that execute without errors, but often produce patterns that provide false confidence.

## Why This Matters

AI-generated tests create a dangerous illusion:
- Coverage reports show 100% line/branch coverage
- CI passes green
- Developers assume code is validated
- **Reality:** Tests verify almost nothing about actual behavior

Stricture must detect these patterns to prevent coverage theater.

## Source Modules Under Test

### 1. PaymentProcessor (TypeScript)

```typescript
// src/payment-processor.ts
export interface PaymentRequest {
  amount: number;
  currency: string;
  userId: string;
  paymentMethodId: string;
  metadata?: Record<string, unknown>;
}

export interface PaymentResult {
  transactionId: string;
  amount: number;
  currency: string;
  status: 'succeeded' | 'failed' | 'pending';
  fee: number;
  netAmount: number;
  processedAt: string;
  metadata: Record<string, unknown>;
}

export class PaymentProcessor {
  async process(request: PaymentRequest): Promise<PaymentResult> {
    if (request.amount <= 0) {
      throw new Error('Amount must be positive');
    }
    if (!['USD', 'EUR', 'GBP'].includes(request.currency)) {
      throw new Error('Unsupported currency');
    }

    const fee = request.amount * 0.029 + 0.30;
    return {
      transactionId: `txn_${Date.now()}`,
      amount: request.amount,
      currency: request.currency,
      status: 'succeeded',
      fee,
      netAmount: request.amount - fee,
      processedAt: new Date().toISOString(),
      metadata: request.metadata || {}
    };
  }

  async refund(transactionId: string, amount?: number): Promise<PaymentResult> {
    if (!transactionId.startsWith('txn_')) {
      throw new Error('Invalid transaction ID');
    }

    return {
      transactionId: `rfnd_${Date.now()}`,
      amount: amount || 0,
      currency: 'USD',
      status: 'succeeded',
      fee: 0,
      netAmount: amount || 0,
      processedAt: new Date().toISOString(),
      metadata: { originalTransaction: transactionId }
    };
  }

  async validate(request: PaymentRequest): Promise<boolean> {
    return request.amount > 0 &&
           request.amount <= 999999.99 &&
           ['USD', 'EUR', 'GBP'].includes(request.currency);
  }
}
```

### 2. DataTransformer (TypeScript)

```typescript
// src/data-transformer.ts
export interface TransformResult {
  records: Array<Record<string, unknown>>;
  errors: Array<{ row: number; message: string }>;
  stats: {
    total: number;
    successful: number;
    failed: number;
    duration: number;
  };
}

export class DataTransformer {
  transform(data: unknown[]): TransformResult {
    const startTime = Date.now();
    const records: Array<Record<string, unknown>> = [];
    const errors: Array<{ row: number; message: string }> = [];

    data.forEach((item, index) => {
      try {
        if (typeof item !== 'object' || item === null) {
          throw new Error('Invalid data type');
        }
        records.push(item as Record<string, unknown>);
      } catch (error) {
        errors.push({
          row: index,
          message: error instanceof Error ? error.message : 'Unknown error'
        });
      }
    });

    return {
      records,
      errors,
      stats: {
        total: data.length,
        successful: records.length,
        failed: errors.length,
        duration: Date.now() - startTime
      }
    };
  }
}
```

### 3. AuthMiddleware (TypeScript)

```typescript
// src/auth-middleware.ts
export interface AuthToken {
  userId: string;
  roles: string[];
  expiresAt: number;
}

export class AuthMiddleware {
  private tokens = new Map<string, AuthToken>();

  verify(token: string): AuthToken {
    const decoded = this.tokens.get(token);
    if (!decoded) {
      throw new Error('Invalid token');
    }
    if (decoded.expiresAt < Date.now()) {
      throw new Error('Token expired');
    }
    return decoded;
  }

  refresh(token: string): string {
    const decoded = this.verify(token);
    const newToken = `tkn_${Date.now()}`;
    this.tokens.set(newToken, {
      ...decoded,
      expiresAt: Date.now() + 3600000
    });
    return newToken;
  }

  hasRole(token: string, role: string): boolean {
    const decoded = this.verify(token);
    return decoded.roles.includes(role);
  }
}
```

## PERFECT: Reference Implementation

These tests demonstrate what AI-generated tests SHOULD look like:

```typescript
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
```

## AI Anti-Pattern 1: toBeDefined Carpet Bomb

**Violates:** TQ-no-shallow-assertions

AI tools love `toBeDefined()` because it always passes if the function returns anything.

### Bad: TypeScript

```typescript
// test/payment-processor.anti1.test.ts (BAD)
import { describe, it, expect } from 'vitest';
import { PaymentProcessor } from '../src/payment-processor';

describe('PaymentProcessor coverage maximizer', () => {
  it('should process payment', async () => {
    const processor = new PaymentProcessor();
    const result = await processor.process({
      amount: 100.00,
      currency: 'USD',
      userId: 'user_123',
      paymentMethodId: 'pm_456'
    });

    expect(result).toBeDefined();
    expect(result.transactionId).toBeDefined();
    expect(result.amount).toBeDefined();
    expect(result.currency).toBeDefined();
    expect(result.status).toBeDefined();
    expect(result.fee).toBeDefined();
    expect(result.netAmount).toBeDefined();
    expect(result.processedAt).toBeDefined();
    expect(result.metadata).toBeDefined();
  });

  it('should refund payment', async () => {
    const processor = new PaymentProcessor();
    const result = await processor.refund('txn_123', 50);

    expect(result).toBeDefined();
    expect(result.transactionId).toBeDefined();
    expect(result.amount).toBeDefined();
  });

  it('should validate payment', async () => {
    const processor = new PaymentProcessor();
    const result = await processor.validate({
      amount: 100,
      currency: 'USD',
      userId: 'user_123',
      paymentMethodId: 'pm_456'
    });

    expect(result).toBeDefined(); // Could be true, false, or anything!
  });
});
```

**Stricture Detection:**
- File: `test/payment-processor.anti1.test.ts`
- Lines: 14-22, 28-30, 38
- Rule: TQ-no-shallow-assertions
- Message: "9 assertions use toBeDefined() without verifying actual values"

### Bad: Go (testify)

```go
// payment_processor_anti1_test.go (BAD)
package payment

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestProcessPayment_CarpetBomb(t *testing.T) {
    processor := NewPaymentProcessor()
    result, err := processor.Process(PaymentRequest{
        Amount:          100.00,
        Currency:        "USD",
        UserID:          "user_123",
        PaymentMethodID: "pm_456",
    })

    assert.NotNil(t, result)
    assert.NotNil(t, result.TransactionID)
    assert.NotNil(t, result.Amount)
    assert.NotNil(t, result.Currency)
    assert.NotNil(t, result.Status)
    assert.NotNil(t, result.Fee)
    assert.NotNil(t, result.NetAmount)
    assert.NotNil(t, result.ProcessedAt)
    assert.NotNil(t, result.Metadata)
    assert.NoError(t, err)
}

func TestRefund_Defined(t *testing.T) {
    processor := NewPaymentProcessor()
    result, err := processor.Refund("txn_123", 50.00)

    assert.NotNil(t, result)
    assert.NotNil(t, result.TransactionID)
    assert.NoError(t, err)
}
```

**Stricture Detection:**
- File: `payment_processor_anti1_test.go`
- Lines: 17-25, 33-34
- Rule: TQ-no-shallow-assertions
- Message: "11 assertions use assert.NotNil without value verification"

### Bad: Python (pytest)

```python
# test_payment_processor_anti1.py (BAD)
import pytest
from payment_processor import PaymentProcessor, PaymentRequest

def test_process_payment_carpet_bomb():
    processor = PaymentProcessor()
    result = processor.process(PaymentRequest(
        amount=100.00,
        currency='USD',
        user_id='user_123',
        payment_method_id='pm_456'
    ))

    assert result is not None
    assert result.transaction_id is not None
    assert result.amount is not None
    assert result.currency is not None
    assert result.status is not None
    assert result.fee is not None
    assert result.net_amount is not None
    assert result.processed_at is not None
    assert result.metadata is not None

def test_refund_defined():
    processor = PaymentProcessor()
    result = processor.refund('txn_123', 50.00)

    assert result is not None
    assert result.transaction_id is not None
```

**Stricture Detection:**
- File: `test_payment_processor_anti1.py`
- Lines: 13-21, 27-28
- Rule: TQ-no-shallow-assertions
- Message: "10 assertions use 'is not None' without value verification"

## AI Anti-Pattern 2: Happy Path Only

**Violates:** TQ-negative-cases

AI generates 10 success tests, 0 error cases. 100% coverage of working code, 0% of error handling.

### Bad: TypeScript

```typescript
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
```

**Stricture Detection:**
- File: `test/payment-processor.anti2.test.ts`
- Lines: 5-104
- Rule: TQ-negative-cases
- Message: "10 tests, 0 error/rejection cases. Missing: negative amount, zero amount, invalid currency, invalid transaction ID"
- Severity: ERROR

## AI Anti-Pattern 3: Global Mock Mess

**Violates:** TQ-mock-scope

AI uses file-scoped `jest.mock()` for convenience, creating cross-test pollution.

### Bad: TypeScript

```typescript
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
```

**Stricture Detection:**
- File: `test/payment-processor.anti3.test.ts`
- Lines: 6-11
- Rule: TQ-mock-scope
- Message: "Global mock defined at file scope affects all 4 tests. Move mocks into test blocks or beforeEach"
- Severity: WARNING

### Bad: Go (testify)

```go
// payment_processor_anti3_test.go (BAD)
package payment

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// GLOBAL MOCK - shared across all tests
type MockGateway struct {
    mock.Mock
}

func (m *MockGateway) Charge(req ChargeRequest) (*ChargeResponse, error) {
    args := m.Called(req)
    return args.Get(0).(*ChargeResponse), args.Error(1)
}

var globalGateway = &MockGateway{}

func TestProcessPayment_Success(t *testing.T) {
    globalGateway.On("Charge", mock.Anything).Return(&ChargeResponse{Success: true}, nil)
    processor := NewPaymentProcessor(globalGateway)

    result, err := processor.Process(PaymentRequest{
        Amount:   100.00,
        Currency: "USD",
    })

    assert.NoError(t, err)
    assert.Equal(t, "succeeded", result.Status)
}

func TestProcessPayment_LargeAmount(t *testing.T) {
    // Uses same globalGateway with stale mock expectations
    processor := NewPaymentProcessor(globalGateway)

    result, err := processor.Process(PaymentRequest{
        Amount:   5000.00,
        Currency: "USD",
    })

    assert.NoError(t, err)
    assert.Equal(t, "succeeded", result.Status)
}

func TestProcessPayment_Failure(t *testing.T) {
    // Trying to override global mock behavior - will conflict
    globalGateway.On("Charge", mock.Anything).Return(nil, errors.New("gateway error"))
    processor := NewPaymentProcessor(globalGateway)

    _, err := processor.Process(PaymentRequest{
        Amount:   100.00,
        Currency: "USD",
    })

    assert.Error(t, err)
}
```

**Stricture Detection:**
- File: `payment_processor_anti3_test.go`
- Lines: 19
- Rule: TQ-mock-scope
- Message: "Global mock 'globalGateway' shared across 3 tests. Create fresh mocks in each test"
- Severity: WARNING

### Bad: Python (pytest)

```python
# test_payment_processor_anti3.py (BAD)
import pytest
from unittest.mock import Mock, patch
from payment_processor import PaymentProcessor

# GLOBAL MOCK - affects all tests in module
@patch('payment_processor.PaymentGateway')
def mock_gateway(MockGateway):
    instance = Mock()
    instance.charge.return_value = {'success': True}
    instance.refund.return_value = {'success': True}
    MockGateway.return_value = instance
    return MockGateway

def test_process_payment_success(mock_gateway):
    processor = PaymentProcessor()
    result = processor.process(
        amount=100.00,
        currency='USD',
        user_id='user_123',
        payment_method_id='pm_456'
    )
    assert result.status == 'succeeded'

def test_process_large_payment(mock_gateway):
    # Uses same mock_gateway fixture
    processor = PaymentProcessor()
    result = processor.process(
        amount=5000.00,
        currency='USD',
        user_id='user_123',
        payment_method_id='pm_456'
    )
    assert result.status == 'succeeded'

def test_process_payment_failure(mock_gateway):
    # Can't easily override global fixture behavior
    processor = PaymentProcessor()
    result = processor.process(
        amount=100.00,
        currency='USD',
        user_id='user_123',
        payment_method_id='pm_456'
    )
    # This will never fail because of global mock
    assert result.status == 'failed'
```

**Stricture Detection:**
- File: `test_payment_processor_anti3.py`
- Lines: 7-13
- Rule: TQ-mock-scope
- Message: "Global mock decorator affects all 3 tests. Use test-scoped patches or fixture factories"

## AI Anti-Pattern 4: Copypasta Test Names

**Violates:** TQ-test-naming

AI generates "should work", "test 1", "handles data" - names that reveal nothing.

### Bad: TypeScript

```typescript
// test/data-transformer.anti4.test.ts (BAD)
import { describe, it, expect } from 'vitest';
import { DataTransformer } from '../src/data-transformer';

describe('DataTransformer', () => {
  it('should work', () => {
    const transformer = new DataTransformer();
    const result = transformer.transform([{ id: 1 }, { id: 2 }]);
    expect(result.records).toHaveLength(2);
  });

  it('test 1', () => {
    const transformer = new DataTransformer();
    const result = transformer.transform([]);
    expect(result.records).toHaveLength(0);
  });

  it('test 2', () => {
    const transformer = new DataTransformer();
    const result = transformer.transform([{ id: 1 }]);
    expect(result.stats.total).toBe(1);
  });

  it('handles data', () => {
    const transformer = new DataTransformer();
    const result = transformer.transform([null, { id: 1 }]);
    expect(result.errors).toHaveLength(1);
  });

  it('it works correctly', () => {
    const transformer = new DataTransformer();
    const result = transformer.transform([{ id: 1 }, null, { id: 2 }]);
    expect(result.records.length).toBe(2);
  });

  it('should handle edge cases', () => {
    const transformer = new DataTransformer();
    const result = transformer.transform([undefined]);
    expect(result.errors.length).toBeGreaterThan(0);
  });

  it('test transform function', () => {
    const transformer = new DataTransformer();
    const result = transformer.transform([{ name: 'test' }]);
    expect(result.stats.successful).toBe(1);
  });
});
```

**Stricture Detection:**
- File: `test/data-transformer.anti4.test.ts`
- Lines: 6, 12, 18, 24, 30, 36, 42
- Rule: TQ-test-naming
- Message: "7 tests with generic names: 'should work', 'test 1', 'test 2', 'handles data', 'it works correctly', 'should handle edge cases', 'test transform function'"
- Suggestion: "Use descriptive names: 'transforms array of valid objects into records', 'returns empty result for empty input array'"

## AI Anti-Pattern 5: Assertion-Free Test

**Violates:** TQ-no-shallow-assertions

AI generates tests that call functions but never verify results.

### Bad: TypeScript

```typescript
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
```

**Stricture Detection:**
- File: `test/payment-processor.anti5.test.ts`
- Lines: 6-14, 17-20, 23-31, 34-44, 47-58
- Rule: TQ-no-shallow-assertions
- Message: "5 tests with 0 assertions"
- Severity: ERROR

## AI Anti-Pattern 6: expect.any() Escape Hatch

**Violates:** TQ-assertion-depth

AI uses `expect.any(Type)` to avoid thinking about actual values.

### Bad: TypeScript

```typescript
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
```

**Stricture Detection:**
- File: `test/payment-processor.anti6.test.ts`
- Lines: 15-23, 30-31, 43, 55-57
- Rule: TQ-assertion-depth
- Message: "12 assertions use expect.any() or typeof without value verification"
- Severity: WARNING

## AI Anti-Pattern 7: Shared State Time Bomb

**Violates:** TQ-test-isolation

AI declares mutable variables in `describe` scope, creating order-dependent tests.

### Bad: TypeScript

```typescript
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
```

**Stricture Detection:**
- File: `test/auth-middleware.anti7.test.ts`
- Lines: 7-9
- Rule: TQ-test-isolation
- Message: "3 shared mutable variables accessed by 6 tests: 'auth', 'testToken', 'testUser'. Tests are order-dependent."
- Severity: ERROR

## AI Anti-Pattern 8: Boundary Avoider

**Violates:** TQ-boundary-tested

AI tests `page=5, limit=20` but never tests 0, 1, 100, -1.

### Bad: TypeScript

```typescript
// test/pagination.anti8.test.ts (BAD)
import { describe, it, expect } from 'vitest';
import { paginate } from '../src/pagination';

describe('Pagination tests', () => {
  it('returns page 2 of results', () => {
    const data = Array.from({ length: 100 }, (_, i) => ({ id: i }));
    const result = paginate(data, { page: 2, limit: 10 });
    expect(result.items).toHaveLength(10);
    expect(result.page).toBe(2);
  });

  it('paginates with limit 20', () => {
    const data = Array.from({ length: 100 }, (_, i) => ({ id: i }));
    const result = paginate(data, { page: 3, limit: 20 });
    expect(result.items).toHaveLength(20);
  });

  it('handles page 5', () => {
    const data = Array.from({ length: 100 }, (_, i) => ({ id: i }));
    const result = paginate(data, { page: 5, limit: 15 });
    expect(result.items.length).toBeGreaterThan(0);
  });

  it('returns middle page', () => {
    const data = Array.from({ length: 200 }, (_, i) => ({ id: i }));
    const result = paginate(data, { page: 4, limit: 25 });
    expect(result.totalPages).toBe(8);
  });

  it('paginates typical dataset', () => {
    const data = Array.from({ length: 75 }, (_, i) => ({ id: i }));
    const result = paginate(data, { page: 2, limit: 10 });
    expect(result.hasNextPage).toBe(true);
  });

  it('works with standard limit', () => {
    const data = Array.from({ length: 100 }, (_, i) => ({ id: i }));
    const result = paginate(data, { page: 3, limit: 25 });
    expect(result.items).toHaveLength(25);
  });

  // Missing: page=0, page=1, page=-1, page=999999
  // Missing: limit=0, limit=1, limit=-1, limit=999999
  // Missing: empty data array
  // Missing: data.length < limit
  // Missing: page > totalPages
});
```

**Stricture Detection:**
- File: `test/pagination.anti8.test.ts`
- Lines: 5-42
- Rule: TQ-boundary-tested
- Message: "Function paginate() tested with typical values only. Missing boundary tests: page ≤ 0, page > totalPages, limit ≤ 0, limit > data.length, empty array"
- Severity: WARNING

## AI Anti-Pattern 9: Missing Return Fields

**Violates:** TQ-return-type-verified

AI checks 2 of 8 fields, declares victory.

### Bad: TypeScript

```typescript
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
```

**Stricture Detection:**
- File: `test/payment-processor.anti9.test.ts`
- Lines: 6-18, 21-27, 30-41, 44-50, 53-58
- Rule: TQ-return-type-verified
- Message: "5 tests verify partial return types. Test 'processes payment' checks 2/8 fields. Test 'refunds payment' checks 1/8 fields."
- Severity: WARNING

## AI Anti-Pattern 10: Combined Disaster

**Violates:** All rules above

The ultimate AI-generated test: 100% coverage, 0% value.

### Bad: TypeScript

```typescript
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
```

**Stricture Detection:**
- File: `test/payment-processor.anti10.test.ts`
- Multiple violations:
  - TQ-mock-scope: Lines 8-9 (global mocks)
  - TQ-test-isolation: Lines 13-17 (shared mutable state)
  - TQ-test-naming: Lines 20, 30, 47, 94, 105 (5 generic names)
  - TQ-no-shallow-assertions: Lines 21-27, 40-42, 51-53, 94-102 (toBeDefined spam)
  - TQ-negative-cases: Lines 11-111 (10 tests, 0 error cases)
  - TQ-assertion-depth: Lines 66, 77, 88 (expect.any() abuse)
  - TQ-boundary-tested: Line 49 (typical values only)
  - TQ-return-type-verified: Lines 40-42, 51-53 (partial checks)
- Severity: ERROR
- Message: "Critical test quality issues: 8/10 anti-patterns detected"

## Coverage Theater Example

All anti-patterns achieve high coverage while verifying nothing:

```typescript
// Coverage report for anti-patterns
// File: src/payment-processor.ts
// Lines: 45/45 (100%)
// Branches: 12/12 (100%)
// Functions: 3/3 (100%)
// Statements: 45/45 (100%)

// But actual verification:
// - Error cases: 0% tested
// - Boundary conditions: 0% tested
// - Return value validation: 25% tested
// - Field completeness: 30% tested
// - State isolation: 0% (order-dependent)
```

## Summary Matrix

| Anti-Pattern | Rule | Severity | Detection Method |
|---|---|---|---|
| 1. toBeDefined Bomb | TQ-no-shallow-assertions | ERROR | Count toBeDefined/NotNil/is not None |
| 2. Happy Path Only | TQ-negative-cases | ERROR | Ratio of success:error tests |
| 3. Global Mock Mess | TQ-mock-scope | WARNING | Mock location analysis |
| 4. Copypasta Names | TQ-test-naming | WARNING | Name pattern matching |
| 5. Assertion-Free | TQ-no-shallow-assertions | ERROR | Zero expect() calls |
| 6. expect.any() | TQ-assertion-depth | WARNING | Count any() matchers |
| 7. Shared State | TQ-test-isolation | ERROR | Mutable describe-scope vars |
| 8. Boundary Avoider | TQ-boundary-tested | WARNING | Missing 0,1,-1,max tests |
| 9. Missing Fields | TQ-return-type-verified | WARNING | Return type coverage % |
| 10. Combined Disaster | ALL | ERROR | Multiple rule violations |

## Expected Stricture Output

When running Stricture on this validation set:

```bash
stricture analyze test/

ERROR: test/payment-processor.anti1.test.ts
  TQ-no-shallow-assertions (lines 14-22, 28-30, 38)
  9 assertions use toBeDefined() without value verification

ERROR: test/payment-processor.anti2.test.ts
  TQ-negative-cases (lines 5-104)
  10 success tests, 0 error cases
  Missing: negative amount, zero amount, invalid currency, invalid transaction ID

WARNING: test/payment-processor.anti3.test.ts
  TQ-mock-scope (lines 6-11)
  Global mock affects all 4 tests. Move to beforeEach or test scope

WARNING: test/data-transformer.anti4.test.ts
  TQ-test-naming (lines 6, 12, 18, 24, 30, 36, 42)
  7 generic test names. Use descriptive names explaining what is tested

ERROR: test/payment-processor.anti5.test.ts
  TQ-no-shallow-assertions (lines 6-58)
  5 tests with 0 assertions

WARNING: test/payment-processor.anti6.test.ts
  TQ-assertion-depth (lines 15-23, 30-31, 43, 55-57)
  12 assertions use expect.any() without value verification

ERROR: test/auth-middleware.anti7.test.ts
  TQ-test-isolation (lines 7-9)
  Shared mutable state: 'auth', 'testToken', 'testUser'
  Tests are order-dependent

WARNING: test/pagination.anti8.test.ts
  TQ-boundary-tested (lines 5-42)
  Missing boundary tests: page ≤ 0, page > totalPages, limit ≤ 0, empty array

WARNING: test/payment-processor.anti9.test.ts
  TQ-return-type-verified (lines 6-18, 21-27, 44-50, 53-58)
  4 tests verify partial return types (checks 1-2 of 8 fields)

ERROR: test/payment-processor.anti10.test.ts
  Multiple violations: 8/10 anti-patterns detected
  Critical: This test provides false confidence

Summary:
  Files analyzed: 10
  Tests analyzed: 63
  Violations: 10 ERROR, 5 WARNING
  Coverage achieved: 100%
  Actual validation: ~15%
```
