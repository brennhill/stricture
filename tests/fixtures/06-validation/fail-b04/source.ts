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

  if (res.status === 422) {
    const err = await res.json();
    throw new Error(`Validation: ${JSON.stringify(err.errors)}`);
  }
  if (res.status === 402) throw new Error("Shop frozen");
  if (!res.ok) throw new Error(`HTTP ${res.status}`);

  const body = await res.json();
  return body.product;
}

// BUG: Only happy-path tests. No 422, 402, 429, 401, 400 tests.
describe("createProduct", () => {
  it("creates a product successfully", async () => {
    const product = await createProduct({ title: "Widget" }, "shpat_token");
    expect(product.id).toBeGreaterThan(0);
    expect(product.title).toBe("Widget");
  });

  it("creates a product with variants", async () => {
    const product = await createProduct(
      { title: "Gadget" },
      "shpat_token"
    );
    expect(product.title).toBe("Gadget");
  });

  // MISSING: No test for 422 validation error (object form)
  // MISSING: No test for 422 validation error (string form)
  // MISSING: No test for 402 shop frozen
  // MISSING: No test for 429 rate limit
  // MISSING: No test for 401 unauthorized
  // MISSING: No test for empty title
});
