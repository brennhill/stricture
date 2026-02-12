# 06 — Shopify Admin REST API

**API:** Shopify Admin REST API (2024-01)
**Why included:** Monetary values as strings (not floats), nullable fulfillment status, webhook HMAC verification, versioned URL paths, Link-header pagination, inventory race conditions.

---

## Manifest Fragment

```yaml
contracts:
  - id: "shopify-products"
    producer: shopify
    consumers: [my-storefront]
    protocol: http
    auth:
      header: X-Shopify-Access-Token
      format: "shpat_*"
    endpoints:
      - path: "/admin/api/2024-01/products.json"
        method: GET
        request:
          query:
            limit:          { type: integer, range: [1, 250], default: 50 }
            page_info:      { type: string, required: false }
            fields:         { type: string, required: false }
        response:
          fields:
            products:
              type: array
              items:
                fields:
                  id:           { type: integer, required: true }
                  title:        { type: string, required: true }
                  body_html:    { type: string, nullable: true }
                  vendor:       { type: string, required: true }
                  product_type: { type: string, required: true }
                  status:       { type: enum, values: ["active", "draft", "archived"], required: true }
                  variants:
                    type: array
                    items:
                      fields:
                        id:                 { type: integer, required: true }
                        price:              { type: string, format: "decimal", required: true }
                        sku:                { type: string, nullable: true }
                        inventory_quantity: { type: integer, required: true }
                        weight:             { type: number, required: false }
                        weight_unit:        { type: enum, values: ["g", "kg", "oz", "lb"], required: false }
                  images:
                    type: array
                    items:
                      fields:
                        id:       { type: integer, required: true }
                        src:      { type: string, format: url, required: true }
                        alt:      { type: string, nullable: true }
          headers:
            Link: { type: string, format: "pagination", required: false }
          status_codes: [200, 401, 402, 403, 406, 429, 500, 503]

      - path: "/admin/api/2024-01/products.json"
        method: POST
        request:
          fields:
            product:
              type: object
              fields:
                title:        { type: string, required: true }
                body_html:    { type: string, required: false }
                vendor:       { type: string, required: false }
                product_type: { type: string, required: false }
                status:       { type: enum, values: ["active", "draft", "archived"], default: "draft" }
                variants:
                  type: array
                  required: false
                  items:
                    fields:
                      price:              { type: string, format: "decimal", required: true }
                      sku:                { type: string, required: false }
                      inventory_quantity: { type: integer, required: false }
                      weight:             { type: number, required: false }
                      weight_unit:        { type: enum, values: ["g", "kg", "oz", "lb"], required: false }
        response:
          fields:
            product:
              type: object
              fields:
                id:           { type: integer, required: true }
                title:        { type: string, required: true }
                status:       { type: enum, values: ["active", "draft", "archived"], required: true }
                variants:     { type: array, required: true }
          status_codes: [201, 400, 401, 402, 403, 422, 429, 500, 503]

  - id: "shopify-orders"
    producer: shopify
    consumers: [my-storefront]
    protocol: http
    auth:
      header: X-Shopify-Access-Token
      format: "shpat_*"
    endpoints:
      - path: "/admin/api/2024-01/orders.json"
        method: GET
        request:
          query:
            limit:            { type: integer, range: [1, 250], default: 50 }
            status:           { type: enum, values: ["open", "closed", "cancelled", "any"], default: "open" }
            financial_status: { type: enum, values: ["authorized", "pending", "paid", "partially_paid", "refunded", "partially_refunded", "voided", "any"], required: false }
        response:
          fields:
            orders:
              type: array
              items:
                fields:
                  id:                 { type: integer, required: true }
                  order_number:       { type: integer, required: true }
                  financial_status:   { type: enum, values: ["pending", "authorized", "partially_paid", "paid", "partially_refunded", "refunded", "voided"], required: true }
                  fulfillment_status: { type: enum, values: ["fulfilled", "partial", "restocked"], nullable: true, required: true }
                  total_price:        { type: string, format: "decimal", required: true }
                  currency:           { type: string, format: "ISO 4217", required: true }
                  line_items:
                    type: array
                    items:
                      fields:
                        id:         { type: integer, required: true }
                        title:      { type: string, required: true }
                        quantity:   { type: integer, required: true }
                        price:      { type: string, format: "decimal", required: true }
          headers:
            Link: { type: string, format: "pagination", required: false }
          status_codes: [200, 401, 402, 403, 406, 429, 500, 503]

      - path: "/admin/api/2024-01/orders/:order_id/fulfillments.json"
        method: POST
        request:
          fields:
            fulfillment:
              type: object
              fields:
                location_id:    { type: integer, required: true }
                tracking_number: { type: string, required: false }
                tracking_company: { type: string, required: false }
                line_items:
                  type: array
                  required: false
                  items:
                    fields:
                      id:       { type: integer, required: true }
                      quantity: { type: integer, range: [1, 99999], required: true }
        response:
          fields:
            fulfillment:
              type: object
              fields:
                id:              { type: integer, required: true }
                order_id:        { type: integer, required: true }
                status:          { type: enum, values: ["pending", "open", "success", "cancelled", "error", "failure"], required: true }
                tracking_number: { type: string, nullable: true }
          status_codes: [201, 400, 401, 402, 403, 404, 422, 429, 500, 503]

  - id: "shopify-webhooks"
    producer: shopify
    consumers: [my-storefront]
    protocol: http
    auth:
      header: X-Shopify-Hmac-Sha256
      format: "base64(hmac-sha256)"
    verification:
      algorithm: HMAC-SHA256
      secret_source: SHOPIFY_WEBHOOK_SECRET
      header: X-Shopify-Hmac-Sha256
      body: raw

  error_shapes:
    validation:
      fields:
        errors:
          oneOf:
            - type: object
              additionalProperties:
                type: array
                items: { type: string }
            - type: string
    rate_limit:
      retry_after: { header: "Retry-After", type: number }
      bucket: { capacity: 40, leak_rate: "2/second" }
```

