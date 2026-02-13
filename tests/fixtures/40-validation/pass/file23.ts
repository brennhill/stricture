// src/validators/email-validator.ts
export function validateEmail(email: string): boolean {
  if (email.length === 0) {
    return false;
  }
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}
