// DEFECT B09: No client-side validation
async function createOrder(req: CreateOrderRequest): Promise<Order> {
  // BUG: Sends request without validating total_amount > 0
  // Client allows negative values, server will reject with 422
  const response = await fetch("/api/orders", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),  // No validation here
  });

  return response.json();
}

// Usage that should fail client-side but doesn't
await createOrder({
  customer_id: "550e8400-e29b-41d4-a716-446655440000",
  total_amount: -5000,  // INVALID: negative amount sent to server
});
