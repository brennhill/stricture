# 01 — Stripe Payments API

**Why included:** Monetary precision, enum states, webhooks, idempotency, prefixed IDs.

---

## Manifest Fragment

```yaml
contracts:
  # ── Charges ──────────────────────────────────────────────
  - id: "stripe-charges-create"
    producer: stripe
    consumers: [my-service]
    protocol: http
    endpoints:
      - path: "/v1/charges"
        method: POST
        request:
          content_type: application/x-www-form-urlencoded
          headers:
            Authorization: { type: string, format: "Bearer sk_*", required: true }
            Idempotency-Key: { type: string, required: false }
          fields:
            amount:      { type: integer, range: [50, 99999999], required: true }
            currency:    { type: enum, values: ["usd","eur","gbp","cad","aud","jpy","chf","sek","nok","dkk","nzd","sgd","hkd","mxn","brl","inr"], required: true }
            source:      { type: string, format: "tok_*|card_*", required: true }
            description: { type: string, required: false }
            metadata:    { type: object, required: false }
            capture:     { type: boolean, required: false, default: true }
        response:
          fields:
            id:                  { type: string, format: "ch_*", required: true }
            object:              { type: literal, value: "charge", required: true }
            amount:              { type: integer, range: [50, 99999999], required: true }
            amount_refunded:     { type: integer, range: [0, 99999999], required: true }
            currency:            { type: enum, values: ["usd","eur","gbp","cad","aud","jpy","chf","sek","nok","dkk","nzd","sgd","hkd","mxn","brl","inr"], required: true }
            status:              { type: enum, values: ["succeeded", "pending", "failed"], required: true }
            paid:                { type: boolean, required: true }
            captured:            { type: boolean, required: true }
            balance_transaction: { type: string, format: "txn_*", required: false, nullable: true }
            failure_code:        { type: string, required: false, nullable: true }
            failure_message:     { type: string, required: false, nullable: true }
            source:              { type: object, required: true }
            created:             { type: integer, required: true }
            livemode:            { type: boolean, required: true }
            metadata:            { type: object, required: true }
          error:
            shape:
              error:
                type:    { type: enum, values: ["api_error","card_error","idempotency_error","invalid_request_error"], required: true }
                code:    { type: string, required: false }
                message: { type: string, required: true }
                param:   { type: string, required: false, nullable: true }
        status_codes: [200, 400, 401, 402, 404, 429, 500]

  - id: "stripe-charges-retrieve"
    producer: stripe
    consumers: [my-service]
    protocol: http
    endpoints:
      - path: "/v1/charges/:id"
        method: GET
        request:
          headers:
            Authorization: { type: string, format: "Bearer sk_*", required: true }
          params:
            id: { type: string, format: "ch_*", required: true }
        response:
          fields: &charge_fields
            id:                  { type: string, format: "ch_*", required: true }
            object:              { type: literal, value: "charge", required: true }
            amount:              { type: integer, range: [50, 99999999], required: true }
            amount_refunded:     { type: integer, range: [0, 99999999], required: true }
            currency:            { type: enum, values: ["usd","eur","gbp","cad","aud","jpy","chf","sek","nok","dkk","nzd","sgd","hkd","mxn","brl","inr"], required: true }
            status:              { type: enum, values: ["succeeded", "pending", "failed"], required: true }
            paid:                { type: boolean, required: true }
            captured:            { type: boolean, required: true }
            balance_transaction: { type: string, format: "txn_*", required: false, nullable: true }
            failure_code:        { type: string, required: false, nullable: true }
            failure_message:     { type: string, required: false, nullable: true }
            source:              { type: object, required: true }
            created:             { type: integer, required: true }
            livemode:            { type: boolean, required: true }
            metadata:            { type: object, required: true }
        status_codes: [200, 401, 404, 429, 500]

  # ── Customers ────────────────────────────────────────────
  - id: "stripe-customers-create"
    producer: stripe
    consumers: [my-service]
    protocol: http
    endpoints:
      - path: "/v1/customers"
        method: POST
        request:
          content_type: application/x-www-form-urlencoded
          headers:
            Authorization: { type: string, format: "Bearer sk_*", required: true }
          fields:
            email:       { type: string, format: "email", required: false }
            name:        { type: string, required: false }
            description: { type: string, required: false }
            metadata:    { type: object, required: false }
            source:      { type: string, format: "tok_*|card_*", required: false }
        response:
          fields:
            id:            { type: string, format: "cus_*", required: true }
            object:        { type: literal, value: "customer", required: true }
            email:         { type: string, required: false, nullable: true }
            name:          { type: string, required: false, nullable: true }
            description:   { type: string, required: false, nullable: true }
            created:       { type: integer, required: true }
            livemode:      { type: boolean, required: true }
            metadata:      { type: object, required: true }
            default_source:{ type: string, required: false, nullable: true }
        status_codes: [200, 400, 401, 429, 500]

  # ── Payment Intents ─────────────────────────────────────
  - id: "stripe-payment-intents-create"
    producer: stripe
    consumers: [my-service]
    protocol: http
    endpoints:
      - path: "/v1/payment_intents"
        method: POST
        request:
          content_type: application/x-www-form-urlencoded
          headers:
            Authorization: { type: string, format: "Bearer sk_*", required: true }
            Idempotency-Key: { type: string, required: false }
          fields:
            amount:               { type: integer, range: [50, 99999999], required: true }
            currency:             { type: enum, values: ["usd","eur","gbp","cad","aud","jpy","chf","sek","nok","dkk","nzd","sgd","hkd","mxn","brl","inr"], required: true }
            payment_method_types: { type: array, items: { type: string }, required: false }
            customer:             { type: string, format: "cus_*", required: false }
            description:          { type: string, required: false }
            metadata:             { type: object, required: false }
        response:
          fields:
            id:                    { type: string, format: "pi_*", required: true }
            object:                { type: literal, value: "payment_intent", required: true }
            amount:                { type: integer, range: [50, 99999999], required: true }
            currency:              { type: enum, values: ["usd","eur","gbp","cad","aud","jpy","chf","sek","nok","dkk","nzd","sgd","hkd","mxn","brl","inr"], required: true }
            status:                { type: enum, values: ["requires_payment_method","requires_confirmation","requires_action","processing","requires_capture","canceled","succeeded"], required: true }
            client_secret:         { type: string, format: "pi_*_secret_*", required: true }
            payment_method_types:  { type: array, items: { type: string }, required: true }
            created:               { type: integer, required: true }
            livemode:              { type: boolean, required: true }
            metadata:              { type: object, required: true }
        status_codes: [200, 400, 401, 402, 429, 500]

  # ── Charges List ─────────────────────────────────────────
  - id: "stripe-charges-list"
    producer: stripe
    consumers: [my-service]
    protocol: http
    endpoints:
      - path: "/v1/charges"
        method: GET
        request:
          headers:
            Authorization: { type: string, format: "Bearer sk_*", required: true }
          query:
            limit:          { type: integer, range: [1, 100], required: false, default: 10 }
            starting_after: { type: string, format: "ch_*", required: false }
            created:        { type: object, required: false }
        response:
          fields:
            object:   { type: literal, value: "list", required: true }
            url:      { type: string, required: true }
            has_more: { type: boolean, required: true }
            data:     { type: array, items: { type: object }, required: true }
        status_codes: [200, 401, 429, 500]

  # ── Webhooks ─────────────────────────────────────────────
  - id: "stripe-webhooks"
    producer: stripe
    consumers: [my-service]
    protocol: http
    direction: inbound
    verification:
      method: hmac-sha256
      header: Stripe-Signature
      format: "t={timestamp},v1={signature}"
      tolerance_seconds: 300
    events:
      - "charge.succeeded"
      - "charge.failed"
      - "payment_intent.succeeded"
      - "payment_intent.payment_failed"
      - "customer.created"
      - "customer.updated"
```

