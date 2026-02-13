async function listProducts(
  token: string,
  limit: number = 500 // BUG: Default exceeds Shopify max of 250
): Promise<ShopifyProduct[]> {
  // BUG: No range validation on limit parameter
  const res = await fetch(
    `${SHOPIFY_BASE}/admin/api/2024-01/products.json?limit=${limit}`,
    { headers: { "X-Shopify-Access-Token": token } }
  );
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  const body = await res.json();
  return body.products;
}

async function createFulfillment(
  orderId: number,
  lineItems: Array<{ id: number; quantity: number }>,
  token: string
): Promise<ShopifyFulfillment> {
  // BUG: No range validation on quantity
  // Manifest says range: [1, 99999] but client accepts any number
  const res = await fetch(
    `${SHOPIFY_BASE}/admin/api/2024-01/orders/${orderId}/fulfillments.json`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Shopify-Access-Token": token,
      },
      body: JSON.stringify({
        fulfillment: {
          location_id: 1,
          line_items: lineItems,
        },
      }),
    }
  );
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  const body = await res.json();
  return body.fulfillment;
}

describe("listProducts", () => {
  it("lists products with custom limit", async () => {
    // BUG: Test uses limit=500, above the 250 max
    const products = await listProducts("shpat_token", 500);
    expect(products.length).toBeLessThanOrEqual(500);
    // Assertion is wrong -- Shopify caps at 250, so this always passes trivially
  });
});

describe("createFulfillment", () => {
  it("creates fulfillment with quantity", async () => {
    // BUG: No test for quantity=0, negative, or exceeding 99999
    const result = await createFulfillment(
      123,
      [{ id: 1, quantity: 5 }],
      "shpat_token"
    );
    expect(result.id).toBeGreaterThan(0);
  });
});
