// src/server/routes/users.ts (validates limit bounds)
router.get("/users", (req: Request, res: Response) => {
  const cursor = req.query.cursor as string | undefined;
  const limit = parseInt(req.query.limit as string || "10", 10);

  // Server validates limit range
  if (limit < 1 || limit > 100) {
    res.status(400).json({ error: "limit must be between 1 and 100" });
    return;
  }

  const allUsers = Array.from(users.values()).sort((a, b) =>
    a.createdAt.localeCompare(b.createdAt)
  );

  let startIndex = 0;
  if (cursor) {
    startIndex = allUsers.findIndex(u => u.id === cursor) + 1;
  }

  const pageUsers = allUsers.slice(startIndex, startIndex + limit);
  // ... rest of handler
});