---

## Types

```typescript
// ── Shopify shared types ──

interface ShopifyProduct {
  id: number;
  title: string;
  body_html: string | null;
  vendor: string;
  product_type: string;
  status: "active" | "draft" | "archived";
  variants: ShopifyVariant[];
  images: ShopifyImage[];
}

interface ShopifyVariant {
  id: number;
  price: string;           // Decimal as string, e.g. "29.99"
  sku: string | null;
  inventory_quantity: number;
  weight: number | null;
  weight_unit: "g" | "kg" | "oz" | "lb" | null;
}

interface ShopifyImage {
  id: number;
  src: string;
  alt: string | null;
}

interface ShopifyOrder {
  id: number;
  order_number: number;
  financial_status:
    | "pending"
    | "authorized"
    | "partially_paid"
    | "paid"
    | "partially_refunded"
    | "refunded"
    | "voided";
  fulfillment_status: "fulfilled" | "partial" | "restocked" | null;
  total_price: string;     // Decimal as string, e.g. "149.95"
  currency: string;        // ISO 4217, e.g. "USD"
  line_items: ShopifyLineItem[];
}

interface ShopifyLineItem {
  id: number;
  title: string;
  quantity: number;
  price: string;           // Decimal as string
}

interface ShopifyFulfillment {
  id: number;
  order_id: number;
  status: "pending" | "open" | "success" | "cancelled" | "error" | "failure";
  tracking_number: string | null;
}

interface ShopifyValidationError {
  errors: Record<string, string[]> | string;
}

interface ShopifyPaginatedResponse<T> {
  data: T;
  linkHeader: string | null;
}
```

---

## PERFECT — Zero Violations

All Stricture rules pass. Handles string monetary values, nullable fulfillment_status, HMAC webhook verification, full Link-header pagination, all status codes, all enum values, and proper range validation.

