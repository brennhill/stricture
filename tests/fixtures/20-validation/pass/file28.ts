// src/client/api-client.ts — Destructures avatar without null check
async getUser(id: string): Promise<User> {
  try {
    const response = await fetch(`${this.baseUrl}/api/users/${id}`, {
      headers: {
        Authorization: `Bearer ${this.authToken}`,
      },
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    const user = await response.json() as User;

    // BUG: Assumes avatar is always a string, crashes when null
    const avatarUrl = user.avatar.toUpperCase();  // TypeError: Cannot read property 'toUpperCase' of null

    return user;
  } catch (err) {
    throw new Error(`Failed to get user: ${(err as Error).message}`);
  }
}
