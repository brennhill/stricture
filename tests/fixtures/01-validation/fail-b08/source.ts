// B08: Only handles "succeeded" and "failed" charge statuses, ignoring "pending".

class StripeClientB08 {
  private readonly baseUrl = "https://api.stripe.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

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

    const charge = (await response.json()) as StripeCharge;
    return { ok: true, data: charge };
  }

  // BUG: Only branches on "succeeded" and "failed".
  // The manifest declares status as enum ["succeeded", "pending", "failed"].
  // A charge using bank-based payment methods (ACH, SEPA) commonly returns
  // "pending", which falls through to the default and throws.
  processChargeResult(charge: StripeCharge): string {
    switch (charge.status) {
      case "succeeded":
        return `Payment complete: ${charge.id}`;
      case "failed":
        return `Payment failed: ${charge.failure_message ?? "Unknown error"}`;
      default:
        // BUG: "pending" hits this branch and throws an error instead of
        // being handled as a valid in-progress state.
        throw new Error(`Unknown charge status: ${charge.status}`);
    }
  }
}
