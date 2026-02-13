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
