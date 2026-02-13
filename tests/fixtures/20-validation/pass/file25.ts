// src/server/routes/users.ts (returns ISO string)
const now = new Date().toISOString();
const user: User = {
  id: uuidv4(),
  name: body.name,
  email: body.email,
  role: body.role,
  avatar: null,
  createdAt: now,  // Server sends ISO string "2026-01-01T00:00:00.000Z"
  updatedAt: now,
};

res.status(201).json(user);
