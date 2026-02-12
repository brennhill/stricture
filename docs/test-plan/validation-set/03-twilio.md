# 03 - Twilio SMS/Voice API

**Why included:** Callback URLs, phone format validation (E.164), SID prefix patterns, status state machines, nullable pricing fields, paginated list endpoints.

---

## Manifest Fragment

```yaml
contracts:
  - id: "twilio-messaging"
    producer: twilio
    consumers: [my-service]
    protocol: http
    base_url: "https://api.twilio.com"
    auth:
      type: basic
      username_format: "AC[a-f0-9]{32}"
      password_format: "[a-f0-9]{32}"
    endpoints:
      - path: "/2010-04-01/Accounts/{AccountSid}/Messages.json"
        method: POST
        request:
          fields:
            To:              { type: string, format: "e164", pattern: "^\\+[1-9]\\d{1,14}$", required: true }
            From:            { type: string, format: "e164", pattern: "^\\+[1-9]\\d{1,14}$", required: true }
            Body:            { type: string, max_length: 1600, required: true }
            StatusCallback:  { type: string, format: "url_https", required: false }
            MaxPrice:        { type: string, format: "decimal", required: false }
        response:
          fields:
            sid:             { type: string, format: "SM[a-f0-9]{32}", required: true }
            account_sid:     { type: string, format: "AC[a-f0-9]{32}", required: true }
            to:              { type: string, format: "e164", required: true }
            from:            { type: string, format: "e164", required: true }
            body:            { type: string, required: true }
            status:          { type: enum, values: ["queued", "sending", "sent", "delivered", "undelivered", "failed", "receiving", "received"], required: true }
            date_created:    { type: string, format: "rfc2822", required: true }
            date_updated:    { type: string, format: "rfc2822", required: true }
            date_sent:       { type: string, format: "rfc2822", nullable: true }
            price:           { type: string, format: "decimal", nullable: true }
            price_unit:      { type: string, format: "iso4217", required: true }
            error_code:      { type: integer, range: [30001, 30999], nullable: true }
            error_message:   { type: string, nullable: true }
            num_segments:    { type: string, format: "integer_string", required: true }
            direction:       { type: enum, values: ["inbound", "outbound-api", "outbound-call", "outbound-reply"], required: true }
            uri:             { type: string, required: true }
        status_codes: [201, 400, 401, 404, 429, 500]

      - path: "/2010-04-01/Accounts/{AccountSid}/Messages/{MessageSid}.json"
        method: GET
        request:
          fields: {}
        response:
          fields:
            sid:             { type: string, format: "SM[a-f0-9]{32}", required: true }
            status:          { type: enum, values: ["queued", "sending", "sent", "delivered", "undelivered", "failed", "receiving", "received"], required: true }
            price:           { type: string, format: "decimal", nullable: true }
            price_unit:      { type: string, format: "iso4217", required: true }
            error_code:      { type: integer, range: [30001, 30999], nullable: true }
            error_message:   { type: string, nullable: true }
        status_codes: [200, 401, 404, 500]

      - path: "/2010-04-01/Accounts/{AccountSid}/Messages.json"
        method: GET
        request:
          fields:
            PageSize:        { type: integer, range: [1, 1000], required: false }
            DateSent:        { type: string, format: "iso8601_date", required: false }
            To:              { type: string, format: "e164", required: false }
            From:            { type: string, format: "e164", required: false }
        response:
          fields:
            messages:        { type: array, items: "Message", required: true }
            page:            { type: integer, required: true }
            page_size:       { type: integer, range: [1, 1000], required: true }
            first_page_uri:  { type: string, required: true }
            next_page_uri:   { type: string, nullable: true }
            previous_page_uri: { type: string, nullable: true }
            uri:             { type: string, required: true }
        status_codes: [200, 401, 500]

  - id: "twilio-voice"
    producer: twilio
    consumers: [my-service]
    protocol: http
    base_url: "https://api.twilio.com"
    auth:
      type: basic
      username_format: "AC[a-f0-9]{32}"
      password_format: "[a-f0-9]{32}"
    endpoints:
      - path: "/2010-04-01/Accounts/{AccountSid}/Calls.json"
        method: POST
        request:
          fields:
            To:              { type: string, format: "e164", pattern: "^\\+[1-9]\\d{1,14}$", required: true }
            From:            { type: string, format: "e164", pattern: "^\\+[1-9]\\d{1,14}$", required: true }
            Url:             { type: string, format: "url_https", required: true }
            StatusCallback:  { type: string, format: "url_https", required: false }
            StatusCallbackEvent: { type: array, items: { type: enum, values: ["initiated", "ringing", "answered", "completed"] }, required: false }
            Timeout:         { type: integer, range: [5, 600], required: false }
        response:
          fields:
            sid:             { type: string, format: "CA[a-f0-9]{32}", required: true }
            account_sid:     { type: string, format: "AC[a-f0-9]{32}", required: true }
            to:              { type: string, format: "e164", required: true }
            from:            { type: string, format: "e164", required: true }
            status:          { type: enum, values: ["queued", "ringing", "in-progress", "completed", "busy", "no-answer", "canceled", "failed"], required: true }
            start_time:      { type: string, format: "rfc2822", nullable: true }
            end_time:        { type: string, format: "rfc2822", nullable: true }
            duration:        { type: string, format: "integer_string", nullable: true }
            price:           { type: string, format: "decimal", nullable: true }
            price_unit:      { type: string, format: "iso4217", required: true }
            direction:       { type: enum, values: ["inbound", "outbound-api", "outbound-dial"], required: true }
            uri:             { type: string, required: true }
        status_codes: [201, 400, 401, 404, 429, 500]

      - path: "/2010-04-01/Accounts/{AccountSid}/Calls/{CallSid}.json"
        method: GET
        request:
          fields: {}
        response:
          fields:
            sid:             { type: string, format: "CA[a-f0-9]{32}", required: true }
            status:          { type: enum, values: ["queued", "ringing", "in-progress", "completed", "busy", "no-answer", "canceled", "failed"], required: true }
            duration:        { type: string, format: "integer_string", nullable: true }
            price:           { type: string, format: "decimal", nullable: true }
            price_unit:      { type: string, format: "iso4217", required: true }
        status_codes: [200, 401, 404, 500]
```

---

## PERFECT - Zero Violations

A production-quality Twilio client with complete error handling, E.164 validation, full status machine coverage, null-safe pricing, and paginated listing.

