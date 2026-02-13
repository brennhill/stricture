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