---

## PERFECT -- Correct Integration

### Implementation (TypeScript)

```typescript
// stripe-client.ts — Production-quality Stripe integration.

// ── Types ──────────────────────────────────────────────────

interface StripeError {
  error: {
    type: "api_error" | "card_error" | "idempotency_error" | "invalid_request_error";
    code?: string;
    message: string;
    param?: string | null;
  };
}

interface StripeCharge {
  id: string;                          // ch_*
  object: "charge";
  amount: number;                      // integer, cents
  amount_refunded: number;
  currency: StripeCurrency;
  status: "succeeded" | "pending" | "failed";
  paid: boolean;
  captured: boolean;
  balance_transaction: string | null;  // txn_* or null
  failure_code: string | null;
  failure_message: string | null;
  source: Record<string, unknown>;
  created: number;
  livemode: boolean;
  metadata: Record<string, string>;
}

interface StripeCustomer {
  id: string;                          // cus_*
  object: "customer";
  email: string | null;
  name: string | null;
  description: string | null;
  created: number;
  livemode: boolean;
  metadata: Record<string, string>;
  default_source: string | null;
}

interface StripePaymentIntent {
  id: string;                          // pi_*
  object: "payment_intent";
  amount: number;
  currency: StripeCurrency;
  status:
    | "requires_payment_method"
    | "requires_confirmation"
    | "requires_action"
    | "processing"
    | "requires_capture"
    | "canceled"
    | "succeeded";
  client_secret: string;              // pi_*_secret_*
  payment_method_types: string[];
  created: number;
  livemode: boolean;
  metadata: Record<string, string>;
}

interface StripeList<T> {
  object: "list";
  url: string;
  has_more: boolean;
  data: T[];
}

type StripeCurrency =
  | "usd" | "eur" | "gbp" | "cad" | "aud"
  | "jpy" | "chf" | "sek" | "nok" | "dkk"
  | "nzd" | "sgd" | "hkd" | "mxn" | "brl" | "inr";

interface CreateChargeParams {
  amount: number;
  currency: StripeCurrency;
  source: string;
  description?: string;
  metadata?: Record<string, string>;
  capture?: boolean;
  idempotencyKey?: string;
}

interface CreateCustomerParams {
  email?: string;
  name?: string;
  description?: string;
  metadata?: Record<string, string>;
  source?: string;
}

interface CreatePaymentIntentParams {
  amount: number;
  currency: StripeCurrency;
  payment_method_types?: string[];
  customer?: string;
  description?: string;
  metadata?: Record<string, string>;
  idempotencyKey?: string;
}

interface ListChargesParams {
  limit?: number;
  starting_after?: string;
  created?: { gt?: number; gte?: number; lt?: number; lte?: number };
}

type StripeResult<T> =
  | { ok: true; data: T }
  | { ok: false; status: number; error: StripeError };

// ── Validation Helpers ──────────────────────────────────────

const CHARGE_ID_PATTERN = /^ch_[a-zA-Z0-9]+$/;
const CUSTOMER_ID_PATTERN = /^cus_[a-zA-Z0-9]+$/;
const TOKEN_PATTERN = /^(tok|card)_[a-zA-Z0-9]+$/;
const PI_ID_PATTERN = /^pi_[a-zA-Z0-9]+$/;

const VALID_CURRENCIES: ReadonlySet<string> = new Set([
  "usd", "eur", "gbp", "cad", "aud",
  "jpy", "chf", "sek", "nok", "dkk",
  "nzd", "sgd", "hkd", "mxn", "brl", "inr",
]);

const MIN_AMOUNT = 50;
const MAX_AMOUNT = 99_999_999;

function validateAmount(amount: number): void {
  if (!Number.isInteger(amount)) {
    throw new Error(`amount must be an integer in cents, got ${amount}`);
  }
  if (amount < MIN_AMOUNT || amount > MAX_AMOUNT) {
    throw new Error(`amount must be between ${MIN_AMOUNT} and ${MAX_AMOUNT}, got ${amount}`);
  }
}

function validateCurrency(currency: string): asserts currency is StripeCurrency {
  if (!VALID_CURRENCIES.has(currency)) {
    throw new Error(`Invalid currency: ${currency}`);
  }
}

function validateChargeId(id: string): void {
  if (!CHARGE_ID_PATTERN.test(id)) {
    throw new Error(`Invalid charge ID format: ${id} (expected ch_*)`);
  }
}

function validateCustomerId(id: string): void {
  if (!CUSTOMER_ID_PATTERN.test(id)) {
    throw new Error(`Invalid customer ID format: ${id} (expected cus_*)`);
  }
}

function validateTokenFormat(source: string): void {
  if (!TOKEN_PATTERN.test(source)) {
    throw new Error(`Invalid source format: ${source} (expected tok_* or card_*)`);
  }
}

function validatePaymentIntentId(id: string): void {
  if (!PI_ID_PATTERN.test(id)) {
    throw new Error(`Invalid payment intent ID format: ${id} (expected pi_*)`);
  }
}

// ── Webhook Signature Verification ──────────────────────────

import * as crypto from "crypto";

interface WebhookEvent {
  id: string;
  object: "event";
  type: string;
  data: {
    object: Record<string, unknown>;
  };
  created: number;
  livemode: boolean;
}

function verifyWebhookSignature(
  payload: string | Buffer,
  signatureHeader: string,
  webhookSecret: string,
  toleranceSeconds: number = 300
): WebhookEvent {
  const payloadStr = typeof payload === "string" ? payload : payload.toString("utf8");

  // Parse "t=TIMESTAMP,v1=SIGNATURE" header
  const parts = signatureHeader.split(",");
  const timestampStr = parts.find((p) => p.startsWith("t="))?.slice(2);
  const signature = parts.find((p) => p.startsWith("v1="))?.slice(3);

  if (!timestampStr || !signature) {
    throw new Error("Invalid Stripe-Signature header format");
  }

  const timestamp = parseInt(timestampStr, 10);
  if (isNaN(timestamp)) {
    throw new Error("Invalid timestamp in Stripe-Signature header");
  }

  // Check tolerance
  const now = Math.floor(Date.now() / 1000);
  if (Math.abs(now - timestamp) > toleranceSeconds) {
    throw new Error(
      `Webhook timestamp too old: ${now - timestamp}s exceeds ${toleranceSeconds}s tolerance`
    );
  }

  // Compute expected signature
  const signedPayload = `${timestamp}.${payloadStr}`;
  const expectedSignature = crypto
    .createHmac("sha256", webhookSecret)
    .update(signedPayload)
    .digest("hex");

  // Constant-time comparison
  const sigBuffer = Buffer.from(signature, "hex");
  const expectedBuffer = Buffer.from(expectedSignature, "hex");

  if (sigBuffer.length !== expectedBuffer.length || !crypto.timingSafeEqual(sigBuffer, expectedBuffer)) {
    throw new Error("Webhook signature verification failed");
  }

  return JSON.parse(payloadStr) as WebhookEvent;
}

// ── Stripe Client ───────────────────────────────────────────

class StripeClient {
  private readonly baseUrl = "https://api.stripe.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    if (!apiKey.startsWith("sk_")) {
      throw new Error("Invalid API key format: must start with sk_");
    }
    this.apiKey = apiKey;
  }

  // ── Create Charge ─────────────────────────────────────

  async createCharge(params: CreateChargeParams): Promise<StripeResult<StripeCharge>> {
    validateAmount(params.amount);
    validateCurrency(params.currency);
    validateTokenFormat(params.source);

    const body = new URLSearchParams();
    body.set("amount", String(params.amount));
    body.set("currency", params.currency);
    body.set("source", params.source);
    if (params.description !== undefined) body.set("description", params.description);
    if (params.capture !== undefined) body.set("capture", String(params.capture));
    if (params.metadata) {
      for (const [k, v] of Object.entries(params.metadata)) {
        body.set(`metadata[${k}]`, v);
      }
    }

    const headers: Record<string, string> = {
      Authorization: `Bearer ${this.apiKey}`,
      "Content-Type": "application/x-www-form-urlencoded",
    };
    if (params.idempotencyKey) {
      headers["Idempotency-Key"] = params.idempotencyKey;
    }

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/charges`, {
        method: "POST",
        headers,
        body: body.toString(),
      });
    } catch (err) {
      throw new Error(`Network error creating charge: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = (await response.json()) as StripeError;
      return { ok: false, status: response.status, error: errorBody };
    }

    const charge = (await response.json()) as StripeCharge;
    return { ok: true, data: charge };
  }

  // ── Retrieve Charge ───────────────────────────────────

  async retrieveCharge(chargeId: string): Promise<StripeResult<StripeCharge>> {
    validateChargeId(chargeId);

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/charges/${chargeId}`, {
        method: "GET",
        headers: { Authorization: `Bearer ${this.apiKey}` },
      });
    } catch (err) {
      throw new Error(`Network error retrieving charge: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = (await response.json()) as StripeError;
      return { ok: false, status: response.status, error: errorBody };
    }

    const charge = (await response.json()) as StripeCharge;
    return { ok: true, data: charge };
  }

  // ── List Charges (with full pagination) ───────────────

  async listAllCharges(params: ListChargesParams = {}): Promise<StripeResult<StripeCharge[]>> {
    const allCharges: StripeCharge[] = [];
    let startingAfter: string | undefined = params.starting_after;
    const limit = params.limit ?? 100;

    // eslint-disable-next-line no-constant-condition
    while (true) {
      const query = new URLSearchParams();
      query.set("limit", String(limit));
      if (startingAfter) query.set("starting_after", startingAfter);
      if (params.created) {
        if (params.created.gt !== undefined)  query.set("created[gt]", String(params.created.gt));
        if (params.created.gte !== undefined) query.set("created[gte]", String(params.created.gte));
        if (params.created.lt !== undefined)  query.set("created[lt]", String(params.created.lt));
        if (params.created.lte !== undefined) query.set("created[lte]", String(params.created.lte));
      }

      let response: Response;
      try {
        response = await fetch(`${this.baseUrl}/v1/charges?${query.toString()}`, {
          method: "GET",
          headers: { Authorization: `Bearer ${this.apiKey}` },
        });
      } catch (err) {
        throw new Error(`Network error listing charges: ${(err as Error).message}`);
      }

      if (!response.ok) {
        const errorBody = (await response.json()) as StripeError;
        return { ok: false, status: response.status, error: errorBody };
      }

      const page = (await response.json()) as StripeList<StripeCharge>;
      allCharges.push(...page.data);

      if (!page.has_more || page.data.length === 0) {
        break;
      }

      startingAfter = page.data[page.data.length - 1].id;
    }

    return { ok: true, data: allCharges };
  }

  // ── Create Customer ───────────────────────────────────

  async createCustomer(params: CreateCustomerParams): Promise<StripeResult<StripeCustomer>> {
    if (params.source) validateTokenFormat(params.source);

    const body = new URLSearchParams();
    if (params.email !== undefined)       body.set("email", params.email);
    if (params.name !== undefined)        body.set("name", params.name);
    if (params.description !== undefined) body.set("description", params.description);
    if (params.source !== undefined)      body.set("source", params.source);
    if (params.metadata) {
      for (const [k, v] of Object.entries(params.metadata)) {
        body.set(`metadata[${k}]`, v);
      }
    }

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/customers`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "application/x-www-form-urlencoded",
        },
        body: body.toString(),
      });
    } catch (err) {
      throw new Error(`Network error creating customer: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = (await response.json()) as StripeError;
      return { ok: false, status: response.status, error: errorBody };
    }

    const customer = (await response.json()) as StripeCustomer;
    return { ok: true, data: customer };
  }

  // ── Create Payment Intent ─────────────────────────────

  async createPaymentIntent(params: CreatePaymentIntentParams): Promise<StripeResult<StripePaymentIntent>> {
    validateAmount(params.amount);
    validateCurrency(params.currency);
    if (params.customer) validateCustomerId(params.customer);

    const body = new URLSearchParams();
    body.set("amount", String(params.amount));
    body.set("currency", params.currency);
    if (params.payment_method_types) {
      for (const [i, pmt] of params.payment_method_types.entries()) {
        body.set(`payment_method_types[${i}]`, pmt);
      }
    }
    if (params.customer !== undefined)    body.set("customer", params.customer);
    if (params.description !== undefined) body.set("description", params.description);
    if (params.metadata) {
      for (const [k, v] of Object.entries(params.metadata)) {
        body.set(`metadata[${k}]`, v);
      }
    }

    const headers: Record<string, string> = {
      Authorization: `Bearer ${this.apiKey}`,
      "Content-Type": "application/x-www-form-urlencoded",
    };
    if (params.idempotencyKey) {
      headers["Idempotency-Key"] = params.idempotencyKey;
    }

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/payment_intents`, {
        method: "POST",
        headers,
        body: body.toString(),
      });
    } catch (err) {
      throw new Error(`Network error creating payment intent: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = (await response.json()) as StripeError;
      return { ok: false, status: response.status, error: errorBody };
    }

    const intent = (await response.json()) as StripePaymentIntent;
    return { ok: true, data: intent };
  }
}

export {
  StripeClient,
  verifyWebhookSignature,
  validateAmount,
  validateCurrency,
  validateChargeId,
  validateCustomerId,
  validateTokenFormat,
  validatePaymentIntentId,
};
export type {
  StripeCharge,
  StripeCustomer,
  StripePaymentIntent,
  StripeList,
  StripeError,
  StripeResult,
  StripeCurrency,
  CreateChargeParams,
  CreateCustomerParams,
  CreatePaymentIntentParams,
  ListChargesParams,
  WebhookEvent,
};
```

### Tests

```typescript
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
```

### Expected Stricture Result

```
PASS  0 violations
```

Stricture must produce zero violations. Every status code (200, 400, 401, 402, 404, 429, 500) has test coverage, all error paths are exercised, all enum values are handled, all assertions check specific field values, all nullable fields are tested for null, webhook signatures are verified with HMAC-SHA256, and pagination follows `has_more` through all pages.

---

## B01 -- No Error Handling

### Implementation

```typescript
// B01: No try/catch on any Stripe API call.

class StripeClientB01 {
  private readonly baseUrl = "https://api.stripe.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async createCharge(amount: number, currency: string, source: string): Promise<StripeCharge> {
    // BUG: No try/catch. If fetch() throws (DNS failure, timeout, network
    // disconnect), the error propagates as an unhandled rejection.
    const response = await fetch(`${this.baseUrl}/v1/charges`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${this.apiKey}`,
        "Content-Type": "application/x-www-form-urlencoded",
      },
      body: `amount=${amount}&currency=${currency}&source=${source}`,
    });
    return (await response.json()) as StripeCharge;
  }

  async retrieveCharge(chargeId: string): Promise<StripeCharge> {
    // BUG: Same issue -- no error handling.
    const response = await fetch(`${this.baseUrl}/v1/charges/${chargeId}`, {
      headers: { Authorization: `Bearer ${this.apiKey}` },
    });
    return (await response.json()) as StripeCharge;
  }
}
```

### Expected Violation

```
TQ-error-path-coverage: fetch() call at line 14 has no error handling.
  Network errors (ECONNREFUSED, ETIMEDOUT, DNS failures) will crash the
  caller with an unhandled promise rejection. All external HTTP calls must
  be wrapped in try/catch or .catch() with meaningful error propagation.
  Locations: createCharge:14, retrieveCharge:25