```typescript
// --- Types ---

const MESSAGE_STATUSES = [
  "queued", "sending", "sent", "delivered",
  "undelivered", "failed", "receiving", "received",
] as const;
type MessageStatus = (typeof MESSAGE_STATUSES)[number];

const CALL_STATUSES = [
  "queued", "ringing", "in-progress", "completed",
  "busy", "no-answer", "canceled", "failed",
] as const;
type CallStatus = (typeof CALL_STATUSES)[number];

const TERMINAL_MESSAGE_STATUSES: ReadonlySet<MessageStatus> = new Set([
  "delivered", "undelivered", "failed", "received",
]);

const TERMINAL_CALL_STATUSES: ReadonlySet<CallStatus> = new Set([
  "completed", "busy", "no-answer", "canceled", "failed",
]);

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
  price: string | null;
  price_unit: string;
  error_code: number | null;
  error_message: string | null;
  num_segments: string;
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
  duration: string | null;
  price: string | null;
  price_unit: string;
  direction: "inbound" | "outbound-api" | "outbound-dial";
  uri: string;
}

interface TwilioMessageList {
  messages: TwilioMessage[];
  page: number;
  page_size: number;
  first_page_uri: string;
  next_page_uri: string | null;
  previous_page_uri: string | null;
  uri: string;
}

interface TwilioError {
  code: number;
  message: string;
  more_info: string;
  status: number;
}

interface SendSmsParams {
  to: string;
  from: string;
  body: string;
  statusCallback?: string;
  maxPrice?: string;
}

interface MakeCallParams {
  to: string;
  from: string;
  url: string;
  statusCallback?: string;
  statusCallbackEvent?: Array<"initiated" | "ringing" | "answered" | "completed">;
  timeout?: number;
}

// --- Validation ---

const E164_REGEX = /^\+[1-9]\d{1,14}$/;
const ACCOUNT_SID_REGEX = /^AC[a-f0-9]{32}$/;
const MESSAGE_SID_REGEX = /^SM[a-f0-9]{32}$/;
const CALL_SID_REGEX = /^CA[a-f0-9]{32}$/;
const AUTH_TOKEN_REGEX = /^[a-f0-9]{32}$/;
const HTTPS_URL_REGEX = /^https:\/\/.+/;

function validateE164(phone: string): boolean {
  return E164_REGEX.test(phone);
}

function validateAccountSid(sid: string): boolean {
  return ACCOUNT_SID_REGEX.test(sid);
}

function validateMessageSid(sid: string): boolean {
  return MESSAGE_SID_REGEX.test(sid);
}

function validateCallSid(sid: string): boolean {
  return CALL_SID_REGEX.test(sid);
}

function validateAuthToken(token: string): boolean {
  return AUTH_TOKEN_REGEX.test(token);
}

function validateHttpsUrl(url: string): boolean {
  try {
    const parsed = new URL(url);
    return parsed.protocol === "https:";
  } catch {
    return false;
  }
}

// --- Price Handling (string-safe, no float math) ---

/**
 * Twilio returns price as a string decimal (e.g., "-0.0075") or null.
 * We keep it as a string to avoid floating-point precision loss.
 * Aggregation uses integer arithmetic in the smallest currency unit.
 */
function priceToMicros(priceStr: string): bigint {
  // Twilio prices are negative (cost to us), e.g. "-0.0075"
  // Convert to micros (millionths of a currency unit) using string parsing
  const negative = priceStr.startsWith("-");
  const abs = negative ? priceStr.slice(1) : priceStr;
  const [intPart, fracPart = ""] = abs.split(".");
  const paddedFrac = fracPart.padEnd(6, "0").slice(0, 6);
  const micros = BigInt(intPart) * 1_000_000n + BigInt(paddedFrac);
  return negative ? -micros : micros;
}

function microsToPrice(micros: bigint): string {
  const negative = micros < 0n;
  const abs = negative ? -micros : micros;
  const intPart = abs / 1_000_000n;
  const fracPart = abs % 1_000_000n;
  const fracStr = fracPart.toString().padStart(6, "0").replace(/0+$/, "") || "0";
  return `${negative ? "-" : ""}${intPart}.${fracStr}`;
}

function aggregatePrices(prices: Array<string | null>): string {
  let totalMicros = 0n;
  for (const p of prices) {
    if (p !== null) {
      totalMicros += priceToMicros(p);
    }
  }
  return microsToPrice(totalMicros);
}

// --- Client ---

class TwilioClient {
  private readonly baseUrl: string;
  private readonly accountSid: string;
  private readonly authToken: string;
  private readonly authHeader: string;

  constructor(accountSid: string, authToken: string) {
    if (!validateAccountSid(accountSid)) {
      throw new Error(
        `Invalid Account SID format: must match AC followed by 32 hex characters`
      );
    }
    if (!validateAuthToken(authToken)) {
      throw new Error(
        `Invalid Auth Token format: must be 32 hex characters`
      );
    }

    this.baseUrl = "https://api.twilio.com/2010-04-01";
    this.accountSid = accountSid;
    this.authToken = authToken;
    this.authHeader =
      "Basic " + Buffer.from(`${accountSid}:${authToken}`).toString("base64");
  }

  // --- SMS ---

  async sendSms(params: SendSmsParams): Promise<TwilioMessage> {
    if (!validateE164(params.to)) {
      throw new Error(
        `Invalid 'To' phone number: must be E.164 format (e.g., +15551234567)`
      );
    }
    if (!validateE164(params.from)) {
      throw new Error(
        `Invalid 'From' phone number: must be E.164 format (e.g., +15551234567)`
      );
    }
    if (params.body.length === 0 || params.body.length > 1600) {
      throw new Error(`Message body must be between 1 and 1600 characters`);
    }
    if (params.statusCallback !== undefined) {
      if (!validateHttpsUrl(params.statusCallback)) {
        throw new Error(`StatusCallback must be a valid HTTPS URL`);
      }
    }

    const formData = new URLSearchParams();
    formData.append("To", params.to);
    formData.append("From", params.from);
    formData.append("Body", params.body);
    if (params.statusCallback !== undefined) {
      formData.append("StatusCallback", params.statusCallback);
    }
    if (params.maxPrice !== undefined) {
      formData.append("MaxPrice", params.maxPrice);
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

    if (response.status === 401) {
      throw new Error("Authentication failed: invalid Account SID or Auth Token");
    }

    if (response.status === 429) {
      throw new Error("Rate limited by Twilio: retry after backoff");
    }

    let errorBody: TwilioError;
    try {
      errorBody = await response.json();
    } catch {
      throw new Error(`Twilio returned HTTP ${response.status} with unparseable body`);
    }

    throw new Error(
      `Twilio error ${errorBody.code}: ${errorBody.message} (see ${errorBody.more_info})`
    );
  }

  async getMessage(messageSid: string): Promise<TwilioMessage> {
    if (!validateMessageSid(messageSid)) {
      throw new Error(
        `Invalid Message SID format: must match SM followed by 32 hex characters`
      );
    }

    let response: Response;
    try {
      response = await fetch(
        `${this.baseUrl}/Accounts/${this.accountSid}/Messages/${messageSid}.json`,
        {
          method: "GET",
          headers: { "Authorization": this.authHeader },
        }
      );
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      throw new Error(`Network error fetching message: ${message}`);
    }

    if (!response.ok) {
      if (response.status === 404) {
        throw new Error(`Message ${messageSid} not found`);
      }
      if (response.status === 401) {
        throw new Error("Authentication failed: invalid Account SID or Auth Token");
      }
      let errorBody: TwilioError;
      try {
        errorBody = await response.json();
      } catch {
        throw new Error(`Twilio returned HTTP ${response.status} with unparseable body`);
      }
      throw new Error(
        `Twilio error ${errorBody.code}: ${errorBody.message} (see ${errorBody.more_info})`
      );
    }

    const msg: TwilioMessage = await response.json();
    if (!validateMessageSid(msg.sid)) {
      throw new Error(`Unexpected message SID format in response: ${msg.sid}`);
    }
    return msg;
  }

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

    const allMessages: TwilioMessage[] = [];
    let nextPageUri: string | null =
      `/2010-04-01/Accounts/${this.accountSid}/Messages.json?PageSize=${pageSize}`;

    if (filters?.to) {
      nextPageUri += `&To=${encodeURIComponent(filters.to)}`;
    }
    if (filters?.from) {
      nextPageUri += `&From=${encodeURIComponent(filters.from)}`;
    }
    if (filters?.dateSent) {
      nextPageUri += `&DateSent=${encodeURIComponent(filters.dateSent)}`;
    }

    while (nextPageUri !== null) {
      const fullUrl = `https://api.twilio.com${nextPageUri}`;

      let response: Response;
      try {
        response = await fetch(fullUrl, {
          method: "GET",
          headers: { "Authorization": this.authHeader },
        });
      } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        throw new Error(`Network error listing messages: ${message}`);
      }

      if (!response.ok) {
        if (response.status === 401) {
          throw new Error("Authentication failed: invalid Account SID or Auth Token");
        }
        let errorBody: TwilioError;
        try {
          errorBody = await response.json();
        } catch {
          throw new Error(
            `Twilio returned HTTP ${response.status} with unparseable body`
          );
        }
        throw new Error(
          `Twilio error ${errorBody.code}: ${errorBody.message} (see ${errorBody.more_info})`
        );
      }

      const page: TwilioMessageList = await response.json();
      allMessages.push(...page.messages);
      nextPageUri = page.next_page_uri;
    }

    return allMessages;
  }

  // --- Voice ---

  async makeCall(params: MakeCallParams): Promise<TwilioCall> {
    if (!validateE164(params.to)) {
      throw new Error(
        `Invalid 'To' phone number: must be E.164 format (e.g., +15551234567)`
      );
    }
    if (!validateE164(params.from)) {
      throw new Error(
        `Invalid 'From' phone number: must be E.164 format (e.g., +15551234567)`
      );
    }
    if (!validateHttpsUrl(params.url)) {
      throw new Error(`TwiML URL must be a valid HTTPS URL`);
    }
    if (params.statusCallback !== undefined) {
      if (!validateHttpsUrl(params.statusCallback)) {
        throw new Error(`StatusCallback must be a valid HTTPS URL`);
      }
    }
    if (params.timeout !== undefined) {
      if (params.timeout < 5 || params.timeout > 600) {
        throw new Error(`Timeout must be between 5 and 600 seconds`);
      }
    }

    const formData = new URLSearchParams();
    formData.append("To", params.to);
    formData.append("From", params.from);
    formData.append("Url", params.url);
    if (params.statusCallback !== undefined) {
      formData.append("StatusCallback", params.statusCallback);
    }
    if (params.statusCallbackEvent !== undefined) {
      for (const event of params.statusCallbackEvent) {
        formData.append("StatusCallbackEvent", event);
      }
    }
    if (params.timeout !== undefined) {
      formData.append("Timeout", params.timeout.toString());
    }

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
      const message = err instanceof Error ? err.message : String(err);
      throw new Error(`Network error initiating call: ${message}`);
    }

    if (response.status === 201) {
      const call: TwilioCall = await response.json();
      if (!validateCallSid(call.sid)) {
        throw new Error(`Unexpected call SID format: ${call.sid}`);
      }
      return call;
    }

    if (response.status === 401) {
      throw new Error("Authentication failed: invalid Account SID or Auth Token");
    }
    if (response.status === 429) {
      throw new Error("Rate limited by Twilio: retry after backoff");
    }

    let errorBody: TwilioError;
    try {
      errorBody = await response.json();
    } catch {
      throw new Error(`Twilio returned HTTP ${response.status} with unparseable body`);
    }

    throw new Error(
      `Twilio error ${errorBody.code}: ${errorBody.message} (see ${errorBody.more_info})`
    );
  }

  async getCall(callSid: string): Promise<TwilioCall> {
    if (!validateCallSid(callSid)) {
      throw new Error(
        `Invalid Call SID format: must match CA followed by 32 hex characters`
      );
    }

    let response: Response;
    try {
      response = await fetch(
        `${this.baseUrl}/Accounts/${this.accountSid}/Calls/${callSid}.json`,
        {
          method: "GET",
          headers: { "Authorization": this.authHeader },
        }
      );
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      throw new Error(`Network error fetching call: ${message}`);
    }

    if (!response.ok) {
      if (response.status === 404) {
        throw new Error(`Call ${callSid} not found`);
      }
      if (response.status === 401) {
        throw new Error("Authentication failed: invalid Account SID or Auth Token");
      }
      let errorBody: TwilioError;
      try {
        errorBody = await response.json();
      } catch {
        throw new Error(`Twilio returned HTTP ${response.status} with unparseable body`);
      }
      throw new Error(
        `Twilio error ${errorBody.code}: ${errorBody.message} (see ${errorBody.more_info})`
      );
    }

    const call: TwilioCall = await response.json();
    if (!validateCallSid(call.sid)) {
      throw new Error(`Unexpected call SID format in response: ${call.sid}`);
    }
    return call;
  }

  // --- Status Machine Helpers ---

  isMessageTerminal(status: MessageStatus): boolean {
    return TERMINAL_MESSAGE_STATUSES.has(status);
  }

  isCallTerminal(status: CallStatus): boolean {
    return TERMINAL_CALL_STATUSES.has(status);
  }

  isMessageFailed(message: TwilioMessage): boolean {
    return message.status === "failed" || message.status === "undelivered";
  }

  isCallFailed(call: TwilioCall): boolean {
    return call.status === "failed";
  }

  getMessageErrorInfo(message: TwilioMessage): string | null {
    if (message.error_code === null) {
      return null;
    }
    return `Error ${message.error_code}: ${message.error_message ?? "unknown"}`;
  }

  getMessagePrice(message: TwilioMessage): string | null {
    // Price is null until the message reaches a terminal state
    if (!this.isMessageTerminal(message.status)) {
      return null;
    }
    return message.price;
  }

  async waitForMessageDelivery(
    messageSid: string,
    maxAttempts: number = 30,
    intervalMs: number = 1000
  ): Promise<TwilioMessage> {
    for (let attempt = 0; attempt < maxAttempts; attempt++) {
      const message = await this.getMessage(messageSid);
      if (this.isMessageTerminal(message.status)) {
        return message;
      }
      await new Promise((resolve) => setTimeout(resolve, intervalMs));
    }
    throw new Error(
      `Message ${messageSid} did not reach terminal state after ${maxAttempts} attempts`
    );
  }
}

