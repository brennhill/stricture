// DEFECT B11: Uses camelCase instead of matching Go json tags
export interface Order {
  id: string;
  customer_id: string;  // CORRECT
  total_amount: number; // CORRECT
  status: OrderStatus;
  note: string | null;
  createdAt: string;    // BUG: Go sends "created_at" but TS expects "createdAt"
  updatedAt: string;    // BUG: Go sends "updated_at" but TS expects "updatedAt"
  version: number;
}

// Client code crashes
async function displayOrder(order: Order) {
  console.log(order.createdAt);  // undefined — field doesn't exist in JSON
  // Should access order.created_at instead
}
