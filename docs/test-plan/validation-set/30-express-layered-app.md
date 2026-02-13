# 30 -- Express.js Layered Application (Architecture Validation)

**Why included:** Exercises all 6 ARCH rules: dependency-direction, import-boundary, no-circular-deps, max-file-lines, layer-violation, module-boundary. Demonstrates a realistic 4-layer Express.js application with module boundaries, shared utilities, and middleware.

---

## Architecture

```
src/
  routes/           (Layer 1 -- HTTP handlers)
    users.ts
    orders.ts
    index.ts
  services/         (Layer 2 -- Business logic)
    user-service.ts
    order-service.ts
  repositories/     (Layer 3 -- Data access)
    user-repo.ts
    order-repo.ts
  models/           (Layer 4 -- Types/entities)
    user.ts
    order.ts
  shared/           (Cross-cutting -- allowed by all layers)
    errors.ts
    logger.ts
    config.ts
  middleware/        (Route-level -- same as Layer 1)
    auth.ts
    validation.ts
  modules/
    auth/
      index.ts          (public API)
      token-validator.ts (internal)
      internal/
        jwt-utils.ts    (internal)
    billing/
      index.ts          (public API)
      internal/
        stripe-adapter.ts (internal)
tests/
  routes/users.test.ts
  services/user-service.test.ts
```

---

## Manifest Fragment

```yaml
# .stricture.yml -- Express.js layered architecture rules.

version: "1.0"
strictness: strict

rules:
  # -- 1. Dependency Direction ------------------------------------------
  ARCH-dependency-direction:
    - error
    - layers:
        - { name: route, patterns: ["src/routes/**", "src/middleware/**"] }
        - { name: service, patterns: ["src/services/**"] }
        - { name: repository, patterns: ["src/repositories/**"] }
        - { name: model, patterns: ["src/models/**"] }
      direction: top-down
      shared: ["src/shared/**"]

  # -- 2. Import Boundary -----------------------------------------------
  ARCH-import-boundary:
    - error
    - boundaries:
        - { name: auth, patterns: ["src/modules/auth/**"], exports: ["src/modules/auth/index.ts"] }
        - { name: billing, patterns: ["src/modules/billing/**"], exports: ["src/modules/billing/index.ts"] }

  # -- 3. No Circular Dependencies --------------------------------------
  ARCH-no-circular-deps:
    - error

  # -- 4. Max File Lines -------------------------------------------------
  ARCH-max-file-lines:
    - error
    - max: 800
    - overrides:
        "**/*.test.ts": 1200

  # -- 5. Layer Violation ------------------------------------------------
  ARCH-layer-violation:
    - error
    - rules:
        - layer: route
          forbidden: [repository, model]
          reason: "Routes must delegate to services, not access repos or models directly"

  # -- 6. Module Boundary ------------------------------------------------
  ARCH-module-boundary:
    - error
    - modules:
        - { name: auth, entry: "src/modules/auth/index.ts" }
        - { name: billing, entry: "src/modules/billing/index.ts" }
```

---

## PERFECT -- Correct Architecture

Every file below demonstrates the correct architectural pattern. Stricture must produce zero violations when scanning this codebase.

### `src/models/user.ts`

```typescript
// user.ts -- User entity and related types.

export interface User {
  id: string;
  email: string;
  name: string;
  passwordHash: string;
  role: UserRole;
  createdAt: Date;
  updatedAt: Date;
}

export type UserRole = "admin" | "member" | "viewer";

export interface CreateUserInput {
  email: string;
  name: string;
  password: string;
  role?: UserRole;
}

export interface UpdateUserInput {
  name?: string;
  email?: string;
  role?: UserRole;
}

export interface UserFilters {
  role?: UserRole;
  search?: string;
  createdAfter?: Date;
  createdBefore?: Date;
}

export interface PaginatedResult<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
  hasMore: boolean;
}
```

**Why correct:** Models have zero imports from other layers. They define pure types and interfaces only. No dependency on routes, services, or repositories.

---

### `src/models/order.ts`

```typescript
// order.ts -- Order entity and related types.

export interface Order {
  id: string;
  userId: string;
  items: OrderItem[];
  status: OrderStatus;
  totalCents: number;
  currency: string;
  shippingAddress: Address;
  createdAt: Date;
  updatedAt: Date;
}

export type OrderStatus =
  | "pending"
  | "confirmed"
  | "processing"
  | "shipped"
  | "delivered"
  | "cancelled"
  | "refunded";

export interface OrderItem {
  productId: string;
  name: string;
  quantity: number;
  unitPriceCents: number;
}

export interface Address {
  line1: string;
  line2?: string;
  city: string;
  state: string;
  postalCode: string;
  country: string;
}

export interface CreateOrderInput {
  userId: string;
  items: OrderItem[];
  shippingAddress: Address;
  currency?: string;
}

export interface OrderFilters {
  userId?: string;
  status?: OrderStatus;
  minTotal?: number;
  maxTotal?: number;
}
```

**Why correct:** Pure types. No imports from any layer.

---

### `src/shared/errors.ts`

```typescript
// errors.ts -- Application error classes used across all layers.

export class AppError extends Error {
  public readonly statusCode: number;
  public readonly code: string;
  public readonly isOperational: boolean;

  constructor(message: string, statusCode: number, code: string) {
    super(message);
    this.name = "AppError";
    this.statusCode = statusCode;
    this.code = code;
    this.isOperational = true;
    Object.setPrototypeOf(this, AppError.prototype);
  }
}

export class NotFoundError extends AppError {
  constructor(resource: string, id: string) {
    super(`${resource} with id ${id} not found`, 404, "NOT_FOUND");
    this.name = "NotFoundError";
  }
}

export class ValidationError extends AppError {
  public readonly fields: Record<string, string>;

  constructor(message: string, fields: Record<string, string> = {}) {
    super(message, 400, "VALIDATION_ERROR");
    this.name = "ValidationError";
    this.fields = fields;
  }
}

export class ConflictError extends AppError {
  constructor(message: string) {
    super(message, 409, "CONFLICT");
    this.name = "ConflictError";
  }
}

export class UnauthorizedError extends AppError {
  constructor(message: string = "Authentication required") {
    super(message, 401, "UNAUTHORIZED");
    this.name = "UnauthorizedError";
  }
}

export class ForbiddenError extends AppError {
  constructor(message: string = "Insufficient permissions") {
    super(message, 403, "FORBIDDEN");
    this.name = "ForbiddenError";
  }
}
```

**Why correct:** Shared utilities are in `src/shared/`, which is declared as cross-cutting in the manifest. All layers may import from shared.

---

### `src/shared/logger.ts`

```typescript
// logger.ts -- Structured logger used across all layers.

export type LogLevel = "debug" | "info" | "warn" | "error";

const LOG_LEVELS: Record<LogLevel, number> = { debug: 0, info: 1, warn: 2, error: 3 };
let currentLevel: LogLevel = "info";

export function setLogLevel(level: LogLevel): void { currentLevel = level; }

export function log(level: LogLevel, message: string, context?: Record<string, unknown>): void {
  if (LOG_LEVELS[level] < LOG_LEVELS[currentLevel]) return;
  const output = JSON.stringify({ level, message, timestamp: new Date().toISOString(), context });
  (level === "error" ? process.stderr : process.stdout).write(output + "\n");
}

export const logger = {
  debug: (msg: string, ctx?: Record<string, unknown>) => log("debug", msg, ctx),
  info: (msg: string, ctx?: Record<string, unknown>) => log("info", msg, ctx),
  warn: (msg: string, ctx?: Record<string, unknown>) => log("warn", msg, ctx),
  error: (msg: string, ctx?: Record<string, unknown>) => log("error", msg, ctx),
};
```

---

### `src/shared/config.ts`

```typescript
// config.ts -- Application configuration loaded from environment.

export interface AppConfig {
  port: number;
  databaseUrl: string;
  jwtSecret: string;
  jwtExpirationSeconds: number;
  logLevel: string;
  corsOrigins: string[];
}

export function loadConfig(): AppConfig {
  const databaseUrl = process.env["DATABASE_URL"] ?? "";
  const jwtSecret = process.env["JWT_SECRET"] ?? "";
  if (!databaseUrl) throw new Error("DATABASE_URL is required");
  if (!jwtSecret) throw new Error("JWT_SECRET is required");

  return {
    port: parseInt(process.env["PORT"] ?? "3000", 10),
    databaseUrl,
    jwtSecret,
    jwtExpirationSeconds: parseInt(process.env["JWT_EXPIRATION"] ?? "3600", 10),
    logLevel: process.env["LOG_LEVEL"] ?? "info",
    corsOrigins: (process.env["CORS_ORIGINS"] ?? "").split(",").filter(Boolean),
  };
}
```

