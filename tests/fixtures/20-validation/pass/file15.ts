// src/server/routes/users.ts (returns all User fields including avatar)
const user: User = {
  id: uuidv4(),
  name: body.name,
  email: body.email,
  role: body.role,
  avatar: null,  // Server DOES return avatar
  createdAt: now,
  updatedAt: now,
};

users.set(user.id, user);
res.status(201).json(user);
