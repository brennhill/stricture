// order-service.ts -- Order business logic.
// V02: Service layer imports from route layer (reverse direction).

import type { Order, CreateOrderInput, OrderFilters, PaginatedResult, OrderStatus } from "../models/order";
import * as orderRepo from "../repositories/order-repo";
import * as userRepo from "../repositories/user-repo";
import { ValidationError } from "../shared/errors";
import { logger } from "../shared/logger";
import { orderRouter } from "../routes/orders";  // <-- VIOLATION: importing from route layer

const VALID_TRANSITIONS: Record<OrderStatus, OrderStatus[]> = {
  pending: ["confirmed", "cancelled"],
  confirmed: ["processing", "cancelled"],
  processing: ["shipped", "cancelled"],
  shipped: ["delivered"],
  delivered: ["refunded"],
  cancelled: [],
  refunded: [],
};

export async function createOrder(input: CreateOrderInput): Promise<Order> {
  await userRepo.findUserById(input.userId);

  if (input.items.length === 0) {
    throw new ValidationError("Order must contain at least one item");
  }

  for (const item of input.items) {
    if (item.quantity < 1) {
      throw new ValidationError(`Invalid quantity for product ${item.productId}`);
    }
    if (item.unitPriceCents < 0) {
      throw new ValidationError(`Invalid price for product ${item.productId}`);
    }
  }

  const order = await orderRepo.createOrder(input);

  // BUG: The service tries to register a webhook on the router.
  // This creates a reverse dependency: service -> route.
  // In practice, this would likely be done via an event emitter or
  // a callback pattern, not by importing the router.
  orderRouter.post(`/${order.id}/webhook`, (_req, res) => {
    res.status(200).send("OK");
  });

  return order;
}

export async function getOrderById(id: string): Promise<Order> {
  return orderRepo.findOrderById(id);
}

export async function transitionOrderStatus(
  orderId: string,
  newStatus: OrderStatus
): Promise<Order> {
  const order = await orderRepo.findOrderById(orderId);
  const allowed = VALID_TRANSITIONS[order.status];

  if (!allowed.includes(newStatus)) {
    throw new ValidationError(
      `Cannot transition from ${order.status} to ${newStatus}. ` +
      `Allowed transitions: ${allowed.join(", ") || "none"}`
    );
  }

  return orderRepo.updateOrderStatus(orderId, newStatus);
}
