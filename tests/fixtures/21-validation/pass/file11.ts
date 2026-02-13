// DEFECT B05: TS sends wrong field name
async function createOrder(req: CreateOrderRequest): Promise<Order> {
  const payload = {
    customerId: req.customer_id,    // BUG: camelCase instead of snake_case
    totalAmount: req.total_amount,  // BUG: camelCase instead of snake_case
    note: req.note,
  };

  const response = await fetch("/api/orders", {
    method: "POST",
    body: JSON.stringify(payload),
  });

  return response.json();  // Go will reject with 400 or ignore fields
}
