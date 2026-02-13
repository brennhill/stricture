// src/server/routes/users.ts (validates If-Match header for optimistic locking)
router.patch("/users/:id", (req: Request, res: Response) => {
  const { id } = req.params;
  const body = req.body as UpdateUserRequest;
  const ifMatch = req.headers["if-match"];

  const user = users.get(id);
  if (!user) {
    res.status(404).json({ error: "user not found" });
    return;
  }

  // Server requires If-Match header to prevent race conditions
  if (!ifMatch) {
    res.status(428).json({ error: "If-Match header required for updates" });
    return;
  }

  // Check version matches (using updatedAt as ETag)
  const currentETag = `"${user.updatedAt}"`;
  if (ifMatch !== currentETag) {
    res.status(412).json({ error: "precondition failed - resource has been modified" });
    return;
  }

  // Apply updates
  if (body.name !== undefined) {
    user.name = body.name;
  }
  if (body.email !== undefined) {
    user.email = body.email;
  }
  if (body.role !== undefined) {
    user.role = body.role;
  }

  user.updatedAt = new Date().toISOString();
  users.set(id, user);
  res.status(200).json(user);
});