---

### `src/repositories/user-repo.ts`

```typescript
// user-repo.ts -- User data access layer.

import type { User, CreateUserInput, UpdateUserInput, UserFilters, PaginatedResult } from "../models/user";
import { NotFoundError, ConflictError } from "../shared/errors";
import { logger } from "../shared/logger";

// In a real app, this would be a database client. For validation purposes,
// we use an in-memory store to demonstrate correct layer patterns.
const users: Map<string, User> = new Map();

export async function createUser(input: CreateUserInput): Promise<User> {
  const existing = Array.from(users.values()).find((u) => u.email === input.email);
  if (existing) {
    throw new ConflictError(`User with email ${input.email} already exists`);
  }

  const now = new Date();
  const user: User = {
    id: generateId(),
    email: input.email,
    name: input.name,
    passwordHash: input.password, // service layer handles hashing before calling repo
    role: input.role ?? "member",
    createdAt: now,
    updatedAt: now,
  };

  users.set(user.id, user);
  logger.info("User created", { userId: user.id });
  return user;
}

export async function findUserById(id: string): Promise<User> {
  const user = users.get(id);
  if (!user) {
    throw new NotFoundError("User", id);
  }
  return user;
}

export async function findUsers(
  filters: UserFilters,
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResult<User>> {
  let results = Array.from(users.values());

  if (filters.role) {
    results = results.filter((u) => u.role === filters.role);
  }
  if (filters.search) {
    const search = filters.search.toLowerCase();
    results = results.filter(
      (u) => u.name.toLowerCase().includes(search) || u.email.toLowerCase().includes(search)
    );
  }
  if (filters.createdAfter) {
    results = results.filter((u) => u.createdAt >= filters.createdAfter!);
  }
  if (filters.createdBefore) {
    results = results.filter((u) => u.createdAt <= filters.createdBefore!);
  }

  const total = results.length;
  const start = (page - 1) * pageSize;
  const data = results.slice(start, start + pageSize);

  return {
    data,
    total,
    page,
    pageSize,
    hasMore: start + pageSize < total,
  };
}

export async function updateUser(id: string, input: UpdateUserInput): Promise<User> {
  const user = await findUserById(id);
  const updated: User = {
    ...user,
    ...input,
    updatedAt: new Date(),
  };
  users.set(id, updated);
  logger.info("User updated", { userId: id });
  return updated;
}

export async function deleteUser(id: string): Promise<void> {
  const exists = users.has(id);
  if (!exists) {
    throw new NotFoundError("User", id);
  }
  users.delete(id);
  logger.info("User deleted", { userId: id });
}

function generateId(): string {
  return `usr_${Date.now()}_${Math.random().toString(36).slice(2, 10)}`;
}
```

**Why correct:** Repository (Layer 3) imports only from models (Layer 4) and shared (cross-cutting). It never imports from routes or services. Dependency direction is strictly top-down.

---

### `src/repositories/order-repo.ts`

```typescript
// order-repo.ts -- Order data access layer.

import type { Order, CreateOrderInput, OrderFilters, PaginatedResult } from "../models/order";
import { NotFoundError } from "../shared/errors";
import { logger } from "../shared/logger";

const orders: Map<string, Order> = new Map();

export async function createOrder(input: CreateOrderInput): Promise<Order> {
  const totalCents = input.items.reduce(
    (sum, item) => sum + item.unitPriceCents * item.quantity,
    0
  );

  const now = new Date();
  const order: Order = {
    id: generateOrderId(),
    userId: input.userId,
    items: input.items,
    status: "pending",
    totalCents,
    currency: input.currency ?? "usd",
    shippingAddress: input.shippingAddress,
    createdAt: now,
    updatedAt: now,
  };

  orders.set(order.id, order);
  logger.info("Order created", { orderId: order.id, userId: input.userId });
  return order;
}

export async function findOrderById(id: string): Promise<Order> {
  const order = orders.get(id);
  if (!order) {
    throw new NotFoundError("Order", id);
  }
  return order;
}

export async function findOrders(
  filters: OrderFilters,
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResult<Order>> {
  let results = Array.from(orders.values());

  if (filters.userId) {
    results = results.filter((o) => o.userId === filters.userId);
  }
  if (filters.status) {
    results = results.filter((o) => o.status === filters.status);
  }
  if (filters.minTotal !== undefined) {
    results = results.filter((o) => o.totalCents >= filters.minTotal!);
  }
  if (filters.maxTotal !== undefined) {
    results = results.filter((o) => o.totalCents <= filters.maxTotal!);
  }

  const total = results.length;
  const start = (page - 1) * pageSize;
  const data = results.slice(start, start + pageSize);

  return { data, total, page, pageSize, hasMore: start + pageSize < total };
}

export async function updateOrderStatus(
  id: string,
  status: Order["status"]
): Promise<Order> {
  const order = await findOrderById(id);
  const updated: Order = { ...order, status, updatedAt: new Date() };
  orders.set(id, updated);
  logger.info("Order status updated", { orderId: id, status });
  return updated;
}

function generateOrderId(): string {
  return `ord_${Date.now()}_${Math.random().toString(36).slice(2, 10)}`;
}
```

**Why correct:** Same pattern as user-repo. Only imports from models and shared.

---

### `src/services/user-service.ts`

```typescript
// user-service.ts -- User business logic.

import type { User, CreateUserInput, UpdateUserInput, UserFilters, PaginatedResult } from "../models/user";
import * as userRepo from "../repositories/user-repo";
import { ValidationError } from "../shared/errors";
import { logger } from "../shared/logger";
import * as crypto from "crypto";

export async function createUser(input: CreateUserInput): Promise<User> {
  validateEmail(input.email);
  validatePassword(input.password);

  const hashedInput: CreateUserInput = {
    ...input,
    password: await hashPassword(input.password),
  };

  logger.info("Creating user", { email: input.email });
  return userRepo.createUser(hashedInput);
}

export async function getUserById(id: string): Promise<User> {
  return userRepo.findUserById(id);
}

export async function listUsers(
  filters: UserFilters,
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResult<User>> {
  if (pageSize < 1 || pageSize > 100) {
    throw new ValidationError("pageSize must be between 1 and 100");
  }
  if (page < 1) {
    throw new ValidationError("page must be >= 1");
  }
  return userRepo.findUsers(filters, page, pageSize);
}

export async function updateUser(id: string, input: UpdateUserInput): Promise<User> {
  if (input.email !== undefined) {
    validateEmail(input.email);
  }
  return userRepo.updateUser(id, input);
}

export async function deleteUser(id: string): Promise<void> {
  // Verify user exists before deleting (findUserById throws NotFoundError)
  await userRepo.findUserById(id);
  return userRepo.deleteUser(id);
}

// -- Validation helpers (business logic, belongs in service layer) -------

function validateEmail(email: string): void {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  if (!emailRegex.test(email)) {
    throw new ValidationError("Invalid email format", { email: "Must be a valid email address" });
  }
}

function validatePassword(password: string): void {
  if (password.length < 8) {
    throw new ValidationError("Password too short", { password: "Must be at least 8 characters" });
  }
  if (!/[A-Z]/.test(password)) {
    throw new ValidationError("Password must contain uppercase letter", {
      password: "Must contain at least one uppercase letter",
    });
  }
  if (!/[0-9]/.test(password)) {
    throw new ValidationError("Password must contain a digit", {
      password: "Must contain at least one digit",
    });
  }
}

async function hashPassword(password: string): Promise<string> {
  const salt = crypto.randomBytes(16).toString("hex");
  return new Promise((resolve, reject) => {
    crypto.scrypt(password, salt, 64, (err, derivedKey) => {
      if (err) {
        reject(err);
        return;
      }
      resolve(`${salt}:${derivedKey.toString("hex")}`);
    });
  });
}
```

**Why correct:** Service (Layer 2) imports from repositories (Layer 3), models (Layer 4), and shared. It never imports from routes (Layer 1). Dependency direction is top-down.

---

### `src/services/order-service.ts`

