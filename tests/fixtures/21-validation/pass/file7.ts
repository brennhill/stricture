// Equivalent shallow assertion in TS tests
test("createOrder returns order", async () => {
  const order = await client.createOrder({
    customer_id: "550e8400-e29b-41d4-a716-446655440000",
    total_amount: 5000,
  });

  // DEFECT B03: Shallow assertion
  expect(order).toBeDefined();  // Should validate order.id is UUID format
  expect(order.id).toBeTruthy();  // Weak — doesn't validate UUID format
});
