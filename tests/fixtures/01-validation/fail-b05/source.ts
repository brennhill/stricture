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
