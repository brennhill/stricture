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