```

### What Makes This Obvious

Any linter or code reviewer catches bare `await fetch()` without error handling. In production, transient network failures are common -- a DNS blip, a load balancer timeout, or Stripe itself being briefly unreachable -- and without a try/catch boundary, the error propagates as an unhandled rejection, crashing the Node.js process or leaving the caller with an opaque stack trace instead of a recoverable error.

---

## B02 -- No Status Code Check

### Implementation

```typescript
// B02: Ignores HTTP status codes; treats all responses as success.

class StripeClientB02 {
  private readonly baseUrl = "https://api.stripe.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async createCharge(amount: number, currency: string, source: string): Promise<StripeCharge> {
    try {
      const response = await fetch(`${this.baseUrl}/v1/charges`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "application/x-www-form-urlencoded",
        },
        body: `amount=${amount}&currency=${currency}&source=${source}`,
      });

      // BUG: No check on response.ok or response.status.
      // A 402 (card declined) or 400 (missing param) response is parsed
      // as if it were a successful charge, returning an error body where
      // the caller expects a StripeCharge object.
      return (await response.json()) as StripeCharge;
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }
  }
}
```

### Expected Violation

```
CTR-status-code-handling: Client does not check response status.
  Manifest declares status_codes [200, 400, 401, 402, 404, 429, 500] but
  client treats all responses as StripeCharge. A 402 card_error response
  will be misinterpreted as a successful charge with missing fields.
  Location: createCharge:24
