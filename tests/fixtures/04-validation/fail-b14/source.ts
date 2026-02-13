// email-service/src/sendgrid-client-b14.ts — Pagination not followed.

export class SendGridClientB14 {
  private readonly baseUrl = "https://api.sendgrid.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async getBounces(): Promise<SendGridBounce[]> {
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
      // BUG: Only returns the first page of bounces.
      // SendGrid paginates bounces with a Link header:
      //   Link: <https://api.sendgrid.com/v3/suppression/bounces?offset=500>; rel="next"
      // Manifest declares: pagination.cursor = header
      // This response may contain only 500 of 5,000 total bounces.
      // The Link header is completely ignored.
      const bounces: SendGridBounce[] = await response.json();
      return bounces;
    }

    throw new Error(`SendGrid error: ${response.status}`);
  }

  async getBouncedEmailCount(): Promise<number> {
    // BUG: Returns count of first page only, not total bounces.
    // If there are 5,000 bounces and the page size is 500,
    // this returns 500 instead of 5,000.
    const bounces = await this.getBounces();
    return bounces.length;
  }

  async isEmailBounced(email: string): Promise<boolean> {
    // BUG: Only checks first page. If the bounce is on page 2+,
    // this returns false even though the email has bounced.
    const bounces = await this.getBounces();
    return bounces.some((b) => b.email === email);
  }
}
