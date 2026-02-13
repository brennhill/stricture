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
