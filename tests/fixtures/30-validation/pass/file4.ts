// user-service.ts -- User business logic.
// V04: Creates a circular dependency with order-service.

import type { User, CreateUserInput, UpdateUserInput, UserFilters, PaginatedResult } from "../models/user";
import * as userRepo from "../repositories/user-repo";
import { ValidationError } from "../shared/errors";
import { logger } from "../shared/logger";
import * as crypto from "crypto";
import { getOrdersByUser } from "./order-service";  // <-- CREATES CYCLE

export async function createUser(input: CreateUserInput): Promise<User> {
  validateEmail(input.email);
  validatePassword(input.password);

  const hashedInput: CreateUserInput = {
    ...input,
    password: await hashPassword(input.password),
  };

  return userRepo.createUser(hashedInput);
}

export async function getUserById(id: string): Promise<User> {
  return userRepo.findUserById(id);
}

// New function that imports from order-service, creating the cycle.
export async function getUserWithOrders(userId: string): Promise<{ user: User; orderCount: number }> {
  const user = await userRepo.findUserById(userId);
  // BUG: This import creates a circular dependency because order-service
  // already imports from user-repo (which is fine), but if order-service
  // also imported from user-service, we'd have user-service -> order-service
  // -> user-service.
  const orders = await getOrdersByUser(userId);
  return { user, orderCount: orders.length };
}

export async function listUsers(
  filters: UserFilters,
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResult<User>> {
  if (pageSize < 1 || pageSize > 100) {
    throw new ValidationError("pageSize must be between 1 and 100");
  }
  return userRepo.findUsers(filters, page, pageSize);
}

// ... (validation helpers same as PERFECT)

function validateEmail(email: string): void {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  if (!emailRegex.test(email)) {
    throw new ValidationError("Invalid email format");
  }
}

function validatePassword(password: string): void {
  if (password.length < 8) {
    throw new ValidationError("Password too short");
  }
}

async function hashPassword(password: string): Promise<string> {
  const salt = crypto.randomBytes(16).toString("hex");
  return new Promise((resolve, reject) => {
    crypto.scrypt(password, salt, 64, (err, key) => {
      if (err) reject(err);
      else resolve(`${salt}:${key.toString("hex")}`);
    });
  });
}