```

### What Makes This Obvious

The manifest explicitly declares that `/v1/charges` can return 400, 401, 402, 404, 429, and 500, each with a distinct error body shape. Without checking `response.ok` or `response.status`, the client casts the error JSON `{ error: { type, message } }` into `StripeCharge`, resulting in a charge object where `id` is `undefined`, `amount` is `undefined`, and `status` is `undefined`. Downstream code that relies on `charge.id` will silently produce corrupted data rather than surfacing the actual problem (e.g., the card was declined).

---

## B03 -- Shallow Test Assertions

### Implementation

```typescript
// B03: Implementation is correct (same as PERFECT).
// The bug is entirely in the tests.

class StripeClientB03 {
  // ... (identical to PERFECT implementation) ...
  async createCharge(params: CreateChargeParams): Promise<StripeResult<StripeCharge>> {
    validateAmount(params.amount);
    validateCurrency(params.currency);
    validateTokenFormat(params.source);

    const body = new URLSearchParams();
    body.set("amount", String(params.amount));
    body.set("currency", params.currency);
    body.set("source", params.source);

    let response: Response;
    try {
      response = await fetch("https://api.stripe.com/v1/charges", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "application/x-www-form-urlencoded",
        },
        body: body.toString(),
      });
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = (await response.json()) as StripeError;
      return { ok: false, status: response.status, error: errorBody };
    }

    return { ok: true, data: (await response.json()) as StripeCharge };
  }

  private readonly apiKey: string;
  constructor(apiKey: string) { this.apiKey = apiKey; }
}
```

### Tests (THE BUG IS HERE)

```typescript
// B03: Tests only check existence, never shape or value.

describe("StripeClientB03", () => {
  it("creates a charge", async () => {
    mockFetch(200, MOCK_CHARGE);
    const client = new StripeClientB03("sk_test_abc");
    const result = await client.createCharge({
      amount: 2000,
      currency: "usd",
      source: "tok_visa",
    });

    // BUG: These assertions prove nothing about correctness.
    // They pass even if result is { ok: true, data: { id: 999, amount: "banana" } }.
    expect(result).toBeDefined();
    expect(result.ok).toBeTruthy();
    if (result.ok) {
      expect(result.data).toBeDefined();
      expect(result.data.id).toBeDefined();
      expect(result.data.amount).toBeDefined();
      expect(result.data.currency).toBeDefined();
      expect(result.data.status).toBeDefined();
    }
  });

  it("handles errors", async () => {
    mockFetch(402, MOCK_ERROR_402);
    const client = new StripeClientB03("sk_test_abc");
    const result = await client.createCharge({
      amount: 2000,
      currency: "usd",
      source: "tok_declined",
    });

    // BUG: Only checks that error exists, not its shape or values.
    expect(result).toBeDefined();
    expect(result.ok).toBeFalsy();
    if (!result.ok) {
      expect(result.error).toBeDefined();
    }
  });
});
```

### Expected Violation

```
TQ-no-shallow-assertions: 8 assertions use toBeDefined()/toBeTruthy()
  instead of checking specific values or types. Assertions like
  expect(result.data.id).toBeDefined() pass for id=0, id="", id=false.
  Must use toBe(), toEqual(), toMatch(), or toStrictEqual() to verify
  the actual contract shape.
  Locations: test:14, test:15, test:17, test:18, test:19, test:20,
             test:31, test:34
```

### What Makes This Obvious

`expect(x).toBeDefined()` only asserts `x !== undefined`. If the API response is malformed -- `{ id: 0 }` instead of `{ id: "ch_abc" }` -- these tests still pass. They provide zero contract validation. The tests give false confidence: they create a green checkmark while the implementation could return any shape at all. Effective tests assert specific values (`toBe("ch_1OxABC123def456")`), types (`typeof result.data.amount === "number"`), and patterns (`toMatch(/^ch_/)`).

---

## B04 -- Missing Negative Tests

### Implementation

```typescript
// B04: Implementation is correct (same as PERFECT).
// The bug is in the test file: only happy-path tests.

// (Same correct implementation as PERFECT -- omitted for brevity)
```

### Tests (THE BUG IS HERE)

```typescript
// B04: Only tests the successful path. No error, edge, or failure tests.