// --- Tests ---

describe("TwilioClient", () => {
  const ACCOUNT_SID = "AC" + "a".repeat(32);
  const AUTH_TOKEN = "b".repeat(32);
  const VALID_TO = "+15551234567";
  const VALID_FROM = "+15559876543";

  const MOCK_MESSAGE: TwilioMessage = {
    sid: "SM" + "c".repeat(32),
    account_sid: ACCOUNT_SID,
    to: VALID_TO,
    from: VALID_FROM,
    body: "Hello, world!",
    status: "queued",
    date_created: "Mon, 10 Feb 2026 12:00:00 +0000",
    date_updated: "Mon, 10 Feb 2026 12:00:00 +0000",
    date_sent: null,
    price: null,
    price_unit: "USD",
    error_code: null,
    error_message: null,
    num_segments: "1",
    direction: "outbound-api",
    uri: `/2010-04-01/Accounts/${ACCOUNT_SID}/Messages/SM${"c".repeat(32)}.json`,
  };

  const MOCK_CALL: TwilioCall = {
    sid: "CA" + "d".repeat(32),
    account_sid: ACCOUNT_SID,
    to: VALID_TO,
    from: VALID_FROM,
    status: "queued",
    start_time: null,
    end_time: null,
    duration: null,
    price: null,
    price_unit: "USD",
    direction: "outbound-api",
    uri: `/2010-04-01/Accounts/${ACCOUNT_SID}/Calls/CA${"d".repeat(32)}.json`,
  };

  let client: TwilioClient;

  beforeEach(() => {
    client = new TwilioClient(ACCOUNT_SID, AUTH_TOKEN);
  });

  // --- Constructor Validation ---

  test("rejects invalid Account SID format", () => {
    expect(() => new TwilioClient("INVALID_SID", AUTH_TOKEN)).toThrow(
      /Invalid Account SID format/
    );
  });

  test("rejects invalid Auth Token format", () => {
    expect(() => new TwilioClient(ACCOUNT_SID, "short")).toThrow(
      /Invalid Auth Token format/
    );
  });

  // --- SMS Send ---

  test("sends SMS successfully and returns complete message", async () => {
    global.fetch = jest.fn().mockResolvedValue({
      status: 201,
      ok: true,
      json: () => Promise.resolve(MOCK_MESSAGE),
    });

    const result = await client.sendSms({
      to: VALID_TO,
      from: VALID_FROM,
      body: "Hello, world!",
    });

    expect(result.sid).toMatch(/^SM[a-f0-9]{32}$/);
    expect(result.account_sid).toBe(ACCOUNT_SID);
    expect(result.to).toBe(VALID_TO);
    expect(result.from).toBe(VALID_FROM);
    expect(result.body).toBe("Hello, world!");
    expect(result.status).toBe("queued");
    expect(result.price).toBeNull();
    expect(result.error_code).toBeNull();
    expect(result.error_message).toBeNull();
    expect(result.num_segments).toBe("1");
    expect(result.direction).toBe("outbound-api");
    expect(MESSAGE_STATUSES).toContain(result.status);
  });

  test("rejects SMS with invalid To phone number", async () => {
    await expect(
      client.sendSms({ to: "5551234567", from: VALID_FROM, body: "test" })
    ).rejects.toThrow(/Invalid 'To' phone number/);
  });

  test("rejects SMS with invalid From phone number", async () => {
    await expect(
      client.sendSms({ to: VALID_TO, from: "invalid", body: "test" })
    ).rejects.toThrow(/Invalid 'From' phone number/);
  });

  test("rejects SMS with empty body", async () => {
    await expect(
      client.sendSms({ to: VALID_TO, from: VALID_FROM, body: "" })
    ).rejects.toThrow(/between 1 and 1600 characters/);
  });

  test("rejects SMS with invalid StatusCallback URL", async () => {
    await expect(
      client.sendSms({
        to: VALID_TO,
        from: VALID_FROM,
        body: "test",
        statusCallback: "http://not-https.com/callback",
      })
    ).rejects.toThrow(/StatusCallback must be a valid HTTPS URL/);
  });

  // --- SMS Error Handling ---

  test("throws on 401 authentication failure", async () => {
    global.fetch = jest.fn().mockResolvedValue({
      status: 401,
      ok: false,
      json: () =>
        Promise.resolve({
          code: 20003,
          message: "Authenticate",
          more_info: "https://www.twilio.com/docs/errors/20003",
          status: 401,
        }),
    });

    await expect(
      client.sendSms({ to: VALID_TO, from: VALID_FROM, body: "test" })
    ).rejects.toThrow(/Authentication failed/);
  });

  test("throws on 429 rate limit", async () => {
    global.fetch = jest.fn().mockResolvedValue({
      status: 429,
      ok: false,
      json: () =>
        Promise.resolve({
          code: 20429,
          message: "Too Many Requests",
          more_info: "https://www.twilio.com/docs/errors/20429",
          status: 429,
        }),
    });

    await expect(
      client.sendSms({ to: VALID_TO, from: VALID_FROM, body: "test" })
    ).rejects.toThrow(/Rate limited/);
  });

  test("throws descriptive error on 400 with Twilio error body", async () => {
    global.fetch = jest.fn().mockResolvedValue({
      status: 400,
      ok: false,
      json: () =>
        Promise.resolve({
          code: 21211,
          message: "Invalid 'To' Phone Number",
          more_info: "https://www.twilio.com/docs/errors/21211",
          status: 400,
        }),
    });

    await expect(
      client.sendSms({ to: VALID_TO, from: VALID_FROM, body: "test" })
    ).rejects.toThrow(/Twilio error 21211/);
  });

  test("throws on network failure", async () => {
    global.fetch = jest.fn().mockRejectedValue(new Error("ECONNREFUSED"));

    await expect(
      client.sendSms({ to: VALID_TO, from: VALID_FROM, body: "test" })
    ).rejects.toThrow(/Network error sending SMS/);
  });

  // --- Message Retrieval ---

  test("retrieves message by SID with full shape validation", async () => {
    const delivered = { ...MOCK_MESSAGE, status: "delivered" as const, price: "-0.0075", date_sent: "Mon, 10 Feb 2026 12:00:01 +0000" };
    global.fetch = jest.fn().mockResolvedValue({
      status: 200,
      ok: true,
      json: () => Promise.resolve(delivered),
    });

    const result = await client.getMessage(MOCK_MESSAGE.sid);
    expect(result.sid).toBe(MOCK_MESSAGE.sid);
    expect(result.status).toBe("delivered");
    expect(result.price).toBe("-0.0075");
    expect(typeof result.price).toBe("string");
    expect(result.date_sent).not.toBeNull();
  });

  test("rejects getMessage with invalid SID format", async () => {
    await expect(client.getMessage("INVALID")).rejects.toThrow(
      /Invalid Message SID format/
    );
  });

  test("throws 404 for unknown message SID", async () => {
    global.fetch = jest.fn().mockResolvedValue({
      status: 404,
      ok: false,
      json: () =>
        Promise.resolve({
          code: 20404,
          message: "The requested resource was not found",
          more_info: "https://www.twilio.com/docs/errors/20404",
          status: 404,
        }),
    });

    await expect(
      client.getMessage("SM" + "f".repeat(32))
    ).rejects.toThrow(/not found/);
  });

  // --- Message Status Machine ---

  test("correctly identifies terminal message statuses", () => {
    expect(client.isMessageTerminal("queued")).toBe(false);
    expect(client.isMessageTerminal("sending")).toBe(false);
    expect(client.isMessageTerminal("sent")).toBe(false);
    expect(client.isMessageTerminal("delivered")).toBe(true);
    expect(client.isMessageTerminal("undelivered")).toBe(true);
    expect(client.isMessageTerminal("failed")).toBe(true);
    expect(client.isMessageTerminal("receiving")).toBe(false);
    expect(client.isMessageTerminal("received")).toBe(true);
  });

  test("correctly identifies failed messages", () => {
    expect(client.isMessageFailed({ ...MOCK_MESSAGE, status: "failed" })).toBe(true);
    expect(client.isMessageFailed({ ...MOCK_MESSAGE, status: "undelivered" })).toBe(true);
    expect(client.isMessageFailed({ ...MOCK_MESSAGE, status: "delivered" })).toBe(false);
    expect(client.isMessageFailed({ ...MOCK_MESSAGE, status: "queued" })).toBe(false);
  });

  // --- Error Info Extraction ---

  test("extracts error info from failed message", () => {
    const failed: TwilioMessage = {
      ...MOCK_MESSAGE,
      status: "failed",
      error_code: 30003,
      error_message: "Unreachable destination handset",
    };
    expect(client.getMessageErrorInfo(failed)).toBe(
      "Error 30003: Unreachable destination handset"
    );
  });

  test("returns null error info for successful message", () => {
    expect(client.getMessageErrorInfo(MOCK_MESSAGE)).toBeNull();
  });

  // --- Price Handling ---

  test("price is null for non-terminal message status", () => {
    expect(client.getMessagePrice({ ...MOCK_MESSAGE, status: "queued" })).toBeNull();
    expect(client.getMessagePrice({ ...MOCK_MESSAGE, status: "sending" })).toBeNull();
  });

  test("price is returned for terminal message status", () => {
    const delivered: TwilioMessage = {
      ...MOCK_MESSAGE,
      status: "delivered",
      price: "-0.0075",
    };
    expect(client.getMessagePrice(delivered)).toBe("-0.0075");
  });

  test("price aggregation preserves precision", () => {
    const prices = ["-0.0075", "-0.0075", "-0.0075", null, "-0.01"];
    const total = aggregatePrices(prices);
    expect(total).toBe("-0.0325");
  });

  // --- Pagination ---

  test("fetches all pages of messages", async () => {
    const page1: TwilioMessageList = {
      messages: [MOCK_MESSAGE],
      page: 0,
      page_size: 1,
      first_page_uri: `/2010-04-01/Accounts/${ACCOUNT_SID}/Messages.json?PageSize=1`,
      next_page_uri: `/2010-04-01/Accounts/${ACCOUNT_SID}/Messages.json?PageSize=1&Page=1`,
      previous_page_uri: null,
      uri: `/2010-04-01/Accounts/${ACCOUNT_SID}/Messages.json?PageSize=1`,
    };
    const msg2 = { ...MOCK_MESSAGE, sid: "SM" + "e".repeat(32) };
    const page2: TwilioMessageList = {
      messages: [msg2],
      page: 1,
      page_size: 1,
      first_page_uri: `/2010-04-01/Accounts/${ACCOUNT_SID}/Messages.json?PageSize=1`,
      next_page_uri: null,
      previous_page_uri: `/2010-04-01/Accounts/${ACCOUNT_SID}/Messages.json?PageSize=1&Page=0`,
      uri: `/2010-04-01/Accounts/${ACCOUNT_SID}/Messages.json?PageSize=1&Page=1`,
    };

    global.fetch = jest
      .fn()
      .mockResolvedValueOnce({ status: 200, ok: true, json: () => Promise.resolve(page1) })
      .mockResolvedValueOnce({ status: 200, ok: true, json: () => Promise.resolve(page2) });

    const all = await client.listAllMessages({ pageSize: 1 });
    expect(all).toHaveLength(2);
    expect(all[0].sid).toBe(MOCK_MESSAGE.sid);
    expect(all[1].sid).toBe(msg2.sid);
    expect(global.fetch).toHaveBeenCalledTimes(2);
  });

  // --- Voice ---

  test("makes call successfully and returns complete call", async () => {
    global.fetch = jest.fn().mockResolvedValue({
      status: 201,
      ok: true,
      json: () => Promise.resolve(MOCK_CALL),
    });

    const result = await client.makeCall({
      to: VALID_TO,
      from: VALID_FROM,
      url: "https://example.com/twiml",
    });

    expect(result.sid).toMatch(/^CA[a-f0-9]{32}$/);
    expect(result.status).toBe("queued");
    expect(result.price).toBeNull();
    expect(result.duration).toBeNull();
    expect(CALL_STATUSES).toContain(result.status);
  });

  test("rejects call with invalid TwiML URL", async () => {
    await expect(
      client.makeCall({
        to: VALID_TO,
        from: VALID_FROM,
        url: "http://not-https.com/twiml",
      })
    ).rejects.toThrow(/TwiML URL must be a valid HTTPS URL/);
  });

  test("rejects call with out-of-range timeout", async () => {
    await expect(
      client.makeCall({
        to: VALID_TO,
        from: VALID_FROM,
        url: "https://example.com/twiml",
        timeout: 3,
      })
    ).rejects.toThrow(/Timeout must be between 5 and 600/);
  });

  // --- Call Status Machine ---

  test("correctly identifies terminal call statuses", () => {
    expect(client.isCallTerminal("queued")).toBe(false);
    expect(client.isCallTerminal("ringing")).toBe(false);
    expect(client.isCallTerminal("in-progress")).toBe(false);
    expect(client.isCallTerminal("completed")).toBe(true);
    expect(client.isCallTerminal("busy")).toBe(true);
    expect(client.isCallTerminal("no-answer")).toBe(true);
    expect(client.isCallTerminal("canceled")).toBe(true);
    expect(client.isCallTerminal("failed")).toBe(true);
  });
});
```

**Expected Stricture result:** PASS (zero violations). All request shapes match the manifest, all response fields are typed correctly, all status codes are handled, all enums are exhaustive, all negative paths are tested, nullable fields are null-checked, SIDs are format-validated, phone numbers are E.164-validated, pagination follows `next_page_uri` to completion, and price aggregation uses integer arithmetic.

---

## B01 - No Error Handling

No try/catch around fetch calls. Network errors and JSON parse failures crash the caller.

```typescript
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
```

**Expected violation:**
- **Rule:** `TQ-error-path-coverage`
- **Message:** `fetch() calls in sendSms, getMessage, makeCall have no try/catch — network errors (ECONNREFUSED, DNS failure, timeout) propagate as unhandled rejections with no contextual error message`
- **Production impact:** Any transient network issue (DNS resolution failure, TCP timeout, TLS handshake error) surfaces as an opaque `TypeError: Failed to fetch` rather than a meaningful error. Callers cannot distinguish network failures from Twilio API errors, making retry logic and alerting impossible to implement correctly.

---

## B02 - No Status Code Check

Treats every HTTP response as a 201 success, never inspecting the status code.

```typescript
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

    // BUG: No status code check — 400, 401, 429, 500 responses are all
    // parsed as if they were a successful 201 TwilioMessage.
    // The error body { code, message, more_info, status } will be silently
    // misinterpreted as a TwilioMessage with undefined fields.
    const message: TwilioMessage = await response.json();
    return message;
  }

  async getMessage(messageSid: string): Promise<TwilioMessage> {
    if (!validateMessageSid(messageSid)) throw new Error(`Invalid Message SID format`);

    let response: Response;
    try {
      response = await fetch(
        `${this.baseUrl}/Accounts/${this.accountSid}/Messages/${messageSid}.json`,
        { method: "GET", headers: { "Authorization": this.authHeader } }
      );
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      throw new Error(`Network error fetching message: ${message}`);
    }

    // BUG: same — no response.ok or status check
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
      throw new Error(`Network error initiating call: ${msg}`);
    }

    // BUG: same — no status check
    return response.json();
  }
}
```

**Expected violation:**
- **Rule:** `CTR-status-code-handling`
- **Message:** `Endpoints sendSms (expects 201), getMessage (expects 200), makeCall (expects 201) never check response.status — error responses (400, 401, 404, 429, 500) are parsed as success types, producing objects with undefined required fields`
- **Production impact:** A 401 response returns `{ code: 20003, message: "Authenticate", ... }` which is silently cast to `TwilioMessage`. Downstream code reads `message.sid` as `undefined`, `message.status` as `undefined`, causing null reference errors far from the original failure point. Rate limiting (429) is invisible, so the client hammers the API until the account is suspended.

---

## B03 - Shallow Test Assertions

Tests exist but only assert `toBeDefined()` / `toBeTruthy()` instead of validating response shapes and field values.

```typescript
// Implementation is identical to PERFECT — only tests are buggy.

