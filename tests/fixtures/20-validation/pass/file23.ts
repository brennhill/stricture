// src/server/routes/users.ts (validates email format)
function isValidEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

function validateCreateUserRequest(req: CreateUserRequest): string | null {
  if (!req.email || !isValidEmail(req.email)) {
    return "email must be valid";  // Server validates email format
  }
  return null;
}
