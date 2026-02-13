// stripe-client.test.ts — Comprehensive tests for Stripe integration.

import { describe, it, expect, beforeEach, vi } from "vitest";
import {
  StripeClient,
  verifyWebhookSignature,
  validateAmount,
  validateCurrency,
  validateChargeId,
  validateCustomerId,
  validateTokenFormat,
  validatePaymentIntentId,
} from "./stripe-client";
import type {
  StripeCharge,
  StripeCustomer,
  StripePaymentIntent,
  StripeList,
  StripeError,
} from "./stripe-client";
import * as crypto from "crypto";

// ── Fixtures ────────────────────────────────────────────────

const MOCK_CHARGE: StripeCharge = {
  id: "ch_1OxABC123def456",
  object: "charge",
  amount: 2000,
  amount_refunded: 0,
  currency: "usd",
  status: "succeeded",
  paid: true,
  captured: true,
  balance_transaction: "txn_1OxABC123def456",
  failure_code: null,
  failure_message: null,
  source: { id: "card_1OxABC", object: "card", last4: "4242" },
  created: 1700000000,
  livemode: false,
  metadata: {},
};

const MOCK_PENDING_CHARGE: StripeCharge = {
  ...MOCK_CHARGE,
  id: "ch_pending789",
  status: "pending",
  paid: false,
  balance_transaction: null,
};

const MOCK_FAILED_CHARGE: StripeCharge = {
  ...MOCK_CHARGE,
  id: "ch_failed012",
  status: "failed",
  paid: false,
  captured: false,
  balance_transaction: null,
  failure_code: "card_declined",
  failure_message: "Your card was declined.",
};

const MOCK_CUSTOMER: StripeCustomer = {
  id: "cus_1OxDEF789ghi012",
  object: "customer",
  email: "test@example.com",
  name: "Jane Doe",
  description: "Test customer",
  created: 1700000000,
  livemode: false,
  metadata: {},
  default_source: null,
};

const MOCK_PAYMENT_INTENT: StripePaymentIntent = {
  id: "pi_1OxGHI345jkl678",
  object: "payment_intent",
  amount: 5000,
  currency: "usd",
  status: "requires_payment_method",
  client_secret: "pi_1OxGHI345jkl678_secret_abc123",
  payment_method_types: ["card"],
  created: 1700000000,
  livemode: false,
  metadata: {},
};

const MOCK_ERROR_400: StripeError = {
  error: {
    type: "invalid_request_error",
    code: "parameter_missing",
    message: "Missing required param: currency.",
    param: "currency",
  },
};

const MOCK_ERROR_402: StripeError = {
  error: {
    type: "card_error",
    code: "card_declined",
    message: "Your card was declined.",
    param: null,
  },
};

const MOCK_ERROR_401: StripeError = {
  error: {
    type: "invalid_request_error",
    message: "Invalid API Key provided: sk_test_****bad",
  },
};

const MOCK_ERROR_429: StripeError = {
  error: {
    type: "api_error",
    message: "Rate limit exceeded. Please retry after a brief wait.",
  },
};

const MOCK_ERROR_500: StripeError = {
  error: {
    type: "api_error",
    message: "An internal error occurred.",
  },
};

// ── Helper ──────────────────────────────────────────────────

function mockFetch(status: number, body: unknown): void {
  global.fetch = vi.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    json: vi.fn().mockResolvedValue(body),
  } as unknown as Response);
}

function mockFetchNetworkError(message: string): void {
  global.fetch = vi.fn().mockRejectedValue(new Error(message));
}

// ── Tests ───────────────────────────────────────────────────

