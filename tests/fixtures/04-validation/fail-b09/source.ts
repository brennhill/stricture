// email-service/src/sendgrid-client-b09.ts — No personalizations limit check.

export class SendGridClientB09 {
  private readonly baseUrl = "https://api.sendgrid.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async sendEmail(request: SendGridSendRequest): Promise<void> {
    // BUG: No validation on personalizations array size.
    // Manifest declares: personalizations.maxItems = 1000.
    // A caller can pass 50,000 personalizations. This:
    // 1. Creates a massive JSON payload (may trigger 413)
    // 2. Gets rejected by SendGrid with a 400
    // 3. Wastes network bandwidth and API rate limit quota
    // 4. Could cause OOM if the serialized body is very large

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

  async sendBulkEmails(
    recipients: Array<{ email: string; name?: string }>,
    from: { email: string },
    subject: string,
    htmlContent: string,
  ): Promise<void> {
    // BUG: Builds a single request with all recipients.
    // Should batch into groups of 1000.
    const request: SendGridSendRequest = {
      personalizations: recipients.map((r) => ({ to: [r] })),
      from,
      subject,
      content: [{ type: "text/html", value: htmlContent }],
    };

    await this.sendEmail(request);
  }
}
