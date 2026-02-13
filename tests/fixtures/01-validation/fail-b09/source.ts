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
