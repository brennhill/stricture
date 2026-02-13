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
