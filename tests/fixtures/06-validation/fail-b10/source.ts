// BUG: API version is accepted as any string, no YYYY-MM format validation
function shopifyUrl(version: string, path: string): string {
  // No format validation on version
  return `${SHOPIFY_BASE}/admin/api/${version}${path}`;
}

async function listProducts(
  version: string,
  token: string
): Promise<ShopifyProduct[]> {
  // BUG: version could be "latest", "v2", "", or any invalid string
  const url = shopifyUrl(version, "/products.json");
  const res = await fetch(url, {
    headers: { "X-Shopify-Access-Token": token },
  });
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  const body = await res.json();
  return body.products;
}

// BUG: Access token format not validated
async function shopifyFetch(
  path: string,
  token: string // No validation that token starts with "shpat_"
): Promise<Response> {
  return fetch(`${SHOPIFY_BASE}/admin/api/2024-01${path}`, {
    headers: { "X-Shopify-Access-Token": token },
  });
}

describe("listProducts", () => {
  it("lists products with version", async () => {
    // BUG: No test that version format is validated
    const products = await listProducts("2024-01", "shpat_token");
    expect(Array.isArray(products)).toBe(true);
  });

  // MISSING: No test for invalid version format
  // MISSING: No test for token format validation
});
