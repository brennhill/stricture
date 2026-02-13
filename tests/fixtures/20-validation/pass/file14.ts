// src/client/api-client.ts — Omits email field
async createUser(req: CreateUserRequest): Promise<User> {
  try {
    const response = await fetch(`${this.baseUrl}/api/users`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${this.authToken}`,
      },
      body: JSON.stringify({
        name: req.name,
        role: req.role,
        // BUG: Missing email field in request body
      }),
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
