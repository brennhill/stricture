// DEFECT B10: No UUID format validation
async function createOrder(req: CreateOrderRequest): Promise<Order> {
  // BUG: Accepts any string for customer_id, no UUID validation
  const response = await fetch("/api/orders", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });

  return response.json();
}

// Usage that should fail client-side but doesn't
await createOrder({
  customer_id: "not-a-uuid",  // INVALID: server will reject with 422
  total_amount: 5000,
});

// Should validate:
// const UUID_REGEX = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
// if (!UUID_REGEX.test(req.customer_id)) throw new Error("Invalid UUID");