describe("TwilioClient", () => {
  const ACCOUNT_SID = "AC" + "a".repeat(32);
  const AUTH_TOKEN = "b".repeat(32);
  const client = new TwilioClient(ACCOUNT_SID, AUTH_TOKEN);

  const MOCK_MESSAGE: TwilioMessage = {
    sid: "SM" + "c".repeat(32),
    account_sid: ACCOUNT_SID,
    to: "+15551234567",
    from: "+15559876543",
    body: "Hello, world!",
    status: "queued",
    date_created: "Mon, 10 Feb 2026 12:00:00 +0000",
    date_updated: "Mon, 10 Feb 2026 12:00:00 +0000",
    date_sent: null,
    price: null,
    price_unit: "USD",
    error_code: null,
    error_message: null,
    num_segments: "1",
    direction: "outbound-api",
    uri: `/2010-04-01/Accounts/${ACCOUNT_SID}/Messages/SM${"c".repeat(32)}.json`,
  };

  // BUG: Shallow assertion — only checks the result is defined,
  // never validates SID format, status enum, field types, or response shape
  test("sends SMS", async () => {
    global.fetch = jest.fn().mockResolvedValue({
      status: 201,
      ok: true,
      json: () => Promise.resolve(MOCK_MESSAGE),
    });

    const result = await client.sendSms({
      to: "+15551234567",
      from: "+15559876543",
      body: "Hello",
    });

    expect(result).toBeDefined();       // BUG: says nothing about shape
    expect(result.sid).toBeTruthy();     // BUG: any truthy string passes
    expect(result.status).toBeTruthy();  // BUG: "invalid_garbage" would pass
  });

  // BUG: Shallow assertion on retrieval
  test("gets message", async () => {
    global.fetch = jest.fn().mockResolvedValue({
      status: 200,
      ok: true,
      json: () => Promise.resolve({ ...MOCK_MESSAGE, status: "delivered", price: "-0.0075" }),
    });

    const result = await client.getMessage(MOCK_MESSAGE.sid);
    expect(result).toBeDefined();  // BUG: no check on status, price type, error fields
  });

  // BUG: Shallow assertion on call
  test("makes call", async () => {
    global.fetch = jest.fn().mockResolvedValue({
      status: 201,
      ok: true,
      json: () =>
        Promise.resolve({
          sid: "CA" + "d".repeat(32),
          account_sid: ACCOUNT_SID,
          to: "+15551234567",
          from: "+15559876543",
          status: "queued",
          start_time: null,
          end_time: null,
          duration: null,
          price: null,
          price_unit: "USD",
          direction: "outbound-api",
          uri: `/2010-04-01/Accounts/${ACCOUNT_SID}/Calls/CA${"d".repeat(32)}.json`,
        }),
    });

    const result = await client.makeCall({
      to: "+15551234567",
      from: "+15559876543",
      url: "https://example.com/twiml",
    });
    expect(result).toBeTruthy();  // BUG: no shape, SID format, or status validation
  });
});
```

**Expected violation:**
- **Rule:** `TQ-no-shallow-assertions`
- **Message:** `Tests use toBeDefined()/toBeTruthy() on response objects without validating field types, SID format patterns (SM*/CA*), enum membership, or nullable field handling — a response with { sid: "WRONG", status: "invalid" } would pass all assertions`
- **Production impact:** Tests provide false confidence. A breaking change in Twilio's API (e.g., field renamed from `error_code` to `errorCode`) would not be caught because no test inspects the actual shape. The test suite greenlights deployments that will crash in production when accessing renamed or removed fields.

---

## B04 - Missing Negative Tests

Only tests happy paths (successful send, successful get). No tests for error responses, invalid inputs, or edge cases.

```typescript
// Implementation is identical to PERFECT — only test suite is incomplete.

