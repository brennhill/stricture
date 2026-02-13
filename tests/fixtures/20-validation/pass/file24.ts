// src/client/api-client.ts — No email format validation
async createUser(req: CreateUserRequest): Promise<User> {
  try {
    // BUG: No validation that req.email matches email format
    // Client sends invalid email, waits for server 400 error instead of failing fast

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
  } catch (err) {
    throw new Error(`Failed to create user: ${(err as Error).message}`);
  }
}
