// src/shared/types.ts (server defines 3 roles)
export type UserRole = "admin" | "editor" | "viewer";

// src/server/routes/users.ts (server validates all 3 roles)
function validateCreateUserRequest(req: CreateUserRequest): string | null {
  if (!req.role || !["admin", "editor", "viewer"].includes(req.role)) {
    return "role must be admin, editor, or viewer";
  }
  return null;
}
