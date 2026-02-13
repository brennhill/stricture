// DEFECT B06: Missing total_amount field
export interface Order {
  id: string;
  customer_id: string;
  // BUG: total_amount field missing — Go sends it but TS doesn't declare it
  status: OrderStatus;
  note: string | null;
  created_at: string;
  updated_at: string;
  version: number;
}

// Client code will crash
function calculateTax(order: Order): number {
  return order.total_amount * 0.08;  // TypeScript error: Property does not exist
}
