// email-service/src/sendgrid-client-b07.ts — Wrong type for bounce.created field.

interface SendGridBounceB07 {
  email: string;
  // BUG: Manifest declares created as { type: integer, format: unix_timestamp }.
  // Storing as string means numeric comparisons fail:
  //   "1700000000" > "200000000" is FALSE (string comparison)
  //   1700000000 > 200000000 is TRUE (numeric comparison)
  created: string;
  reason: string;
  status: string;
}

export class SendGridClientB07 {
  private readonly baseUrl = "https://api.sendgrid.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async getBounces(): Promise<SendGridBounceB07[]> {
    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v3/suppression/bounces`, {
        method: "GET",
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "application/json",
        },
      });
    } catch (err) {
      throw new Error(`SendGrid network error: ${(err as Error).message}`);
    }

    if (response.status === 200) {
      const bounces: SendGridBounceB07[] = await response.json();
      return bounces;
    }

    throw new Error(`SendGrid error: ${response.status}`);
  }

  async getBouncesAfter(timestamp: number): Promise<SendGridBounceB07[]> {
    const bounces = await this.getBounces();
    // BUG: String comparison instead of numeric comparison.
    // "9" > "1700000000" is TRUE in string comparison because "9" > "1".
    // This filter returns wrong results.
    return bounces.filter((b) => b.created > String(timestamp));
  }
}
