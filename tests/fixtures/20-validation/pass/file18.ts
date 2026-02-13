// src/shared/types.ts — Client type expects number ID
export interface User {
  id: number;  // BUG: Server sends string, client expects number
  name: string;
  email: string;
  role: UserRole;
  avatar: string | null;
  createdAt: string;
  updatedAt: string;
}
