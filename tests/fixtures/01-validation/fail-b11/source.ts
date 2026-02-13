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
