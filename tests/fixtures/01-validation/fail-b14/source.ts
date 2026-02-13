// B14: Fetches first page of charges but ignores has_more flag.

class StripeClientB14 {
  private readonly baseUrl = "https://api.stripe.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  // BUG: Returns only the first page of results.
  // If there are 250 charges, this returns at most 100 (Stripe's max
  // page size) and silently discards the remaining 150.
  async listCharges(limit: number = 100): Promise<StripeResult<StripeCharge[]>> {
    const query = new URLSearchParams();
    query.set("limit", String(limit));

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/charges?${query.toString()}`, {
        headers: { Authorization: `Bearer ${this.apiKey}` },
      });
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = (await response.json()) as StripeError;
      return { ok: false, status: response.status, error: errorBody };
    }

    const page = (await response.json()) as StripeList<StripeCharge>;

    // BUG: page.has_more is never checked.
    // The manifest response includes has_more: { type: boolean, required: true }
    // which indicates whether more pages exist. This implementation returns
    // page.data directly, ignoring any subsequent pages.
    return { ok: true, data: page.data };
  }

  // Example of how this causes real damage:
  async calculateTotalRevenue(): Promise<number> {
    const result = await this.listCharges();
    if (!result.ok) throw new Error("Failed to list charges");

    // BUG: Only sums the first page. If there are 500 charges, this
    // reports revenue for only the first 100, potentially underreporting
    // by 80% or more.
    return result.data
      .filter((c) => c.status === "succeeded")
      .reduce((sum, c) => sum + c.amount, 0);
  }
}
