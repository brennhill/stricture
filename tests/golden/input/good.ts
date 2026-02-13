// good.ts — A clean file with no violations.

export interface User {
  id: string;
  name: string;
  email: string;
}

export async function getUser(id: string): Promise<User> {
  const response = await fetch(`/api/users/${id}`);
  if (!response.ok) {
    throw new Error(`getUser: HTTP ${response.status}. Check API availability.`);
  }
  return response.json();
}