```typescript
import crypto from "node:crypto";

// ── Client ──

const SHOPIFY_BASE = "https://my-store.myshopify.com";
const API_VERSION = "2024-01";

function shopifyUrl(path: string): string {
  if (!/^\d{4}-\d{2}$/.test(API_VERSION)) {
    throw new Error(`Invalid API version format: ${API_VERSION}`);
  }
  return `${SHOPIFY_BASE}/admin/api/${API_VERSION}${path}`;
}

async function shopifyFetch(
  path: string,
  options: RequestInit = {},
  token: string
): Promise<Response> {
  const url = shopifyUrl(path);
  const res = await fetch(url, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      "X-Shopify-Access-Token": token,
      ...options.headers,
    },
  });

  if (res.status === 401) {
    throw new Error("Unauthorized: invalid or expired access token");
  }
  if (res.status === 402) {
    throw new Error("Payment required: shop is frozen, pay outstanding charges");
  }
  if (res.status === 403) {
    throw new Error("Forbidden: access token lacks required scope");
  }
  if (res.status === 406) {
    throw new Error("Not Acceptable: request Accept header invalid");
  }
  if (res.status === 429) {
    const retryAfter = res.headers.get("Retry-After");
    const waitMs = retryAfter ? parseFloat(retryAfter) * 1000 : 2000;
    throw new Error(`Rate limited: retry after ${waitMs}ms`);
  }
  if (res.status === 503) {
    throw new Error("Service unavailable: Shopify is temporarily down");
  }
  if (res.status >= 500) {
    throw new Error(`Shopify server error: ${res.status}`);
  }

  return res;
}

// ── List products (paginated) ──

function parseLinkHeader(header: string | null): { next?: string; previous?: string } {
  const result: { next?: string; previous?: string } = {};
  if (!header) return result;

  const parts = header.split(",");
  for (const part of parts) {
    const match = part.match(/<([^>]+)>;\s*rel="(next|previous)"/);
    if (match) {
      const [, url, rel] = match;
      if (rel === "next") result.next = url;
      if (rel === "previous") result.previous = url;
    }
  }
  return result;
}

async function listAllProducts(token: string): Promise<ShopifyProduct[]> {
  const allProducts: ShopifyProduct[] = [];
  let nextUrl: string | undefined = shopifyUrl("/products.json?limit=250");

  while (nextUrl) {
    const res = await fetch(nextUrl, {
      headers: {
        "Content-Type": "application/json",
        "X-Shopify-Access-Token": token,
      },
    });

    if (!res.ok) {
      if (res.status === 401) throw new Error("Unauthorized");
      if (res.status === 402) throw new Error("Shop frozen");
      if (res.status === 429) {
        const retryAfter = res.headers.get("Retry-After");
        const waitMs = retryAfter ? parseFloat(retryAfter) * 1000 : 2000;
        await new Promise((resolve) => setTimeout(resolve, waitMs));
        continue; // Retry same page
      }
      throw new Error(`Failed to list products: ${res.status}`);
    }

    const body = (await res.json()) as { products: ShopifyProduct[] };

    for (const product of body.products) {
      if (!["active", "draft", "archived"].includes(product.status)) {
        throw new Error(`Unknown product status: ${product.status}`);
      }
      for (const variant of product.variants) {
        // Validate price is a valid decimal string
        if (!/^\d+\.\d{2}$/.test(variant.price)) {
          throw new Error(`Invalid price format: ${variant.price}`);
        }
      }
      allProducts.push(product);
    }

    const linkHeader = res.headers.get("Link");
    const links = parseLinkHeader(linkHeader);
    nextUrl = links.next; // Follow pagination until no next link
  }

  return allProducts;
}

// ── Create product ──

interface CreateProductInput {
  title: string;
  body_html?: string;
  vendor?: string;
  product_type?: string;
  status?: "active" | "draft" | "archived";
  variants?: Array<{
    price: string;
    sku?: string;
    inventory_quantity?: number;
    weight?: number;
    weight_unit?: "g" | "kg" | "oz" | "lb";
  }>;
}

async function createProduct(
  input: CreateProductInput,
  token: string
): Promise<ShopifyProduct> {
  if (!input.title || input.title.trim().length === 0) {
    throw new Error("Product title is required");
  }

  if (input.variants) {
    for (const variant of input.variants) {
      if (!/^\d+\.\d{2}$/.test(variant.price)) {
        throw new Error(`Invalid variant price format: ${variant.price}`);
      }
      if (variant.weight_unit && !["g", "kg", "oz", "lb"].includes(variant.weight_unit)) {
        throw new Error(`Invalid weight unit: ${variant.weight_unit}`);
      }
    }
  }

  const res = await shopifyFetch("/products.json", {
    method: "POST",
    body: JSON.stringify({ product: input }),
  }, token);

  if (res.status === 422) {
    const error = (await res.json()) as ShopifyValidationError;
    if (typeof error.errors === "string") {
      throw new Error(`Validation error: ${error.errors}`);
    }
    const messages = Object.entries(error.errors)
      .map(([field, msgs]) => `${field}: ${msgs.join(", ")}`)
      .join("; ");
    throw new Error(`Validation error: ${messages}`);
  }
  if (res.status === 400) {
    throw new Error("Bad request: malformed product data");
  }
  if (res.status !== 201) {
    throw new Error(`Unexpected status creating product: ${res.status}`);
  }

  const body = (await res.json()) as { product: ShopifyProduct };
  return body.product;
}

// ── List orders ──

async function listAllOrders(token: string): Promise<ShopifyOrder[]> {
  const allOrders: ShopifyOrder[] = [];
  let nextUrl: string | undefined = shopifyUrl("/orders.json?limit=250&status=any");

  while (nextUrl) {
    const res = await fetch(nextUrl, {
      headers: {
        "Content-Type": "application/json",
        "X-Shopify-Access-Token": token,
      },
    });

    if (!res.ok) {
      if (res.status === 401) throw new Error("Unauthorized");
      if (res.status === 402) throw new Error("Shop frozen");
      if (res.status === 429) {
        const retryAfter = res.headers.get("Retry-After");
        const waitMs = retryAfter ? parseFloat(retryAfter) * 1000 : 2000;
        await new Promise((resolve) => setTimeout(resolve, waitMs));
        continue;
      }
      throw new Error(`Failed to list orders: ${res.status}`);
    }

    const body = (await res.json()) as { orders: ShopifyOrder[] };

    for (const order of body.orders) {
      // Validate financial_status enum exhaustively
      switch (order.financial_status) {
        case "pending":
        case "authorized":
        case "partially_paid":
        case "paid":
        case "partially_refunded":
        case "refunded":
        case "voided":
          break;
        default: {
          const exhaustive: never = order.financial_status;
          throw new Error(`Unknown financial_status: ${exhaustive}`);
        }
      }

      // fulfillment_status is nullable -- handle null safely
      if (order.fulfillment_status !== null) {
        switch (order.fulfillment_status) {
          case "fulfilled":
          case "partial":
          case "restocked":
            break;
          default: {
            const exhaustive: never = order.fulfillment_status;
            throw new Error(`Unknown fulfillment_status: ${exhaustive}`);
          }
        }
      }

      // total_price is a string decimal -- compare as strings, never parseFloat
      if (!/^\d+\.\d{2}$/.test(order.total_price)) {
        throw new Error(`Invalid total_price format: ${order.total_price}`);
      }

      allOrders.push(order);
    }

    const linkHeader = res.headers.get("Link");
    const links = parseLinkHeader(linkHeader);
    nextUrl = links.next;
  }

  return allOrders;
}

// ── Create fulfillment ──

interface CreateFulfillmentInput {
  location_id: number;
  tracking_number?: string;
  tracking_company?: string;
  line_items?: Array<{ id: number; quantity: number }>;
}

async function createFulfillment(
  orderId: number,
  input: CreateFulfillmentInput,
  token: string
): Promise<ShopifyFulfillment> {
  if (input.line_items) {
    for (const item of input.line_items) {
      if (item.quantity < 1 || item.quantity > 99999) {
        throw new Error(`Line item quantity out of range: ${item.quantity}`);
      }
    }
  }

  const res = await shopifyFetch(`/orders/${orderId}/fulfillments.json`, {
    method: "POST",
    body: JSON.stringify({ fulfillment: input }),
  }, token);

  if (res.status === 422) {
    const error = (await res.json()) as ShopifyValidationError;
    if (typeof error.errors === "string") {
      throw new Error(`Fulfillment error: ${error.errors}`);
    }
    const messages = Object.entries(error.errors)
      .map(([field, msgs]) => `${field}: ${msgs.join(", ")}`)
      .join("; ");
    throw new Error(`Fulfillment validation error: ${messages}`);
  }
  if (res.status === 404) {
    throw new Error(`Order not found: ${orderId}`);
  }
  if (res.status === 400) {
    throw new Error("Bad request: malformed fulfillment data");
  }
  if (res.status !== 201) {
    throw new Error(`Unexpected status creating fulfillment: ${res.status}`);
  }

  const body = (await res.json()) as { fulfillment: ShopifyFulfillment };
  return body.fulfillment;
}

// ── Webhook verification ──

function verifyShopifyWebhook(
  rawBody: Buffer,
  hmacHeader: string,
  secret: string
): boolean {
  const computed = crypto
    .createHmac("sha256", secret)
    .update(rawBody)
    .digest("base64");
  return crypto.timingSafeEqual(
    Buffer.from(computed),
    Buffer.from(hmacHeader)
  );
}

async function handleWebhook(
  rawBody: Buffer,
  headers: Record<string, string>,
  secret: string
): Promise<void> {
  const hmac = headers["x-shopify-hmac-sha256"];
  if (!hmac) {
    throw new Error("Missing X-Shopify-Hmac-Sha256 header");
  }

  if (!verifyShopifyWebhook(rawBody, hmac, secret)) {
    throw new Error("Invalid webhook signature");
  }

  const payload = JSON.parse(rawBody.toString("utf-8"));
  // Process verified webhook payload
  await processWebhookEvent(payload);
}

async function processWebhookEvent(payload: unknown): Promise<void> {
  // Placeholder for webhook processing logic
}

// ── Monetary comparison utility ──

function compareDecimalStrings(a: string, b: string): number {
  // Compare decimal strings without floating-point conversion
  const [aInt, aDec = "00"] = a.split(".");
  const [bInt, bDec = "00"] = b.split(".");
  const aWhole = BigInt(aInt);
  const bWhole = BigInt(bInt);
  if (aWhole !== bWhole) return aWhole < bWhole ? -1 : 1;
  const aPadded = aDec.padEnd(10, "0");
  const bPadded = bDec.padEnd(10, "0");
  if (aPadded < bPadded) return -1;
  if (aPadded > bPadded) return 1;
  return 0;
}

// ── Tests ──

describe("Shopify Admin API", () => {
  describe("listAllProducts", () => {
    it("returns all products across paginated pages", async () => {
      const products = await listAllProducts("shpat_test_token");

      expect(products.length).toBeGreaterThan(0);
      for (const product of products) {
        expect(typeof product.id).toBe("number");
        expect(typeof product.title).toBe("string");
        expect(["active", "draft", "archived"]).toContain(product.status);
        expect(product.body_html === null || typeof product.body_html === "string").toBe(true);

        for (const variant of product.variants) {
          expect(typeof variant.id).toBe("number");
          expect(typeof variant.price).toBe("string");
          expect(variant.price).toMatch(/^\d+\.\d{2}$/);
          expect(typeof variant.inventory_quantity).toBe("number");
          expect(Number.isInteger(variant.inventory_quantity)).toBe(true);
          if (variant.weight_unit !== null) {
            expect(["g", "kg", "oz", "lb"]).toContain(variant.weight_unit);
          }
        }
      }
    });

    it("handles 401 unauthorized", async () => {
      await expect(listAllProducts("invalid_token")).rejects.toThrow("Unauthorized");
    });

    it("handles 402 shop frozen", async () => {
      await expect(listAllProducts("shpat_frozen_shop")).rejects.toThrow("Shop frozen");
    });

    it("handles 429 rate limit with retry", async () => {
      // Integration test: verify retry loop respects Retry-After header
      const products = await listAllProducts("shpat_rate_limited_token");
      expect(Array.isArray(products)).toBe(true);
    });
  });

  describe("createProduct", () => {
    it("creates a product with valid fields", async () => {
      const product = await createProduct(
        { title: "Test Widget", status: "draft", variants: [{ price: "29.99" }] },
        "shpat_test_token"
      );

      expect(typeof product.id).toBe("number");
      expect(product.title).toBe("Test Widget");
      expect(product.status).toBe("draft");
      expect(Array.isArray(product.variants)).toBe(true);
    });

    it("rejects product without title", async () => {
      await expect(
        createProduct({ title: "" }, "shpat_test_token")
      ).rejects.toThrow("Product title is required");
    });

    it("handles 422 validation error (object form)", async () => {
      await expect(
        createProduct({ title: "x".repeat(1000) }, "shpat_test_token")
      ).rejects.toThrow(/Validation error:/);
    });

    it("handles 422 validation error (string form)", async () => {
      await expect(
        createProduct({ title: "Bad\x00Product" }, "shpat_test_token")
      ).rejects.toThrow(/Validation error:/);
    });

    it("validates variant price format before sending", () => {
      expect(() =>
        createProduct(
          { title: "Test", variants: [{ price: "not-a-price" }] },
          "shpat_test_token"
        )
      ).rejects.toThrow("Invalid variant price format");
    });
  });

  describe("listAllOrders", () => {
    it("returns orders with all financial_status values handled", async () => {
      const orders = await listAllOrders("shpat_test_token");

      for (const order of orders) {
        expect(typeof order.id).toBe("number");
        expect(typeof order.order_number).toBe("number");
        expect([
          "pending", "authorized", "partially_paid", "paid",
          "partially_refunded", "refunded", "voided",
        ]).toContain(order.financial_status);

        // fulfillment_status is nullable
        if (order.fulfillment_status === null) {
          expect(order.fulfillment_status).toBeNull();
        } else {
          expect(["fulfilled", "partial", "restocked"]).toContain(
            order.fulfillment_status
          );
        }

        // total_price is a string decimal
        expect(typeof order.total_price).toBe("string");
        expect(order.total_price).toMatch(/^\d+\.\d{2}$/);
        expect(typeof order.currency).toBe("string");
        expect(order.currency).toMatch(/^[A-Z]{3}$/);
      }
    });
  });

  describe("createFulfillment", () => {
    it("creates a fulfillment with valid data", async () => {
      const fulfillment = await createFulfillment(
        12345,
        { location_id: 1, tracking_number: "1Z999AA10123456784" },
        "shpat_test_token"
      );

      expect(typeof fulfillment.id).toBe("number");
      expect(fulfillment.order_id).toBe(12345);
      expect([
        "pending", "open", "success", "cancelled", "error", "failure",
      ]).toContain(fulfillment.status);
      expect(
        fulfillment.tracking_number === null ||
        typeof fulfillment.tracking_number === "string"
      ).toBe(true);
    });

    it("handles 404 for nonexistent order", async () => {
      await expect(
        createFulfillment(999999, { location_id: 1 }, "shpat_test_token")
      ).rejects.toThrow("Order not found");
    });

    it("validates line item quantity range", async () => {
      await expect(
        createFulfillment(
          12345,
          { location_id: 1, line_items: [{ id: 1, quantity: 0 }] },
          "shpat_test_token"
        )
      ).rejects.toThrow("Line item quantity out of range");
    });
  });

  describe("verifyShopifyWebhook", () => {
    it("accepts valid HMAC signature", () => {
      const secret = "test_secret";
      const body = Buffer.from('{"id":1}');
      const hmac = crypto.createHmac("sha256", secret).update(body).digest("base64");

      expect(verifyShopifyWebhook(body, hmac, secret)).toBe(true);
    });

    it("rejects invalid HMAC signature", () => {
      const body = Buffer.from('{"id":1}');
      expect(verifyShopifyWebhook(body, "invalid_base64", "secret")).toBe(false);
    });

    it("rejects missing HMAC header", async () => {
      await expect(
        handleWebhook(Buffer.from("{}"), {}, "secret")
      ).rejects.toThrow("Missing X-Shopify-Hmac-Sha256 header");
    });
  });

  describe("compareDecimalStrings", () => {
    it("compares string decimals without float conversion", () => {
      expect(compareDecimalStrings("29.99", "29.99")).toBe(0);
      expect(compareDecimalStrings("29.99", "30.00")).toBe(-1);
      expect(compareDecimalStrings("100.00", "99.99")).toBe(1);
      expect(compareDecimalStrings("19.999", "19.998")).toBe(1);
    });
  });
});
```

