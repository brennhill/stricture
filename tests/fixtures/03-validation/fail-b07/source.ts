interface TwilioMessage {
  sid: string;
  account_sid: string;
  to: string;
  from: string;
  body: string;
  status: MessageStatus;
  date_created: string;
  date_updated: string;
  date_sent: string | null;
  price: number | null;       // BUG: Twilio returns string (e.g., "-0.0075"), not number
  price_unit: string;
  error_code: number | null;
  error_message: string | null;
  num_segments: number;        // BUG: Twilio returns string (e.g., "1"), not number
  direction: "inbound" | "outbound-api" | "outbound-call" | "outbound-reply";
  uri: string;
}

interface TwilioCall {
  sid: string;
  account_sid: string;
  to: string;
  from: string;
  status: CallStatus;
  start_time: string | null;
  end_time: string | null;
  duration: number | null;     // BUG: Twilio returns string (e.g., "45"), not number
  price: number | null;        // BUG: same as above
  price_unit: string;
  direction: "inbound" | "outbound-api" | "outbound-dial";
  uri: string;
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
    if (!validateE164(params.to)) throw new Error(`Invalid 'To' phone number`);
    if (!validateE164(params.from)) throw new Error(`Invalid 'From' phone number`);
    if (params.body.length === 0 || params.body.length > 1600) {
      throw new Error(`Message body must be between 1 and 1600 characters`);
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
      throw new Error(`Network error sending SMS: ${msg}`);
    }

    if (response.status === 201) {
      const message: TwilioMessage = await response.json();
      if (!validateMessageSid(message.sid)) {
        throw new Error(`Unexpected message SID format: ${message.sid}`);
      }
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
}
