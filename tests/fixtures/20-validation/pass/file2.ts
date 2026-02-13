// src/server/routes/users.ts — Express route handlers for User API

import express, { Request, Response } from "express";
import { User, CreateUserRequest, UpdateUserRequest, PaginatedResponse } from "../../shared/types.js";
import { v4 as uuidv4 } from "uuid";

const router = express.Router();

// In-memory storage for demo purposes
const users: Map<string, User> = new Map();

// Validation helpers
function isValidEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

function validateCreateUserRequest(req: CreateUserRequest): string | null {
  if (!req.name || req.name.length === 0 || req.name.length > 255) {
    return "name must be between 1 and 255 characters";
  }
  if (!req.email || !isValidEmail(req.email)) {
    return "email must be valid";
  }
  if (!req.role || !["admin", "editor", "viewer"].includes(req.role)) {
    return "role must be admin, editor, or viewer";
  }
  return null;
}

// POST /api/users — Create user
router.post("/users", (req: Request, res: Response) => {
  const body = req.body as CreateUserRequest;

  const validationError = validateCreateUserRequest(body);
  if (validationError) {
    res.status(400).json({ error: validationError });
    return;
  }

  // Check for duplicate email
  for (const user of users.values()) {
    if (user.email === body.email) {
      res.status(409).json({ error: "email already exists" });
      return;
    }
  }

  const now = new Date().toISOString();
  const user: User = {
    id: uuidv4(),
    name: body.name,
    email: body.email,
    role: body.role,
    avatar: null,
    createdAt: now,
    updatedAt: now,
  };

  users.set(user.id, user);
  res.status(201).json(user);
});

// GET /api/users/:id — Get user by ID
router.get("/users/:id", (req: Request, res: Response) => {
  const { id } = req.params;
  const user = users.get(id);

  if (!user) {
    res.status(404).json({ error: "user not found" });
    return;
  }

  res.status(200).json(user);
});

// GET /api/users — List users with cursor pagination
router.get("/users", (req: Request, res: Response) => {
  const cursor = req.query.cursor as string | undefined;
  const limit = parseInt(req.query.limit as string || "10", 10);

  const allUsers = Array.from(users.values()).sort((a, b) =>
    a.createdAt.localeCompare(b.createdAt)
  );

  let startIndex = 0;
  if (cursor) {
    startIndex = allUsers.findIndex(u => u.id === cursor) + 1;
  }

  const pageUsers = allUsers.slice(startIndex, startIndex + limit);
  const hasMore = startIndex + limit < allUsers.length;
  const nextCursor = hasMore ? pageUsers[pageUsers.length - 1].id : null;

  const response: PaginatedResponse<User> = {
    data: pageUsers,
    cursor: nextCursor,
    hasMore,
  };

  res.status(200).json(response);
});

// PATCH /api/users/:id — Update user
router.patch("/users/:id", (req: Request, res: Response) => {
  const { id } = req.params;
  const body = req.body as UpdateUserRequest;

  const user = users.get(id);
  if (!user) {
    res.status(404).json({ error: "user not found" });
    return;
  }

  // Validate updated fields
  if (body.name !== undefined) {
    if (body.name.length === 0 || body.name.length > 255) {
      res.status(400).json({ error: "name must be between 1 and 255 characters" });
      return;
    }
    user.name = body.name;
  }

  if (body.email !== undefined) {
    if (!isValidEmail(body.email)) {
      res.status(400).json({ error: "email must be valid" });
      return;
    }
    // Check for duplicate email
    for (const [otherId, otherUser] of users.entries()) {
      if (otherId !== id && otherUser.email === body.email) {
        res.status(409).json({ error: "email already exists" });
        return;
      }
    }
    user.email = body.email;
  }

  if (body.role !== undefined) {
    if (!["admin", "editor", "viewer"].includes(body.role)) {
      res.status(400).json({ error: "role must be admin, editor, or viewer" });
      return;
    }
    user.role = body.role;
  }

  user.updatedAt = new Date().toISOString();
  users.set(id, user);
  res.status(200).json(user);
});

export default router;
