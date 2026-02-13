class TwilioClient {
  private readonly baseUrl: string;
  private readonly accountSid: string;
  private readonly authHeader: string;

  constructor(accountSid: string, authToken: string) {
    if (!validateAccountSid(accountSid)) throw new Error(`Invalid Account SID format`);
    if (!validateAuthToken(authToken)) throw new Error(`Invalid Auth Token format`);
    this.baseUrl = "https://api.twilio.com/2010-04-01";
    this.accountSid = accountSid;
    this.authHeader =
      "Basic " + Buffer.from(`${accountSid}:${authToken}`).toString("base64");
  }

  // BUG: Aggregates prices using parseFloat + floating-point addition
  // instead of integer arithmetic or BigInt
  async getTotalSpend(messageSids: string[]): Promise<number> {
    let total = 0;

    for (const sid of messageSids) {
      const message = await this.getMessage(sid);

      if (message.price !== null) {
        // BUG: parseFloat("-0.0075") works for single values but
        // floating-point addition accumulates errors:
        //   -0.0075 + -0.0075 + -0.0075 = -0.022499999999999996
        //   (expected: -0.0225)
        total += parseFloat(message.price);
      }
    }

    return total;
  }

  // BUG: Compares price against threshold using floating-point comparison
  isPriceAboveThreshold(message: TwilioMessage, thresholdCents: number): boolean {
    if (message.price === null) return false;

    // BUG: parseFloat("-0.0075") * 100 = -0.75 (correct here by luck)
    // but parseFloat("-0.0033") * 100 = -0.32999999999999996 (not -0.33)
    const priceCents = parseFloat(message.price) * 100;
    return Math.abs(priceCents) > thresholdCents;
  }

  // BUG: Formats aggregated price for display — accumulated errors show up
  formatTotalSpend(prices: Array<string | null>): string {
    let total = 0;
    for (const p of prices) {
      if (p !== null) {
        total += parseFloat(p);  // BUG: precision loss on each addition
      }
    }
    // toFixed(4) masks some errors but not all:
    // 10000 messages at -0.0075 each:
    //   expected: -75.0000
    //   actual:   -74.9999 (or -75.0001 depending on accumulation order)
    return total.toFixed(4);
  }

  async sendSms(params: SendSmsParams): Promise<TwilioMessage> { /* identical to PERFECT */ }
  async getMessage(messageSid: string): Promise<TwilioMessage> { /* identical to PERFECT */ }
  async makeCall(params: MakeCallParams): Promise<TwilioCall> { /* identical to PERFECT */ }
  async getCall(callSid: string): Promise<TwilioCall> { /* identical to PERFECT */ }
}