describe("TwilioClient", () => {
  const ACCOUNT_SID = "AC" + "a".repeat(32);
  const AUTH_TOKEN = "b".repeat(32);
  const client = new TwilioClient(ACCOUNT_SID, AUTH_TOKEN);

  // BUG: Only happy-path tests. No tests for:
  // - Invalid phone numbers (non-E.164)
  // - 401 authentication failures
  // - 429 rate limiting
  // - 400 bad requests
  // - 404 not found
  // - 500 server errors
  // - Network failures (ECONNREFUSED)
  // - Invalid SID formats
  // - Failed/undelivered message statuses
  // - Null price before delivery
  // - Empty body, body > 1600 chars
  // - Invalid StatusCallback URL

  test("sends SMS successfully", async () => {
    global.fetch = jest.fn().mockResolvedValue({
      status: 201,
      ok: true,
      json: () =>
        Promise.resolve({
          sid: "SM" + "c".repeat(32),
          account_sid: ACCOUNT_SID,
          to: "+15551234567",
          from: "+15559876543",
          body: "Hello",
          status: "queued",
          date_created: "Mon, 10 Feb 2026 12:00:00 +0000",
          date_updated: "Mon, 10 Feb 2026 12:00:00 +0000",
          date_sent: null,
          price: null,
          price_unit: "USD",
          error_code: null,
          error_message: null,
          num_segments: "1",
          direction: "outbound-api",
          uri: "/2010-04-01/Accounts/AC" + "a".repeat(32) + "/Messages/SM" + "c".repeat(32) + ".json",
        }),
    });

    const result = await client.sendSms({
      to: "+15551234567",
      from: "+15559876543",
      body: "Hello",
    });

    expect(result.sid).toMatch(/^SM[a-f0-9]{32}$/);
    expect(result.status).toBe("queued");
  });

  test("gets message successfully", async () => {
    global.fetch = jest.fn().mockResolvedValue({
      status: 200,
      ok: true,
      json: () =>
        Promise.resolve({
          sid: "SM" + "c".repeat(32),
          account_sid: ACCOUNT_SID,
          to: "+15551234567",
          from: "+15559876543",
          body: "Hello",
          status: "delivered",
          date_created: "Mon, 10 Feb 2026 12:00:00 +0000",
          date_updated: "Mon, 10 Feb 2026 12:00:01 +0000",
          date_sent: "Mon, 10 Feb 2026 12:00:01 +0000",
          price: "-0.0075",
          price_unit: "USD",
          error_code: null,
          error_message: null,
          num_segments: "1",
          direction: "outbound-api",
          uri: "/2010-04-01/Accounts/AC" + "a".repeat(32) + "/Messages/SM" + "c".repeat(32) + ".json",
        }),
    });

    const result = await client.getMessage("SM" + "c".repeat(32));
    expect(result.status).toBe("delivered");
    expect(result.price).toBe("-0.0075");
  });

  test("makes call successfully", async () => {
    global.fetch = jest.fn().mockResolvedValue({
      status: 201,
      ok: true,
      json: () =>
        Promise.resolve({
          sid: "CA" + "d".repeat(32),
          account_sid: ACCOUNT_SID,
          to: "+15551234567",
          from: "+15559876543",
          status: "queued",
          start_time: null,
          end_time: null,
          duration: null,
          price: null,
          price_unit: "USD",
          direction: "outbound-api",
          uri: "/2010-04-01/Accounts/AC" + "a".repeat(32) + "/Calls/CA" + "d".repeat(32) + ".json",
        }),
    });

    const result = await client.makeCall({
      to: "+15551234567",
      from: "+15559876543",
      url: "https://example.com/twiml",
    });

    expect(result.sid).toMatch(/^CA[a-f0-9]{32}$/);
    expect(result.status).toBe("queued");
  });
});
```

**Expected violation:**
- **Rule:** `TQ-negative-cases`
- **Message:** `Test suite covers only 3 happy-path scenarios (sendSms success, getMessage success, makeCall success). Missing negative tests for: 401 auth failure, 429 rate limit, 400 bad request, 404 not found, network errors, invalid E.164 numbers, invalid SID formats, failed/undelivered statuses, null price access, body length bounds, invalid callback URLs`
- **Production impact:** Without negative tests, error handling code paths are never exercised. A refactor that accidentally removes the `status === 401` check would not be caught. The team has no way to verify that authentication failures, rate limiting, or invalid input produce meaningful errors rather than silent data corruption.

---

## B05 - Request Missing Required Fields

Omits the `To` field from SMS send, which the Twilio API requires.

```typescript
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
```

**Expected violation:**
- **Rule:** `CTR-request-shape`
- **Message:** `POST /Accounts/{AccountSid}/Messages.json request body is missing required field "To" — manifest declares To as { type: string, format: "e164", required: true } but the URLSearchParams body only includes From, Body, and optional StatusCallback`
- **Production impact:** Every SMS send request results in a 400 error from Twilio: `{ code: 21604, message: "'To' is required" }`. The client validates the `to` parameter locally but never transmits it, so the E.164 validation gives false confidence that the value is being used. This is a 100% failure rate on the primary operation.

---

## B06 - Response Type Mismatch

Client's `TwilioMessage` type is missing the `error_code` and `error_message` fields that Twilio always returns.

```typescript
// BUG: TwilioMessage type is incomplete — missing error_code and error_message
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
  price: string | null;
  price_unit: string;
  // BUG: error_code field missing — Twilio always returns this (null or integer)
  // BUG: error_message field missing — Twilio always returns this (null or string)
  num_segments: string;
  direction: "inbound" | "outbound-api" | "outbound-call" | "outbound-reply";
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

  // BUG: getMessageErrorInfo cannot work — the type has no error_code field
  getMessageErrorInfo(message: TwilioMessage): string | null {
    // TypeScript error: Property 'error_code' does not exist on type 'TwilioMessage'
    // At runtime, this reads undefined from the JSON (which does have the field),
    // but the type system gives no guidance.
    return null; // Falls back to always returning null
  }

  async getMessage(messageSid: string): Promise<TwilioMessage> { /* identical to PERFECT */ }
  async makeCall(params: MakeCallParams): Promise<TwilioCall> { /* identical to PERFECT */ }
}
```

**Expected violation:**
- **Rule:** `CTR-response-shape`
- **Message:** `Client type TwilioMessage is missing fields declared in the manifest response schema: error_code (integer|null, required) and error_message (string|null, required). The Twilio API always returns these fields but the client type cannot access them, making error diagnosis impossible for failed/undelivered messages`
- **Production impact:** When a message fails delivery (status "failed" or "undelivered"), the application cannot report why. Twilio provides error codes (30003 = "Unreachable handset", 30005 = "Unknown destination", 30006 = "Landline unreachable") but the client type has no field to read them. Support teams cannot diagnose delivery failures.

---

## B07 - Wrong Field Types

Stores `price` as `number` instead of `string | null`, and `num_segments` as `number` instead of `string`.

```typescript
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
```

**Expected violation:**
- **Rule:** `CTR-manifest-conformance`
- **Message:** `Field type mismatches between client types and manifest: TwilioMessage.price declared as number|null but manifest specifies string (format: decimal, nullable: true); TwilioMessage.num_segments declared as number but manifest specifies string (format: integer_string); TwilioCall.duration declared as number|null but manifest specifies string (format: integer_string, nullable: true); TwilioCall.price declared as number|null but manifest specifies string (format: decimal, nullable: true)`
- **Production impact:** At runtime, `JSON.parse` correctly returns strings for these fields (Twilio sends `"price": "-0.0075"`, not `"price": -0.0075`). The TypeScript type says `number` but the value is actually `string`, so arithmetic operations like `message.price * 100` silently produce string concatenation (`"-0.007500"`) instead of multiplication. Database writes with the wrong type cause schema validation failures.

---

## B08 - Incomplete Enum Handling

Handles "delivered" and "failed" message statuses but ignores "undelivered", "sending", "queued", "sent", "receiving", and "received".

```typescript
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
```

**Expected violation:**
- **Rule:** `CTR-strictness-parity`
- **Message:** `processMessageStatus handles 2 of 8 enum values (delivered, failed) — missing: queued, sending, sent, undelivered, receiving, received. processCallStatus handles 2 of 8 enum values (completed, failed) — missing: queued, ringing, in-progress, busy, no-answer, canceled. Default case returns empty string with no logging, silently dropping 75% of possible states`
- **Production impact:** The "undelivered" status (carrier rejected, e.g., invalid number on carrier side) is silently ignored and returns `""`. This means messages that were rejected by the carrier appear as successfully processed. The "busy" and "no-answer" call statuses are not handled, so the system cannot retry calls or notify users that the callee was unavailable.

---

## B09 - Missing Range Validation

No check that phone numbers match E.164 format length constraints. Accepts any string starting with `+`.

```typescript
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
```

**Expected violation:**
- **Rule:** `CTR-strictness-parity`
- **Message:** `E.164 phone validation accepts any string starting with "+" — manifest declares pattern "^\+[1-9]\d{1,14}$" requiring 2-15 total characters, first digit after + must be 1-9, digits only. Current validation accepts "+", "+0123" (leading zero), "+abc" (non-digits), and "+1234567890123456789" (>15 digits)`
- **Production impact:** Invalid phone numbers like `+0123456789` (starts with 0, which is not a valid country code) or `+1` (too short for any real number) are sent to Twilio, which rejects them with error 21211. This wastes API calls, pollutes error logs, and delays error detection to the network round-trip instead of catching it locally. The `+abc` case could potentially cause injection issues in downstream logging systems.

---

## B10 - Format Not Validated

SIDs are accepted as any string with no prefix pattern check. `SM` and `CA` prefix validation is absent.

```typescript
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
```

**Expected violation:**
- **Rule:** `CTR-strictness-parity`
- **Message:** `SID format validation checks length (34 chars) but not prefix pattern — manifest declares Account SID as "AC[a-f0-9]{32}", Message SID as "SM[a-f0-9]{32}", Call SID as "CA[a-f0-9]{32}". Current validation accepts cross-resource SIDs (a Call SID "CA..." passes as a Message SID) and arbitrary 34-character strings`
- **Production impact:** A Call SID accidentally passed to `getMessage()` is sent to the Messages endpoint, resulting in a 404. More critically, using a Message SID as an Account SID in the constructor (`new TwilioClient("SM...", token)`) constructs a valid Basic auth header that authenticates against the wrong resource scope, potentially leaking data or causing 401 errors that are difficult to diagnose because the SID "looks valid."

---

## B11 - Precision Loss

Parses Twilio price strings with `parseFloat` and aggregates with floating-point arithmetic, losing precision.

```typescript
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

  // BUG: Aggregates prices using parseFloat + floating-point addition
  // instead of integer arithmetic or BigInt
  async getTotalSpend(messageSids: string[]): Promise<number> {
    let total = 0;

    for (const sid of messageSids) {
      const message = await this.getMessage(sid);

      if (message.price !== null) {
        // BUG: parseFloat("-0.0075") works for single values but
        // floating-point addition accumulates errors:
        //   -0.0075 + -0.0075 + -0.0075 = -0.022499999999999996
        //   (expected: -0.0225)
        total += parseFloat(message.price);
      }
    }

    return total;
  }

  // BUG: Compares price against threshold using floating-point comparison
  isPriceAboveThreshold(message: TwilioMessage, thresholdCents: number): boolean {
    if (message.price === null) return false;

    // BUG: parseFloat("-0.0075") * 100 = -0.75 (correct here by luck)
    // but parseFloat("-0.0033") * 100 = -0.32999999999999996 (not -0.33)
    const priceCents = parseFloat(message.price) * 100;
    return Math.abs(priceCents) > thresholdCents;
  }

  // BUG: Formats aggregated price for display — accumulated errors show up
  formatTotalSpend(prices: Array<string | null>): string {
    let total = 0;
    for (const p of prices) {
      if (p !== null) {
        total += parseFloat(p);  // BUG: precision loss on each addition
      }
    }
    // toFixed(4) masks some errors but not all:
    // 10000 messages at -0.0075 each:
    //   expected: -75.0000
    //   actual:   -74.9999 (or -75.0001 depending on accumulation order)
    return total.toFixed(4);
  }

  async sendSms(params: SendSmsParams): Promise<TwilioMessage> { /* identical to PERFECT */ }
  async getMessage(messageSid: string): Promise<TwilioMessage> { /* identical to PERFECT */ }
  async makeCall(params: MakeCallParams): Promise<TwilioCall> { /* identical to PERFECT */ }
  async getCall(callSid: string): Promise<TwilioCall> { /* identical to PERFECT */ }
}
```

**Expected violation:**
- **Rule:** `CTR-strictness-parity`
- **Message:** `Price field (manifest: string, format: decimal) is parsed with parseFloat() and aggregated with floating-point addition. Price values like "-0.0075" lose precision when accumulated: 3 * -0.0075 = -0.022499999999999996 (expected -0.0225). Manifest format "decimal" requires string-safe arithmetic (e.g., BigInt in micros or a decimal library)`
- **Production impact:** Monthly billing reports show incorrect totals. With 10,000 messages at $0.0075 each, the expected total is $75.00 but floating-point accumulation produces $74.9999 or $75.0001. This causes reconciliation failures with Twilio invoices, accounting discrepancies, and failed automated balance checks. The error grows linearly with volume.

---

## B12 - Nullable Field Crash

Accesses `message.price.toString()` and `message.date_sent.substring()` without null checks, crashing when these fields are null (which they are before delivery).

```typescript
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
```

**Expected violation:**
- **Rule:** `CTR-response-shape`
- **Message:** `Nullable fields accessed without null guards: TwilioMessage.price (nullable: true) accessed via .toString(), TwilioMessage.date_sent (nullable: true) via .substring(), TwilioMessage.error_message (nullable: true) via .toUpperCase(), TwilioCall.duration (nullable: true) via .toString(), TwilioCall.price (nullable: true) via .toString(), TwilioCall.start_time (nullable: true) via .substring(), TwilioCall.end_time (nullable: true) via .substring(). All crash with TypeError when the field is null`
- **Production impact:** Calling `formatMessageSummary()` on a freshly queued message (the most common case, since messages start in "queued" status) immediately throws `TypeError: Cannot read properties of null (reading 'toString')`. This crashes the message tracking pipeline on every send, because the message is always "queued" before it's "delivered." The function only works after the message reaches a terminal state, but it's called in the send confirmation flow.

---

## B13 - Missing Auth Validation

Constructs the client without validating the Account SID or Auth Token format. Accepts empty strings, wrong prefixes, and malformed credentials.

```typescript
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
```

**Expected violation:**
- **Rule:** `CTR-request-shape`
- **Message:** `Constructor does not validate credential formats: accountSid must match "AC[a-f0-9]{32}" (manifest: auth.username_format) and authToken must match "[a-f0-9]{32}" (manifest: auth.password_format). Unvalidated accountSid is interpolated into URL path "/Accounts/{AccountSid}/..." — empty strings, path traversal sequences ("../"), or wrong-prefix SIDs are accepted`
- **Production impact:** Instantiating `new TwilioClient("", "")` creates a client that constructs URLs like `https://api.twilio.com/2010-04-01/Accounts//Messages.json` (double slash). Every API call returns 401 but the error manifests only at request time, not at construction time. Using `new TwilioClient("../Calls", token)` constructs `https://api.twilio.com/2010-04-01/Accounts/../Calls/Messages.json` which resolves to the Calls endpoint instead of Messages, causing silent resource confusion.

