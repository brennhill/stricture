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
