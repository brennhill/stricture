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
