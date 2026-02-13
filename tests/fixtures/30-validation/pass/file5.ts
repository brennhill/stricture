// order-service.ts -- Order business logic.
// V04: Creates a circular dependency with user-service.

import type { Order, CreateOrderInput, OrderFilters, PaginatedResult, OrderStatus } from "../models/order";
import * as orderRepo from "../repositories/order-repo";
import * as userRepo from "../repositories/user-repo";
import { ValidationError } from "../shared/errors";
import { logger } from "../shared/logger";
import { getUserById } from "./user-service";  // <-- COMPLETES THE CYCLE

// ... (same as PERFECT, but with one new export)

export async function createOrder(input: CreateOrderInput): Promise<Order> {
  // BUG: Uses getUserById from user-service instead of userRepo.findUserById.
  // This itself is not the violation (services can import from each other if
  // it does not create a cycle), but combined with user-service importing
  // getOrdersByUser from this file, it creates: user-service -> order-service -> user-service.
  const user = await getUserById(input.userId);
  logger.info("Creating order for user", { userName: user.name });

  if (input.items.length === 0) {
    throw new ValidationError("Order must contain at least one item");
  }

  return orderRepo.createOrder(input);
}

export async function getOrderById(id: string): Promise<Order> {
  return orderRepo.findOrderById(id);
}

export async function getOrdersByUser(userId: string): Promise<Order[]> {
  const result = await orderRepo.findOrders({ userId }, 1, 1000);
  return result.data;
}

export async function transitionOrderStatus(
  orderId: string,
  newStatus: OrderStatus
): Promise<Order> {
  const order = await orderRepo.findOrderById(orderId);
  const VALID_TRANSITIONS: Record<OrderStatus, OrderStatus[]> = {
    pending: ["confirmed", "cancelled"],
    confirmed: ["processing", "cancelled"],
    processing: ["shipped", "cancelled"],
    shipped: ["delivered"],
    delivered: ["refunded"],
    cancelled: [],
    refunded: [],
  };
  const allowed = VALID_TRANSITIONS[order.status];

  if (!allowed.includes(newStatus)) {
    throw new ValidationError(
      `Cannot transition from ${order.status} to ${newStatus}`
    );
  }

  return orderRepo.updateOrderStatus(orderId, newStatus);
}
