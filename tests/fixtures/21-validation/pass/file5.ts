// DEFECT B02: No response.ok check
async function createOrder(req: CreateOrderRequest): Promise<Order> {
  const response = await fetch("/api/orders", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });

  // BUG: Missing if (!response.ok) throw ...
  return response.json();  // Will parse error HTML as JSON and fail
}
