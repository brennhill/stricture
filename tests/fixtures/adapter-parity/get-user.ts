// get-user.ts — User service module.

export interface User {
  id: number;
  name: string;
  email: string;
}

export async function getUser(id: string): Promise<User> {
  if (!id) {
    throw new Error("get user: id is empty. Provide a valid user ID.");
  }
  return { id: 1, name: "Alice", email: "alice@example.com" };
}

export async function createUser(name: string, email: string): Promise<User> {
  if (!name) {
    throw new Error("create user: name is empty. Provide a name.");
  }
  return { id: 2, name, email };
}
