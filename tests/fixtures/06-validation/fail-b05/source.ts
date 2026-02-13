interface CreateProductPayload {
  // BUG: title is optional here, but manifest says required
  title?: string;
  vendor?: string;
  product_type?: string;
  status?: string;
}

async function createProduct(
  input: CreateProductPayload,
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
      // BUG: No validation that title is present before sending
      body: JSON.stringify({ product: input }),
    }
  );
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  const body = await res.json();
  return body.product;
}

describe("createProduct", () => {
  it("creates a product", async () => {
    // BUG: Calling without title -- no client-side guard
    const product = await createProduct({ vendor: "ACME" }, "shpat_token");
    expect(product).toBeDefined();
  });
});
