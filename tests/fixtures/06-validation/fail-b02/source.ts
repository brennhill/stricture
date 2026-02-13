async function createProduct(
  input: { title: string },
  token: string
): Promise<ShopifyProduct> {
  try {
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
    // BUG: No status code check -- 422, 402, 429 all parsed as product
    const body = await res.json();
    return body.product;
  } catch (err) {
    throw new Error(`Failed to create product: ${err}`);
  }
}

async function listOrders(token: string): Promise<ShopifyOrder[]> {
  try {
    const res = await fetch(
      `${SHOPIFY_BASE}/admin/api/2024-01/orders.json`,
      { headers: { "X-Shopify-Access-Token": token } }
    );
    // BUG: No status code check
    const body = await res.json();
    return body.orders;
  } catch (err) {
    throw new Error(`Failed to list orders: ${err}`);
  }
}

describe("createProduct", () => {
  it("creates product", async () => {
    const product = await createProduct({ title: "Widget" }, "shpat_token");
    expect(product.title).toBe("Widget");
  });
});
