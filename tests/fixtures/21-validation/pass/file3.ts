// DEFECT B01: No try/catch — network errors will crash
async function createOrder(req: CreateOrderRequest): Promise<Order> {
  const response = await fetch("/api/orders", {  // UNHANDLED rejection
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });

  return response.json();  // No error path
}
