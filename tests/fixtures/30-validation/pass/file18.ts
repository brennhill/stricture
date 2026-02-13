// index.ts -- Route aggregator.

import { Router } from "express";
import { userRouter } from "./users";
import { orderRouter } from "./orders";

export const apiRouter = Router();

apiRouter.use("/users", userRouter);
apiRouter.use("/orders", orderRouter);

apiRouter.get("/health", (_req, res) => {
  res.json({ status: "ok", timestamp: new Date().toISOString() });
});