```typescript
// order-service.ts -- Order business logic.

import type { Order, CreateOrderInput, OrderFilters, PaginatedResult, OrderStatus } from "../models/order";
import * as orderRepo from "../repositories/order-repo";
import * as userRepo from "../repositories/user-repo";
import { ValidationError } from "../shared/errors";
import { logger } from "../shared/logger";

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
  // Verify user exists
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

  logger.info("Creating order", { userId: input.userId, itemCount: input.items.length });
  return orderRepo.createOrder(input);
}

export async function getOrderById(id: string): Promise<Order> {
  return orderRepo.findOrderById(id);
}

export async function listOrders(
  filters: OrderFilters,
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResult<Order>> {
  if (pageSize < 1 || pageSize > 100) {
    throw new ValidationError("pageSize must be between 1 and 100");
  }
  return orderRepo.findOrders(filters, page, pageSize);
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

  logger.info("Order status transition", {
    orderId,
    from: order.status,
    to: newStatus,
  });

  return orderRepo.updateOrderStatus(orderId, newStatus);
}

export async function cancelOrder(orderId: string): Promise<Order> {
  return transitionOrderStatus(orderId, "cancelled");
}
```

**Why correct:** Service imports from repos (one layer down), models (two layers down), and shared. All valid top-down imports.

---

### `src/modules/auth/index.ts`

```typescript
// index.ts -- Auth module public API. All external consumers import from here.

export { validateToken, generateToken, type TokenPayload } from "./token-validator";
```

**Why correct:** Module entry point re-exports the public API. External consumers use `../modules/auth` which resolves to this file.

---

### `src/modules/auth/token-validator.ts`

```typescript
// token-validator.ts -- JWT token creation and validation.

import { loadConfig } from "../../shared/config";
import { UnauthorizedError } from "../../shared/errors";
import * as crypto from "crypto";

export interface TokenPayload {
  userId: string;
  role: string;
  exp: number;
}

export function generateToken(userId: string, role: string): string {
  const config = loadConfig();
  const header = base64url(JSON.stringify({ alg: "HS256", typ: "JWT" }));
  const now = Math.floor(Date.now() / 1000);
  const payload = base64url(
    JSON.stringify({
      userId,
      role,
      iat: now,
      exp: now + config.jwtExpirationSeconds,
    })
  );
  const signature = sign(`${header}.${payload}`, config.jwtSecret);
  return `${header}.${payload}.${signature}`;
}

export function validateToken(token: string): TokenPayload {
  const config = loadConfig();
  const parts = token.split(".");
  if (parts.length !== 3) {
    throw new UnauthorizedError("Malformed token");
  }

  const [header, payload, signature] = parts;
  const expectedSig = sign(`${header}.${payload}`, config.jwtSecret);

  const sigBuf = Buffer.from(signature, "base64url");
  const expectedBuf = Buffer.from(expectedSig, "base64url");

  if (sigBuf.length !== expectedBuf.length || !crypto.timingSafeEqual(sigBuf, expectedBuf)) {
    throw new UnauthorizedError("Invalid token signature");
  }

  const decoded = JSON.parse(Buffer.from(payload, "base64url").toString()) as TokenPayload;

  if (decoded.exp < Math.floor(Date.now() / 1000)) {
    throw new UnauthorizedError("Token expired");
  }

  return decoded;
}

function base64url(str: string): string {
  return Buffer.from(str).toString("base64url");
}

function sign(data: string, secret: string): string {
  return crypto.createHmac("sha256", secret).update(data).digest("base64url");
}
```

---

### `src/modules/billing/index.ts`

```typescript
// index.ts -- Billing module public API.

export {
  createPaymentIntent,
  capturePayment,
  refundPayment,
  type PaymentResult,
  type RefundResult,
} from "./internal/stripe-adapter";
```

---

### `src/modules/billing/internal/stripe-adapter.ts`

```typescript
// stripe-adapter.ts -- Stripe integration (internal to billing module).

import { logger } from "../../../shared/logger";
import { AppError } from "../../../shared/errors";

export interface PaymentResult {
  paymentId: string;
  status: "succeeded" | "pending" | "failed";
  amountCents: number;
}

export interface RefundResult {
  refundId: string;
  status: "succeeded" | "pending" | "failed";
  amountCents: number;
}

async function stripeRequest(path: string, method: string, body?: string): Promise<Record<string, unknown>> {
  try {
    const response = await fetch(`https://api.stripe.com${path}`, {
      method,
      headers: {
        Authorization: `Bearer ${process.env["STRIPE_SECRET_KEY"]}`,
        "Content-Type": "application/x-www-form-urlencoded",
      },
      body,
    });
    if (!response.ok) {
      const error = (await response.json()) as Record<string, Record<string, string>>;
      throw new AppError(`Stripe error: ${error.error?.message}`, response.status, "STRIPE_ERROR");
    }
    return (await response.json()) as Record<string, unknown>;
  } catch (err) {
    if (err instanceof AppError) throw err;
    throw new AppError(`Stripe request failed: ${(err as Error).message}`, 500, "STRIPE_FAILED");
  }
}

export async function createPaymentIntent(amountCents: number, currency: string, customerId: string): Promise<PaymentResult> {
  if (amountCents < 50) throw new AppError("Amount must be at least 50 cents", 400, "INVALID_AMOUNT");
  logger.info("Creating payment intent", { amountCents, currency, customerId });
  const intent = await stripeRequest("/v1/payment_intents", "POST", `amount=${amountCents}&currency=${currency}&customer=${customerId}`);
  return { paymentId: intent.id as string, status: intent.status === "succeeded" ? "succeeded" : "pending", amountCents };
}

export async function capturePayment(paymentId: string): Promise<PaymentResult> {
  logger.info("Capturing payment", { paymentId });
  const result = await stripeRequest(`/v1/payment_intents/${paymentId}/capture`, "POST");
  return { paymentId: result.id as string, status: "succeeded", amountCents: result.amount as number };
}

export async function refundPayment(paymentId: string, amountCents?: number): Promise<RefundResult> {
  logger.info("Refunding payment", { paymentId, amountCents });
  const body = amountCents !== undefined ? `payment_intent=${paymentId}&amount=${amountCents}` : `payment_intent=${paymentId}`;
  const result = await stripeRequest("/v1/refunds", "POST", body);
  return { refundId: result.id as string, status: "succeeded", amountCents: result.amount as number };
}
```

---

### `src/middleware/auth.ts`

```typescript
// auth.ts -- Authentication middleware (route-level, Layer 1).

import type { Request, Response, NextFunction } from "express";
import { validateToken } from "../modules/auth";
import { UnauthorizedError, ForbiddenError } from "../shared/errors";
import { logger } from "../shared/logger";
import type { UserRole } from "../models/user";

// Extend Express Request to include authenticated user
declare global {
  namespace Express {
    interface Request {
      user?: {
        userId: string;
        role: string;
      };
    }
  }
}

export function authenticate(req: Request, _res: Response, next: NextFunction): void {
  const authHeader = req.headers.authorization;
  if (!authHeader || !authHeader.startsWith("Bearer ")) {
    next(new UnauthorizedError("Missing or invalid Authorization header"));
    return;
  }

  const token = authHeader.slice(7);

  try {
    const payload = validateToken(token);
    req.user = { userId: payload.userId, role: payload.role };
    logger.debug("Authenticated request", { userId: payload.userId });
    next();
  } catch (err) {
    next(err);
  }
}

export function requireRole(...roles: UserRole[]) {
  return (req: Request, _res: Response, next: NextFunction): void => {
    if (!req.user) {
      next(new UnauthorizedError());
      return;
    }

    if (!roles.includes(req.user.role as UserRole)) {
      next(new ForbiddenError(`Requires one of: ${roles.join(", ")}`));
      return;
    }

    next();
  };
}
```

**Why correct:** Middleware is at the route level (Layer 1) in the manifest. It imports from `modules/auth` through the module's public entry point (`index.ts`), from shared, and uses a type from models. The `import type` from models is acceptable because middleware is in the route layer which is above model.

---

### `src/middleware/validation.ts`

```typescript
// validation.ts -- Request validation middleware.

import type { Request, Response, NextFunction } from "express";
import { ValidationError } from "../shared/errors";

type Validator = (body: unknown) => Record<string, string> | null;

export function validateBody(validator: Validator) {
  return (req: Request, _res: Response, next: NextFunction): void => {
    const errors = validator(req.body);
    if (errors) {
      next(new ValidationError("Request validation failed", errors));
      return;
    }
    next();
  };
}

export function requireFields(...fields: string[]) {
  return (req: Request, _res: Response, next: NextFunction): void => {
    const missing: Record<string, string> = {};
    for (const field of fields) {
      if (req.body[field] === undefined || req.body[field] === null) {
        missing[field] = `${field} is required`;
      }
    }

    if (Object.keys(missing).length > 0) {
      next(new ValidationError("Missing required fields", missing));
      return;
    }
    next();
  };
}
```

---

### `src/routes/users.ts`

```typescript
// users.ts -- User route handlers.

