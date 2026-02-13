// DEFECT B07: Treats total_amount as string when Go sends number
export interface Order {
  id: string;
  customer_id: string;
  total_amount: string;  // BUG: Go sends number, TS expects string
  status: OrderStatus;
  // ...
}

// Client code crashes
async function displayPrice(order: Order) {
  // BUG: order.total_amount is number from JSON, but TS thinks it's string
  const dollars = order.total_amount / 100;  // Runtime error: NaN
  console.log(`$${dollars.toFixed(2)}`);     // "NaN"
}