---

## B14 - Pagination Terminated Early

Fetches only the first page of messages from the list endpoint, ignoring `next_page_uri`.

```typescript
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
```

**Expected violation:**
- **Rule:** `CTR-response-shape`
- **Message:** `listAllMessages ignores next_page_uri field from pagination response — manifest declares next_page_uri (nullable: true) as the continuation token, but the implementation fetches only the first page and returns page.messages without checking if next_page_uri is non-null. With pageSize=100 and 500 total messages, returns 100 and silently drops 400`
- **Production impact:** Any account with more than `pageSize` messages (default 100) gets silently truncated results. A compliance audit that uses `listAllMessages()` to verify all messages were delivered will miss 80% of messages on a 500-message account. Billing reconciliation undercounts by the same factor. The bug is invisible because the function returns a valid array with no indication that data was omitted.

---

## B15 - Race Condition

Reads message status, then acts on it in a separate step. The status can change between the read and the action, causing stale-state decisions.

```typescript
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
```

**Expected violation:**
- **Rule:** `CTR-request-shape`
- **Message:** `TOCTOU (time-of-check-to-time-of-use) race in resendIfFailed, getFailedMessageSids, and cancelIfStillQueued: message/call status is read via GET, then a subsequent action is taken based on the stale status. Twilio status transitions are asynchronous — "queued" can become "failed" or "delivered" between the GET and the action. resendIfFailed can double-send when a "failed" message is retried internally by Twilio. cancelIfStillQueued can attempt to cancel a call that has already started ringing`
- **Production impact:** `resendIfFailed()` causes double-delivery: user receives an SMS twice (once from Twilio's internal retry, once from our resend). For transactional messages (OTP codes, password resets), this confuses users and may trigger fraud alerts. `cancelIfStillQueued()` can fail with an unhandled 400/409 when the call transitions to "ringing" between the status check and the cancel request, crashing the call management pipeline. `getFailedMessageSids()` returns an inconsistent snapshot when processing hundreds of SIDs sequentially, leading to incorrect failure counts in dashboards.