import { Router } from "express";
import type { Request, Response, NextFunction } from "express";
import * as userService from "../services/user-service";
import { authenticate, requireRole } from "../middleware/auth";
import { validateBody } from "../middleware/validation";
import { AppError } from "../shared/errors";
import { logger } from "../shared/logger";

export const userRouter = Router();

// -- GET /users ---------------------------------------------------------

userRouter.get("/", authenticate, async (req: Request, res: Response, next: NextFunction) => {
  try {
    const page = parseInt(req.query["page"] as string, 10) || 1;
    const pageSize = parseInt(req.query["pageSize"] as string, 10) || 20;
    const filters = {
      role: req.query["role"] as string | undefined,
      search: req.query["search"] as string | undefined,
    };

    const result = await userService.listUsers(filters, page, pageSize);
    res.json(result);
  } catch (err) {
    next(err);
  }
});

// -- GET /users/:id -----------------------------------------------------

userRouter.get("/:id", authenticate, async (req: Request, res: Response, next: NextFunction) => {
  try {
    const user = await userService.getUserById(req.params["id"]);
    res.json(user);
  } catch (err) {
    next(err);
  }
});

// -- POST /users --------------------------------------------------------

const createUserValidator = (body: unknown): Record<string, string> | null => {
  const errors: Record<string, string> = {};
  const b = body as Record<string, unknown>;
  if (typeof b.email !== "string") errors["email"] = "email is required";
  if (typeof b.name !== "string") errors["name"] = "name is required";
  if (typeof b.password !== "string") errors["password"] = "password is required";
  return Object.keys(errors).length > 0 ? errors : null;
};

userRouter.post(
  "/",
  authenticate,
  requireRole("admin"),
  validateBody(createUserValidator),
  async (req: Request, res: Response, next: NextFunction) => {
    try {
      const user = await userService.createUser(req.body);
      logger.info("User created via API", { userId: user.id });
      res.status(201).json(user);
    } catch (err) {
      next(err);
    }
  }
);

// -- PATCH /users/:id ---------------------------------------------------

userRouter.patch("/:id", authenticate, async (req: Request, res: Response, next: NextFunction) => {
  try {
    const user = await userService.updateUser(req.params["id"], req.body);
    res.json(user);
  } catch (err) {
    next(err);
  }
});

// -- DELETE /users/:id --------------------------------------------------

userRouter.delete(
  "/:id",
  authenticate,
  requireRole("admin"),
  async (req: Request, res: Response, next: NextFunction) => {
    try {
      await userService.deleteUser(req.params["id"]);
      res.status(204).send();
    } catch (err) {
      next(err);
    }
  }
);

// -- Error handler for this router --------------------------------------

userRouter.use((err: Error, _req: Request, res: Response, _next: NextFunction) => {
  if (err instanceof AppError) {
    res.status(err.statusCode).json({
      error: { code: err.code, message: err.message },
    });
    return;
  }
  logger.error("Unhandled error in user routes", { error: err.message });
  res.status(500).json({ error: { code: "INTERNAL_ERROR", message: "Internal server error" } });
});
```

**Why correct:** Routes (Layer 1) import only from services (Layer 2), middleware (same layer), and shared. They never import directly from repositories or models for data access. The route delegates all business logic and data access to the service layer.

---

### `src/routes/orders.ts`

```typescript
// orders.ts -- Order route handlers.

import { Router } from "express";
import type { Request, Response, NextFunction } from "express";
import * as orderService from "../services/order-service";
import { authenticate } from "../middleware/auth";
import { AppError, ValidationError } from "../shared/errors";
import { logger } from "../shared/logger";

export const orderRouter = Router();

// -- POST /orders -------------------------------------------------------

orderRouter.post("/", authenticate, async (req: Request, res: Response, next: NextFunction) => {
  try {
    if (!req.user) {
      next(new ValidationError("User context required"));
      return;
    }

    const order = await orderService.createOrder({
      ...req.body,
      userId: req.user.userId,
    });
    logger.info("Order created via API", { orderId: order.id });
    res.status(201).json(order);
  } catch (err) {
    next(err);
  }
});

// -- GET /orders --------------------------------------------------------

orderRouter.get("/", authenticate, async (req: Request, res: Response, next: NextFunction) => {
  try {
    const page = parseInt(req.query["page"] as string, 10) || 1;
    const pageSize = parseInt(req.query["pageSize"] as string, 10) || 20;
    const filters = {
      userId: req.user?.userId,
      status: req.query["status"] as string | undefined,
    };

    const result = await orderService.listOrders(filters, page, pageSize);
    res.json(result);
  } catch (err) {
    next(err);
  }
});

// -- GET /orders/:id ----------------------------------------------------

orderRouter.get("/:id", authenticate, async (req: Request, res: Response, next: NextFunction) => {
  try {
    const order = await orderService.getOrderById(req.params["id"]);
    res.json(order);
  } catch (err) {
    next(err);
  }
});

// -- PATCH /orders/:id/status -------------------------------------------

orderRouter.patch(
  "/:id/status",
  authenticate,
  async (req: Request, res: Response, next: NextFunction) => {
    try {
      const { status } = req.body;
      if (!status) {
        next(new ValidationError("status is required"));
        return;
      }

      const order = await orderService.transitionOrderStatus(req.params["id"], status);
      logger.info("Order status updated via API", { orderId: order.id, status: order.status });
      res.json(order);
    } catch (err) {
      next(err);
    }
  }
);

// -- POST /orders/:id/cancel -------------------------------------------

orderRouter.post(
  "/:id/cancel",
  authenticate,
  async (req: Request, res: Response, next: NextFunction) => {
    try {
      const order = await orderService.cancelOrder(req.params["id"]);
      res.json(order);
    } catch (err) {
      next(err);
    }
  }
);

// -- Error handler ------------------------------------------------------

orderRouter.use((err: Error, _req: Request, res: Response, _next: NextFunction) => {
  if (err instanceof AppError) {
    res.status(err.statusCode).json({
      error: { code: err.code, message: err.message },
    });
    return;
  }
  logger.error("Unhandled error in order routes", { error: err.message });
  res.status(500).json({ error: { code: "INTERNAL_ERROR", message: "Internal server error" } });
});
```

---

### `src/routes/index.ts`

```typescript
// index.ts -- Route aggregator.

import { Router } from "express";
import { userRouter } from "./users";
import { orderRouter } from "./orders";

export const apiRouter = Router();

apiRouter.use("/users", userRouter);
apiRouter.use("/orders", orderRouter);

apiRouter.get("/health", (_req, res) => {
  res.json({ status: "ok", timestamp: new Date().toISOString() });
});
```

---

### `tests/routes/users.test.ts`

```typescript
// users.test.ts -- User route handler tests.

import { describe, it, expect, beforeEach, vi } from "vitest";
import request from "supertest";
import express from "express";
import { userRouter } from "../../src/routes/users";

// Mock the service layer (not the repository -- route tests validate
// the HTTP interface, not data access)
vi.mock("../../src/services/user-service", () => ({
  createUser: vi.fn(),
  getUserById: vi.fn(),
  listUsers: vi.fn(),
  updateUser: vi.fn(),
  deleteUser: vi.fn(),
}));

vi.mock("../../src/modules/auth", () => ({
  validateToken: vi.fn().mockReturnValue({ userId: "usr_123", role: "admin" }),
}));

const app = express();
app.use(express.json());
app.use("/users", userRouter);

describe("GET /users", () => {
  it("returns paginated user list", async () => {
    const { listUsers } = await import("../../src/services/user-service");
    (listUsers as ReturnType<typeof vi.fn>).mockResolvedValue({
      data: [{ id: "usr_1", name: "Alice", email: "alice@test.com" }],
      total: 1,
      page: 1,
      pageSize: 20,
      hasMore: false,
    });

    const res = await request(app)
      .get("/users")
      .set("Authorization", "Bearer valid-token");

    expect(res.status).toBe(200);
    expect(res.body.data).toHaveLength(1);
    expect(res.body.data[0].name).toBe("Alice");
    expect(res.body.total).toBe(1);
    expect(res.body.hasMore).toBe(false);
  });

  it("passes query params to service", async () => {
    const { listUsers } = await import("../../src/services/user-service");
    (listUsers as ReturnType<typeof vi.fn>).mockResolvedValue({
      data: [],
      total: 0,
      page: 2,
      pageSize: 10,
      hasMore: false,
    });

    await request(app)
      .get("/users?page=2&pageSize=10&role=admin")
      .set("Authorization", "Bearer valid-token");

    expect(listUsers).toHaveBeenCalledWith(
      { role: "admin", search: undefined },
      2,
      10
    );
  });
});

