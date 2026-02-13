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

  // BUG: Fetches only the first page and returns it as the complete list
  async listAllMessages(filters?: {
    to?: string;
    from?: string;
    dateSent?: string;
    pageSize?: number;
  }): Promise<TwilioMessage[]> {
    const pageSize = filters?.pageSize ?? 100;
    if (pageSize < 1 || pageSize > 1000) {
      throw new Error("PageSize must be between 1 and 1000");
    }

    let url = `${this.baseUrl}/Accounts/${this.accountSid}/Messages.json?PageSize=${pageSize}`;
    if (filters?.to) url += `&To=${encodeURIComponent(filters.to)}`;
    if (filters?.from) url += `&From=${encodeURIComponent(filters.from)}`;
    if (filters?.dateSent) url += `&DateSent=${encodeURIComponent(filters.dateSent)}`;

    let response: Response;
    try {
      response = await fetch(url, {
        method: "GET",
        headers: { "Authorization": this.authHeader },
      });
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      throw new Error(`Network error listing messages: ${msg}`);
    }

    if (!response.ok) {
      if (response.status === 401) throw new Error("Authentication failed");
      let errorBody: TwilioError;
      try { errorBody = await response.json(); } catch { throw new Error(`HTTP ${response.status}`); }
      throw new Error(`Twilio error ${errorBody.code}: ${errorBody.message}`);
    }

    const page: TwilioMessageList = await response.json();

    // BUG: Returns only page.messages from the first page.
    // page.next_page_uri is completely ignored.
    // If the account has 500 messages and pageSize is 100,
    // this returns only the first 100 and silently discards 400.
    return page.messages;

    // The PERFECT implementation has:
    //   while (nextPageUri !== null) { ... nextPageUri = page.next_page_uri; }
    // This version fetches one page and stops.
  }

  async sendSms(params: SendSmsParams): Promise<TwilioMessage> { /* identical to PERFECT */ }
  async getMessage(messageSid: string): Promise<TwilioMessage> { /* identical to PERFECT */ }
  async makeCall(params: MakeCallParams): Promise<TwilioCall> { /* identical to PERFECT */ }
  async getCall(callSid: string): Promise<TwilioCall> { /* identical to PERFECT */ }
}
