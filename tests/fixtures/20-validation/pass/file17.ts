// src/server/routes/users.ts (returns string ID)
const user: User = {
  id: uuidv4(),  // Returns UUID string like "a1b2c3d4-..."
  name: body.name,
  email: body.email,
  role: body.role,
  avatar: null,
  createdAt: now,
  updatedAt: now,
};

res.status(201).json(user);
