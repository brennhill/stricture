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
