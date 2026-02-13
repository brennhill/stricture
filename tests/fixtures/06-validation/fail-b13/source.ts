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
