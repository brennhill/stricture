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

  // BUG: Accesses nullable fields without null checks
  formatMessageSummary(message: TwilioMessage): string {
    // BUG: message.price is null when status is "queued", "sending", or "sent"
    // This crashes with: TypeError: Cannot read properties of null (reading 'toString')
    const priceDisplay = message.price.toString();

    // BUG: message.date_sent is null until the message is actually sent
    // This crashes with: TypeError: Cannot read properties of null (reading 'substring')
    const sentDate = message.date_sent.substring(0, 16);

    // BUG: message.error_message is null when the message hasn't failed
    // This crashes with: TypeError: Cannot read properties of null (reading 'toUpperCase')
    const errorDisplay = message.error_message.toUpperCase();

    return `SMS ${message.sid}: sent ${sentDate}, cost ${priceDisplay}, error: ${errorDisplay}`;
  }

  // BUG: Same issue for calls — nullable fields accessed without checks
  formatCallSummary(call: TwilioCall): string {
    // BUG: call.duration is null while the call is in progress or queued
    const durationStr = call.duration.toString() + " seconds";

    // BUG: call.price is null until the call completes
    const priceStr = call.price.toString();

    // BUG: call.start_time is null when the call is queued
    const startStr = call.start_time.substring(0, 16);

    // BUG: call.end_time is null while the call is active
    const endStr = call.end_time.substring(0, 16);

    return `Call ${call.sid}: ${startStr} - ${endStr}, ${durationStr}, cost ${priceStr}`;
  }

  async sendSms(params: SendSmsParams): Promise<TwilioMessage> { /* identical to PERFECT */ }
  async getMessage(messageSid: string): Promise<TwilioMessage> { /* identical to PERFECT */ }
  async makeCall(params: MakeCallParams): Promise<TwilioCall> { /* identical to PERFECT */ }
  async getCall(callSid: string): Promise<TwilioCall> { /* identical to PERFECT */ }
}