describe("StripeClientB04", () => {
  it("creates a charge successfully", async () => {
    mockFetch(200, MOCK_CHARGE);
    const client = new StripeClient("sk_test_abc");
    const result = await client.createCharge({
      amount: 2000,
      currency: "usd",
      source: "tok_visa",
    });
    expect(result.ok).toBe(true);
    if (result.ok) {
      expect(result.data.id).toBe("ch_1OxABC123def456");
      expect(result.data.amount).toBe(2000);
    }
  });

  it("retrieves a charge successfully", async () => {
    mockFetch(200, MOCK_CHARGE);
    const client = new StripeClient("sk_test_abc");
    const result = await client.retrieveCharge("ch_1OxABC123def456");
    expect(result.ok).toBe(true);
    if (result.ok) {
      expect(result.data.id).toBe("ch_1OxABC123def456");
    }
  });

  it("creates a customer successfully", async () => {
    mockFetch(200, MOCK_CUSTOMER);
    const client = new StripeClient("sk_test_abc");
    const result = await client.createCustomer({ email: "test@example.com" });
    expect(result.ok).toBe(true);
  });

  // BUG: No tests for:
  //   - 400, 401, 402, 404, 429, 500 responses
  //   - Network failures (ECONNREFUSED, ETIMEDOUT)
  //   - Invalid input (amount < 50, bad currency, bad charge ID format)
  //   - Null/missing fields (failure_code, balance_transaction)
  //   - Pending and failed charge statuses
  //   - Webhook signature verification failures
  //   - Pagination edge cases
});
```

### Expected Violation

```
TQ-negative-cases: createCharge has 0 negative tests.
  Manifest declares 6 non-200 status codes and multiple validation
  constraints. Functions with external dependencies require tests for
  each failure mode: network errors, HTTP errors, validation rejections.
  Missing: 400 test, 401 test, 402 test, 429 test, 500 test,
           network error test, invalid amount test, invalid currency test,
           invalid source test.
  Functions without negative tests: createCharge, retrieveCharge,
  createCustomer, createPaymentIntent, verifyWebhookSignature
```

### What Makes This Obvious

A test suite that only verifies success tells you the feature works when everything goes right -- but says nothing about what happens when things go wrong. Stripe API calls can fail in seven distinct ways (network error + six HTTP status codes), and the charge creation endpoint has three input validation constraints. Without negative tests, regressions in error handling go undetected. If someone removes the try/catch or the `response.ok` check, these tests still pass.

---

## B05 -- Request Missing Required Fields

### Implementation

```typescript
// B05: Omits the required `currency` field from charge creation.

class StripeClientB05 {
  private readonly baseUrl = "https://api.stripe.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async createCharge(amount: number, source: string): Promise<StripeResult<StripeCharge>> {
    if (!Number.isInteger(amount) || amount < 50 || amount > 99_999_999) {
      throw new Error("Invalid amount");
    }

    const body = new URLSearchParams();
    body.set("amount", String(amount));
    // BUG: `currency` is required by the manifest but is never sent.
    // Stripe will return a 400: "Missing required param: currency."
    body.set("source", source);

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/charges`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "application/x-www-form-urlencoded",
        },
        body: body.toString(),
      });
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = (await response.json()) as StripeError;
      return { ok: false, status: response.status, error: errorBody };
    }

    return { ok: true, data: (await response.json()) as StripeCharge };
  }
}
```

### Expected Violation

```
CTR-request-shape: POST /v1/charges request is missing required field
  `currency`. Manifest declares currency as { type: enum, required: true }
  but the client never includes it in the request body. Every call to
  createCharge will receive a 400 response.
  Location: createCharge (missing: currency)
```

### What Makes This Obvious

The manifest explicitly declares `currency: { required: true }` on the `POST /v1/charges` endpoint. The implementation's function signature omits currency entirely -- there is no parameter for it, and no hard-coded default. Every single call will fail with Stripe's `400: Missing required param: currency`. This is a structural mismatch between what the client sends and what the server requires.

---

## B06 -- Response Type Mismatch

### Implementation

```typescript
// B06: Client's StripeCharge type is missing the `balance_transaction` field.

interface StripeChargeB06 {
  id: string;
  object: "charge";
  amount: number;
  amount_refunded: number;
  currency: string;
  status: "succeeded" | "pending" | "failed";
  paid: boolean;
  captured: boolean;
  // BUG: `balance_transaction` field is missing from the type.
  // The API always returns it (as string | null), but the client type
  // doesn't include it, so TypeScript code can never access it.
  // Any logic that needs the transaction ID (reconciliation, refund
  // tracking) will fail silently -- the field exists at runtime but
  // TypeScript reports `Property 'balance_transaction' does not exist`.
  failure_code: string | null;
  failure_message: string | null;
  source: Record<string, unknown>;
  created: number;
  livemode: boolean;
  metadata: Record<string, string>;
}

class StripeClientB06 {
  private readonly baseUrl = "https://api.stripe.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async createCharge(params: CreateChargeParams): Promise<StripeResult<StripeChargeB06>> {
    validateAmount(params.amount);
    validateCurrency(params.currency);
    validateTokenFormat(params.source);

    const body = new URLSearchParams();
    body.set("amount", String(params.amount));
    body.set("currency", params.currency);
    body.set("source", params.source);

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/charges`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "application/x-www-form-urlencoded",
        },
        body: body.toString(),
      });
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = (await response.json()) as StripeError;
      return { ok: false, status: response.status, error: errorBody };
    }

    return { ok: true, data: (await response.json()) as StripeChargeB06 };
  }
}
```

### Expected Violation

```
CTR-response-shape: Client type StripeChargeB06 is missing field
  `balance_transaction` declared in manifest as
  { type: string, format: "txn_*", nullable: true }.
  The server always returns this field but the client cannot access it
  through its type system. Downstream reconciliation or refund logic
  that needs the balance transaction ID will fail.
  Location: StripeChargeB06 interface (missing: balance_transaction)
```

### What Makes This Obvious

The manifest declares `balance_transaction` as a field on charge responses. The API sends it on every response. But the client type definition omits it. In TypeScript strict mode, any code that tries to access `charge.balance_transaction` will get a compile error. The fix is trivial (add the field), but the consequence of missing it is severe: financial reconciliation requires the `txn_*` ID to match charges to balance movements, and without it, automated accounting breaks.

---

## B07 -- Wrong Field Types

### Implementation

```typescript
// B07: `amount` is stored as a string, not a number.

interface StripeChargeB07 {
  id: string;
  object: "charge";
  amount: string;           // BUG: Should be number (integer cents).
  amount_refunded: string;  // BUG: Should be number.
  currency: string;
  status: "succeeded" | "pending" | "failed";
  paid: boolean;
  captured: boolean;
  balance_transaction: string | null;
  failure_code: string | null;
  failure_message: string | null;
  source: Record<string, unknown>;
  created: number;
  livemode: boolean;
  metadata: Record<string, string>;
}

