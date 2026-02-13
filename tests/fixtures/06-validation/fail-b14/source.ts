async function listProducts(token: string): Promise<ShopifyProduct[]> {
  const res = await fetch(
    `${SHOPIFY_BASE}/admin/api/2024-01/products.json?limit=50`,
    { headers: { "X-Shopify-Access-Token": token } }
  );
  if (!res.ok) throw new Error(`HTTP ${res.status}`);

  // BUG: Ignores Link header -- only returns first page
  // Shopify paginates via Link header with rel="next"
  // A shop with 500 products only gets the first 50
  const body = await res.json();
  return body.products;
}

async function listOrders(token: string): Promise<ShopifyOrder[]> {
  const res = await fetch(
    `${SHOPIFY_BASE}/admin/api/2024-01/orders.json?limit=50`,
    { headers: { "X-Shopify-Access-Token": token } }
  );
  if (!res.ok) throw new Error(`HTTP ${res.status}`);

  // BUG: Same pagination issue -- only first 50 orders returned
  const body = await res.json();
  return body.orders;
}

function countTotalProducts(products: ShopifyProduct[]): number {
  // This count is wrong -- it only reflects the first page
  return products.length;
}

describe("listProducts", () => {
  it("lists products", async () => {
    const products = await listProducts("shpat_token");
    expect(Array.isArray(products)).toBe(true);
    // BUG: This test passes even if only 50 of 500 products are returned
    expect(products.length).toBeGreaterThan(0);
    expect(products.length).toBeLessThanOrEqual(50);
    // The assertion actually CONFIRMS the bug -- it expects at most 50
  });
});

describe("countTotalProducts", () => {
  it("counts products", async () => {
    const products = await listProducts("shpat_token");
    const count = countTotalProducts(products);
    // BUG: Test doesn't know the real total, so it can't detect missing products
    expect(count).toBeGreaterThan(0);
  });
});
