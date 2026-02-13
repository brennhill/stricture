// B02: Ignores HTTP status codes; treats all responses as success.

class SupabaseClientB02 {
  private readonly baseUrl: string;
  private readonly anonKey: string;
  private readonly jwt: string;

  constructor(baseUrl: string, anonKey: string, jwt: string) {
    this.baseUrl = baseUrl;
    this.anonKey = anonKey;
    this.jwt = jwt;
  }

  async queryUsers(): Promise<Record<string, unknown>[]> {
    try {
      const response = await fetch(`${this.baseUrl}/rest/v1/users?select=*`, {
        headers: {
          apikey: this.anonKey,
          Authorization: `Bearer ${this.jwt}`,
        },
      });

      // BUG: No check on response.ok or response.status.
      // A 401 (expired JWT), 403 (RLS violation), or 416 (bad range)
      // response is parsed as if it were a user array.
      // Supabase returns { message, details, hint, code } on errors,
      // which gets cast to Record<string, unknown>[] -- a single-element
      // "array" that is actually an error object.
      return (await response.json()) as Record<string, unknown>[];
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }
  }

  async insertUser(email: string, role: string): Promise<Record<string, unknown>> {
    try {
      const response = await fetch(`${this.baseUrl}/rest/v1/users`, {
        method: "POST",
        headers: {
          apikey: this.anonKey,
          Authorization: `Bearer ${this.jwt}`,
          "Content-Type": "application/json",
          Prefer: "return=representation",
        },
        body: JSON.stringify({ email, user_role: role }),
      });

      // BUG: Does not check for 201 vs 409 (duplicate) vs 400.
      // On 409, the response body is an error object, not a user row array.
      const rows = (await response.json()) as Record<string, unknown>[];
      return rows[0];
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }
  }
}