describe("POST /users", () => {
  it("creates a user and returns 201", async () => {
    const { createUser } = await import("../../src/services/user-service");
    (createUser as ReturnType<typeof vi.fn>).mockResolvedValue({
      id: "usr_new",
      name: "Bob",
      email: "bob@test.com",
      role: "member",
    });

    const res = await request(app)
      .post("/users")
      .set("Authorization", "Bearer valid-token")
      .send({ email: "bob@test.com", name: "Bob", password: "Secret123" });

    expect(res.status).toBe(201);
    expect(res.body.id).toBe("usr_new");
  });

  it("returns 400 for missing required fields", async () => {
    const res = await request(app)
      .post("/users")
      .set("Authorization", "Bearer valid-token")
      .send({ email: "bob@test.com" }); // missing name and password

    expect(res.status).toBe(400);
  });
});

describe("DELETE /users/:id", () => {
  it("returns 204 on successful delete", async () => {
    const { deleteUser } = await import("../../src/services/user-service");
    (deleteUser as ReturnType<typeof vi.fn>).mockResolvedValue(undefined);

    const res = await request(app)
      .delete("/users/usr_123")
      .set("Authorization", "Bearer valid-token");

    expect(res.status).toBe(204);
  });
});
```

**Why correct:** Test file imports from routes and services (for mocking). Test files are exempt from layer rules. Test file is well under the 1200-line override limit.

---

### `tests/services/user-service.test.ts`

```typescript
// user-service.test.ts -- User service unit tests.

import { describe, it, expect, beforeEach, vi } from "vitest";
import * as userService from "../../src/services/user-service";

vi.mock("../../src/repositories/user-repo");

describe("createUser", () => {
  it("rejects invalid email format", async () => {
    await expect(
      userService.createUser({
        email: "not-an-email",
        name: "Test",
        password: "ValidPass1",
      })
    ).rejects.toThrow("Invalid email format");
  });

  it("rejects short passwords", async () => {
    await expect(
      userService.createUser({
        email: "test@example.com",
        name: "Test",
        password: "short",
      })
    ).rejects.toThrow("Password too short");
  });

  it("rejects passwords without uppercase", async () => {
    await expect(
      userService.createUser({
        email: "test@example.com",
        name: "Test",
        password: "nouppercase1",
      })
    ).rejects.toThrow("uppercase");
  });

  it("rejects passwords without digits", async () => {
    await expect(
      userService.createUser({
        email: "test@example.com",
        name: "Test",
        password: "NoDigitsHere",
      })
    ).rejects.toThrow("digit");
  });
});

describe("listUsers", () => {
  it("rejects pageSize > 100", async () => {
    await expect(
      userService.listUsers({}, 1, 200)
    ).rejects.toThrow("pageSize must be between 1 and 100");
  });

  it("rejects page < 1", async () => {
    await expect(
      userService.listUsers({}, 0)
    ).rejects.toThrow("page must be >= 1");
  });
});
```

---

## PERFECT -- Summary

The PERFECT codebase has the following properties that Stricture must verify with zero violations:

1. **ARCH-dependency-direction**: All imports flow top-down (route -> service -> repository -> model). Shared is accessible by all.
2. **ARCH-import-boundary**: The auth and billing modules are accessed only through their `index.ts` entry points.
3. **ARCH-no-circular-deps**: No file imports create a cycle anywhere in the dependency graph.
4. **ARCH-max-file-lines**: All source files are well under 800 lines. Test files are under 1200 lines.
5. **ARCH-layer-violation**: Routes never access repositories or models directly. Business logic stays in services.
6. **ARCH-module-boundary**: No external file reaches into `modules/auth/internal/` or `modules/billing/internal/` directly.

---

## Violations

Each violation below modifies one or more PERFECT files to introduce a specific architectural rule violation. The violation ID uses the `V##` format (not `B##`, which is reserved for contract/test-quality bugs).

---

## V01 -- Route Imports Repository Directly (ARCH-dependency-direction)

**Rule violated:** `ARCH-dependency-direction`

### Violating File: `src/routes/users.ts`

```typescript
// users.ts -- User route handlers.
// V01: Route layer imports directly from repository layer, skipping services.

import { Router } from "express";
import type { Request, Response, NextFunction } from "express";
import * as userService from "../services/user-service";
import * as userRepo from "../repositories/user-repo";  // <-- VIOLATION
import { authenticate, requireRole } from "../middleware/auth";
import { AppError } from "../shared/errors";
import { logger } from "../shared/logger";

export const userRouter = Router();

// The GET /users endpoint bypasses the service layer and queries the repo directly.
userRouter.get("/", authenticate, async (req: Request, res: Response, next: NextFunction) => {
  try {
    const page = parseInt(req.query["page"] as string, 10) || 1;
    const pageSize = parseInt(req.query["pageSize"] as string, 10) || 20;

    // BUG: Route directly calls repository. This skips business logic
    // (validation, authorization, pagination bounds) in user-service.
    const result = await userRepo.findUsers({}, page, pageSize);  // <-- VIOLATION
    res.json(result);
  } catch (err) {
    next(err);
  }
});

// Other handlers still use the service (partial violation).
userRouter.get("/:id", authenticate, async (req: Request, res: Response, next: NextFunction) => {
  try {
    const user = await userService.getUserById(req.params["id"]);
    res.json(user);
  } catch (err) {
    next(err);
  }
});
```

### Expected Violation

```
ARCH-dependency-direction: src/routes/users.ts imports from src/repositories/user-repo.ts.
  Route layer (Layer 1) imports from repository layer (Layer 3), skipping the
  service layer. In a top-down architecture with layers [route, service,
  repository, model], the route layer should only import from service or shared.
  Location: line 6, import * as userRepo from "../repositories/user-repo"
```

### What Makes This Detectable

The manifest defines four layers with `direction: top-down`. The route layer is Layer 1 and the repository layer is Layer 3. While top-down direction technically allows higher layers to import lower ones, the `ARCH-layer-violation` rule separately forbids routes from accessing repositories directly (the `forbidden: [repository, model]` config). However, the `dependency-direction` rule also flags this because skipping an intermediate layer signals an architectural smell. Stricture resolves the import path `../repositories/user-repo` to a file matching the `src/repositories/**` pattern and reports the layer skip.

---

## V02 -- Service Imports from Route (ARCH-dependency-direction, Reverse)

**Rule violated:** `ARCH-dependency-direction`

### Violating File: `src/services/order-service.ts`

```typescript
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
```

### Expected Violation

```
ARCH-dependency-direction: src/services/order-service.ts imports from src/routes/orders.ts.
  Service layer (Layer 2) imports from route layer (Layer 1). In a top-down
  architecture, lower layers must not import from higher layers. The service
  layer should never depend on route handlers.
  Location: line 8, import { orderRouter } from "../routes/orders"
```

### What Makes This Detectable

This is a textbook reverse dependency. The manifest declares `direction: top-down` with routes above services. Stricture resolves `../routes/orders` to a file matching `src/routes/**` (Layer 1) and the importing file matches `src/services/**` (Layer 2). Since Layer 2 is below Layer 1 in the declared order, this is a reverse import.

---

## V03 -- Repository Imports from Service (ARCH-dependency-direction, Reverse)

**Rule violated:** `ARCH-dependency-direction`

### Violating File: `src/repositories/user-repo.ts`

```typescript
// user-repo.ts -- User data access layer.
// V03: Repository imports from service layer (reverse direction).

import type { User, CreateUserInput, UpdateUserInput, UserFilters, PaginatedResult } from "../models/user";
import { NotFoundError, ConflictError } from "../shared/errors";
import { logger } from "../shared/logger";
import { createUser as createUserInService } from "../services/user-service";  // <-- VIOLATION

const users: Map<string, User> = new Map();

export async function createUser(input: CreateUserInput): Promise<User> {
  const existing = Array.from(users.values()).find((u) => u.email === input.email);
  if (existing) {
    throw new ConflictError(`User with email ${input.email} already exists`);
  }

  const now = new Date();
  const user: User = {
    id: generateId(),
    email: input.email,
    name: input.name,
    passwordHash: input.password,
    role: input.role ?? "member",
    createdAt: now,
    updatedAt: now,
  };

  users.set(user.id, user);
  logger.info("User created", { userId: user.id });

  // BUG: The repository calls back into the service layer to "notify"
  // about the new user. This creates a reverse dependency and, combined
  // with the service calling the repo, forms a circular dependency.
  // The correct approach would be an event emitter or callback injection.
  await createUserInService({
    email: `audit-${input.email}`,
    name: `Audit: ${input.name}`,
    password: "AuditPlaceholder1",
  });

  return user;
}

export async function findUserById(id: string): Promise<User> {
  const user = users.get(id);
  if (!user) {
    throw new NotFoundError("User", id);
  }
  return user;
}

export async function findUsers(
  filters: UserFilters,
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResult<User>> {
  let results = Array.from(users.values());

  if (filters.role) {
    results = results.filter((u) => u.role === filters.role);
  }
  if (filters.search) {
    const search = filters.search.toLowerCase();
    results = results.filter(
      (u) => u.name.toLowerCase().includes(search) || u.email.toLowerCase().includes(search)
    );
  }

  const total = results.length;
  const start = (page - 1) * pageSize;
  const data = results.slice(start, start + pageSize);

  return { data, total, page, pageSize, hasMore: start + pageSize < total };
}

export async function updateUser(id: string, input: UpdateUserInput): Promise<User> {
  const user = await findUserById(id);
  const updated: User = { ...user, ...input, updatedAt: new Date() };
  users.set(id, updated);
  return updated;
}

export async function deleteUser(id: string): Promise<void> {
  if (!users.has(id)) {
    throw new NotFoundError("User", id);
  }
  users.delete(id);
}

function generateId(): string {
  return `usr_${Date.now()}_${Math.random().toString(36).slice(2, 10)}`;
}
```

