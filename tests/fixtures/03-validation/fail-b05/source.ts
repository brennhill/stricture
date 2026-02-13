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
    if (!validateE164(params.to)) throw new Error(`Invalid 'To' phone number`);
    if (!validateE164(params.from)) throw new Error(`Invalid 'From' phone number`);
    if (params.body.length === 0 || params.body.length > 1600) {
      throw new Error(`Message body must be between 1 and 1600 characters`);
    }
    if (params.statusCallback !== undefined && !validateHttpsUrl(params.statusCallback)) {
      throw new Error(`StatusCallback must be a valid HTTPS URL`);
    }

    const formData = new URLSearchParams();
    // BUG: "To" field is never appended to the form data.
    // The params.to value is validated above but never included in the request body.
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
      const message = err instanceof Error ? err.message : String(err);
      throw new Error(`Network error sending SMS: ${message}`);
    }

    if (response.status === 201) {
      const message: TwilioMessage = await response.json();
      if (!validateMessageSid(message.sid)) {
        throw new Error(`Unexpected message SID format: ${message.sid}`);
      }
      return message;
    }

    if (response.status === 401) throw new Error("Authentication failed");
    if (response.status === 429) throw new Error("Rate limited by Twilio");

    let errorBody: TwilioError;
    try {
      errorBody = await response.json();
    } catch {
      throw new Error(`Twilio returned HTTP ${response.status} with unparseable body`);
    }
    throw new Error(`Twilio error ${errorBody.code}: ${errorBody.message}`);
  }

  // getMessage and makeCall are identical to PERFECT (omitted for brevity)
  async getMessage(messageSid: string): Promise<TwilioMessage> { /* identical to PERFECT */ }
  async makeCall(params: MakeCallParams): Promise<TwilioCall> { /* identical to PERFECT */ }
}
