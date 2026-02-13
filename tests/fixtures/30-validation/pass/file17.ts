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
