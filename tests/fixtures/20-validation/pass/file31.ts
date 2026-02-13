// src/server/routes/users.ts (returns hasMore cursor pagination)
router.get("/users", (req: Request, res: Response) => {
  const cursor = req.query.cursor as string | undefined;
  const limit = parseInt(req.query.limit as string || "10", 10);

  const allUsers = Array.from(users.values()).sort((a, b) =>
    a.createdAt.localeCompare(b.createdAt)
  );

  let startIndex = 0;
  if (cursor) {
    startIndex = allUsers.findIndex(u => u.id === cursor) + 1;
  }

  const pageUsers = allUsers.slice(startIndex, startIndex + limit);
  const hasMore = startIndex + limit < allUsers.length;
  const nextCursor = hasMore ? pageUsers[pageUsers.length - 1].id : null;

  const response: PaginatedResponse<User> = {
    data: pageUsers,
    cursor: nextCursor,
    hasMore,  // Server signals if more pages exist
  };

  res.status(200).json(response);
});
