// src/services/order-service.ts
export interface OrderItem {
  productId: string;
  quantity: number;
  price: number;
  discount?: number;
}

export function getOrderItems(orderId: string): OrderItem[] {
  return database.fetchOrderItems(orderId);
}
