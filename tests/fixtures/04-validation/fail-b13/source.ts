// email-service/src/sendgrid-client-b13.ts — No API key format validation.

export class SendGridClientB13 {
  private readonly baseUrl = "https://api.sendgrid.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    // BUG: No format validation on API key.
    // Manifest declares: Authorization header format: "Bearer SG\\..*"
    // The constructor should verify the key matches the SG.* pattern.
    // Without this check:
    // 1. A Stripe key ("sk_live_...") is silently accepted
    // 2. An empty string is accepted
    // 3. A key with trailing whitespace is accepted
    // 4. The error only surfaces on the first API call as a 401
    this.apiKey = apiKey;
  }

  async sendEmail(request: SendGridSendRequest): Promise<void> {
    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v3/mail/send`, {
        method: "POST",
        headers: {
          // BUG: If apiKey is "sk_live_stripe_key", this sends:
          // Authorization: Bearer sk_live_stripe_key
          // SendGrid returns 401 with a cryptic error.
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

    if (response.status === 401) {
      // The error message is correct but the root cause is unclear.
      // Was the key revoked? Was the wrong key used? Was it malformed?
      throw new Error("SendGrid authentication failed");
    }

    throw new Error(`SendGrid error: ${response.status}`);
  }
}
