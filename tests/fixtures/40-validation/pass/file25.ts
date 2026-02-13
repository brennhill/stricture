// src/validators/age-validator.ts
export function validateAge(age: number): boolean {
  if (age < 0 || age > 150) {
    throw new Error('Age must be between 0 and 150');
  }
  return true;
}
