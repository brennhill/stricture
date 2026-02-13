// src/client/api-client.ts — No limit validation
async listUsers(cursor?: string, limit = 10): Promise<PaginatedResponse<User>> {
  try {
    const params = new URLSearchParams();
    if (cursor) params.set("cursor", cursor);
    params.set("limit", limit.toString());  // BUG: No check for limit < 1 or limit > 100

    const response = await fetch(`${this.baseUrl}/api/users?${params}`, {
      headers: {
        Authorization: `Bearer ${this.authToken}`,
      },
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    return await response.json() as PaginatedResponse<User>;
  } catch (err) {
    throw new Error(`Failed to list users: ${(err as Error).message}`);
  }
}
