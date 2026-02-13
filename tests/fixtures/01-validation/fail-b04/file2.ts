// B04: Only tests the successful path. No error, edge, or failure tests.

describe("StripeClientB04", () => {
  it("creates a charge successfully", async () => {
    mockFetch(200, MOCK_CHARGE);
    const client = new StripeClient("sk_test_abc");
    const result = await client.createCharge({
      amount: 2000,
      currency: "usd",
      source: "tok_visa",
    });
    expect(result.ok).toBe(true);
    if (result.ok) {
      expect(result.data.id).toBe("ch_1OxABC123def456");
      expect(result.data.amount).toBe(2000);
    }
  });

  it("retrieves a charge successfully", async () => {
    mockFetch(200, MOCK_CHARGE);
    const client = new StripeClient("sk_test_abc");
    const result = await client.retrieveCharge("ch_1OxABC123def456");
    expect(result.ok).toBe(true);
    if (result.ok) {
      expect(result.data.id).toBe("ch_1OxABC123def456");
    }
  });

  it("creates a customer successfully", async () => {
    mockFetch(200, MOCK_CUSTOMER);
    const client = new StripeClient("sk_test_abc");
    const result = await client.createCustomer({ email: "test@example.com" });
    expect(result.ok).toBe(true);
  });

  // BUG: No tests for:
  //   - 400, 401, 402, 404, 429, 500 responses
  //   - Network failures (ECONNREFUSED, ETIMEDOUT)
  //   - Invalid input (amount < 50, bad currency, bad charge ID format)
  //   - Null/missing fields (failure_code, balance_transaction)
  //   - Pending and failed charge statuses
  //   - Webhook signature verification failures
  //   - Pagination edge cases
});
