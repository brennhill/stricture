// src/client/api-client.ts — Missing Authorization header
async createUser(req: CreateUserRequest): Promise<User> {
  try {
    const response = await fetch(`${this.baseUrl}/api/users`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        // BUG: Missing Authorization header
        // Server will return 401 Unauthorized
      },
      body: JSON.stringify(req),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    return await response.json() as User;
  } catch (err) {
    throw new Error(`Failed to create user: ${(err as Error).message}`);
  }
}
