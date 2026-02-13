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
