// users.ts -- User route handlers.
// V10: Route layer imports from model layer and performs business logic
// (filtering, transformation) that belongs in the service layer.

import { Router } from "express";
import type { Request, Response, NextFunction } from "express";
import * as userService from "../services/user-service";
import { authenticate } from "../middleware/auth";
import { AppError } from "../shared/errors";
import { logger } from "../shared/logger";
import type { User, UserRole } from "../models/user";  // <-- VIOLATION: importing from model

export const userRouter = Router();

// This endpoint performs business logic directly in the route handler
// by importing and using the User type for data manipulation that
// should live in the service layer.
userRouter.get("/", authenticate, async (req: Request, res: Response, next: NextFunction) => {
  try {
    const result = await userService.listUsers({}, 1, 1000);

    // BUG: The route performs filtering and transformation logic that
    // belongs in the service layer. It imports the User type and UserRole
    // enum to do complex business logic inline. This violates the
    // layer-violation rule which forbids routes from accessing models.
    const filtered: User[] = result.data.filter((user: User) => {
      if (req.query["minAge"]) {
        const minAge = parseInt(req.query["minAge"] as string, 10);
        const ageMs = Date.now() - user.createdAt.getTime();
        const ageDays = ageMs / (1000 * 60 * 60 * 24);
        if (ageDays < minAge) return false;
      }

      if (req.query["role"]) {
        const allowedRoles: UserRole[] = (req.query["role"] as string).split(",") as UserRole[];
        if (!allowedRoles.includes(user.role)) return false;
      }

      return true;
    });

    // Also performing aggregation logic that belongs in the service
    const roleDistribution: Record<UserRole, number> = {
      admin: 0,
      member: 0,
      viewer: 0,
    };
    for (const user of filtered) {
      roleDistribution[user.role]++;
    }

    res.json({
      users: filtered,
      meta: {
        total: filtered.length,
        roleDistribution,
      },
    });
  } catch (err) {
    next(err);
  }
});
