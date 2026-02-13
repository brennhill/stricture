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
