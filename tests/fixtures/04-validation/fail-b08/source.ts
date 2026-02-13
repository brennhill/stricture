// email-service/src/sendgrid-client-b08.ts — Missing 413 and 429 handling.

export class SendGridClientB08 {
  private readonly baseUrl = "https://api.sendgrid.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async sendEmail(request: SendGridSendRequest): Promise<void> {
    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v3/mail/send`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(request),
      });
    } catch (err) {
      throw new Error(`SendGrid network error: ${(err as Error).message}`);
    }

    if (response.status === 202) {
      return;
    }

    // BUG: Only handles 400 and 401. Manifest declares status codes:
    // [202, 400, 401, 403, 413, 429, 500].
    // Missing: 403, 413, 429, 500 — each requires different remediation.
    switch (response.status) {
      case 400: {
        const body = await response.json();
        throw new Error(`Bad request: ${body.errors[0].message}`);
      }
      case 401:
        throw new Error("Unauthorized: check API key");
      default:
        // 413 and 429 fall through to this generic handler.
        // 413: caller does not know to split the batch.
        // 429: caller does not know to wait and retry with backoff.
        throw new Error(`SendGrid error: ${response.status}`);
    }
  }
}
