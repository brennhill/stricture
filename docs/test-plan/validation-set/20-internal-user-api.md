# 20 — Internal User API (TypeScript Server + Client)

**Why included:** Both sides of contract in one repo. CTR-shared-type-sync, CTR-dual-test, CTR-json-tag-match.

## Architecture

This validation set demonstrates a common pattern: TypeScript Express server + TypeScript fetch client sharing type definitions in the same repository.

**File Structure:**
- `src/shared/types.ts` — Shared TypeScript types for User, CreateUserRequest, UpdateUserRequest
- `src/server/routes/users.ts` — Express route handlers (POST, GET, PATCH)
- `src/client/api-client.ts` — Fetch-based API client
- `tests/server/users.test.ts` — Server-side tests (supertest)
- `tests/client/api-client.test.ts` — Client-side tests (fetch mock)

**Contracts Exercised:**
- **CTR-shared-type-sync:** Both server and client reference `src/shared/types.ts`
- **CTR-dual-test:** Same scenarios tested on both server and client sides
- **CTR-json-tag-match:** Field names in JSON must match TypeScript property names

---

## Manifest Fragment

```json
{
  "endpoints": [
    {
      "id": "internal-users-create",
      "method": "POST",
      "path": "/api/users",
      "description": "Create new user",
      "required_fields": ["name", "email", "role"],
      "shared_types": ["User", "CreateUserRequest"],
      "shared_types_file": "src/shared/types.ts"
    },
    {
      "id": "internal-users-get",
      "method": "GET",
      "path": "/api/users/:id",
      "description": "Get user by ID",
      "shared_types": ["User"],
      "shared_types_file": "src/shared/types.ts"
    },
    {
      "id": "internal-users-list",
      "method": "GET",
      "path": "/api/users",
      "description": "List users with cursor pagination",
      "query_params": ["cursor", "limit"],
      "shared_types": ["User", "PaginatedResponse"],
      "shared_types_file": "src/shared/types.ts"
    },
    {
      "id": "internal-users-update",
      "method": "PATCH",
      "path": "/api/users/:id",
      "description": "Update user fields",
      "shared_types": ["User", "UpdateUserRequest"],
      "shared_types_file": "src/shared/types.ts"
    }
  ]
}
```

---

## PERFECT — All Files Correct

### src/shared/types.ts

```typescript
// src/shared/types.ts — Shared type definitions for User API

export type UserRole = "admin" | "editor" | "viewer";

export interface User {
  id: string;
  name: string;
  email: string;
  role: UserRole;
  avatar: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface CreateUserRequest {
  name: string;
  email: string;
  role: UserRole;
}

export interface UpdateUserRequest {
  name?: string;
  email?: string;
  role?: UserRole;
}

export interface PaginatedResponse<T> {
  data: T[];
  cursor: string | null;
  hasMore: boolean;
}
```

### src/server/routes/users.ts

```typescript
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
```

### src/client/api-client.ts

```typescript
// src/client/api-client.ts — Fetch-based API client for User API

import { User, CreateUserRequest, UpdateUserRequest, PaginatedResponse } from "../shared/types.js";

export class UserApiClient {
  constructor(private baseUrl: string, private authToken: string) {}

  async createUser(req: CreateUserRequest): Promise<User> {
    try {
      const response = await fetch(`${this.baseUrl}/api/users`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${this.authToken}`,
        },
        body: JSON.stringify(req),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || `HTTP ${response.status}`);
      }

      return await response.json() as User;
    } catch (err) {
      throw new Error(`Failed to create user: ${(err as Error).message}`);
    }
  }

  async getUser(id: string): Promise<User> {
    try {
      const response = await fetch(`${this.baseUrl}/api/users/${id}`, {
        headers: {
          Authorization: `Bearer ${this.authToken}`,
        },
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || `HTTP ${response.status}`);
      }

      return await response.json() as User;
    } catch (err) {
      throw new Error(`Failed to get user: ${(err as Error).message}`);
    }
  }

  async listUsers(cursor?: string, limit = 10): Promise<PaginatedResponse<User>> {
    try {
      const params = new URLSearchParams();
      if (cursor) params.set("cursor", cursor);
      params.set("limit", limit.toString());

      const response = await fetch(`${this.baseUrl}/api/users?${params}`, {
        headers: {
          Authorization: `Bearer ${this.authToken}`,
        },
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || `HTTP ${response.status}`);
      }

      return await response.json() as PaginatedResponse<User>;
    } catch (err) {
      throw new Error(`Failed to list users: ${(err as Error).message}`);
    }
  }

  async updateUser(id: string, req: UpdateUserRequest): Promise<User> {
    try {
      const response = await fetch(`${this.baseUrl}/api/users/${id}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${this.authToken}`,
        },
        body: JSON.stringify(req),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || `HTTP ${response.status}`);
      }

      return await response.json() as User;
    } catch (err) {
      throw new Error(`Failed to update user: ${(err as Error).message}`);
    }
  }
}
```

### tests/server/users.test.ts

```typescript
// tests/server/users.test.ts — Server-side tests using supertest

