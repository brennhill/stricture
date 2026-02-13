// BUG: total_price, price are number -- Shopify returns string decimals
interface Order {
  id: number;
  order_number: number;
  financial_status: string;
  fulfillment_status: string | null;
  total_price: number;     // BUG: Should be string. Shopify sends "29.99"
  currency: string;
}

interface Variant {
  id: number;
  price: number;           // BUG: Should be string. Shopify sends "19.99"
  sku: string | null;
  inventory_quantity: number;
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

function calculateOrderTotal(orders: Order[]): number {
  // BUG: Summing float-typed prices introduces precision errors
  return orders.reduce((sum, order) => sum + order.total_price, 0);
}

describe("listOrders", () => {
  it("lists orders with total", async () => {
    const orders = await listOrders("shpat_token");
    expect(typeof orders[0].total_price).toBe("number");
    // BUG: This assertion enforces the wrong type
    const total = calculateOrderTotal(orders);
    expect(total).toBeGreaterThan(0);
  });
});
