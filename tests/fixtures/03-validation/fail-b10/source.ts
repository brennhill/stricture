// BUG: SID validation only checks length, not prefix pattern
function validateMessageSid(sid: string): boolean {
  return sid.length === 34;  // BUG: "XX" + 32 hex chars would pass
}

function validateCallSid(sid: string): boolean {
  return sid.length === 34;  // BUG: "SM" + 32 hex chars would pass (wrong resource type)
}

function validateAccountSid(sid: string): boolean {
  return sid.length === 34;  // BUG: "SM" + 32 hex chars would pass (wrong resource type)
}

class TwilioClient {
  private readonly baseUrl: string;
  private readonly accountSid: string;
  private readonly authHeader: string;

  constructor(accountSid: string, authToken: string) {
    // BUG: validateAccountSid only checks length — "SM" + 32 hex chars passes
    // as an Account SID, so a Message SID could be used as auth credentials
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
    formData.append("To", params.to);
    formData.append("From", params.from);
    formData.append("Body", params.body);

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
      // BUG: validateMessageSid only checks length — a response with
      // sid "CA..." (call SID) or "XX..." (garbage) would pass
      if (!validateMessageSid(message.sid)) throw new Error(`Bad SID: ${message.sid}`);
      return message;
    }
    if (response.status === 401) throw new Error("Authentication failed");
    if (response.status === 429) throw new Error("Rate limited");
    let errorBody: TwilioError;
    try { errorBody = await response.json(); } catch { throw new Error(`HTTP ${response.status}`); }
    throw new Error(`Twilio error ${errorBody.code}: ${errorBody.message}`);
  }

  async getMessage(messageSid: string): Promise<TwilioMessage> {
    // BUG: accepts Call SIDs, Account SIDs, or any 34-char string
    if (!validateMessageSid(messageSid)) throw new Error(`Invalid Message SID format`);

    let response: Response;
    try {
      response = await fetch(
        `${this.baseUrl}/Accounts/${this.accountSid}/Messages/${messageSid}.json`,
        { method: "GET", headers: { "Authorization": this.authHeader } }
      );
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      throw new Error(`Network error: ${msg}`);
    }

    if (!response.ok) {
      if (response.status === 404) throw new Error(`Message not found`);
      if (response.status === 401) throw new Error("Authentication failed");
      let errorBody: TwilioError;
      try { errorBody = await response.json(); } catch { throw new Error(`HTTP ${response.status}`); }
      throw new Error(`Twilio error ${errorBody.code}: ${errorBody.message}`);
    }

    return response.json();
  }

  async getCall(callSid: string): Promise<TwilioCall> {
    // BUG: accepts Message SIDs or any 34-char string as a Call SID
    if (!validateCallSid(callSid)) throw new Error(`Invalid Call SID format`);

    let response: Response;
    try {
      response = await fetch(
        `${this.baseUrl}/Accounts/${this.accountSid}/Calls/${callSid}.json`,
        { method: "GET", headers: { "Authorization": this.authHeader } }
      );
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      throw new Error(`Network error: ${msg}`);
    }

    if (!response.ok) {
      if (response.status === 404) throw new Error(`Call not found`);
      if (response.status === 401) throw new Error("Authentication failed");
      let errorBody: TwilioError;
      try { errorBody = await response.json(); } catch { throw new Error(`HTTP ${response.status}`); }
      throw new Error(`Twilio error ${errorBody.code}: ${errorBody.message}`);
    }

    return response.json();
  }

  async makeCall(params: MakeCallParams): Promise<TwilioCall> { /* identical to PERFECT */ }
}
