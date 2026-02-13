// DEFECT B13: Missing Authorization header
async function createOrder(req: CreateOrderRequest): Promise<Order> {
  const response = await fetch("/api/orders", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      // BUG: Missing Authorization header
      // Should have: "Authorization": `Bearer ${getToken()}`
    },
    body: JSON.stringify(req),
  });

  return response.json();  // Server returns 401, client crashes
}