---

## B01 — No Error Handling

**Bug:** No try/catch or error handling around fetch calls. Network failures, JSON parse errors, and non-200 responses all crash the caller with unhandled exceptions.

**Stricture rule:** `TQ-error-path-coverage`

```typescript
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
```

**Expected violation:** `TQ-error-path-coverage` -- no try/catch around fetch, no .catch() on promise chain. Network errors, DNS failures, and Shopify 5xx responses all produce unhandled rejections.

**Production impact:** Any Shopify outage, rate limit, or authentication failure crashes the calling service. The process may terminate from an unhandled promise rejection. No error telemetry is captured because the error never reaches a handler.

---

## B02 — No Status Code Check

**Bug:** Fetch response is used directly without checking `res.ok` or `res.status`. A 422 validation error or 402 frozen-shop response is parsed as if it were a successful product.

**Stricture rule:** `CTR-status-code-handling`

```typescript
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
```

**Expected violation:** `CTR-status-code-handling` -- manifest declares status codes [201, 400, 401, 402, 403, 422, 429, 500, 503] for POST /products.json but client checks none of them. All responses are treated as 201 success.

**Production impact:** A 422 response body `{"errors":{"title":["is too long"]}}` is assigned to `body.product`, which is `undefined`. Downstream code receives `undefined` instead of a `ShopifyProduct`, causing cascading null reference errors far from the original API call. A 402 frozen-shop response silently corrupts data.

