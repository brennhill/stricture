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