### Expected Violation

```
ARCH-dependency-direction: src/repositories/user-repo.ts imports from src/services/user-service.ts.
  Repository layer (Layer 3) imports from service layer (Layer 2). In a top-down
  architecture, lower layers must not import from higher layers.
  Location: line 7, import { createUser as createUserInService } from "../services/user-service"
```

### What Makes This Detectable

Stricture resolves the import to `src/services/user-service.ts`, which matches the `src/services/**` pattern (Layer 2). The importing file is in `src/repositories/**` (Layer 3). Layer 3 importing from Layer 2 is a reverse dependency.

---

## V04 -- Circular Dependency A <-> B (ARCH-no-circular-deps)

**Rule violated:** `ARCH-no-circular-deps`

### Violating File: `src/services/user-service.ts`

```typescript
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
```

### Violating File: `src/services/order-service.ts`

```typescript
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
```

### Expected Violation

```
ARCH-no-circular-deps: Circular dependency detected.
  Cycle: src/services/user-service.ts -> src/services/order-service.ts -> src/services/user-service.ts
  user-service.ts imports getOrdersByUser from order-service.ts (line 9),
  order-service.ts imports getUserById from user-service.ts (line 8).
  Resolution: Extract shared logic into a separate module or use dependency injection.
```

### What Makes This Detectable

Stricture builds a dependency graph from all import/require statements. A depth-first search finds the strongly connected component containing both `user-service.ts` and `order-service.ts`. The cycle is reported with the shortest path.

---

## V05 -- Circular Dependency A -> B -> C -> A (ARCH-no-circular-deps, 3-Node Cycle)

**Rule violated:** `ARCH-no-circular-deps`

### Violating Files

**`src/services/user-service.ts`** adds:

```typescript
import { sendWelcomeNotification } from "./notification-service";

export async function createUser(input: CreateUserInput): Promise<User> {
  // ... (create user logic) ...
  const user = await userRepo.createUser(hashedInput);
  await sendWelcomeNotification(user.id, user.email);  // <-- starts the chain
  return user;
}
```

**`src/services/notification-service.ts`** (new file):

```typescript
// notification-service.ts -- Notification dispatch.
// V05: Part of a 3-node circular dependency.

import { logger } from "../shared/logger";
import { getUserById } from "./user-service";  // <-- COMPLETES THE CYCLE

export async function sendWelcomeNotification(userId: string, email: string): Promise<void> {
  logger.info("Sending welcome notification", { userId, email });

  // BUG: This service imports from user-service to "enrich" the notification
  // with user details. But user-service imports from notification-service,
  // creating: user-service -> notification-service -> user-service.
  const user = await getUserById(userId);
  logger.info("Notification sent", { userName: user.name });
}

export async function sendOrderNotification(userId: string, orderId: string): Promise<void> {
  const user = await getUserById(userId);
  logger.info("Order notification sent", { userName: user.name, orderId });
}
```

### Expected Violation

```
ARCH-no-circular-deps: Circular dependency detected.
  Cycle: src/services/user-service.ts -> src/services/notification-service.ts -> src/services/user-service.ts
  This is a 2-hop cycle. user-service imports sendWelcomeNotification from
  notification-service, which imports getUserById from user-service.
```

### What Makes This Detectable

Even though this looks like a 3-node cycle from the description (user -> notification -> user), it is actually a 2-hop cycle in the import graph. Stricture detects cycles of any length. The fix is to pass the user data as a parameter instead of re-importing it, or to use an event emitter for decoupled notification dispatch.

---

## V06 -- Module Boundary Violation: Direct Internal Import (ARCH-module-boundary)

**Rule violated:** `ARCH-module-boundary`

### Violating File: `src/routes/orders.ts`

```typescript
// orders.ts -- Order route handlers.
// V06: Imports directly from auth module internals instead of through index.ts.

import { Router } from "express";
import type { Request, Response, NextFunction } from "express";
import * as orderService from "../services/order-service";
import { validateToken } from "../modules/auth/token-validator";  // <-- VIOLATION
import { AppError, ValidationError } from "../shared/errors";
import { logger } from "../shared/logger";

export const orderRouter = Router();

// -- POST /orders -------------------------------------------------------

orderRouter.post("/", async (req: Request, res: Response, next: NextFunction) => {
  try {
    // Inline auth instead of using middleware
    const authHeader = req.headers.authorization;
    if (!authHeader) {
      next(new ValidationError("Missing authorization"));
      return;
    }

    // BUG: Importing validateToken from the internal file token-validator.ts
    // bypasses the auth module's public API (index.ts). This couples the route
    // to the module's internal structure. If the auth module refactors
    // token-validator.ts into multiple files, this import breaks.
    const payload = validateToken(authHeader.slice(7));

    const order = await orderService.createOrder({
      ...req.body,
      userId: payload.userId,
    });
    res.status(201).json(order);
  } catch (err) {
    next(err);
  }
});
```

### Expected Violation

```
ARCH-module-boundary: src/routes/orders.ts imports from src/modules/auth/token-validator.ts.
  The auth module defines its public API at src/modules/auth/index.ts.
  External consumers must import through the module entry point, not from
  internal files. Use: import { validateToken } from "../modules/auth"
  Location: line 7, import { validateToken } from "../modules/auth/token-validator"
```

### What Makes This Detectable

The manifest declares the auth module with entry point `src/modules/auth/index.ts`. Stricture checks every import that resolves to a file within `src/modules/auth/**`. If the resolved path is anything other than `src/modules/auth/index.ts` and the importing file is outside `src/modules/auth/`, it is a module boundary violation. The import resolves to `src/modules/auth/token-validator.ts`, which is not the entry point.

---

## V07 -- Deep Import into Module Internals (ARCH-import-boundary)

**Rule violated:** `ARCH-import-boundary`

### Violating File: `src/services/billing-service.ts` (new file)

```typescript
// billing-service.ts -- Billing orchestration.
// V07: Imports directly from billing module's internal directory.

import { logger } from "../shared/logger";
import { ValidationError } from "../shared/errors";

// VIOLATION: Importing from the internal directory of the billing module.
// The billing module's public API is at src/modules/billing/index.ts, which
// re-exports createPaymentIntent, capturePayment, refundPayment.
// This import bypasses the public API and reaches into the internal directory.
import { createPaymentIntent } from "../modules/billing/internal/stripe-adapter";  // <-- VIOLATION

export async function processPayment(
  orderId: string,
  amountCents: number,
  currency: string,
  customerId: string
): Promise<{ paymentId: string; status: string }> {
  if (amountCents <= 0) {
    throw new ValidationError("Amount must be positive");
  }

  logger.info("Processing payment", { orderId, amountCents, currency });

  // BUG: By importing from internal/stripe-adapter directly, this service
  // is coupled to the Stripe implementation. If billing switches to a
  // different payment provider, this import path breaks even though the
  // public API (index.ts) would remain stable.
  const result = await createPaymentIntent(amountCents, currency, customerId);

  return { paymentId: result.paymentId, status: result.status };
}

export async function refundOrder(
  orderId: string,
  paymentId: string,
  amountCents?: number
): Promise<void> {
  logger.info("Refunding order", { orderId, paymentId, amountCents });

  // This import is also wrong but we only need to demonstrate one.
  // In the PERFECT version, this would be:
  //   import { refundPayment } from "../modules/billing";
  const { refundPayment } = await import("../modules/billing/internal/stripe-adapter");
  await refundPayment(paymentId, amountCents);
}
```

### Expected Violation

