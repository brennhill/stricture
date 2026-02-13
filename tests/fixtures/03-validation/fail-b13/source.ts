class TwilioClient {
  private readonly baseUrl: string;
  private readonly accountSid: string;
  private readonly authHeader: string;

  // BUG: No validation of Account SID or Auth Token format in constructor
  constructor(accountSid: string, authToken: string) {
    // BUG: No check that accountSid matches /^AC[a-f0-9]{32}$/
    // Empty string, "hello", "SM...", "12345" all accepted
    this.baseUrl = "https://api.twilio.com/2010-04-01";
    this.accountSid = accountSid;
    // BUG: No check that authToken matches /^[a-f0-9]{32}$/
    // Empty string creates "Basic <base64 of ':'>", which is valid HTTP
    // but always returns 401 from Twilio
    this.authHeader =
      "Basic " + Buffer.from(`${accountSid}:${authToken}`).toString("base64");
  }

  async sendSms(params: SendSmsParams): Promise<TwilioMessage> {
    if (!validateE164(params.to)) throw new Error(`Invalid 'To' phone number`);
    if (!validateE164(params.from)) throw new Error(`Invalid 'From' phone number`);
    if (params.body.length === 0 || params.body.length > 1600) {
      throw new Error(`Message body must be between 1 and 1600 characters`);
    }
    if (params.statusCallback !== undefined && !validateHttpsUrl(params.statusCallback)) {
      throw new Error(`StatusCallback must be a valid HTTPS URL`);
    }

    const formData = new URLSearchParams();
    formData.append("To", params.to);
    formData.append("From", params.from);
    formData.append("Body", params.body);

    let response: Response;
    try {
      // BUG: URL includes unvalidated accountSid — if it contains special
      // characters (e.g., "../"), this could construct a malformed URL
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

  async getMessage(messageSid: string): Promise<TwilioMessage> { /* identical to PERFECT */ }
  async makeCall(params: MakeCallParams): Promise<TwilioCall> { /* identical to PERFECT */ }
  async getCall(callSid: string): Promise<TwilioCall> { /* identical to PERFECT */ }
}
