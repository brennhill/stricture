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