```
ARCH-import-boundary: src/services/billing-service.ts imports from src/modules/billing/internal/stripe-adapter.ts.
  The billing module boundary restricts external access to src/modules/billing/index.ts.
  Importing from src/modules/billing/internal/** violates the module boundary.
  Location: line 10, import { createPaymentIntent } from "../modules/billing/internal/stripe-adapter"
```

### What Makes This Detectable

The manifest declares boundaries for the billing module: `patterns: ["src/modules/billing/**"]` with `exports: ["src/modules/billing/index.ts"]`. Any import from outside `src/modules/billing/` that resolves to a file within `src/modules/billing/**` must target `src/modules/billing/index.ts`. The import targets `src/modules/billing/internal/stripe-adapter.ts`, which is inside the boundary but not the allowed export.

---

## V08 -- File Exceeds 800 Lines (ARCH-max-file-lines)

**Rule violated:** `ARCH-max-file-lines`

### Violating File: `src/services/order-service.ts`

```typescript
// order-service.ts -- Order business logic.
// V08: This file has grown to 847 lines due to accumulated business rules,
// validation logic, and status transition handling that should have been
// refactored into separate modules.

import type { Order, CreateOrderInput, OrderFilters, PaginatedResult, OrderStatus } from "../models/order";
import * as orderRepo from "../repositories/order-repo";
import * as userRepo from "../repositories/user-repo";
import { ValidationError, AppError } from "../shared/errors";
import { logger } from "../shared/logger";

// -- Status transition matrix (lines 12-45) -----------------------------
// ... (34 lines of transition configuration)

// -- Order creation with extensive validation (lines 47-120) ------------
// ... (74 lines including item validation, price calculation, tax logic)

// -- Order retrieval with permission checks (lines 122-180) -------------
// ... (59 lines)

// -- Order listing with complex filters (lines 182-260) -----------------
// ... (79 lines of filter building, sorting, cursor-based pagination)

// -- Status transition with audit logging (lines 262-350) ---------------
// ... (89 lines including pre/post transition hooks)

// -- Cancellation with refund orchestration (lines 352-440) -------------
// ... (89 lines)

// -- Order total recalculation (lines 442-510) --------------------------
// ... (69 lines including discount rules, tax brackets, rounding)

// -- Shipping cost calculation (lines 512-590) --------------------------
// ... (79 lines including weight-based, zone-based, flat-rate logic)

// -- Order export/reporting (lines 592-680) -----------------------------
// ... (89 lines of CSV/JSON export, date range queries)

// -- Email notification triggers (lines 682-750) ------------------------
// ... (69 lines of template selection, variable substitution)

// -- Inventory reservation (lines 752-820) ------------------------------
// ... (69 lines of stock checking, reservation, timeout handling)

// -- Retry logic for external calls (lines 822-847) ---------------------
// ... (26 lines of exponential backoff for payment and shipping APIs)

// The file would contain 847 lines of real code. The sections above
// represent realistic function groups that accumulate in a service file
// over time. The fix is to extract shipping calculation, notification
// dispatch, inventory management, and reporting into separate service files.
```

### Expected Violation

```
ARCH-max-file-lines: src/services/order-service.ts has 847 lines, maximum is 800.
  The file exceeds the configured maximum of 800 lines for source files.
  Consider refactoring: extract shipping logic, notification dispatch,
  inventory management, or reporting into separate service files.
```

### What Makes This Detectable

Stricture counts the lines in each file and compares against the configured `max` value (800 for source files). The file has 847 lines, which exceeds the limit by 47 lines. This is the simplest ARCH rule to implement -- it requires no import analysis, just line counting.

---

## V09 -- Test File Exceeds 1200 Lines (ARCH-max-file-lines)

**Rule violated:** `ARCH-max-file-lines`

### Violating File: `tests/services/order-service.test.ts`

```typescript
// order-service.test.ts -- Order service comprehensive tests.
// V09: Test file has grown to 1350 lines, exceeding the 1200-line override.

import { describe, it, expect, beforeEach, vi } from "vitest";
import * as orderService from "../../src/services/order-service";

vi.mock("../../src/repositories/order-repo");
vi.mock("../../src/repositories/user-repo");

// -- Creation tests (lines 10-180) --------------------------------------
// describe("createOrder", () => { ... 170 lines of test cases ... });

// -- Retrieval tests (lines 182-300) ------------------------------------
// describe("getOrderById", () => { ... 119 lines ... });

// -- Listing/filter tests (lines 302-500) -------------------------------
// describe("listOrders", () => { ... 199 lines of filter combination tests ... });

// -- Status transition tests (lines 502-800) ----------------------------
// describe("transitionOrderStatus", () => {
//   ... 299 lines testing every valid and invalid transition
//   ... 7 source states x 7 target states = 49 combinations
// });

// -- Cancellation tests (lines 802-950) ---------------------------------
// describe("cancelOrder", () => { ... 149 lines ... });

// -- Edge case tests (lines 952-1150) -----------------------------------
// describe("edge cases", () => {
//   ... 199 lines: concurrent modifications, empty items, max amounts,
//   ... unicode in product names, timezone boundaries
// });

// -- Integration-style tests (lines 1152-1350) --------------------------
// describe("order lifecycle", () => {
//   ... 199 lines: full lifecycle from creation through delivery and refund
// });

// The file would contain 1350 lines of real test code. Test files have a
// higher limit (1200 lines via override), but 1350 still exceeds it.
// The fix is to split into order-creation.test.ts, order-transitions.test.ts,
// and order-lifecycle.test.ts.
```

### Expected Violation

```
ARCH-max-file-lines: tests/services/order-service.test.ts has 1350 lines, maximum is 1200.
  The file matches the override pattern "**/*.test.ts" with a limit of 1200
  lines, but the file has 1350 lines. Split into focused test files.
```

### What Makes This Detectable

The manifest configures `overrides: { "**/*.test.ts": 1200 }`. The file matches `**/*.test.ts`, so the override limit of 1200 applies instead of the default 800. At 1350 lines, it exceeds even the elevated limit.

---

## V10 -- Route Accesses Model for Business Logic (ARCH-layer-violation)

**Rule violated:** `ARCH-layer-violation`

### Violating File: `src/routes/users.ts`

```typescript
// users.ts -- User route handlers.
// V10: Route layer imports from model layer and performs business logic
// (filtering, transformation) that belongs in the service layer.

import { Router } from "express";
import type { Request, Response, NextFunction } from "express";
import * as userService from "../services/user-service";
import { authenticate } from "../middleware/auth";
import { AppError } from "../shared/errors";
import { logger } from "../shared/logger";
import type { User, UserRole } from "../models/user";  // <-- VIOLATION: importing from model

export const userRouter = Router();

// This endpoint performs business logic directly in the route handler
// by importing and using the User type for data manipulation that
// should live in the service layer.
userRouter.get("/", authenticate, async (req: Request, res: Response, next: NextFunction) => {
  try {
    const result = await userService.listUsers({}, 1, 1000);

    // BUG: The route performs filtering and transformation logic that
    // belongs in the service layer. It imports the User type and UserRole
    // enum to do complex business logic inline. This violates the
    // layer-violation rule which forbids routes from accessing models.
    const filtered: User[] = result.data.filter((user: User) => {
      if (req.query["minAge"]) {
        const minAge = parseInt(req.query["minAge"] as string, 10);
        const ageMs = Date.now() - user.createdAt.getTime();
        const ageDays = ageMs / (1000 * 60 * 60 * 24);
        if (ageDays < minAge) return false;
      }

      if (req.query["role"]) {
        const allowedRoles: UserRole[] = (req.query["role"] as string).split(",") as UserRole[];
        if (!allowedRoles.includes(user.role)) return false;
      }

      return true;
    });

    // Also performing aggregation logic that belongs in the service
    const roleDistribution: Record<UserRole, number> = {
      admin: 0,
      member: 0,
      viewer: 0,
    };
    for (const user of filtered) {
      roleDistribution[user.role]++;
    }

    res.json({
      users: filtered,
      meta: {
        total: filtered.length,
        roleDistribution,
      },
    });
  } catch (err) {
    next(err);
  }
});
```

### Expected Violation

```
ARCH-layer-violation: src/routes/users.ts imports from model layer.
  The route layer has forbidden access to the model layer. Routes must
  delegate data access and business logic to the service layer.
  Manifest rule: { layer: route, forbidden: [repository, model] }
  Location: line 11, import type { User, UserRole } from "../models/user"
```

### What Makes This Detectable