class StripeClientB07 {
  private readonly baseUrl = "https://api.stripe.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async createCharge(
    amount: string,  // BUG: accepts string instead of number
    currency: string,
    source: string
  ): Promise<StripeResult<StripeChargeB07>> {
    // BUG: String amount means arithmetic operations produce wrong results.
    // "2000" + 500 === "2000500" (string concatenation, not addition)

    const body = new URLSearchParams();
    body.set("amount", amount);
    body.set("currency", currency);
    body.set("source", source);

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/charges`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "application/x-www-form-urlencoded",
        },
        body: body.toString(),
      });
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = (await response.json()) as StripeError;
      return { ok: false, status: response.status, error: errorBody };
    }

    return { ok: true, data: (await response.json()) as StripeChargeB07 };
  }
}
```

### Expected Violation

```
CTR-manifest-conformance: Field `amount` declared as { type: integer }
  in manifest but typed as `string` in client StripeChargeB07.
  Field `amount_refunded` declared as { type: integer } in manifest but
  typed as `string` in client.
  String amounts cause arithmetic bugs: "2000" + 500 === "2000500".
  Locations: StripeChargeB07.amount, StripeChargeB07.amount_refunded,
             createCharge parameter `amount`
```

### What Makes This Obvious

The manifest says `amount: { type: integer }`. The client types it as `string`. This is a direct contradiction. The practical impact is severe: any arithmetic on the amount field silently produces wrong results due to JavaScript string concatenation. `charge.amount + tip` produces `"200050"` instead of `2050`. Refund calculations, tax computations, and running totals all break silently.

---

## B08 -- Incomplete Enum Handling

### Implementation

```typescript
// B08: Only handles "succeeded" and "failed" charge statuses, ignoring "pending".

class StripeClientB08 {
  private readonly baseUrl = "https://api.stripe.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async createCharge(params: CreateChargeParams): Promise<StripeResult<StripeCharge>> {
    validateAmount(params.amount);
    validateCurrency(params.currency);
    validateTokenFormat(params.source);

    const body = new URLSearchParams();
    body.set("amount", String(params.amount));
    body.set("currency", params.currency);
    body.set("source", params.source);

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/charges`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "application/x-www-form-urlencoded",
        },
        body: body.toString(),
      });
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = (await response.json()) as StripeError;
      return { ok: false, status: response.status, error: errorBody };
    }

    const charge = (await response.json()) as StripeCharge;
    return { ok: true, data: charge };
  }

  // BUG: Only branches on "succeeded" and "failed".
  // The manifest declares status as enum ["succeeded", "pending", "failed"].
  // A charge using bank-based payment methods (ACH, SEPA) commonly returns
  // "pending", which falls through to the default and throws.
  processChargeResult(charge: StripeCharge): string {
    switch (charge.status) {
      case "succeeded":
        return `Payment complete: ${charge.id}`;
      case "failed":
        return `Payment failed: ${charge.failure_message ?? "Unknown error"}`;
      default:
        // BUG: "pending" hits this branch and throws an error instead of
        // being handled as a valid in-progress state.
        throw new Error(`Unknown charge status: ${charge.status}`);
    }
  }
}
```

### Expected Violation

```
CTR-strictness-parity: Charge status enum has 3 values in manifest
  ["succeeded", "pending", "failed"] but processChargeResult only
  handles 2 of 3. Missing: "pending". Bank-based payment methods
  (ACH, SEPA) commonly produce "pending" charges, which will throw
  an unexpected error.
  Location: processChargeResult switch statement (missing case: "pending")
```

### What Makes This Obvious

The manifest explicitly declares three valid charge statuses. The switch statement only handles two. The "pending" status is not rare -- it is the normal result for ACH debits, SEPA transfers, and other bank-based payments that take days to clear. When a pending charge is processed, the code throws `Unknown charge status: pending`, crashing the handler. The fix is adding a `case "pending"` that returns an appropriate message like `Payment processing: ${charge.id}`.

---

## B09 -- Missing Range Validation

### Implementation

```typescript
// B09: No validation that amount is >= 50 cents (Stripe minimum).

class StripeClientB09 {
  private readonly baseUrl = "https://api.stripe.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async createCharge(params: CreateChargeParams): Promise<StripeResult<StripeCharge>> {
    // BUG: No range validation on amount.
    // The manifest declares amount: { range: [50, 99999999] } but this
    // implementation sends whatever amount is provided.
    // amount=1 will be accepted locally but rejected by Stripe with
    // a 400: "Amount must be at least 50 cents."
    // amount=0 or amount=-100 will also reach Stripe without any guard.

    validateCurrency(params.currency);
    validateTokenFormat(params.source);

    const body = new URLSearchParams();
    body.set("amount", String(params.amount));
    body.set("currency", params.currency);
    body.set("source", params.source);

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/charges`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "application/x-www-form-urlencoded",
        },
        body: body.toString(),
      });
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = (await response.json()) as StripeError;
      return { ok: false, status: response.status, error: errorBody };
    }

    return { ok: true, data: (await response.json()) as StripeCharge };
  }
}
```

### Expected Violation

```
CTR-strictness-parity: Field `amount` has range [50, 99999999] in manifest
  but client performs no range validation before sending the request.
  Values below 50 (e.g., 1, 0, -100) will be sent to Stripe, consuming
  API quota and returning a 400 that could have been prevented locally.
  Location: createCharge (missing: amount range check)
```

### What Makes This Obvious

The manifest declares a strict range constraint: `amount: { range: [50, 99999999] }`. The PERFECT implementation validates this before making the API call. This buggy version omits the check entirely. The consequence is unnecessary API calls that will always fail -- every sub-50-cent charge request wastes a round trip and counts against rate limits, and the resulting 400 error message from Stripe is less informative than a local validation error. In batch operations processing thousands of charges, this could exhaust the rate limit quota with guaranteed-to-fail requests.

---

## B10 -- Format Not Validated

### Implementation

```typescript
// B10: Charge ID accepted as any string, no ch_* format validation.

class StripeClientB10 {
  private readonly baseUrl = "https://api.stripe.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async retrieveCharge(chargeId: string): Promise<StripeResult<StripeCharge>> {
    // BUG: No format validation on chargeId.
    // The manifest declares id: { format: "ch_*" } but this method
    // accepts any string. Passing a customer ID (cus_*), payment intent
    // ID (pi_*), or arbitrary string results in a confusing 404 from
    // Stripe instead of a clear local validation error.

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/charges/${chargeId}`, {
        headers: { Authorization: `Bearer ${this.apiKey}` },
      });
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = (await response.json()) as StripeError;
      return { ok: false, status: response.status, error: errorBody };
    }

    return { ok: true, data: (await response.json()) as StripeCharge };
  }

  async createPaymentIntent(params: CreatePaymentIntentParams): Promise<StripeResult<StripePaymentIntent>> {
    validateAmount(params.amount);
    validateCurrency(params.currency);

    // BUG: No validation on params.customer format.
    // cus_* format is required by manifest but any string is accepted.
    // Passing "customer@email.com" instead of "cus_abc123" will cause
    // a confusing 400 from Stripe.

    const body = new URLSearchParams();
    body.set("amount", String(params.amount));
    body.set("currency", params.currency);
    if (params.customer !== undefined) body.set("customer", params.customer);

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/payment_intents`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "application/x-www-form-urlencoded",
        },
        body: body.toString(),
      });
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = (await response.json()) as StripeError;
      return { ok: false, status: response.status, error: errorBody };
    }

    return { ok: true, data: (await response.json()) as StripePaymentIntent };
  }
}
```

### Expected Violation

```
CTR-strictness-parity: Manifest declares format constraints that client
  does not enforce:
  - retrieveCharge param `chargeId`: format "ch_*" not validated
  - createPaymentIntent param `customer`: format "cus_*" not validated
  Accepts arbitrary strings (email addresses, UUIDs, other ID types)
  instead of validating the required prefix pattern.
  Locations: retrieveCharge:chargeId, createPaymentIntent:customer
```

### What Makes This Obvious

Stripe's ID system uses prefixed strings to prevent cross-resource confusion: `ch_` for charges, `cus_` for customers, `pi_` for payment intents. The manifest encodes these as format constraints. Without local validation, a developer can accidentally pass a customer ID to `retrieveCharge("cus_abc123")` and receive a mysterious 404 "No such charge" error, when the real problem is that they passed the wrong type of ID. Format validation catches these mix-ups instantly with a clear error message.

---

## B11 -- Precision Loss

### Implementation

```typescript
// B11: Uses float dollars instead of integer cents for monetary amounts.

class StripeClientB11 {
  private readonly baseUrl = "https://api.stripe.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  // BUG: Accepts dollars as a float and converts to cents internally.
  // Floating-point arithmetic causes precision errors:
  //   0.1 + 0.2 = 0.30000000000000004
  //   Math.round(0.30000000000000004 * 100) = 30  (correct by luck)
  //   BUT: 1.005 * 100 = 100.49999999999999
  //   Math.round(100.49999999999999) = 100  (should be 101 -- LOST A CENT)

  async createCharge(
    amountDollars: number,  // BUG: float dollars instead of integer cents
    currency: string,
    source: string
  ): Promise<StripeResult<StripeCharge>> {
    // BUG: Float-to-int conversion loses precision.
    const amountCents = Math.round(amountDollars * 100);

    validateCurrency(currency);
    validateTokenFormat(source);

    const body = new URLSearchParams();
    body.set("amount", String(amountCents));
    body.set("currency", currency);
    body.set("source", source);

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/charges`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "application/x-www-form-urlencoded",
        },
        body: body.toString(),
      });
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = (await response.json()) as StripeError;
      return { ok: false, status: response.status, error: errorBody };
    }

    return { ok: true, data: (await response.json()) as StripeCharge };
  }

  // BUG: Summing charges in dollar floats accumulates rounding errors.
  calculateTotal(charges: StripeCharge[]): number {
    let total = 0;
    for (const charge of charges) {
      // Converting from cents to dollars for display then back for sum
      total += charge.amount / 100;  // BUG: float division
    }
    // After 1000 charges of $19.99 each:
    // Expected: 19990.00
    // Actual:   19989.999999999996 (off by fractions of a cent)
    return Math.round(total * 100) / 100;  // Rounding hides some errors
  }
}
```

### Expected Violation

```
CTR-strictness-parity: Field `amount` is declared as { type: integer }
  in manifest (cents) but client accepts float dollars and converts with
  Math.round(amountDollars * 100). Floating-point multiplication causes
  precision loss: 1.005 * 100 = 100.49999999999999, rounds to 100 instead
  of 101. Monetary amounts must use integer cents throughout.
  Locations: createCharge parameter (float dollars), calculateTotal (float division)
