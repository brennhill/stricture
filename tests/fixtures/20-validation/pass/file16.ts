// src/shared/types.ts — Client type expects "profilePicture" instead of "avatar"
export interface User {
  id: string;
  name: string;
  email: string;
  role: UserRole;
  profilePicture: string | null;  // BUG: Field name mismatch (avatar → profilePicture)
  createdAt: string;
  updatedAt: string;
}
