// email-service/src/sendgrid-client-b02.ts — Wrong success status code.

export class SendGridClientB02 {
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

    // BUG: SendGrid returns 202 Accepted on success, not 200 OK.
    // This condition is never true for successful sends.
    if (response.status === 200) {
      return;
    }

    throw new Error(`SendGrid error: ${response.status}`);
  }
}
