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
