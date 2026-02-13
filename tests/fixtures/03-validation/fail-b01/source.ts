class TwilioClient {
  private readonly baseUrl: string;
  private readonly accountSid: string;
  private readonly authHeader: string;

  constructor(accountSid: string, authToken: string) {
    if (!validateAccountSid(accountSid)) {
      throw new Error(`Invalid Account SID format`);
    }
    if (!validateAuthToken(authToken)) {
      throw new Error(`Invalid Auth Token format`);
    }
    this.baseUrl = "https://api.twilio.com/2010-04-01";
    this.accountSid = accountSid;
    this.authHeader =
      "Basic " + Buffer.from(`${accountSid}:${authToken}`).toString("base64");
  }

  // BUG: No try/catch around fetch — network errors propagate as unhandled
  async sendSms(params: SendSmsParams): Promise<TwilioMessage> {
    if (!validateE164(params.to)) {
      throw new Error(`Invalid 'To' phone number`);
    }
    if (!validateE164(params.from)) {
      throw new Error(`Invalid 'From' phone number`);
    }
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

    // BUG: bare fetch — ECONNREFUSED, DNS failure, timeout all become
    // unhandled promise rejections with no context for the caller
    const response = await fetch(
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

    if (response.status === 201) {
      const message: TwilioMessage = await response.json();
      if (!validateMessageSid(message.sid)) {
        throw new Error(`Unexpected message SID format: ${message.sid}`);
      }
      return message;
    }

    if (response.status === 401) {
      throw new Error("Authentication failed");
    }
    if (response.status === 429) {
      throw new Error("Rate limited by Twilio");
    }

    const errorBody: TwilioError = await response.json();
    throw new Error(`Twilio error ${errorBody.code}: ${errorBody.message}`);
  }

  // BUG: same issue — bare fetch with no catch
  async getMessage(messageSid: string): Promise<TwilioMessage> {
    if (!validateMessageSid(messageSid)) {
      throw new Error(`Invalid Message SID format`);
    }

    const response = await fetch(
      `${this.baseUrl}/Accounts/${this.accountSid}/Messages/${messageSid}.json`,
      { method: "GET", headers: { "Authorization": this.authHeader } }
    );

    if (!response.ok) {
      if (response.status === 404) throw new Error(`Message not found`);
      if (response.status === 401) throw new Error("Authentication failed");
      const errorBody: TwilioError = await response.json();
      throw new Error(`Twilio error ${errorBody.code}: ${errorBody.message}`);
    }

    return response.json();
  }

  async makeCall(params: MakeCallParams): Promise<TwilioCall> {
    if (!validateE164(params.to)) throw new Error(`Invalid 'To' phone number`);
    if (!validateE164(params.from)) throw new Error(`Invalid 'From' phone number`);
    if (!validateHttpsUrl(params.url)) throw new Error(`TwiML URL must be HTTPS`);

    const formData = new URLSearchParams();
    formData.append("To", params.to);
    formData.append("From", params.from);
    formData.append("Url", params.url);

    const response = await fetch(
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

    if (response.status === 201) {
      const call: TwilioCall = await response.json();
      return call;
    }

    const errorBody: TwilioError = await response.json();
    throw new Error(`Twilio error ${errorBody.code}: ${errorBody.message}`);
  }
}
