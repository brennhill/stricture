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
