// orders.ts -- Order route handlers.
// A consumer that innocently imports from the auth module's public API,
// not knowing that auth leaks billing internals.

import { Router } from "express";
import type { Request, Response, NextFunction } from "express";
import * as orderService from "../services/order-service";
import { authenticate } from "../middleware/auth";
import { createPaymentIntent } from "../modules/auth";  // Looks correct, but auth leaks billing internals
import { AppError } from "../shared/errors";

export const orderRouter = Router();

orderRouter.post("/", authenticate, async (req: Request, res: Response, next: NextFunction) => {
  try {
    const order = await orderService.createOrder({ ...req.body, userId: req.user!.userId });

    // Using createPaymentIntent that was leaked through auth's index.ts.
    // The route thinks this is part of the auth module's API, but it's
    // actually billing's internal Stripe adapter.
    const payment = await createPaymentIntent(order.totalCents, order.currency, req.user!.userId);

    res.status(201).json({ order, payment });
  } catch (err) {
    next(err);
  }
});