describe("StripeClient", () => {
  let client: StripeClient;

  beforeEach(() => {
    client = new StripeClient("sk_test_abc123");
    vi.restoreAllMocks();
  });

  // ── Constructor ─────────────────────────────────────

  describe("constructor", () => {
    it("accepts valid API key format", () => {
      expect(() => new StripeClient("sk_test_abc123")).not.toThrow();
      expect(() => new StripeClient("sk_live_xyz789")).not.toThrow();
    });

    it("rejects invalid API key format", () => {
      expect(() => new StripeClient("pk_test_abc123")).toThrow("Invalid API key format");
      expect(() => new StripeClient("")).toThrow("Invalid API key format");
      expect(() => new StripeClient("random_string")).toThrow("Invalid API key format");
    });
  });

  // ── Validation Helpers ──────────────────────────────

  describe("validateAmount", () => {
    it("accepts valid amounts", () => {
      expect(() => validateAmount(50)).not.toThrow();
      expect(() => validateAmount(2000)).not.toThrow();
      expect(() => validateAmount(99_999_999)).not.toThrow();
    });

    it("rejects amounts below minimum", () => {
      expect(() => validateAmount(0)).toThrow("must be between 50 and 99999999");
      expect(() => validateAmount(49)).toThrow("must be between 50 and 99999999");
      expect(() => validateAmount(-100)).toThrow("must be between 50 and 99999999");
    });

    it("rejects amounts above maximum", () => {
      expect(() => validateAmount(100_000_000)).toThrow("must be between 50 and 99999999");
    });

    it("rejects non-integer amounts", () => {
      expect(() => validateAmount(19.99)).toThrow("must be an integer in cents");
      expect(() => validateAmount(0.5)).toThrow("must be an integer in cents");
    });
  });

  describe("validateCurrency", () => {
    it("accepts valid currencies", () => {
      expect(() => validateCurrency("usd")).not.toThrow();
      expect(() => validateCurrency("eur")).not.toThrow();
      expect(() => validateCurrency("jpy")).not.toThrow();
    });

    it("rejects invalid currencies", () => {
      expect(() => validateCurrency("USD")).toThrow("Invalid currency");
      expect(() => validateCurrency("xyz")).toThrow("Invalid currency");
      expect(() => validateCurrency("")).toThrow("Invalid currency");
    });
  });

  describe("validateChargeId", () => {
    it("accepts valid charge IDs", () => {
      expect(() => validateChargeId("ch_1OxABC123")).not.toThrow();
    });

    it("rejects invalid charge IDs", () => {
      expect(() => validateChargeId("cus_123")).toThrow("expected ch_*");
      expect(() => validateChargeId("random")).toThrow("expected ch_*");
      expect(() => validateChargeId("")).toThrow("expected ch_*");
    });
  });

  describe("validateCustomerId", () => {
    it("accepts valid customer IDs", () => {
      expect(() => validateCustomerId("cus_1OxDEF789")).not.toThrow();
    });

    it("rejects invalid customer IDs", () => {
      expect(() => validateCustomerId("ch_123")).toThrow("expected cus_*");
      expect(() => validateCustomerId("")).toThrow("expected cus_*");
    });
  });

  describe("validateTokenFormat", () => {
    it("accepts valid token formats", () => {
      expect(() => validateTokenFormat("tok_visa123")).not.toThrow();
      expect(() => validateTokenFormat("card_1OxABC")).not.toThrow();
    });

    it("rejects invalid token formats", () => {
      expect(() => validateTokenFormat("src_123")).toThrow("expected tok_* or card_*");
      expect(() => validateTokenFormat("")).toThrow("expected tok_* or card_*");
    });
  });

  describe("validatePaymentIntentId", () => {
    it("accepts valid payment intent IDs", () => {
      expect(() => validatePaymentIntentId("pi_1OxGHI345")).not.toThrow();
    });

    it("rejects invalid payment intent IDs", () => {
      expect(() => validatePaymentIntentId("ch_123")).toThrow("expected pi_*");
    });
  });

  // ── createCharge ────────────────────────────────────

  describe("createCharge", () => {
    it("returns charge on success", async () => {
      mockFetch(200, MOCK_CHARGE);
      const result = await client.createCharge({
        amount: 2000,
        currency: "usd",
        source: "tok_visa",
      });
      expect(result.ok).toBe(true);
      if (result.ok) {
        expect(result.data.id).toBe("ch_1OxABC123def456");
        expect(result.data.object).toBe("charge");
        expect(result.data.amount).toBe(2000);
        expect(result.data.currency).toBe("usd");
        expect(result.data.status).toBe("succeeded");
        expect(result.data.paid).toBe(true);
        expect(result.data.captured).toBe(true);
        expect(result.data.balance_transaction).toBe("txn_1OxABC123def456");
        expect(result.data.failure_code).toBeNull();
        expect(result.data.failure_message).toBeNull();
        expect(typeof result.data.created).toBe("number");
        expect(typeof result.data.livemode).toBe("boolean");
        expect(result.data.metadata).toEqual({});
      }
    });

    it("sends idempotency key header when provided", async () => {
      mockFetch(200, MOCK_CHARGE);
      await client.createCharge({
        amount: 2000,
        currency: "usd",
        source: "tok_visa",
        idempotencyKey: "unique-key-123",
      });
      const fetchCall = (global.fetch as ReturnType<typeof vi.fn>).mock.calls[0];
      expect(fetchCall[1].headers["Idempotency-Key"]).toBe("unique-key-123");
    });

    it("sends metadata as bracket-encoded fields", async () => {
      mockFetch(200, MOCK_CHARGE);
      await client.createCharge({
        amount: 2000,
        currency: "usd",
        source: "tok_visa",
        metadata: { order_id: "ord_123" },
      });
      const fetchCall = (global.fetch as ReturnType<typeof vi.fn>).mock.calls[0];
      const body = fetchCall[1].body as string;
      expect(body).toContain("metadata%5Border_id%5D=ord_123");
    });

    it("returns error on 400 (missing field)", async () => {
      mockFetch(400, MOCK_ERROR_400);
      const result = await client.createCharge({
        amount: 2000,
        currency: "usd",
        source: "tok_visa",
      });
      expect(result.ok).toBe(false);
      if (!result.ok) {
        expect(result.status).toBe(400);
        expect(result.error.error.type).toBe("invalid_request_error");
        expect(result.error.error.message).toContain("currency");
        expect(result.error.error.param).toBe("currency");
      }
    });

    it("returns error on 401 (bad API key)", async () => {
      mockFetch(401, MOCK_ERROR_401);
      const result = await client.createCharge({
        amount: 2000,
        currency: "usd",
        source: "tok_visa",
      });
      expect(result.ok).toBe(false);
      if (!result.ok) {
        expect(result.status).toBe(401);
        expect(result.error.error.type).toBe("invalid_request_error");
      }
    });

    it("returns error on 402 (card declined)", async () => {
      mockFetch(402, MOCK_ERROR_402);
      const result = await client.createCharge({
        amount: 2000,
        currency: "usd",
        source: "tok_chargeDeclined",
      });
      expect(result.ok).toBe(false);
      if (!result.ok) {
        expect(result.status).toBe(402);
        expect(result.error.error.type).toBe("card_error");
        expect(result.error.error.code).toBe("card_declined");
      }
    });

    it("returns error on 429 (rate limit)", async () => {
      mockFetch(429, MOCK_ERROR_429);
      const result = await client.createCharge({
        amount: 2000,
        currency: "usd",
        source: "tok_visa",
      });
      expect(result.ok).toBe(false);
      if (!result.ok) {
        expect(result.status).toBe(429);
        expect(result.error.error.type).toBe("api_error");
      }
    });

    it("returns error on 500 (server error)", async () => {
      mockFetch(500, MOCK_ERROR_500);
      const result = await client.createCharge({
        amount: 2000,
        currency: "usd",
        source: "tok_visa",
      });
      expect(result.ok).toBe(false);
      if (!result.ok) {
        expect(result.status).toBe(500);
      }
    });

    it("throws on network failure", async () => {
      mockFetchNetworkError("ECONNREFUSED");
      await expect(
        client.createCharge({ amount: 2000, currency: "usd", source: "tok_visa" })
      ).rejects.toThrow("Network error creating charge: ECONNREFUSED");
    });

    it("rejects amount below minimum before calling API", async () => {
      await expect(
        client.createCharge({ amount: 10, currency: "usd", source: "tok_visa" })
      ).rejects.toThrow("must be between 50 and 99999999");
      expect(global.fetch).toBeUndefined();
    });

    it("rejects non-integer amount before calling API", async () => {
      await expect(
        client.createCharge({ amount: 19.99, currency: "usd", source: "tok_visa" })
      ).rejects.toThrow("must be an integer in cents");
    });

    it("rejects invalid currency before calling API", async () => {
      await expect(
        client.createCharge({ amount: 2000, currency: "xxx" as never, source: "tok_visa" })
      ).rejects.toThrow("Invalid currency");
    });

    it("rejects invalid source format before calling API", async () => {
      await expect(
        client.createCharge({ amount: 2000, currency: "usd", source: "invalid_source" })
      ).rejects.toThrow("expected tok_* or card_*");
    });
  });

  // ── retrieveCharge ──────────────────────────────────

  describe("retrieveCharge", () => {
    it("returns charge on success", async () => {
      mockFetch(200, MOCK_CHARGE);
      const result = await client.retrieveCharge("ch_1OxABC123def456");
      expect(result.ok).toBe(true);
      if (result.ok) {
        expect(result.data.id).toBe("ch_1OxABC123def456");
        expect(result.data.amount).toBe(2000);
        expect(result.data.status).toBe("succeeded");
      }
    });

    it("handles pending charge status", async () => {
      mockFetch(200, MOCK_PENDING_CHARGE);
      const result = await client.retrieveCharge("ch_pending789");
      expect(result.ok).toBe(true);
      if (result.ok) {
        expect(result.data.status).toBe("pending");
        expect(result.data.paid).toBe(false);
        expect(result.data.balance_transaction).toBeNull();
      }
    });

    it("handles failed charge with failure fields", async () => {
      mockFetch(200, MOCK_FAILED_CHARGE);
      const result = await client.retrieveCharge("ch_failed012");
      expect(result.ok).toBe(true);
      if (result.ok) {
        expect(result.data.status).toBe("failed");
        expect(result.data.failure_code).toBe("card_declined");
        expect(result.data.failure_message).toBe("Your card was declined.");
      }
    });

    it("returns 404 for non-existent charge", async () => {
      const notFoundError: StripeError = {
        error: {
          type: "invalid_request_error",
          message: "No such charge: ch_nonexistent",
          param: null,
        },
      };
      mockFetch(404, notFoundError);
      const result = await client.retrieveCharge("ch_nonexistent");
      expect(result.ok).toBe(false);
      if (!result.ok) {
        expect(result.status).toBe(404);
      }
    });

    it("rejects invalid charge ID format", async () => {
      await expect(client.retrieveCharge("not_a_charge_id")).rejects.toThrow("expected ch_*");
    });

    it("throws on network failure", async () => {
      mockFetchNetworkError("ETIMEDOUT");
      await expect(client.retrieveCharge("ch_1OxABC123")).rejects.toThrow(
        "Network error retrieving charge: ETIMEDOUT"
      );
    });
  });

  // ── listAllCharges (pagination) ─────────────────────

  describe("listAllCharges", () => {
    it("fetches single page when has_more is false", async () => {
      const page: StripeList<StripeCharge> = {
        object: "list",
        url: "/v1/charges",
        has_more: false,
        data: [MOCK_CHARGE],
      };
      mockFetch(200, page);
      const result = await client.listAllCharges();
      expect(result.ok).toBe(true);
      if (result.ok) {
        expect(result.data).toHaveLength(1);
        expect(result.data[0].id).toBe("ch_1OxABC123def456");
      }
    });

    it("paginates through multiple pages", async () => {
      const page1: StripeList<StripeCharge> = {
        object: "list",
        url: "/v1/charges",
        has_more: true,
        data: [MOCK_CHARGE],
      };
      const page2: StripeList<StripeCharge> = {
        object: "list",
        url: "/v1/charges",
        has_more: false,
        data: [MOCK_PENDING_CHARGE],
      };

      global.fetch = vi.fn()
        .mockResolvedValueOnce({
          ok: true, status: 200, json: vi.fn().mockResolvedValue(page1),
        } as unknown as Response)
        .mockResolvedValueOnce({
          ok: true, status: 200, json: vi.fn().mockResolvedValue(page2),
        } as unknown as Response);

      const result = await client.listAllCharges();
      expect(result.ok).toBe(true);
      if (result.ok) {
        expect(result.data).toHaveLength(2);
        expect(result.data[0].id).toBe("ch_1OxABC123def456");
        expect(result.data[1].id).toBe("ch_pending789");
      }

      // Verify second call used starting_after
      const secondCall = (global.fetch as ReturnType<typeof vi.fn>).mock.calls[1];
      expect(secondCall[0]).toContain("starting_after=ch_1OxABC123def456");
    });

    it("returns error from first page failure", async () => {
      mockFetch(401, MOCK_ERROR_401);
      const result = await client.listAllCharges();
      expect(result.ok).toBe(false);
      if (!result.ok) {
        expect(result.status).toBe(401);
      }
    });
  });

  // ── createCustomer ──────────────────────────────────

  describe("createCustomer", () => {
    it("creates customer successfully", async () => {
      mockFetch(200, MOCK_CUSTOMER);
      const result = await client.createCustomer({
        email: "test@example.com",
        name: "Jane Doe",
        description: "Test customer",
      });
      expect(result.ok).toBe(true);
      if (result.ok) {
        expect(result.data.id).toMatch(/^cus_/);
        expect(result.data.object).toBe("customer");
        expect(result.data.email).toBe("test@example.com");
        expect(result.data.name).toBe("Jane Doe");
        expect(result.data.default_source).toBeNull();
      }
    });

    it("creates customer with minimal params", async () => {
      mockFetch(200, { ...MOCK_CUSTOMER, email: null, name: null, description: null });
      const result = await client.createCustomer({});
      expect(result.ok).toBe(true);
    });

    it("returns error on 400", async () => {
      mockFetch(400, MOCK_ERROR_400);
      const result = await client.createCustomer({ email: "not-an-email" });
      expect(result.ok).toBe(false);
      if (!result.ok) {
        expect(result.status).toBe(400);
      }
    });

    it("validates source format when provided", async () => {
      await expect(
        client.createCustomer({ source: "invalid_src" })
      ).rejects.toThrow("expected tok_* or card_*");
    });
  });

  // ── createPaymentIntent ─────────────────────────────

  describe("createPaymentIntent", () => {
    it("creates payment intent successfully", async () => {
      mockFetch(200, MOCK_PAYMENT_INTENT);
      const result = await client.createPaymentIntent({
        amount: 5000,
        currency: "usd",
        payment_method_types: ["card"],
      });
      expect(result.ok).toBe(true);
      if (result.ok) {
        expect(result.data.id).toMatch(/^pi_/);
        expect(result.data.object).toBe("payment_intent");
        expect(result.data.amount).toBe(5000);
        expect(result.data.status).toBe("requires_payment_method");
        expect(result.data.client_secret).toContain("_secret_");
        expect(result.data.payment_method_types).toEqual(["card"]);
      }
    });

    it("handles all 7 payment intent statuses", async () => {
      const statuses = [
        "requires_payment_method",
        "requires_confirmation",
        "requires_action",
        "processing",
        "requires_capture",
        "canceled",
        "succeeded",
      ] as const;

      for (const status of statuses) {
        mockFetch(200, { ...MOCK_PAYMENT_INTENT, status });
        const result = await client.createPaymentIntent({
          amount: 5000,
          currency: "usd",
        });
        expect(result.ok).toBe(true);
        if (result.ok) {
          expect(result.data.status).toBe(status);
        }
      }
    });

    it("validates amount range", async () => {
      await expect(
        client.createPaymentIntent({ amount: 10, currency: "usd" })
      ).rejects.toThrow("must be between 50 and 99999999");
    });

    it("validates currency", async () => {
      await expect(
        client.createPaymentIntent({ amount: 5000, currency: "abc" as never })
      ).rejects.toThrow("Invalid currency");
    });

    it("validates customer ID format when provided", async () => {
      await expect(
        client.createPaymentIntent({
          amount: 5000,
          currency: "usd",
          customer: "not_a_customer_id",
        })
      ).rejects.toThrow("expected cus_*");
    });

    it("returns error on 402", async () => {
      mockFetch(402, MOCK_ERROR_402);
      const result = await client.createPaymentIntent({
        amount: 5000,
        currency: "usd",
      });
      expect(result.ok).toBe(false);
      if (!result.ok) {
        expect(result.status).toBe(402);
        expect(result.error.error.type).toBe("card_error");
      }
    });
  });

  // ── Webhook Verification ────────────────────────────

  describe("verifyWebhookSignature", () => {
    const webhookSecret = "whsec_test_secret_key";

    function signPayload(payload: string, secret: string, timestamp: number): string {
      const signedPayload = `${timestamp}.${payload}`;
      const signature = crypto
        .createHmac("sha256", secret)
        .update(signedPayload)
        .digest("hex");
      return `t=${timestamp},v1=${signature}`;
    }

    it("verifies valid webhook signature", () => {
      const payload = JSON.stringify({
        id: "evt_test123",
        object: "event",
        type: "charge.succeeded",
        data: { object: { id: "ch_1OxABC123def456" } },
        created: 1700000000,
        livemode: false,
      });
      const timestamp = Math.floor(Date.now() / 1000);
      const header = signPayload(payload, webhookSecret, timestamp);

      const event = verifyWebhookSignature(payload, header, webhookSecret);
      expect(event.type).toBe("charge.succeeded");
      expect(event.id).toBe("evt_test123");
    });

    it("rejects invalid signature", () => {
      const payload = '{"id":"evt_test","type":"charge.succeeded"}';
      const timestamp = Math.floor(Date.now() / 1000);
      const header = `t=${timestamp},v1=0000000000000000000000000000000000000000000000000000000000000000`;

      expect(() => verifyWebhookSignature(payload, header, webhookSecret)).toThrow(
        "Webhook signature verification failed"
      );
    });

    it("rejects expired timestamp", () => {
      const payload = '{"id":"evt_test","type":"charge.succeeded"}';
      const staleTimestamp = Math.floor(Date.now() / 1000) - 600; // 10 min ago
      const header = signPayload(payload, webhookSecret, staleTimestamp);

      expect(() => verifyWebhookSignature(payload, header, webhookSecret)).toThrow(
        "Webhook timestamp too old"
      );
    });

    it("rejects malformed signature header", () => {
      expect(() =>
        verifyWebhookSignature("{}", "garbage_header", webhookSecret)
      ).toThrow("Invalid Stripe-Signature header format");
    });

    it("rejects header missing v1 signature", () => {
      expect(() =>
        verifyWebhookSignature("{}", "t=1234567890", webhookSecret)
      ).toThrow("Invalid Stripe-Signature header format");
    });

    it("rejects header missing timestamp", () => {
      expect(() =>
        verifyWebhookSignature("{}", "v1=abc123", webhookSecret)
      ).toThrow("Invalid Stripe-Signature header format");
    });

    it("accepts Buffer payload", () => {
      const payloadStr = JSON.stringify({
        id: "evt_buf123",
        object: "event",
        type: "customer.created",
        data: { object: { id: "cus_123" } },
        created: 1700000000,
        livemode: false,
      });
      const timestamp = Math.floor(Date.now() / 1000);
      const header = signPayload(payloadStr, webhookSecret, timestamp);

      const event = verifyWebhookSignature(Buffer.from(payloadStr), header, webhookSecret);
      expect(event.type).toBe("customer.created");
    });
  });
});
