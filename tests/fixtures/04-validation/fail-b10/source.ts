// email-service/src/sendgrid-client-b10.ts — No email format validation.

export class SendGridClientB10 {
  private readonly baseUrl = "https://api.sendgrid.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async sendEmail(request: SendGridSendRequest): Promise<void> {
    // BUG: No email format validation on from.email or personalizations[].to[].email.
    // Manifest declares email fields as: { type: string, format: email_rfc5322 }.
    // SendGrid validates server-side and returns 400, but:
    // 1. The error message from SendGrid may be cryptic
    // 2. We waste an API call and rate limit quota
    // 3. We cannot provide a user-friendly error message
    // 4. Batch sends fail entirely if one email is invalid

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

    throw new Error(`SendGrid error: ${response.status}`);
  }

  async sendWelcomeEmail(userEmail: string, userName: string): Promise<void> {
    // BUG: userEmail is used directly without validation.
    // If userEmail is "John Doe" (a name, not an email), this sends
    // a malformed request to SendGrid.
    await this.sendEmail({
      personalizations: [{ to: [{ email: userEmail, name: userName }] }],
      from: { email: "welcome@myapp.com" },
      subject: `Welcome, ${userName}!`,
      content: [{ type: "text/html", value: `<h1>Welcome, ${userName}!</h1>` }],
    });
  }
}
