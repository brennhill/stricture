// src/server/routes/users.ts (no changes, server returns 404 correctly)
router.get("/users/:id", (req: Request, res: Response) => {
  const { id } = req.params;
  const user = users.get(id);

  if (!user) {
    res.status(404).json({ error: "user not found" });
    return;
  }

  res.status(200).json(user);
});
