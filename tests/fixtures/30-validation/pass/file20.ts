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
