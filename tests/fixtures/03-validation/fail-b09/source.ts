// BUG: E.164 "validation" only checks for leading + sign
// Does not enforce the 1-15 digit range or reject leading zero after +
function validateE164(phone: string): boolean {
  return phone.startsWith("+");  // BUG: "+", "+0123", "+1234567890123456789" all pass
}

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
    // BUG: validateE164 is too permissive — accepts "+", "+0", "+abc",
    // "+1234567890123456789" (>15 digits), etc.
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

  async makeCall(params: MakeCallParams): Promise<TwilioCall> {
    // BUG: same weak validateE164 used here
    if (!validateE164(params.to)) throw new Error(`Invalid 'To' phone number`);
    if (!validateE164(params.from)) throw new Error(`Invalid 'From' phone number`);
    if (!validateHttpsUrl(params.url)) throw new Error(`TwiML URL must be HTTPS`);
    if (params.timeout !== undefined && (params.timeout < 5 || params.timeout > 600)) {
      throw new Error(`Timeout must be between 5 and 600`);
    }

    const formData = new URLSearchParams();
    formData.append("To", params.to);
    formData.append("From", params.from);
    formData.append("Url", params.url);

    let response: Response;
    try {
      response = await fetch(
        `${this.baseUrl}/Accounts/${this.accountSid}/Calls.json`,
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

    if (response.status === 201) return response.json();
    if (response.status === 401) throw new Error("Authentication failed");
    if (response.status === 429) throw new Error("Rate limited");
    let errorBody: TwilioError;
    try { errorBody = await response.json(); } catch { throw new Error(`HTTP ${response.status}`); }
    throw new Error(`Twilio error ${errorBody.code}: ${errorBody.message}`);
  }

  async getMessage(messageSid: string): Promise<TwilioMessage> { /* identical to PERFECT */ }
  async getCall(callSid: string): Promise<TwilioCall> { /* identical to PERFECT */ }
}