```

### What Makes This Obvious

The manifest declares `amount: { type: integer }` -- amounts are in cents, always. The buggy code accepts dollars as a float and multiplies by 100, which is the classic floating-point money bug. IEEE 754 cannot represent all decimal fractions exactly: `1.005 * 100` evaluates to `100.49999999999999`, not `100.5`. `Math.round()` then rounds down to `100` instead of `101`, losing a cent. Over thousands of transactions, these errors accumulate. This is why every payment API uses integer cents and the manifest enforces it.

---

## B12 -- Nullable Field Crash

### Implementation

```typescript
// B12: Accesses charge.failure_code without null check.

class StripeClientB12 {
  private readonly baseUrl = "https://api.stripe.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async createCharge(params: CreateChargeParams): Promise<StripeResult<StripeCharge>> {
    validateAmount(params.amount);
    validateCurrency(params.currency);
    validateTokenFormat(params.source);

    const body = new URLSearchParams();
    body.set("amount", String(params.amount));
    body.set("currency", params.currency);
    body.set("source", params.source);

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/charges`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "application/x-www-form-urlencoded",
        },
        body: body.toString(),
      });
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = (await response.json()) as StripeError;
      return { ok: false, status: response.status, error: errorBody };
    }

    return { ok: true, data: (await response.json()) as StripeCharge };
  }

  // BUG: Assumes failure_code and balance_transaction are always present.
  // The manifest declares both as nullable, meaning they are null on
  // successful charges. Accessing .toUpperCase() on null throws TypeError.
  formatChargeReport(charge: StripeCharge): string {
    const lines: string[] = [
      `Charge: ${charge.id}`,
      `Amount: ${charge.amount} ${charge.currency}`,
      `Status: ${charge.status}`,
      // BUG: failure_code is null on succeeded/pending charges.
      // charge.failure_code.toUpperCase() throws:
      //   TypeError: Cannot read properties of null (reading 'toUpperCase')
      `Failure Code: ${charge.failure_code.toUpperCase()}`,
      // BUG: failure_message is null on succeeded/pending charges.
      `Failure Message: ${charge.failure_message.trim()}`,
      // BUG: balance_transaction is null on pending/failed charges.
      `Transaction: ${charge.balance_transaction.replace("txn_", "TXN-")}`,
    ];
    return lines.join("\n");
  }
}
```

### Expected Violation

```
CTR-response-shape: Accessing methods on nullable fields without null
  check. Manifest declares these fields as nullable:
  - failure_code: { nullable: true } -- .toUpperCase() crashes when null
  - failure_message: { nullable: true } -- .trim() crashes when null
  - balance_transaction: { nullable: true } -- .replace() crashes when null
  On a succeeded charge, failure_code and failure_message are null.
  On a pending charge, balance_transaction is null.
  Location: formatChargeReport lines 11-14
```

### What Makes This Obvious

The manifest marks three fields as `nullable: true`: `failure_code`, `failure_message`, and `balance_transaction`. The buggy code calls string methods (`.toUpperCase()`, `.trim()`, `.replace()`) directly on these fields without checking for null. A succeeded charge has `failure_code: null` and `failure_message: null`, so calling `.toUpperCase()` on null throws a `TypeError`. This is exactly the kind of crash that happens in production but not in tests (because test fixtures typically use failed charges where these fields are populated).

---

## B13 -- Missing Webhook Signature Verification

### Implementation

```typescript
// B13: Accepts Stripe webhook payloads without verifying the signature.

import type { Request, Response as ExpressResponse } from "express";

// BUG: No signature verification at all.
// An attacker can POST a fake webhook event to trigger actions:
//   - Fake "charge.succeeded" to mark unpaid orders as paid
//   - Fake "customer.updated" to change account details
//   - Fake "payment_intent.succeeded" to fulfill unprocessed orders

async function handleStripeWebhook(req: Request, res: ExpressResponse): Promise<void> {
  // BUG: The Stripe-Signature header is completely ignored.
  // The manifest specifies verification:
  //   method: hmac-sha256
  //   header: Stripe-Signature
  //   format: "t={timestamp},v1={signature}"
  //   tolerance_seconds: 300
  // None of this is checked.

  const event = req.body;  // BUG: Raw body parsed as trusted event.

  switch (event.type) {
    case "charge.succeeded": {
      const charge = event.data.object as StripeCharge;
      await markOrderAsPaid(charge.id, charge.amount);
      break;
    }
    case "charge.failed": {
      const charge = event.data.object as StripeCharge;
      await notifyPaymentFailed(charge.id, charge.failure_message);
      break;
    }
    case "payment_intent.succeeded": {
      const intent = event.data.object as StripePaymentIntent;
      await fulfillOrder(intent.id, intent.amount);
      break;
    }
    default:
      // Ignore unhandled event types
      break;
  }

  res.status(200).json({ received: true });
}

// Placeholder functions -- in production these modify real data.
async function markOrderAsPaid(chargeId: string, amount: number): Promise<void> { /* ... */ }
async function notifyPaymentFailed(chargeId: string, message: string | null): Promise<void> { /* ... */ }
async function fulfillOrder(intentId: string, amount: number): Promise<void> { /* ... */ }

export { handleStripeWebhook };
```

### Expected Violation

```
CTR-request-shape: Webhook endpoint accepts inbound POST without
  verifying HMAC-SHA256 signature. Manifest declares:
    verification.method: hmac-sha256
    verification.header: Stripe-Signature
    verification.tolerance_seconds: 300
  The handler reads req.body as a trusted event without checking the
  Stripe-Signature header. An attacker can forge webhook events to
  mark unpaid orders as paid or trigger unauthorized fulfillment.
  Location: handleStripeWebhook (missing: signature verification)
```

### What Makes This Obvious

Webhook signature verification is a critical security control. Stripe signs every webhook payload with HMAC-SHA256 and sends the signature in the `Stripe-Signature` header. The manifest documents this explicitly. Without verification, any HTTP client can POST a fabricated event to the webhook endpoint. A `charge.succeeded` event for a charge that never happened would trigger `markOrderAsPaid`, shipping goods without collecting payment. This is not a theoretical risk -- it is a documented attack vector that Stripe warns about in their security guidelines.

---

## B14 -- Pagination

### Implementation

```typescript
// B14: Fetches first page of charges but ignores has_more flag.

class StripeClientB14 {
  private readonly baseUrl = "https://api.stripe.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  // BUG: Returns only the first page of results.
  // If there are 250 charges, this returns at most 100 (Stripe's max
  // page size) and silently discards the remaining 150.
  async listCharges(limit: number = 100): Promise<StripeResult<StripeCharge[]>> {
    const query = new URLSearchParams();
    query.set("limit", String(limit));

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/charges?${query.toString()}`, {
        headers: { Authorization: `Bearer ${this.apiKey}` },
      });
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = (await response.json()) as StripeError;
      return { ok: false, status: response.status, error: errorBody };
    }

    const page = (await response.json()) as StripeList<StripeCharge>;

    // BUG: page.has_more is never checked.
    // The manifest response includes has_more: { type: boolean, required: true }
    // which indicates whether more pages exist. This implementation returns
    // page.data directly, ignoring any subsequent pages.
    return { ok: true, data: page.data };
  }

  // Example of how this causes real damage:
  async calculateTotalRevenue(): Promise<number> {
    const result = await this.listCharges();
    if (!result.ok) throw new Error("Failed to list charges");

    // BUG: Only sums the first page. If there are 500 charges, this
    // reports revenue for only the first 100, potentially underreporting
    // by 80% or more.
    return result.data
      .filter((c) => c.status === "succeeded")
      .reduce((sum, c) => sum + c.amount, 0);
  }
}
```

### Expected Violation

```
CTR-response-shape: Client ignores `has_more` field from list response.
  Manifest declares has_more: { type: boolean, required: true } on
  GET /v1/charges. When has_more is true, additional pages exist and
  must be fetched using `starting_after` parameter. Client returns only
  first page, silently discarding remaining data.
  Location: listCharges (ignores: has_more, missing: pagination loop)
