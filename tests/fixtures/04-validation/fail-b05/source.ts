// email-service/src/sendgrid-client-b05.ts — Missing required 'subject' field.

interface SendGridSendRequestB05 {
  personalizations: Array<{
    to: Array<{ email: string; name?: string }>;
  }>;
  from: { email: string; name?: string };
  // BUG: 'subject' field is missing from the interface entirely.
  // The manifest declares subject as required: true.
  content: Array<{ type: string; value: string }>;
}

export class SendGridClientB05 {
  private readonly baseUrl = "https://api.sendgrid.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async sendEmail(request: SendGridSendRequestB05): Promise<void> {
    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v3/mail/send`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "application/json",
        },
        // BUG: Serialized body never contains 'subject' because the
        // interface doesn't have it. SendGrid returns 400.
        body: JSON.stringify(request),
      });
    } catch (err) {
      throw new Error(`SendGrid network error: ${(err as Error).message}`);
    }

    if (response.status === 202) {
      return;
    }

    throw new Error(`SendGrid error: ${response.status}`);
  }
}
