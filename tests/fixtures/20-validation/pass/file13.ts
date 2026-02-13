// src/server/routes/users.ts (validates required fields)
function validateCreateUserRequest(req: CreateUserRequest): string | null {
  if (!req.name || req.name.length === 0 || req.name.length > 255) {
    return "name must be between 1 and 255 characters";
  }
  if (!req.email || !isValidEmail(req.email)) {
    return "email must be valid";  // Server REQUIRES email
  }
  if (!req.role || !["admin", "editor", "viewer"].includes(req.role)) {
    return "role must be admin, editor, or viewer";
  }
  return null;
}
