// BUG: Missing fulfillment_status, missing line_items
interface Order {
  id: number;
  order_number: number;
  financial_status: string;
  // MISSING: fulfillment_status -- Shopify always sends this (null or string)
  // MISSING: line_items -- Shopify always includes line items
  total_price: string;
  currency: string;
}

async function listOrders(token: string): Promise<Order[]> {
  const res = await fetch(
    `${SHOPIFY_BASE}/admin/api/2024-01/orders.json`,
    { headers: { "X-Shopify-Access-Token": token } }
  );
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  const body = await res.json();
  return body.orders as Order[];
}

function getUnfulfilledOrders(orders: Order[]): Order[] {
  // BUG: Cannot filter by fulfillment_status because it is not in the type
  // This function has no way to distinguish fulfilled from unfulfilled
  return orders; // Returns ALL orders
}

describe("listOrders", () => {
  it("lists orders", async () => {
    const orders = await listOrders("shpat_token");
    expect(orders.length).toBeGreaterThan(0);
    expect(typeof orders[0].id).toBe("number");
    expect(typeof orders[0].total_price).toBe("string");
    // No assertion for fulfillment_status or line_items
  });
});
