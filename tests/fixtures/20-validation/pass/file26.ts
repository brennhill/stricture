// src/shared/types.ts — Client type expects Date object
export interface User {
  id: string;
  name: string;
  email: string;
  role: UserRole;
  avatar: string | null;
  createdAt: Date;  // BUG: Type says Date, but JSON sends string
  updatedAt: Date;  // BUG: Type says Date, but JSON sends string
}

// src/client/api-client.ts
async getUser(id: string): Promise<User> {
  const response = await fetch(`${this.baseUrl}/api/users/${id}`);
  return await response.json() as User;
  // Runtime: createdAt is string "2026-01-01T00:00:00.000Z", not Date object
  // Calling user.createdAt.getTime() will crash with "getTime is not a function"
}
