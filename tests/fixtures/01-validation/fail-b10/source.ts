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
