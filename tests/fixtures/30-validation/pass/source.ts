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
