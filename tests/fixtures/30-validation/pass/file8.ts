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