import request from "supertest";
import express from "express";
import userRouter from "../../src/server/routes/users.js";
import { User } from "../../src/shared/types.js";

const app = express();
app.use(express.json());
app.use("/api", userRouter);

describe("User API - Server Tests", () => {
  describe("POST /api/users", () => {
    it("should create user with valid fields", async () => {
      const response = await request(app)
        .post("/api/users")
        .send({ name: "Alice", email: "alice@example.com", role: "editor" });

      expect(response.status).toBe(201);
      expect(response.body).toMatchObject({
        name: "Alice",
        email: "alice@example.com",
        role: "editor",
      });
      expect(response.body.id).toBeDefined();
      expect(response.body.avatar).toBeNull();
    });

    it("should reject missing name", async () => {
      const response = await request(app)
        .post("/api/users")
        .send({ email: "bob@example.com", role: "admin" });

      expect(response.status).toBe(400);
      expect(response.body.error).toContain("name");
    });

    it("should reject name longer than 255 characters", async () => {
      const response = await request(app)
        .post("/api/users")
        .send({ name: "x".repeat(256), email: "long@example.com", role: "viewer" });

      expect(response.status).toBe(400);
      expect(response.body.error).toContain("name");
    });

    it("should reject invalid email", async () => {
      const response = await request(app)
        .post("/api/users")
        .send({ name: "Charlie", email: "not-an-email", role: "editor" });

      expect(response.status).toBe(400);
      expect(response.body.error).toContain("email");
    });

    it("should reject duplicate email", async () => {
      await request(app)
        .post("/api/users")
        .send({ name: "Diana", email: "diana@example.com", role: "admin" });

      const response = await request(app)
        .post("/api/users")
        .send({ name: "Diana2", email: "diana@example.com", role: "editor" });

      expect(response.status).toBe(409);
      expect(response.body.error).toContain("email already exists");
    });

    it("should reject invalid role", async () => {
      const response = await request(app)
        .post("/api/users")
        .send({ name: "Eve", email: "eve@example.com", role: "superadmin" });

      expect(response.status).toBe(400);
      expect(response.body.error).toContain("role");
    });
  });

  describe("GET /api/users/:id", () => {
    it("should return user by ID", async () => {
      const createResponse = await request(app)
        .post("/api/users")
        .send({ name: "Frank", email: "frank@example.com", role: "viewer" });

      const userId = createResponse.body.id;

      const response = await request(app).get(`/api/users/${userId}`);
      expect(response.status).toBe(200);
      expect(response.body).toMatchObject({
        id: userId,
        name: "Frank",
        email: "frank@example.com",
        role: "viewer",
      });
    });

    it("should return 404 for non-existent user", async () => {
      const response = await request(app).get("/api/users/non-existent-id");
      expect(response.status).toBe(404);
      expect(response.body.error).toContain("user not found");
    });
  });

  describe("GET /api/users", () => {
    it("should list users with cursor pagination", async () => {
      await request(app).post("/api/users").send({ name: "User1", email: "u1@example.com", role: "admin" });
      await request(app).post("/api/users").send({ name: "User2", email: "u2@example.com", role: "editor" });
      await request(app).post("/api/users").send({ name: "User3", email: "u3@example.com", role: "viewer" });

      const response = await request(app).get("/api/users?limit=2");
      expect(response.status).toBe(200);
      expect(response.body.data).toHaveLength(2);
      expect(response.body.hasMore).toBe(true);
      expect(response.body.cursor).toBeDefined();
    });

    it("should paginate to next page using cursor", async () => {
      const firstPage = await request(app).get("/api/users?limit=2");
      const cursor = firstPage.body.cursor;

      const secondPage = await request(app).get(`/api/users?cursor=${cursor}&limit=2`);
      expect(secondPage.status).toBe(200);
      expect(secondPage.body.data.length).toBeGreaterThan(0);
    });
  });

  describe("PATCH /api/users/:id", () => {
    it("should update user name", async () => {
      const createResponse = await request(app)
        .post("/api/users")
        .send({ name: "George", email: "george@example.com", role: "editor" });

      const userId = createResponse.body.id;

      const response = await request(app)
        .patch(`/api/users/${userId}`)
        .send({ name: "George Updated" });

      expect(response.status).toBe(200);
      expect(response.body.name).toBe("George Updated");
      expect(response.body.email).toBe("george@example.com");
    });

    it("should reject name longer than 255 characters", async () => {
      const createResponse = await request(app)
        .post("/api/users")
        .send({ name: "Hannah", email: "hannah@example.com", role: "viewer" });

      const userId = createResponse.body.id;

      const response = await request(app)
        .patch(`/api/users/${userId}`)
        .send({ name: "x".repeat(256) });

      expect(response.status).toBe(400);
      expect(response.body.error).toContain("name");
    });

    it("should reject invalid email", async () => {
      const createResponse = await request(app)
        .post("/api/users")
        .send({ name: "Ian", email: "ian@example.com", role: "admin" });

      const userId = createResponse.body.id;

      const response = await request(app)
        .patch(`/api/users/${userId}`)
        .send({ email: "not-valid" });

      expect(response.status).toBe(400);
      expect(response.body.error).toContain("email");
    });

    it("should reject duplicate email on update", async () => {
      await request(app).post("/api/users").send({ name: "Jack", email: "jack@example.com", role: "editor" });
      const createResponse = await request(app)
        .post("/api/users")
        .send({ name: "Jill", email: "jill@example.com", role: "viewer" });

      const jillId = createResponse.body.id;

      const response = await request(app)
        .patch(`/api/users/${jillId}`)
        .send({ email: "jack@example.com" });

      expect(response.status).toBe(409);
      expect(response.body.error).toContain("email already exists");
    });

    it("should return 404 for non-existent user", async () => {
      const response = await request(app)
        .patch("/api/users/non-existent-id")
        .send({ name: "Nobody" });

      expect(response.status).toBe(404);
      expect(response.body.error).toContain("user not found");
    });
  });
});
```

---

## BUG CASES — Contract Violations

### B01 — No Error Handling (TQ-error-path-coverage)

**Bug:** Client fetch() call with no try/catch wrapper
**Expected violation:** `TQ-error-path-coverage`

**Server:**
```typescript
// src/server/routes/users.ts (no changes, server is correct)
router.post("/users", (req: Request, res: Response) => {
  const body = req.body as CreateUserRequest;
  const validationError = validateCreateUserRequest(body);
  if (validationError) {
    res.status(400).json({ error: validationError });
    return;
  }
  // ... rest of handler
});
```

**Client (BUGGY):**
```typescript
// src/client/api-client.ts — Missing try/catch
async createUser(req: CreateUserRequest): Promise<User> {
  const response = await fetch(`${this.baseUrl}/api/users`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${this.authToken}`,
    },
    body: JSON.stringify(req),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || `HTTP ${response.status}`);
  }

  return await response.json() as User;
  // No try/catch — network errors will crash caller
}
```

**Why Stricture catches this:** Server test file has error path coverage (400, 409, 404 tests). Client test file must match. If client has no try/catch, Stricture flags `TQ-error-path-coverage` violation because contract requires both sides handle errors.

---

### B02 — No Status Code Check (CTR-status-code-handling)

**Bug:** Client ignores response.ok, treats all responses as success
**Expected violation:** `CTR-status-code-handling`

**Server:**
```typescript
// src/server/routes/users.ts (no changes, server returns 404 correctly)
router.get("/users/:id", (req: Request, res: Response) => {
  const { id } = req.params;
  const user = users.get(id);

  if (!user) {
    res.status(404).json({ error: "user not found" });
    return;
  }

  res.status(200).json(user);
});
```

**Client (BUGGY):**
```typescript
// src/client/api-client.ts — Missing response.ok check
async getUser(id: string): Promise<User> {
  try {
    const response = await fetch(`${this.baseUrl}/api/users/${id}`, {
      headers: {
        Authorization: `Bearer ${this.authToken}`,
      },
    });

    // BUG: No check for response.ok
    // Will try to parse 404 error JSON as User type
    return await response.json() as User;
  } catch (err) {
    throw new Error(`Failed to get user: ${(err as Error).message}`);
  }
}
```

**Why Stricture catches this:** Server returns 404 status code. Client must check `response.ok` or `response.status` before parsing. Stricture compares server status codes (200, 404) against client handling logic. Missing check = `CTR-status-code-handling` violation.

---

### B03 — Shallow Test Assertions (TQ-no-shallow-assertions)

**Bug:** Test only checks if result is defined, not actual values
**Expected violation:** `TQ-no-shallow-assertions`

**Server Test:**
```typescript
// tests/server/users.test.ts (correct, deep assertions)
it("should create user with valid fields", async () => {
  const response = await request(app)
    .post("/api/users")
    .send({ name: "Alice", email: "alice@example.com", role: "editor" });

  expect(response.status).toBe(201);
  expect(response.body).toMatchObject({
    name: "Alice",
    email: "alice@example.com",
    role: "editor",
  });
  expect(response.body.id).toBeDefined();
  expect(response.body.avatar).toBeNull();
});
```

**Client Test (BUGGY):**
```typescript
// tests/client/api-client.test.ts — Shallow assertions
it("should create user", async () => {
  const client = new UserApiClient("http://localhost:3000", "test-token");
  const result = await client.createUser({
    name: "Alice",
    email: "alice@example.com",
    role: "editor",
  });

  // BUG: Only checks if result exists, not field values
  expect(result).toBeDefined();
  expect(result.id).toBeDefined();
  // Missing: name, email, role, avatar, createdAt, updatedAt checks
});
```

**Why Stricture catches this:** Server test validates all fields (name, email, role, avatar). Client test must match same assertion depth per `CTR-dual-test` rule. Shallow assertions = `TQ-no-shallow-assertions` violation.

---

### B04 — Missing Negative Tests (TQ-negative-cases)

**Bug:** No test coverage for 404, 500, or empty list scenarios
**Expected violation:** `TQ-negative-cases`

**Server Test:**
```typescript
// tests/server/users.test.ts (has negative test)
it("should return 404 for non-existent user", async () => {
  const response = await request(app).get("/api/users/non-existent-id");
  expect(response.status).toBe(404);
  expect(response.body.error).toContain("user not found");
});
```

**Client Test (BUGGY):**
```typescript
// tests/client/api-client.test.ts — Missing negative tests
describe("UserApiClient", () => {
  it("should get user by ID", async () => {
    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ id: "123", name: "Alice", email: "alice@example.com", role: "editor", avatar: null, createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z" }),
    });

    const client = new UserApiClient("http://localhost:3000", "test-token");
    const user = await client.getUser("123");
    expect(user.id).toBe("123");
  });

  // BUG: No test for 404 case
  // Missing: it("should throw error for non-existent user", ...)
});
```

**Why Stricture catches this:** Server test suite has 404 test case. Per `CTR-dual-test`, client must have matching negative test. Missing negative test = `TQ-negative-cases` violation.

---

### B05 — Request Missing Required Fields (CTR-request-shape)

**Bug:** Client POST omits required "email" field
**Expected violation:** `CTR-request-shape`

**Server:**
```typescript
// src/server/routes/users.ts (validates required fields)
function validateCreateUserRequest(req: CreateUserRequest): string | null {
  if (!req.name || req.name.length === 0 || req.name.length > 255) {
    return "name must be between 1 and 255 characters";
  }
  if (!req.email || !isValidEmail(req.email)) {
    return "email must be valid";  // Server REQUIRES email
  }
  if (!req.role || !["admin", "editor", "viewer"].includes(req.role)) {
    return "role must be admin, editor, or viewer";
  }
  return null;
}
```

**Client (BUGGY):**
```typescript
// src/client/api-client.ts — Omits email field
async createUser(req: CreateUserRequest): Promise<User> {
  try {
    const response = await fetch(`${this.baseUrl}/api/users`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${this.authToken}`,
      },
      body: JSON.stringify({
        name: req.name,
        role: req.role,
        // BUG: Missing email field in request body
      }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    return await response.json() as User;
  } catch (err) {
    throw new Error(`Failed to create user: ${(err as Error).message}`);
  }
}
```

**Why Stricture catches this:** Manifest declares `required_fields: ["name", "email", "role"]`. Server validates all three. Client only sends name and role. Stricture compares request body against manifest + server validation = `CTR-request-shape` violation.

---

### B06 — Response Type Mismatch (CTR-response-shape)

**Bug:** Client type has "avatar" field server doesn't return
**Expected violation:** `CTR-response-shape`

**Server:**
```typescript
// src/server/routes/users.ts (returns all User fields including avatar)
const user: User = {
  id: uuidv4(),
  name: body.name,
  email: body.email,
  role: body.role,
  avatar: null,  // Server DOES return avatar
  createdAt: now,
  updatedAt: now,
};

