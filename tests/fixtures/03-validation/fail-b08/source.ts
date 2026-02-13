class TwilioClient {
  private readonly baseUrl: string;
  private readonly accountSid: string;
  private readonly authHeader: string;

  constructor(accountSid: string, authToken: string) {
    if (!validateAccountSid(accountSid)) throw new Error(`Invalid Account SID format`);
    if (!validateAuthToken(authToken)) throw new Error(`Invalid Auth Token format`);
    this.baseUrl = "https://api.twilio.com/2010-04-01";
    this.accountSid = accountSid;
    this.authHeader =
      "Basic " + Buffer.from(`${accountSid}:${authToken}`).toString("base64");
  }

  async sendSms(params: SendSmsParams): Promise<TwilioMessage> {
    // ... identical to PERFECT (full validation, error handling, etc.)
    if (!validateE164(params.to)) throw new Error(`Invalid 'To' phone number`);
    if (!validateE164(params.from)) throw new Error(`Invalid 'From' phone number`);
    if (params.body.length === 0 || params.body.length > 1600) {
      throw new Error(`Message body must be between 1 and 1600 characters`);
    }

    const formData = new URLSearchParams();
    formData.append("To", params.to);
    formData.append("From", params.from);
    formData.append("Body", params.body);
    if (params.statusCallback !== undefined) {
      formData.append("StatusCallback", params.statusCallback);
    }

    let response: Response;
    try {
      response = await fetch(
        `${this.baseUrl}/Accounts/${this.accountSid}/Messages.json`,
        {
          method: "POST",
          headers: {
            "Authorization": this.authHeader,
            "Content-Type": "application/x-www-form-urlencoded",
          },
          body: formData.toString(),
        }
      );
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      throw new Error(`Network error: ${msg}`);
    }

    if (response.status === 201) {
      const message: TwilioMessage = await response.json();
      if (!validateMessageSid(message.sid)) throw new Error(`Bad SID: ${message.sid}`);
      return message;
    }
    if (response.status === 401) throw new Error("Authentication failed");
    if (response.status === 429) throw new Error("Rate limited");
    let errorBody: TwilioError;
    try { errorBody = await response.json(); } catch { throw new Error(`HTTP ${response.status}`); }
    throw new Error(`Twilio error ${errorBody.code}: ${errorBody.message}`);
  }

  // BUG: Only handles 2 of 8 possible message statuses
  processMessageStatus(message: TwilioMessage): string {
    switch (message.status) {
      case "delivered":
        return "Message delivered successfully";
      case "failed":
        return `Message failed: ${message.error_message ?? "unknown error"}`;
      // BUG: Missing cases:
      // - "queued"       — message accepted, waiting to send
      // - "sending"      — message being transmitted
      // - "sent"         — message sent to carrier (not yet delivered)
      // - "undelivered"  — carrier could not deliver (different from "failed")
      // - "receiving"    — inbound message being processed
      // - "received"     — inbound message delivered to our account
      default:
        // Falls through silently — no logging, no error, just returns empty string
        return "";
    }
  }

  // BUG: Only handles 2 of 8 possible call statuses
  processCallStatus(call: TwilioCall): string {
    switch (call.status) {
      case "completed":
        return `Call completed, duration: ${call.duration ?? "unknown"}s`;
      case "failed":
        return "Call failed";
      // BUG: Missing cases:
      // - "queued"       — call waiting to be placed
      // - "ringing"      — phone is ringing
      // - "in-progress"  — call is active
      // - "busy"         — callee's line is busy
      // - "no-answer"    — callee did not pick up
      // - "canceled"     — call was canceled before connection
      default:
        return "";
    }
  }

  async getMessage(messageSid: string): Promise<TwilioMessage> { /* identical to PERFECT */ }
  async makeCall(params: MakeCallParams): Promise<TwilioCall> { /* identical to PERFECT */ }
  async getCall(callSid: string): Promise<TwilioCall> { /* identical to PERFECT */ }
}
