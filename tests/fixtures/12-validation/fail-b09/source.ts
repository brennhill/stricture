// B09: No validation on Range header bounds or page size.

class SupabaseClientB09 {
  private readonly baseUrl: string;
  private readonly anonKey: string;
  private readonly jwt: string;

  constructor(baseUrl: string, anonKey: string, jwt: string) {
    this.baseUrl = baseUrl;
    this.anonKey = anonKey;
    this.jwt = jwt;
  }

  // BUG: No validation that pageSize does not exceed PostgREST max-rows
  // (default 1000). Also no validation on offset being non-negative.
  async queryUsers(pageSize: number = 10000): Promise<Record<string, unknown>[]> {
    try {
      const response = await fetch(
        `${this.baseUrl}/rest/v1/users?select=*`,
        {
          headers: {
            apikey: this.anonKey,
            Authorization: `Bearer ${this.jwt}`,
            Range: `0-${pageSize - 1}`, // BUG: range(0, 9999) -- way beyond server max
          },
        }
      );

      if (!response.ok) throw new Error(`Query failed: ${response.status}`);
      return (await response.json()) as Record<string, unknown>[];
    } catch (err) {
      throw err;
    }
  }

  // BUG: Negative offset not validated. range(-1, 10) is nonsensical.
  async queryUsersOffset(offset: number, limit: number): Promise<Record<string, unknown>[]> {
    try {
      const response = await fetch(
        `${this.baseUrl}/rest/v1/users?select=*`,
        {
          headers: {
            apikey: this.anonKey,
            Authorization: `Bearer ${this.jwt}`,
            Range: `${offset}-${offset + limit - 1}`, // No bounds check
          },
        }
      );

      if (!response.ok) throw new Error(`Query failed: ${response.status}`);
      return (await response.json()) as Record<string, unknown>[];
    } catch (err) {
      throw err;
    }
  }

  // BUG: Uses Number.MAX_SAFE_INTEGER as page size.
  async getAllUsers(): Promise<Record<string, unknown>[]> {
    return this.queryUsers(Number.MAX_SAFE_INTEGER);
  }
}
