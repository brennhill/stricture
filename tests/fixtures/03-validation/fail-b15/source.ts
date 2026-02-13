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

  // BUG: TOCTOU race — reads status, then acts on it, but the status
  // can change between the two operations.
  async resendIfFailed(messageSid: string, from: string): Promise<TwilioMessage | null> {
    // Step 1: Read the current status
    const message = await this.getMessage(messageSid);

    // BUG: Between this read and the action below, the message status could change:
    //
    // Timeline:
    //   T0: getMessage() returns status="sending" (not failed)
    //   T1: Twilio delivers the message, status becomes "delivered"
    //   T2: We check `message.status` — it says "sending" (stale!)
    //   T3: We decide not to resend (correct, but for the wrong reason)
    //
    // Or worse:
    //   T0: getMessage() returns status="queued"
    //   T1: Twilio fails delivery, status becomes "failed"
    //   T2: We check `message.status` — it says "queued" (stale!)
    //   T3: We skip the resend — user never gets the message
    //
    // Or the double-send:
    //   T0: getMessage() returns status="failed"
    //   T1: Twilio retries internally and delivers the message
    //   T2: We check `message.status` — it says "failed" (stale!)
    //   T3: We resend — user gets the message TWICE

    if (message.status === "failed" || message.status === "undelivered") {
      // Step 2: Act on the stale status — resend to the same number
      // By this point, Twilio may have already retried internally,
      // or the carrier may have delivered it late, or another resend
      // process may have already sent a duplicate.
      const resent = await this.sendSms({
        to: message.to,
        from: from,
        body: message.body,
      });
      return resent;
    }

    return null;
  }

  // BUG: Same TOCTOU race in batch status check
  async getFailedMessageSids(sids: string[]): Promise<string[]> {
    const failed: string[] = [];

    // BUG: Sequential reads mean the first message's status could change
    // by the time we read the last message's status. The "snapshot" of
    // statuses is inconsistent across the batch.
    for (const sid of sids) {
      const message = await this.getMessage(sid);
      if (message.status === "failed" || message.status === "undelivered") {
        // BUG: This status was true at read time but may not be true now.
        // Twilio may have retried internally between reads.
        failed.push(sid);
      }
    }

    return failed;
  }

  // BUG: Same race in call monitoring — status transitions are non-reversible
  // but reading and acting are separated
  async cancelIfStillQueued(callSid: string): Promise<boolean> {
    const call = await this.getCall(callSid);

    // BUG: Between getCall() and the cancel request below, the call could
    // transition from "queued" to "ringing" or "in-progress".
    // Canceling a ringing or in-progress call has different semantics
    // than canceling a queued call, and may fail or cause unexpected behavior.
    if (call.status === "queued") {
      // Attempt to cancel by updating the call status
      const formData = new URLSearchParams();
      formData.append("Status", "canceled");

      let response: Response;
      try {
        response = await fetch(
          `${this.baseUrl}/Accounts/${this.accountSid}/Calls/${callSid}.json`,
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
        throw new Error(`Network error canceling call: ${msg}`);
      }

      if (response.status === 200) return true;
      if (response.status === 401) throw new Error("Authentication failed");

      // BUG: If the call transitioned to "ringing" between our read and this
      // POST, Twilio may return 400 or 409. We don't handle that case.
      let errorBody: TwilioError;
      try { errorBody = await response.json(); } catch { throw new Error(`HTTP ${response.status}`); }
      throw new Error(`Twilio error ${errorBody.code}: ${errorBody.message}`);
    }

    return false;
  }

  async sendSms(params: SendSmsParams): Promise<TwilioMessage> { /* identical to PERFECT */ }
  async getMessage(messageSid: string): Promise<TwilioMessage> { /* identical to PERFECT */ }
  async makeCall(params: MakeCallParams): Promise<TwilioCall> { /* identical to PERFECT */ }
  async getCall(callSid: string): Promise<TwilioCall> { /* identical to PERFECT */ }
}
