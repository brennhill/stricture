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
