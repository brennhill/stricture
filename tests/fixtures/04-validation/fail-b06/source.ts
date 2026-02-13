// email-service/src/sendgrid-client-b06.ts — Tries to parse empty 202 body.

interface SendGridSendResponseB06 {
  // BUG: This type does not exist — 202 Accepted has no body.
  // The manifest declares: success.body = null
  message_id: string;
  status: string;
}

export class SendGridClientB06 {
  private readonly baseUrl = "https://api.sendgrid.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async sendEmail(request: SendGridSendRequest): Promise<SendGridSendResponseB06> {
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
      // BUG: 202 Accepted returns an empty body. response.json() will throw:
      // SyntaxError: Unexpected end of JSON input
      const data: SendGridSendResponseB06 = await response.json();
      return data;
    }

    throw new Error(`SendGrid error: ${response.status}`);
  }
}