users.set(user.id, user);
res.status(201).json(user);
```

**Client (BUGGY):**
```typescript
// src/shared/types.ts — Client type expects "profilePicture" instead of "avatar"
export interface User {
  id: string;
  name: string;
  email: string;
  role: UserRole;
  profilePicture: string | null;  // BUG: Field name mismatch (avatar → profilePicture)
  createdAt: string;
  updatedAt: string;
}
```

**Why Stricture catches this:** Server returns `avatar` field in JSON response. Shared type expects `profilePicture`. Field name mismatch = `CTR-response-shape` violation (JSON field "avatar" not present in TypeScript type).

---

### B07 — Wrong Field Types (CTR-manifest-conformance)

**Bug:** Client treats ID as number, server sends string
**Expected violation:** `CTR-manifest-conformance`

**Server:**
```typescript
// src/server/routes/users.ts (returns string ID)
const user: User = {
  id: uuidv4(),  // Returns UUID string like "a1b2c3d4-..."
  name: body.name,
  email: body.email,
  role: body.role,
  avatar: null,
  createdAt: now,
  updatedAt: now,
};

res.status(201).json(user);
```

**Client (BUGGY):**
```typescript
// src/shared/types.ts — Client type expects number ID
export interface User {
  id: number;  // BUG: Server sends string, client expects number
  name: string;
  email: string;
  role: UserRole;
  avatar: string | null;
  createdAt: string;
  updatedAt: string;
}
```

**Why Stricture catches this:** Manifest declares `shared_types: ["User"]`. Server implementation sends `id: string` (UUID). Client type expects `id: number`. Type mismatch on shared type = `CTR-manifest-conformance` violation.

---

### B08 — Incomplete Enum Handling (CTR-strictness-parity)

**Bug:** Client handles "active"/"inactive", server also sends "suspended"
**Expected violation:** `CTR-strictness-parity`

**Server:**
```typescript
// src/shared/types.ts (server defines 3 roles)
export type UserRole = "admin" | "editor" | "viewer";

