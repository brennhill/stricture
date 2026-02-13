async function listProducts(token: string): Promise<ShopifyProduct[]> {
  const res = await fetch(
    `${SHOPIFY_BASE}/admin/api/2024-01/products.json`,
    { headers: { "X-Shopify-Access-Token": token } }
  );
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  const body = await res.json();
  return body.products;
}

describe("listProducts", () => {
  it("lists products", async () => {
    const products = await listProducts("shpat_token");
    // BUG: Shallow assertions -- proves nothing about product shape
    expect(products).toBeDefined();
    expect(products).toBeTruthy();
    expect(Array.isArray(products)).toBe(true);
  });

  it("products have data", async () => {
    const products = await listProducts("shpat_token");
    if (products.length > 0) {
      const product = products[0];
      // BUG: Only checks existence, not types or values
      expect(product.id).toBeDefined();
      expect(product.title).toBeDefined();
      expect(product.variants).toBeDefined();
      expect(product.status).toBeTruthy();
    }
  });
});
