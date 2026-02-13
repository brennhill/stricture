// email-service/src/sendgrid-client-b01.ts — Missing error handling on fetch.

export class SendGridClientB01 {
  private readonly baseUrl = "https://api.sendgrid.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async sendEmail(request: SendGridSendRequest): Promise<void> {
    // BUG: No try/catch — fetch() can throw on network errors, DNS failures,
    // timeouts, or AbortController signals. These propagate as unhandled
    // rejections and crash the calling service.
    const response = await fetch(`${this.baseUrl}/v3/mail/send`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${this.apiKey}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });

    if (response.status === 202) {
      return;
    }

    throw new Error(`SendGrid error: ${response.status}`);
  }
}
