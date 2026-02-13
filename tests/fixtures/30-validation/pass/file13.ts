// auth.ts -- Authentication middleware.
// V11: Uses a dynamic import to reach into auth module internals.

import type { Request, Response, NextFunction } from "express";
import { UnauthorizedError, ForbiddenError } from "../shared/errors";
import { logger } from "../shared/logger";
import type { UserRole } from "../models/user";

export async function authenticate(req: Request, _res: Response, next: NextFunction): Promise<void> {
  const authHeader = req.headers.authorization;
  if (!authHeader || !authHeader.startsWith("Bearer ")) {
    next(new UnauthorizedError("Missing or invalid Authorization header"));
    return;
  }

  const token = authHeader.slice(7);

  try {
    // VIOLATION: Dynamic import reaching into module internals.
    // Instead of importing from "../modules/auth" (the public entry point),
    // this reaches directly into the internal jwt-utils file.
    // Dynamic imports are still imports -- they create a dependency on
    // the internal file structure of the auth module.
    const { decodePayload } = await import("../modules/auth/internal/jwt-utils");  // <-- VIOLATION

    // Even worse: jwt-utils is not even exported through the module's
    // public API. The auth module exports validateToken and generateToken,
    // but not the low-level decodePayload function. This creates a
    // dependency on an implementation detail.
    const payload = decodePayload(token);

    req.user = { userId: payload.userId, role: payload.role };
    logger.debug("Authenticated request", { userId: payload.userId });
    next();
  } catch (err) {
    next(err);
  }
}

export function requireRole(...roles: UserRole[]) {
  return (req: Request, _res: Response, next: NextFunction): void => {
    if (!req.user) {
      next(new UnauthorizedError());
      return;
    }

    if (!roles.includes(req.user.role as UserRole)) {
      next(new ForbiddenError(`Requires one of: ${roles.join(", ")}`));
      return;
    }

    next();
  };
}
