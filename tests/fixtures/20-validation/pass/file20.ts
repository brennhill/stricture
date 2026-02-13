// src/shared/types.ts — Client only handles 2 of 3 roles
export type UserRole = "admin" | "editor";  // BUG: Missing "viewer" from enum

// src/client/api-client.ts
async createUser(req: CreateUserRequest): Promise<User> {
  // Client can only create admin/editor, but server accepts viewer
  // Runtime: client receives "viewer" role in response, TypeScript type doesn't allow it
}
