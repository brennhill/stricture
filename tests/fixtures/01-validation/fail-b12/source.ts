// B12: Accesses charge.failure_code without null check.

class StripeClientB12 {
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

    return { ok: true, data: (await response.json()) as StripeCharge };
  }

  // BUG: Assumes failure_code and balance_transaction are always present.
  // The manifest declares both as nullable, meaning they are null on
  // successful charges. Accessing .toUpperCase() on null throws TypeError.
  formatChargeReport(charge: StripeCharge): string {
    const lines: string[] = [
      `Charge: ${charge.id}`,
      `Amount: ${charge.amount} ${charge.currency}`,
      `Status: ${charge.status}`,
      // BUG: failure_code is null on succeeded/pending charges.
      // charge.failure_code.toUpperCase() throws:
      //   TypeError: Cannot read properties of null (reading 'toUpperCase')
      `Failure Code: ${charge.failure_code.toUpperCase()}`,
      // BUG: failure_message is null on succeeded/pending charges.
      `Failure Message: ${charge.failure_message.trim()}`,
      // BUG: balance_transaction is null on pending/failed charges.
      `Transaction: ${charge.balance_transaction.replace("txn_", "TXN-")}`,
    ];
    return lines.join("\n");
  }
}