---

## B03 — Shallow Assertions

**Bug:** Tests use `toBeDefined()` and `toBeTruthy()` without validating the actual shape, types, or values of the Shopify response. The test passes even if the implementation returns a completely wrong object.

**Stricture rule:** `TQ-no-shallow-assertions`

```typescript
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
```

**Expected violation:** `TQ-no-shallow-assertions` -- `expect(products).toBeDefined()`, `expect(product.id).toBeDefined()`, etc. are shallow. The function returns `ShopifyProduct[]` with specific fields (id: number, title: string, status: "active"|"draft"|"archived", variants with string prices), but none of the assertions check types, values, or structural invariants.

**Production impact:** If the Shopify API changes its response format or if the parsing logic introduces a bug (e.g., returning `{id: "string"}` instead of `{id: number}`), these tests still pass. The real failure surfaces only in production when downstream code breaks on the wrong types.

---

## B04 — Missing Negative Tests

**Bug:** Tests only cover the happy path (200 success). No tests for 422 validation errors, 402 shop frozen, 429 rate limiting, or any error scenario specific to Shopify.

**Stricture rule:** `TQ-negative-cases`

```typescript
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
```

**Expected violation:** `TQ-negative-cases` -- implementation handles 422, 402, and generic errors, but no test exercises any of these paths. Manifest declares status codes [201, 400, 401, 402, 403, 422, 429, 500, 503]. Tests only cover 201. Zero negative test coverage.

**Production impact:** The 422 error-parsing logic may be broken (e.g., both `errors` shapes -- object and string -- are stringified identically) and nobody discovers it until a real validation failure occurs in production. The 402 shop-frozen path is completely untested.

---

## B05 — Request Missing Required Fields

**Bug:** Client sends a product creation request without the required `title` field. The manifest declares `title` as required for POST /products.json.

**Stricture rule:** `CTR-request-shape`

```typescript
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
```

**Expected violation:** `CTR-request-shape` -- manifest declares `product.title` as `required: true` for POST /admin/api/2024-01/products.json. Client type `CreateProductPayload` marks `title` as optional. The test case invokes `createProduct` without a title, sending `{"product":{"vendor":"ACME"}}`.

**Production impact:** The request reaches Shopify and returns a 422 with `{"errors":{"title":["can't be blank"]}}`. Since the client has no pre-flight validation, every call without a title wastes an API request and counts against the 40 req/s rate limit. If the 422 is not handled (see B02), the error is silently swallowed.

---

## B06 — Response Type Mismatch

**Bug:** Client-side `Order` type is missing the `fulfillment_status` field entirely. The manifest and Shopify API always return this field (as `null` or a string), but the client type does not declare it, so it is silently dropped from typed access.

**Stricture rule:** `CTR-response-shape`

```typescript
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
```

**Expected violation:** `CTR-response-shape` -- manifest declares `fulfillment_status` (enum, nullable, required) and `line_items` (array, required) on the Order response. Client type `Order` is missing both fields. The API always sends them, but the client cannot access them through the type system.

**Production impact:** The `getUnfulfilledOrders` function cannot work. It returns all orders because it has no access to `fulfillment_status`. The fulfillment dashboard shows orders that have already been shipped as needing fulfillment, causing duplicate shipments.

---

## B07 — Wrong Field Types

**Bug:** Client stores `total_price` as a `number` instead of a `string`. Shopify returns prices as string decimals (e.g., `"29.99"`), but the client converts them to floats, introducing floating-point precision errors.

