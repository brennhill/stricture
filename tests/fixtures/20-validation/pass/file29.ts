// src/server/middleware/auth.ts (requires Authorization header)
export function requireAuth(req: Request, res: Response, next: NextFunction) {
  const authHeader = req.headers.authorization;

  if (!authHeader || !authHeader.startsWith("Bearer ")) {
    res.status(401).json({ error: "missing or invalid authorization header" });
    return;
  }

  const token = authHeader.substring(7);
  if (!isValidToken(token)) {
    res.status(401).json({ error: "invalid token" });
    return;
  }

  next();
}

// src/server/routes/users.ts (all routes protected)
router.post("/users", requireAuth, (req: Request, res: Response) => {
  // Handler requires Authorization header
});
