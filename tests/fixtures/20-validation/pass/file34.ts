// src/client/api-client.ts — No If-Match header, race condition possible
async updateUser(id: string, req: UpdateUserRequest): Promise<User> {
  try {
    const response = await fetch(`${this.baseUrl}/api/users/${id}`, {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${this.authToken}`,
        // BUG: Missing If-Match header
        // Server returns 428 Precondition Required
        // Concurrent updates can overwrite each other (lost update problem)
      },
      body: JSON.stringify(req),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    return await response.json() as User;
  } catch (err) {
    throw new Error(`Failed to update user: ${(err as Error).message}`);
  }
}

// Correct implementation:
// 1. GET user to retrieve current updatedAt (ETag)
// 2. PATCH with If-Match: "<updatedAt>" header
// 3. If 412 Precondition Failed, retry from step 1
