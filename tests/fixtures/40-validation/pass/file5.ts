// src/auth/login.ts
export async function login(email: string, password: string): Promise<boolean> {
  const response = await authService.authenticate({ email, password });
  return response.success;
}