**Stricture rule:** `CTR-manifest-conformance`

```typescript
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
```

**Expected violation:** `CTR-manifest-conformance` -- manifest declares `total_price` as `type: string, format: "decimal"` and `variants[].price` as `type: string, format: "decimal"`. Client types declare them as `number`. The JSON deserialization will actually produce a number (JSON has no string/number distinction for `"29.99"` vs `29.99`), but the Shopify API sends `"29.99"` (quoted string), so `typeof order.total_price` is `"string"` at runtime, contradicting the TypeScript type.

**Production impact:** `calculateOrderTotal` adds string values via `+`, which in JavaScript performs string concatenation: `"29.99" + "15.00"` = `"29.9915.00"`. The "total" becomes a nonsensical concatenated string coerced to NaN. Financial reports show incorrect totals. If the JSON parser does convert to number, floating-point arithmetic introduces rounding errors: `29.99 + 15.01` may equal `45.00000000000001`.

---

## B08 — Incomplete Enum Handling

**Bug:** Client handles only 2 of 7 `financial_status` values ("paid" and "refunded"). Orders with "pending", "authorized", "partially_paid", "partially_refunded", or "voided" status fall through the switch with no handling.

**Stricture rule:** `CTR-strictness-parity`

```typescript
function processOrderByFinancialStatus(order: ShopifyOrder): string {
  // BUG: Only handles 2 of 7 financial_status values
  switch (order.financial_status) {
    case "paid":
      return "ready_to_ship";
    case "refunded":
      return "refund_processed";
    // MISSING: "pending" -- order awaiting payment
    // MISSING: "authorized" -- payment authorized but not captured
    // MISSING: "partially_paid" -- installment or split payment
    // MISSING: "partially_refunded" -- partial refund issued
    // MISSING: "voided" -- authorization voided
    // MISSING: default case
  }
  // BUG: Falls through with undefined return for 5 of 7 values
  return undefined as unknown as string;
}

function categorizeOrder(order: ShopifyOrder): string {
  // BUG: Only checks two fulfillment statuses, ignores null and "restocked"
  if (order.fulfillment_status === "fulfilled") {
    return "complete";
  } else if (order.fulfillment_status === "partial") {
    return "in_progress";
  }
  // MISSING: null (unfulfilled) -- most common state
  // MISSING: "restocked" -- items returned to inventory
  return "unknown";
}

describe("processOrderByFinancialStatus", () => {
  it("handles paid orders", () => {
    const result = processOrderByFinancialStatus({
      financial_status: "paid",
    } as ShopifyOrder);
    expect(result).toBe("ready_to_ship");
  });

  it("handles refunded orders", () => {
    const result = processOrderByFinancialStatus({
      financial_status: "refunded",
    } as ShopifyOrder);
    expect(result).toBe("refund_processed");
  });

  // MISSING: Tests for pending, authorized, partially_paid, partially_refunded, voided
});
```

**Expected violation:** `CTR-strictness-parity` -- manifest declares `financial_status` as enum with 7 values: ["pending", "authorized", "partially_paid", "paid", "partially_refunded", "refunded", "voided"]. Consumer handles only 2 (paid, refunded). Missing: 5 values. Similarly, `fulfillment_status` has 3 values plus null, and consumer handles only 2 of 4 possible states.

**Production impact:** An order with `financial_status: "authorized"` (common in credit card flows where capture is deferred) returns `undefined` from `processOrderByFinancialStatus`. The order status dashboard shows these as blank. Authorized orders are never shipped because they are not recognized as "ready_to_ship". Revenue is delayed until someone manually investigates.

---

## B09 — Missing Range Validation

**Bug:** Client sends `limit=500` in the products query. Shopify's maximum is 250. The API will either reject the request or silently cap at 250, but the client assumes it received 500 products per page and may miscalculate pagination.

**Stricture rule:** `CTR-strictness-parity`

```typescript
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
```

**Expected violation:** `CTR-strictness-parity` -- manifest declares `limit` with range [1, 250] but client defaults to 500 with no bounds checking. Manifest declares `line_items[].quantity` with range [1, 99999] but client performs no validation. Consumer does not enforce ranges that the manifest specifies.

**Production impact:** Sending `limit=500` causes Shopify to silently cap at 250. If the client uses `limit` to calculate total pages (e.g., `totalPages = totalProducts / limit`), pagination is wrong: the client thinks it needs 2 pages for 1000 products but actually needs 4. Half the product catalog is never synced. For quantity, sending `quantity: 0` or negative values causes a 422 that wastes an API call.

---

## B10 — Format Not Validated

**Bug:** API version string in the URL path is not validated as a `YYYY-MM` format. An arbitrary string like `"latest"` or `"v2"` is inserted into the URL, producing a 404 from Shopify.

**Stricture rule:** `CTR-strictness-parity`

```typescript
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
```

**Expected violation:** `CTR-strictness-parity` -- manifest declares API version as path parameter with format `YYYY-MM` (e.g., "2024-01"). Client accepts any string without regex validation. Manifest declares auth token format as `"shpat_*"`. Client accepts any string.

**Production impact:** A configuration error sets `version` to `"latest"` (common in other APIs). The constructed URL becomes `/admin/api/latest/products.json`, which Shopify does not recognize, returning 404. The error message is cryptic ("HTTP 404") with no indication that the version format is wrong. With unvalidated tokens, a missing `shpat_` prefix causes 401 errors that are hard to debug.

---

## B11 — Precision Loss

**Bug:** Variant prices (Shopify strings like `"19.999"`) are compared using `parseFloat`, introducing floating-point precision errors. Price comparisons and totals produce incorrect results.

**Stricture rule:** `CTR-strictness-parity`

