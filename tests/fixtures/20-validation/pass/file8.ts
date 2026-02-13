// src/client/api-client.ts — Missing response.ok check
async getUser(id: string): Promise<User> {
  try {
    const response = await fetch(`${this.baseUrl}/api/users/${id}`, {
      headers: {
        Authorization: `Bearer ${this.authToken}`,
      },
    });

    // BUG: No check for response.ok
    // Will try to parse 404 error JSON as User type
    return await response.json() as User;
  } catch (err) {
    throw new Error(`Failed to get user: ${(err as Error).message}`);
  }
}