// src/server/routes/users.ts (server validates all 3 roles)
function validateCreateUserRequest(req: CreateUserRequest): string | null {
  if (!req.role || !["admin", "editor", "viewer"].includes(req.role)) {
    return "role must be admin, editor, or viewer";
  }
  return null;
}
```

**Client (BUGGY):**
```typescript
// src/shared/types.ts — Client only handles 2 of 3 roles
export type UserRole = "admin" | "editor";  // BUG: Missing "viewer" from enum

// src/client/api-client.ts
async createUser(req: CreateUserRequest): Promise<User> {
  // Client can only create admin/editor, but server accepts viewer
  // Runtime: client receives "viewer" role in response, TypeScript type doesn't allow it
}
```

**Why Stricture catches this:** Server validates 3 role values (admin, editor, viewer). Client type only defines 2 (admin, editor). Enum completeness mismatch on shared type = `CTR-strictness-parity` violation.

---

### B09 — Missing Range Validation (CTR-strictness-parity)

**Bug:** No validation on page/limit bounds, allows negative or extreme values
**Expected violation:** `CTR-strictness-parity`

**Server:**
```typescript
// src/server/routes/users.ts (validates limit bounds)
router.get("/users", (req: Request, res: Response) => {
  const cursor = req.query.cursor as string | undefined;
  const limit = parseInt(req.query.limit as string || "10", 10);

  // Server validates limit range
  if (limit < 1 || limit > 100) {
    res.status(400).json({ error: "limit must be between 1 and 100" });
    return;
  }

  const allUsers = Array.from(users.values()).sort((a, b) =>
    a.createdAt.localeCompare(b.createdAt)
  );

  let startIndex = 0;
  if (cursor) {
    startIndex = allUsers.findIndex(u => u.id === cursor) + 1;
  }

  const pageUsers = allUsers.slice(startIndex, startIndex + limit);
  // ... rest of handler
});
```

**Client (BUGGY):**
```typescript
// src/client/api-client.ts — No limit validation
async listUsers(cursor?: string, limit = 10): Promise<PaginatedResponse<User>> {
  try {
    const params = new URLSearchParams();
    if (cursor) params.set("cursor", cursor);
    params.set("limit", limit.toString());  // BUG: No check for limit < 1 or limit > 100

    const response = await fetch(`${this.baseUrl}/api/users?${params}`, {
      headers: {
        Authorization: `Bearer ${this.authToken}`,
      },
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    return await response.json() as PaginatedResponse<User>;
  } catch (err) {
    throw new Error(`Failed to list users: ${(err as Error).message}`);
  }
}
```

**Why Stricture catches this:** Server validates `limit` parameter (1-100 range). Client accepts any number without validation. Strictness parity violation = client should reject invalid limits before sending request. `CTR-strictness-parity` violation.

---

### B10 — Format Not Validated (CTR-strictness-parity)

**Bug:** Email not validated with regex on client side
**Expected violation:** `CTR-strictness-parity`

**Server:**
```typescript
// src/server/routes/users.ts (validates email format)
function isValidEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

function validateCreateUserRequest(req: CreateUserRequest): string | null {
  if (!req.email || !isValidEmail(req.email)) {
    return "email must be valid";  // Server validates email format
  }
  return null;
}
```

**Client (BUGGY):**
```typescript
// src/client/api-client.ts — No email format validation
async createUser(req: CreateUserRequest): Promise<User> {
  try {
    // BUG: No validation that req.email matches email format
    // Client sends invalid email, waits for server 400 error instead of failing fast

    const response = await fetch(`${this.baseUrl}/api/users`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${this.authToken}`,
      },
      body: JSON.stringify(req),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    return await response.json() as User;
  } catch (err) {
    throw new Error(`Failed to create user: ${(err as Error).message}`);
  }
}
```

**Why Stricture catches this:** Server has `isValidEmail()` helper that validates email format. Client has no equivalent check. Per `CTR-strictness-parity`, client should validate email format before sending (fail-fast). Missing client validation = violation.

---

### B11 — Precision Loss (CTR-strictness-parity)

**Bug:** Shared type uses Date object but JSON sends string, causing serialization mismatch
**Expected violation:** `CTR-strictness-parity`

**Server:**
```typescript
// src/server/routes/users.ts (returns ISO string)
const now = new Date().toISOString();
const user: User = {
  id: uuidv4(),
  name: body.name,
  email: body.email,
  role: body.role,
  avatar: null,
  createdAt: now,  // Server sends ISO string "2026-01-01T00:00:00.000Z"
  updatedAt: now,
};

res.status(201).json(user);
```

**Client (BUGGY):**
```typescript
// src/shared/types.ts — Client type expects Date object
export interface User {
  id: string;
  name: string;
  email: string;
  role: UserRole;
  avatar: string | null;
  createdAt: Date;  // BUG: Type says Date, but JSON sends string
  updatedAt: Date;  // BUG: Type says Date, but JSON sends string
}

// src/client/api-client.ts
async getUser(id: string): Promise<User> {
  const response = await fetch(`${this.baseUrl}/api/users/${id}`);
  return await response.json() as User;
  // Runtime: createdAt is string "2026-01-01T00:00:00.000Z", not Date object
  // Calling user.createdAt.getTime() will crash with "getTime is not a function"
}
```

**Why Stricture catches this:** Server returns `createdAt: string` in JSON. Shared type declares `createdAt: Date`. JSON can't transport Date objects (only strings). Type mismatch on serialization = `CTR-strictness-parity` violation.

---

### B12 — Nullable Field Crashes (CTR-response-shape)

**Bug:** Client destructures optional middleName without null check
**Expected violation:** `CTR-response-shape`

**Server:**
```typescript
// src/shared/types.ts (avatar is nullable)
export interface User {
  id: string;
  name: string;
  email: string;
  role: UserRole;
  avatar: string | null;  // Server declares nullable
  createdAt: string;
  updatedAt: string;
}

// src/server/routes/users.ts
const user: User = {
  id: uuidv4(),
  name: body.name,
  email: body.email,
  role: body.role,
  avatar: null,  // Server returns null for avatar
  createdAt: now,
  updatedAt: now,
};
```

**Client (BUGGY):**
```typescript
// src/client/api-client.ts — Destructures avatar without null check
async getUser(id: string): Promise<User> {
  try {
    const response = await fetch(`${this.baseUrl}/api/users/${id}`, {
      headers: {
        Authorization: `Bearer ${this.authToken}`,
      },
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    const user = await response.json() as User;

    // BUG: Assumes avatar is always a string, crashes when null
    const avatarUrl = user.avatar.toUpperCase();  // TypeError: Cannot read property 'toUpperCase' of null

    return user;
  } catch (err) {
    throw new Error(`Failed to get user: ${(err as Error).message}`);
  }
}
```

**Why Stricture catches this:** Shared type declares `avatar: string | null`. Server returns `null`. Client code dereferences `user.avatar` without null check. Stricture detects nullable field access without guard = `CTR-response-shape` violation (client doesn't handle full response shape).

---

### B13 — Missing Auth (CTR-request-shape)

**Bug:** Client doesn't send Authorization header
**Expected violation:** `CTR-request-shape`

**Server:**
```typescript
// src/server/middleware/auth.ts (requires Authorization header)
export function requireAuth(req: Request, res: Response, next: NextFunction) {
  const authHeader = req.headers.authorization;

  if (!authHeader || !authHeader.startsWith("Bearer ")) {
    res.status(401).json({ error: "missing or invalid authorization header" });
    return;
  }

  const token = authHeader.substring(7);
  if (!isValidToken(token)) {
    res.status(401).json({ error: "invalid token" });
    return;
  }

  next();
}

// src/server/routes/users.ts (all routes protected)
router.post("/users", requireAuth, (req: Request, res: Response) => {
  // Handler requires Authorization header
});
```

**Client (BUGGY):**
```typescript
// src/client/api-client.ts — Missing Authorization header
async createUser(req: CreateUserRequest): Promise<User> {
  try {
    const response = await fetch(`${this.baseUrl}/api/users`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        // BUG: Missing Authorization header
        // Server will return 401 Unauthorized
      },
      body: JSON.stringify(req),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    return await response.json() as User;
  } catch (err) {
    throw new Error(`Failed to create user: ${(err as Error).message}`);
  }
}
```

**Why Stricture catches this:** Server middleware requires `Authorization` header on all endpoints. Manifest should declare `required_headers: ["Authorization"]`. Client omits header. Request shape incomplete = `CTR-request-shape` violation.

---

### B14 — Pagination Terminated Early (CTR-response-shape)

**Bug:** Client ignores totalPages/hasMore, only fetches page 1
**Expected violation:** `CTR-response-shape`

**Server:**
```typescript
// src/server/routes/users.ts (returns hasMore cursor pagination)
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
    hasMore,  // Server signals if more pages exist
  };

  res.status(200).json(response);
});
```

**Client (BUGGY):**
```typescript
// src/client/api-client.ts — Ignores hasMore flag
async getAllUsers(): Promise<User[]> {
  try {
    const firstPage = await this.listUsers(undefined, 10);

    // BUG: Only returns first page, ignores hasMore flag
    // If hasMore=true, client should loop and fetch next pages using cursor
    return firstPage.data;

    // Correct implementation:
    // const allUsers = [...firstPage.data];
    // let cursor = firstPage.cursor;
    // while (firstPage.hasMore && cursor) {
    //   const nextPage = await this.listUsers(cursor, 10);
    //   allUsers.push(...nextPage.data);
    //   cursor = nextPage.cursor;
    // }
    // return allUsers;
  } catch (err) {
    throw new Error(`Failed to get all users: ${(err as Error).message}`);
  }
}
```

**Why Stricture catches this:** Server returns `PaginatedResponse<User>` with `hasMore` and `cursor` fields. Client has `getAllUsers()` method that ignores `hasMore` flag. Response shape incomplete handling = `CTR-response-shape` violation (client doesn't use all response fields).

---

### B15 — Race Condition (CTR-request-shape)

**Bug:** PUT update with no If-Match/version check, allows concurrent update overwrites
**Expected violation:** `CTR-request-shape`

**Server:**
```typescript
// src/server/routes/users.ts (validates If-Match header for optimistic locking)
router.patch("/users/:id", (req: Request, res: Response) => {
  const { id } = req.params;
  const body = req.body as UpdateUserRequest;
  const ifMatch = req.headers["if-match"];

  const user = users.get(id);
  if (!user) {
    res.status(404).json({ error: "user not found" });
    return;
  }

  // Server requires If-Match header to prevent race conditions
  if (!ifMatch) {
    res.status(428).json({ error: "If-Match header required for updates" });
    return;
  }

  // Check version matches (using updatedAt as ETag)
  const currentETag = `"${user.updatedAt}"`;
  if (ifMatch !== currentETag) {
    res.status(412).json({ error: "precondition failed - resource has been modified" });
    return;
  }

  // Apply updates
  if (body.name !== undefined) {
    user.name = body.name;
  }
  if (body.email !== undefined) {
    user.email = body.email;
  }
  if (body.role !== undefined) {
    user.role = body.role;
  }

  user.updatedAt = new Date().toISOString();
  users.set(id, user);
  res.status(200).json(user);
});
```

**Client (BUGGY):**
```typescript
// src/client/api-client.ts — No If-Match header, race condition possible
async updateUser(id: string, req: UpdateUserRequest): Promise<User> {
  try {
    const response = await fetch(`${this.baseUrl}/api/users/${id}`, {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${this.authToken}`,
        // BUG: Missing If-Match header
        // Server returns 428 Precondition Required
        // Concurrent updates can overwrite each other (lost update problem)
      },
      body: JSON.stringify(req),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    return await response.json() as User;
  } catch (err) {
    throw new Error(`Failed to update user: ${(err as Error).message}`);
  }
}

// Correct implementation:
// 1. GET user to retrieve current updatedAt (ETag)
// 2. PATCH with If-Match: "<updatedAt>" header
// 3. If 412 Precondition Failed, retry from step 1
```

**Why Stricture catches this:** Server requires `If-Match` header for PATCH requests (returns 428 if missing). Manifest should declare `required_headers: ["If-Match"]` for UPDATE operations. Client omits header. Request shape incomplete = `CTR-request-shape` violation. Also violates concurrency safety (no test for 412 Precondition Failed = `TQ-negative-cases`).

