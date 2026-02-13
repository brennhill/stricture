// src/shared/types.ts (avatar is nullable)
export interface User {
  id: string;
  name: string;
  email: string;
  role: UserRole;
  avatar: string | null;  // Server declares nullable
  createdAt: string;
  updatedAt: string;
}

// src/server/routes/users.ts
const user: User = {
  id: uuidv4(),
  name: body.name,
  email: body.email,
  role: body.role,
  avatar: null,  // Server returns null for avatar
  createdAt: now,
  updatedAt: now,
};
