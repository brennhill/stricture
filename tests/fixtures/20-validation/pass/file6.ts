// src/client/api-client.ts — Missing try/catch
async createUser(req: CreateUserRequest): Promise<User> {
  const response = await fetch(`${this.baseUrl}/api/users`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${this.authToken}`,
    },
    body: JSON.stringify(req),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || `HTTP ${response.status}`);
  }

  return await response.json() as User;
  // No try/catch — network errors will crash caller
}