The manifest declares `rules: [{ layer: route, forbidden: [repository, model] }]`. The route file imports from `src/models/user.ts`, which matches the `src/models/**` pattern (model layer). Even though it uses `import type`, the manifest rule does not exempt type-only imports by default. The route is performing business logic (filtering, aggregation) that belongs in the service layer, and the type import is a signal of this layer violation.

---

## V11 -- Dynamic Import Violating Module Boundary (ARCH-module-boundary)

**Rule violated:** `ARCH-module-boundary`

### Violating File: `src/middleware/auth.ts`

```typescript
// auth.ts -- Authentication middleware.
// V11: Uses a dynamic import to reach into auth module internals.

import type { Request, Response, NextFunction } from "express";
import { UnauthorizedError, ForbiddenError } from "../shared/errors";
import { logger } from "../shared/logger";
import type { UserRole } from "../models/user";

export async function authenticate(req: Request, _res: Response, next: NextFunction): Promise<void> {
  const authHeader = req.headers.authorization;
  if (!authHeader || !authHeader.startsWith("Bearer ")) {
    next(new UnauthorizedError("Missing or invalid Authorization header"));
    return;
  }

  const token = authHeader.slice(7);

  try {
    // VIOLATION: Dynamic import reaching into module internals.
    // Instead of importing from "../modules/auth" (the public entry point),
    // this reaches directly into the internal jwt-utils file.
    // Dynamic imports are still imports -- they create a dependency on
    // the internal file structure of the auth module.
    const { decodePayload } = await import("../modules/auth/internal/jwt-utils");  // <-- VIOLATION

    // Even worse: jwt-utils is not even exported through the module's
    // public API. The auth module exports validateToken and generateToken,
    // but not the low-level decodePayload function. This creates a
    // dependency on an implementation detail.
    const payload = decodePayload(token);

    req.user = { userId: payload.userId, role: payload.role };
    logger.debug("Authenticated request", { userId: payload.userId });
    next();
  } catch (err) {
    next(err);
  }
}

export function requireRole(...roles: UserRole[]) {
  return (req: Request, _res: Response, next: NextFunction): void => {
    if (!req.user) {
      next(new UnauthorizedError());
      return;
    }

    if (!roles.includes(req.user.role as UserRole)) {
      next(new ForbiddenError(`Requires one of: ${roles.join(", ")}`));
      return;
    }

    next();
  };
}
```

### Expected Violation

```
ARCH-module-boundary: src/middleware/auth.ts dynamically imports src/modules/auth/internal/jwt-utils.ts.
  The auth module defines its public API at src/modules/auth/index.ts.
  Dynamic imports (await import()) still create dependencies on internal
  module structure. Use the public API: import { validateToken } from "../modules/auth"
  Location: line 23, await import("../modules/auth/internal/jwt-utils")
```

### What Makes This Detectable

Stricture parses both static `import` declarations and dynamic `import()` expressions. The dynamic import resolves to `src/modules/auth/internal/jwt-utils.ts`, which is inside the auth module boundary. Since the importing file is outside `src/modules/auth/`, and the resolved path is not the module's entry point, this is a boundary violation. Dynamic imports do not receive special treatment -- they create the same dependency as static imports.

---

## V12 -- Re-export Chain Leaking Internal Module (ARCH-import-boundary)

**Rule violated:** `ARCH-import-boundary`

### Violating File: `src/modules/auth/index.ts`

```typescript
// index.ts -- Auth module public API.
// V12: Re-exports from another module's internal directory.

export { validateToken, generateToken, type TokenPayload } from "./token-validator";

// VIOLATION: This re-export reaches into the billing module's internal
// directory. Even though it goes through auth's own index.ts (which is the
// auth module's public API), it creates a dependency from the auth module
// into billing's internals -- violating billing's module boundary.
export { createPaymentIntent } from "../billing/internal/stripe-adapter";  // <-- VIOLATION
```

### Violating Context: `src/routes/orders.ts`

```typescript
// orders.ts -- Order route handlers.
// A consumer that innocently imports from the auth module's public API,
// not knowing that auth leaks billing internals.

import { Router } from "express";
import type { Request, Response, NextFunction } from "express";
import * as orderService from "../services/order-service";
import { authenticate } from "../middleware/auth";
import { createPaymentIntent } from "../modules/auth";  // Looks correct, but auth leaks billing internals
import { AppError } from "../shared/errors";

export const orderRouter = Router();

orderRouter.post("/", authenticate, async (req: Request, res: Response, next: NextFunction) => {
  try {
    const order = await orderService.createOrder({ ...req.body, userId: req.user!.userId });

    // Using createPaymentIntent that was leaked through auth's index.ts.
    // The route thinks this is part of the auth module's API, but it's
    // actually billing's internal Stripe adapter.
    const payment = await createPaymentIntent(order.totalCents, order.currency, req.user!.userId);

    res.status(201).json({ order, payment });
  } catch (err) {
    next(err);
  }
});
```

### Expected Violation

```
ARCH-import-boundary: src/modules/auth/index.ts imports from src/modules/billing/internal/stripe-adapter.ts.
  The billing module boundary restricts external access to src/modules/billing/index.ts.
  The re-export in auth/index.ts creates a cross-module boundary violation.
  Even though auth/index.ts is the auth module's public entry point, it is
  reaching into billing's internals, which violates billing's boundary.
  Location: line 7, export { createPaymentIntent } from "../billing/internal/stripe-adapter"
```

### What Makes This Detectable

Stricture tracks the full dependency graph, including re-exports. When `src/modules/auth/index.ts` has an export statement that resolves to `src/modules/billing/internal/stripe-adapter.ts`, Stricture checks whether this resolved path is the billing module's allowed export (`src/modules/billing/index.ts`). It is not -- it is an internal file. The violation is reported at the re-export site (auth's index.ts), not at the downstream consumer (orders.ts), because auth's index.ts is where the boundary is crossed.

This is a particularly insidious pattern because consumers of the auth module believe they are using auth's public API. The leak is invisible to them -- only the auth module's maintainer (or Stricture) can detect that billing internals are being exposed.

---

## Violation Summary

| ID | Rule | File | Import/Issue |
|----|------|------|--------------|
| V01 | ARCH-dependency-direction | `src/routes/users.ts` | Imports `repositories/user-repo` (skips service layer) |
| V02 | ARCH-dependency-direction | `src/services/order-service.ts` | Imports `routes/orders` (reverse direction) |
| V03 | ARCH-dependency-direction | `src/repositories/user-repo.ts` | Imports `services/user-service` (reverse direction) |
| V04 | ARCH-no-circular-deps | `services/user-service.ts` <-> `services/order-service.ts` | Mutual import creates 2-node cycle |
| V05 | ARCH-no-circular-deps | `user-service` -> `notification-service` -> `user-service` | 3-node cycle through notification |
| V06 | ARCH-module-boundary | `src/routes/orders.ts` | Imports `modules/auth/token-validator` (bypasses index.ts) |
| V07 | ARCH-import-boundary | `src/services/billing-service.ts` | Imports `modules/billing/internal/stripe-adapter` (deep internal) |
| V08 | ARCH-max-file-lines | `src/services/order-service.ts` | 847 lines (max: 800) |
| V09 | ARCH-max-file-lines | `tests/services/order-service.test.ts` | 1350 lines (max: 1200 for test files) |
| V10 | ARCH-layer-violation | `src/routes/users.ts` | Imports `models/user` and performs business logic in route |
| V11 | ARCH-module-boundary | `src/middleware/auth.ts` | Dynamic import into `modules/auth/internal/jwt-utils` |
| V12 | ARCH-import-boundary | `src/modules/auth/index.ts` | Re-exports from `modules/billing/internal/stripe-adapter` |

---

## Detection Strategy

### How to Run This Validation

1. **False positive test:** Scan the PERFECT codebase. Stricture must report zero ARCH violations.
2. **Detection test:** For each V01-V12, apply the modification to the PERFECT codebase and scan. Stricture must report exactly the expected violation.
3. **Isolation test:** Each violation is independent. Applying V01 must not trigger V02's rule, and vice versa.
4. **Combination test:** Apply V01 + V04 + V08 simultaneously. Stricture must report all three violations without interference.

### Rule Coverage Matrix

| Rule | PERFECT (TN) | Violations (TP) |
|------|--------------|-----------------|
| ARCH-dependency-direction | All imports flow top-down | V01, V02, V03 |
| ARCH-no-circular-deps | No cycles in dependency graph | V04, V05 |
| ARCH-module-boundary | All external imports use module entry points | V06, V11 |
| ARCH-import-boundary | No imports from internal module directories | V07, V12 |
| ARCH-max-file-lines | All files under limits | V08, V09 |
| ARCH-layer-violation | Routes never access repos or models | V01 (secondary), V10 |
