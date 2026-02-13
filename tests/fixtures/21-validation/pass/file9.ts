// DEFECT B04: Test suite missing negative cases
describe("OrderClient", () => {
  test("createOrder success", async () => {
    const order = await client.createOrder({
      customer_id: "550e8400-e29b-41d4-a716-446655440000",
      total_amount: 5000,
    });
    expect(order.status).toBe("pending");
  });

  // BUG: Missing tests for:
  // - Invalid UUID → 422
  // - Negative amount → 422
  // - Missing auth → 401
  // - Server error → 500
});
