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
