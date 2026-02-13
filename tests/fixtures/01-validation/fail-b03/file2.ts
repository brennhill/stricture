// B03: Tests only check existence, never shape or value.

describe("StripeClientB03", () => {
  it("creates a charge", async () => {
    mockFetch(200, MOCK_CHARGE);
    const client = new StripeClientB03("sk_test_abc");
    const result = await client.createCharge({
      amount: 2000,
      currency: "usd",
      source: "tok_visa",
    });

    // BUG: These assertions prove nothing about correctness.
    // They pass even if result is { ok: true, data: { id: 999, amount: "banana" } }.
    expect(result).toBeDefined();
    expect(result.ok).toBeTruthy();
    if (result.ok) {
      expect(result.data).toBeDefined();
      expect(result.data.id).toBeDefined();
      expect(result.data.amount).toBeDefined();
      expect(result.data.currency).toBeDefined();
      expect(result.data.status).toBeDefined();
    }
  });

  it("handles errors", async () => {
    mockFetch(402, MOCK_ERROR_402);
    const client = new StripeClientB03("sk_test_abc");
    const result = await client.createCharge({
      amount: 2000,
      currency: "usd",
      source: "tok_declined",
    });

    // BUG: Only checks that error exists, not its shape or values.
    expect(result).toBeDefined();
    expect(result.ok).toBeFalsy();
    if (!result.ok) {
      expect(result.error).toBeDefined();
    }
  });
});