```typescript
function findCheapestVariant(variants: ShopifyVariant[]): ShopifyVariant | null {
  if (variants.length === 0) return null;

  let cheapest = variants[0];
  for (let i = 1; i < variants.length; i++) {
    // BUG: parseFloat on string prices introduces floating-point errors
    if (parseFloat(variants[i].price) < parseFloat(cheapest.price)) {
      cheapest = variants[i];
    }
  }
  return cheapest;
}

function calculateTotalRevenue(orders: ShopifyOrder[]): number {
  // BUG: Summing parseFloat results accumulates floating-point errors
  return orders.reduce((sum, order) => {
    return sum + parseFloat(order.total_price);
  }, 0);
}

function isPriceEqual(a: string, b: string): boolean {
  // BUG: parseFloat("19.999") === 19.999 but
  // parseFloat("0.1") + parseFloat("0.2") !== parseFloat("0.3")
  return parseFloat(a) === parseFloat(b);
}

describe("findCheapestVariant", () => {
  it("finds cheapest", () => {
    const variants = [
      { id: 1, price: "19.99" } as ShopifyVariant,
      { id: 2, price: "9.99" } as ShopifyVariant,
    ];
    const cheapest = findCheapestVariant(variants);
    expect(cheapest?.id).toBe(2);
  });

  // BUG: No test for prices that trigger float precision issues
  // e.g., "0.10" + "0.20" !== "0.30" in float arithmetic
});

describe("calculateTotalRevenue", () => {
  it("sums order totals", () => {
    const orders = [
      { total_price: "10.00" } as ShopifyOrder,
      { total_price: "20.00" } as ShopifyOrder,
    ];
    const total = calculateTotalRevenue(orders);
    // BUG: Uses float comparison -- works for these values but breaks for others
    expect(total).toBe(30);
  });
});
```

**Expected violation:** `CTR-strictness-parity` -- manifest declares `price` and `total_price` as `type: string, format: "decimal"`. Client uses `parseFloat()` to convert strings to IEEE 754 floats for comparison and arithmetic. This violates the manifest's string-decimal contract. Decimal string operations must preserve exact precision.

**Production impact:** For 1,000 orders with `total_price: "19.99"`, `calculateTotalRevenue` returns `19989.999999999996` instead of `19990.00`. Financial reports are off by fractions of a cent, which compounds across millions of transactions. `isPriceEqual("0.30", "0.30")` works, but summing `"0.10" + "0.20"` and comparing to `"0.30"` fails because `0.1 + 0.2 = 0.30000000000000004` in IEEE 754.

---

## B12 — Nullable Field Crash

**Bug:** Code calls `.toLowerCase()` on `order.fulfillment_status`, which is `null` for unfulfilled orders. Shopify returns `fulfillment_status: null` when no fulfillment has been created, which is the default state for new orders.

**Stricture rule:** `CTR-response-shape`

```typescript
function getFulfillmentLabel(order: ShopifyOrder): string {
  // BUG: fulfillment_status is null for unfulfilled orders
  // .toLowerCase() on null throws: "Cannot read properties of null"
  const status = order.fulfillment_status.toLowerCase();

  switch (status) {
    case "fulfilled":
      return "Shipped";
    case "partial":
      return "Partially Shipped";
    case "restocked":
      return "Returned";
    default:
      return "Unknown";
  }
}

function getOrderSummary(order: ShopifyOrder): string {
  // BUG: String interpolation with null calls .toString() implicitly
  // but template literal with null property access crashes
  return `Order #${order.order_number}: ${order.fulfillment_status.toUpperCase()} - ${order.total_price} ${order.currency}`;
}

function groupOrdersByFulfillment(
  orders: ShopifyOrder[]
): Record<string, ShopifyOrder[]> {
  const groups: Record<string, ShopifyOrder[]> = {};
  for (const order of orders) {
    // BUG: null.toString() throws -- unfulfilled orders crash the loop
    const key = order.fulfillment_status.toString();
    if (!groups[key]) groups[key] = [];
    groups[key].push(order);
  }
  return groups;
}

describe("getFulfillmentLabel", () => {
  it("labels fulfilled orders", () => {
    const label = getFulfillmentLabel({
      fulfillment_status: "fulfilled",
    } as ShopifyOrder);
    expect(label).toBe("Shipped");
  });

  // BUG: No test for fulfillment_status === null (the most common case)
});
```

**Expected violation:** `CTR-response-shape` -- manifest declares `fulfillment_status` as `nullable: true`. Client code calls `.toLowerCase()`, `.toUpperCase()`, and `.toString()` on this field without null checks. Three functions crash on the most common order state (unfulfilled, where `fulfillment_status` is `null`).

**Production impact:** New orders always have `fulfillment_status: null`. The order listing page crashes with `TypeError: Cannot read properties of null (reading 'toLowerCase')` as soon as any unfulfilled order exists. This is not an edge case -- it is the default state. The entire orders dashboard is broken for any shop that has not fulfilled 100% of its orders.

---

## B13 — Missing Webhook HMAC Verification

**Bug:** Webhook handler processes the Shopify payload without verifying the `X-Shopify-Hmac-Sha256` signature. Any HTTP client can forge webhook payloads and trigger order processing, inventory updates, or refunds.

**Stricture rule:** `CTR-request-shape`

```typescript
interface WebhookPayload {
  id: number;
  topic: string;
  [key: string]: unknown;
}

async function handleShopifyWebhook(
  req: { body: string; headers: Record<string, string> }
): Promise<void> {
  // BUG: No HMAC signature verification
  // Anyone can POST to this endpoint and forge webhook events

  // Directly parse and process the payload
  const payload: WebhookPayload = JSON.parse(req.body);

  switch (req.headers["x-shopify-topic"]) {
    case "orders/create":
      await processNewOrder(payload);
      break;
    case "orders/paid":
      await processPayment(payload);
      break;
    case "products/update":
      await syncProduct(payload);
      break;
    case "app/uninstalled":
      await handleUninstall(payload);
      break;
    default:
      console.log(`Unhandled webhook topic: ${req.headers["x-shopify-topic"]}`);
  }
}

async function processNewOrder(payload: WebhookPayload): Promise<void> {
  // Creates order in local database based on unverified payload
  await db.createOrder(payload);
}

