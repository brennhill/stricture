// email-service/src/sendgrid-client-b15.ts — Race condition: check then send.

export class SendGridClientB15 {
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
      return await response.json();
    }

    throw new Error(`SendGrid error: ${response.status}`);
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

    throw new Error(`SendGrid error: ${response.status}`);
  }

  async sendEmailIfNotBounced(
    recipientEmail: string,
    request: SendGridSendRequest,
  ): Promise<{ sent: boolean; reason?: string }> {
    // BUG: TOCTOU race condition (Time Of Check vs Time Of Use).
    //
    // Step 1: Check suppression list (takes ~200ms network round trip)
    const bounces = await this.getBounces();
    const isBounced = bounces.some((b) => b.email === recipientEmail);

    // RACE WINDOW: Between step 1 and step 2, another service or webhook
    // could add this email to the suppression list. Scenarios:
    //   - A simultaneous email send bounced and the bounce webhook fired
    //   - An admin manually added the email to suppressions
    //   - A spam complaint was processed
    //   - Another instance of this service processed a bounce

    if (isBounced) {
      return { sent: false, reason: "Email is on bounce suppression list" };
    }

    // Step 2: Send email (~200ms later — the suppression state may have changed)
    // BUG: The correct approach is to:
    // 1. Send the email optimistically (SendGrid handles suppression server-side)
    // 2. Check the Activity API afterward to confirm delivery status
    // 3. Or use SendGrid's server-side suppression which atomically checks
    //    suppression status at send time
    await this.sendEmail(request);

    return { sent: true };
  }

  async sendBulkIfNotBounced(
    recipients: Array<{ email: string; name?: string }>,
    from: { email: string },
    subject: string,
    content: string,
  ): Promise<{ sent: string[]; skipped: string[] }> {
    // BUG: Fetches bounces once, then iterates and sends sequentially.
    // By the time the last recipient is processed (could be minutes later
    // for a large list), the bounce list is stale.
    const bounces = await this.getBounces();
    const bouncedEmails = new Set(bounces.map((b) => b.email));

    const sent: string[] = [];
    const skipped: string[] = [];

    for (const recipient of recipients) {
      if (bouncedEmails.has(recipient.email)) {
        skipped.push(recipient.email);
        continue;
      }

      await this.sendEmail({
        personalizations: [{ to: [recipient] }],
        from,
        subject,
        content: [{ type: "text/html", value: content }],
      });
      sent.push(recipient.email);
    }

    return { sent, skipped };
  }
}
