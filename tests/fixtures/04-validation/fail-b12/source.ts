// email-service/src/sendgrid-client-b12.ts — Nullable field access crash.

export class SendGridClientB12 {
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

    if (response.status === 400 || response.status === 500) {
      const body: SendGridErrorResponse = await response.json();
      this.formatAndThrowErrors(body.errors);
    }

    throw new Error(`SendGrid error: ${response.status}`);
  }

  private formatAndThrowErrors(errors: SendGridErrorDetail[]): never {
    const formatted = errors.map((err) => {
      // BUG: err.field can be null according to the manifest:
      //   field: { type: string, nullable: true, required: false }
      // When SendGrid returns an error like "Rate limit exceeded" or
      // "Authorization required", the field property is null.
      // Calling .toLowerCase() on null throws TypeError.
      const fieldName = err.field.toLowerCase(); // CRASH when field is null

      // BUG: Same issue with err.help — also nullable.
      const helpLink = err.help.trim(); // CRASH when help is null

      return `[${fieldName}] ${err.message} (see: ${helpLink})`;
    });

    throw new Error(`SendGrid validation errors:\n${formatted.join("\n")}`);
  }
}