async function processPayment(payload: WebhookPayload): Promise<void> {
  // Marks order as paid based on unverified payload
  await db.updateOrderStatus(payload.id, "paid");
}

async function syncProduct(payload: WebhookPayload): Promise<void> {
  await db.upsertProduct(payload);
}

async function handleUninstall(payload: WebhookPayload): Promise<void> {
  // Deletes all shop data based on unverified payload
  await db.deleteShopData(payload.id);
}

// Placeholder DB
const db = {
  createOrder: async (_p: unknown) => {},
  updateOrderStatus: async (_id: number, _status: string) => {},
  upsertProduct: async (_p: unknown) => {},
  deleteShopData: async (_id: number) => {},
};

describe("handleShopifyWebhook", () => {
  it("processes order creation webhook", async () => {
    await handleShopifyWebhook({
      body: JSON.stringify({ id: 1, topic: "orders/create" }),
      headers: { "x-shopify-topic": "orders/create" },
    });
    // BUG: Test does not include or verify HMAC signature
    // This test passes with forged payloads
  });
});
```

**Expected violation:** `CTR-request-shape` -- manifest declares webhook auth via `X-Shopify-Hmac-Sha256` header with `HMAC-SHA256` verification algorithm. Handler does not read or verify this header. The request shape contract requires signature verification before payload processing.

**Production impact:** An attacker can POST `{"id":1}` to the webhook endpoint with `x-shopify-topic: app/uninstalled` and trigger `deleteShopData`, wiping the entire shop's local database. Forged `orders/paid` webhooks can mark orders as paid without actual payment, causing products to ship without payment. This is a critical security vulnerability. Shopify's documentation explicitly warns: "You should always verify webhooks."

---

## B14 — Pagination Terminated Early

**Bug:** Client fetches only the first page of products and ignores the `Link` header. For shops with more than 250 products, this returns an incomplete catalog.

**Stricture rule:** `CTR-response-shape`

```typescript
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
```

**Expected violation:** `CTR-response-shape` -- manifest declares `Link` header with `format: "pagination"` in the response. Client reads the response body but never reads `res.headers.get("Link")`. The `rel="next"` link is discarded, so subsequent pages are never fetched. The response contract includes pagination metadata that the client ignores.

**Production impact:** A shop with 10,000 products syncs only the first 50 to the local catalog. The product search shows 50 results. Customers cannot find 99.5% of the inventory. The `countTotalProducts` function reports 50, leading the shop owner to believe their catalog is nearly empty. This bug is invisible in development with small test datasets but catastrophic in production.

---

## B15 — Race Condition

**Bug:** Inventory update reads `inventory_quantity`, decrements it, and writes back without any concurrency control. Two concurrent orders for the last item both read quantity=1, both decrement to 0, and both succeed -- overselling the product.

**Stricture rule:** `CTR-request-shape`

```typescript
async function processOrder(
  orderId: number,
  lineItems: Array<{ variant_id: number; quantity: number }>,
  token: string
): Promise<void> {
  for (const item of lineItems) {
    // Step 1: Read current inventory
    const res = await fetch(
      `${SHOPIFY_BASE}/admin/api/2024-01/variants/${item.variant_id}.json`,
      { headers: { "X-Shopify-Access-Token": token } }
    );
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const { variant } = (await res.json()) as { variant: ShopifyVariant };

    // BUG: Race condition -- another request can modify inventory between read and write
    const currentQty = variant.inventory_quantity;
    const newQty = currentQty - item.quantity;

    if (newQty < 0) {
      throw new Error(
        `Insufficient inventory for variant ${item.variant_id}: have ${currentQty}, need ${item.quantity}`
      );
    }

    // Step 2: Write updated inventory (no optimistic locking, no atomic operation)
    const updateRes = await fetch(
      `${SHOPIFY_BASE}/admin/api/2024-01/variants/${item.variant_id}.json`,
      {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          "X-Shopify-Access-Token": token,
        },
        // BUG: Sets absolute inventory -- concurrent reads both see quantity=1
        // Both set newQty=0, both succeed. Actual inventory should be -1.
        body: JSON.stringify({ variant: { inventory_quantity: newQty } }),
      }
    );
    if (!updateRes.ok) throw new Error(`HTTP ${updateRes.status}`);
  }
}

// BUG: No use of Shopify's inventory_levels/adjust.json endpoint
// which provides atomic inventory adjustments
// Correct approach: POST /admin/api/2024-01/inventory_levels/adjust.json
// { "location_id": 123, "inventory_item_id": 456, "available_adjustment": -1 }

describe("processOrder", () => {
  it("decrements inventory", async () => {
    // BUG: Test runs sequentially -- cannot detect concurrency issues
    await processOrder(
      1,
      [{ variant_id: 100, quantity: 1 }],
      "shpat_token"
    );
    // No assertion that inventory was actually decremented atomically
  });

  // MISSING: No concurrent test
  // MISSING: No test for two simultaneous orders on last item
});
```

**Expected violation:** `CTR-request-shape` -- the read-modify-write pattern on `inventory_quantity` lacks atomicity. The manifest declares `inventory_quantity` as `type: integer` on the variant. The client reads this value, performs arithmetic client-side, and writes back the absolute result. No optimistic locking (ETag/If-Match), no atomic adjustment API, and no version field is checked between read and write. Shopify provides `POST /inventory_levels/adjust.json` for atomic adjustments, but the client uses the non-atomic PUT variant endpoint.

**Production impact:** During a flash sale, two concurrent requests both read `inventory_quantity: 1`. Both compute `newQty = 0`. Both PUT `inventory_quantity: 0`. Both orders succeed, but only one item exists. The customer who receives nothing files a chargeback. Over time, inventory counts drift further from reality. The shop oversells systematically during high-traffic periods. Shopify's own documentation recommends using the `inventory_levels/adjust` endpoint with relative adjustments to avoid this exact problem.
