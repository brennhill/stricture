// types.ts — Type definitions matching Go server EXACTLY.

export type OrderStatus =
  | "pending"
  | "confirmed"
  | "processing"
  | "shipped"
  | "delivered";

export interface Order {
  id: string;
  customer_id: string;        // CORRECT: matches Go json tag
  total_amount: number;        // CORRECT: number type (cents)
  status: OrderStatus;
  note: string | null;         // CORRECT: nullable
  created_at: string;          // CORRECT: ISO 8601 timestamp
  updated_at: string;
  version: number;
}

export interface CreateOrderRequest {
  customer_id: string;         // CORRECT: snake_case
  total_amount: number;
  note?: string | null;
}

export interface UpdateStatusRequest {
  status: OrderStatus;
  version: number;             // CORRECT: optimistic locking
}

export interface ListOrdersResponse {
  orders: Order[];
  next_cursor?: string | null; // CORRECT: cursor pagination
}

export class OrderAPIError extends Error {
  constructor(
    message: string,
    public status: number,
    public body: unknown
  ) {
    super(message);
    this.name = "OrderAPIError";
  }
}
