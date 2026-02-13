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