```

### What Makes This Obvious

The manifest declares `has_more` as a required boolean on list responses. Its purpose is explicit: tell the client whether more data exists. Ignoring it means the client returns partial data and presents it as complete. For a revenue calculation, this means reporting $10,000 in revenue when the actual total is $50,000, simply because only the first 100 of 500 charges were fetched. The PERFECT implementation uses a `while` loop that continues fetching with `starting_after` until `has_more` is false.

---

## B15 -- Race Condition

### Implementation

```typescript
// B15: Read-then-update on customer without any concurrency protection.

class StripeClientB15 {
  private readonly baseUrl = "https://api.stripe.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async getCustomer(customerId: string): Promise<StripeResult<StripeCustomer>> {
    validateCustomerId(customerId);

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/customers/${customerId}`, {
        headers: { Authorization: `Bearer ${this.apiKey}` },
      });
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = (await response.json()) as StripeError;
      return { ok: false, status: response.status, error: errorBody };
    }

    return { ok: true, data: (await response.json()) as StripeCustomer };
  }

  async updateCustomer(
    customerId: string,
    updates: Partial<Pick<StripeCustomer, "email" | "name" | "description">>
  ): Promise<StripeResult<StripeCustomer>> {
    validateCustomerId(customerId);

    const body = new URLSearchParams();
    if (updates.email !== undefined) body.set("email", updates.email ?? "");
    if (updates.name !== undefined) body.set("name", updates.name ?? "");
    if (updates.description !== undefined) body.set("description", updates.description ?? "");

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/customers/${customerId}`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "application/x-www-form-urlencoded",
        },
        body: body.toString(),
      });
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = (await response.json()) as StripeError;
      return { ok: false, status: response.status, error: errorBody };
    }

    return { ok: true, data: (await response.json()) as StripeCustomer };
  }

  // BUG: Read-modify-write without any concurrency protection.
  // Two concurrent calls to addTagToCustomer can overwrite each other:
  //
  // Timeline:
  //   T1: reads customer.metadata = { tier: "gold" }
  //   T2: reads customer.metadata = { tier: "gold" }
  //   T1: writes metadata = { tier: "gold", region: "us" }
  //   T2: writes metadata = { tier: "gold", cohort: "2024" }
  //   Result: metadata = { tier: "gold", cohort: "2024" }
  //   Lost: region: "us" (T1's write was silently overwritten)
  //
  // The manifest does not provide an ETag or version field for optimistic
  // locking, so the client must use Idempotency-Key or serialize access.

  async addMetadataToCustomer(
    customerId: string,
    key: string,
    value: string
  ): Promise<StripeResult<StripeCustomer>> {
    // Step 1: Read current customer
    const current = await this.getCustomer(customerId);
    if (!current.ok) return current;

    // BUG: No lock, no version check, no idempotency key.
    // Between the read above and the write below, another process
    // can modify the customer, and this write will silently overwrite
    // those changes.

    // Step 2: Merge new metadata with existing
    const updatedMetadata = { ...current.data.metadata, [key]: value };

    // Step 3: Write back (overwrites any concurrent modifications)
    const body = new URLSearchParams();
    for (const [k, v] of Object.entries(updatedMetadata)) {
      body.set(`metadata[${k}]`, v);
    }

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/customers/${customerId}`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "application/x-www-form-urlencoded",
        },
        body: body.toString(),
      });
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = (await response.json()) as StripeError;
      return { ok: false, status: response.status, error: errorBody };
    }

    return { ok: true, data: (await response.json()) as StripeCustomer };
  }
}
```

### Expected Violation

```
CTR-request-shape: Read-modify-write on POST /v1/customers/:id without
  concurrency control. addMetadataToCustomer reads metadata, merges
  locally, then writes back. Concurrent calls create a race condition
  where the second write overwrites the first, silently losing metadata.
  Stripe's customer update API accepts partial metadata updates
  (metadata[key]=value) without needing to read-modify-write. Use
  atomic partial updates or Idempotency-Key to prevent data loss.
  Location: addMetadataToCustomer (race window: between getCustomer
  and POST update, ~50-200ms network round trip)
```

### What Makes This Obvious

The read-modify-write pattern without a lock is a classic concurrency bug. The timeline in the code comments illustrates it precisely: two concurrent calls read the same state, each adds a different key, and the second write wins, silently discarding the first write's key. In production, this manifests as intermittent metadata loss that is difficult to reproduce and diagnose. The irony is that Stripe's API already supports atomic partial metadata updates -- `metadata[region]=us` appends the key without needing to read first -- so the read-modify-write pattern is both dangerous and unnecessary.
