async function listProducts(token: string): Promise<ShopifyProduct[]> {
  const res = await fetch(
    `${SHOPIFY_BASE}/admin/api/2024-01/products.json?limit=50`,
    { headers: { "X-Shopify-Access-Token": token } }
  );
  const body = await res.json();
  return body.products;
}

async function createProduct(
  input: { title: string },
  token: string
): Promise<ShopifyProduct> {
  const res = await fetch(
    `${SHOPIFY_BASE}/admin/api/2024-01/products.json`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Shopify-Access-Token": token,
      },
      body: JSON.stringify({ product: input }),
    }
  );
  const body = await res.json();
  return body.product;
}

// Tests
describe("listProducts", () => {
  it("lists products", async () => {
    const products = await listProducts("shpat_token");
    expect(products.length).toBeGreaterThan(0);
  });
});
