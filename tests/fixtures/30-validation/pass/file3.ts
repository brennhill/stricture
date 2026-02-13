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
