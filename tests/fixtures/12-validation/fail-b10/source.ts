// B10: UUID values accepted as any string. JWT format not checked.

class SupabaseClientB10 {
  private readonly baseUrl: string;
  private readonly anonKey: string;
  private readonly jwt: string;

  constructor(baseUrl: string, anonKey: string, jwt: string) {
    // BUG: No validation that anonKey or jwt are in the expected format.
    // anonKey should start with "eyJ" (base64 JSON header).
    // jwt should be a valid JWT (three dot-separated base64 segments).
    this.baseUrl = baseUrl;
    this.anonKey = anonKey;
    this.jwt = jwt;
  }

  // BUG: No UUID format validation. Accepts any string.
  async getUser(userId: string): Promise<Record<string, unknown> | null> {
    try {
      const response = await fetch(
        `${this.baseUrl}/rest/v1/users?id=eq.${userId}&select=*`,
        {
          headers: {
            apikey: this.anonKey,
            Authorization: `Bearer ${this.jwt}`,
          },
        }
      );

      if (!response.ok) throw new Error(`Query failed: ${response.status}`);

      // BUG: "hello" becomes ?id=eq.hello -- PostgreSQL uuid cast fails with
      // error code 22P02: "invalid input syntax for type uuid: hello"
      const rows = (await response.json()) as Record<string, unknown>[];
      return rows[0] ?? null;
    } catch (err) {
      throw err;
    }
  }

  // BUG: No email format validation on sign-up input.
  async signUp(email: string, password: string): Promise<Record<string, unknown>> {
    try {
      const response = await fetch(`${this.baseUrl}/auth/v1/signup`, {
        method: "POST",
        headers: {
          apikey: this.anonKey,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ email, password }),
      });

      if (!response.ok) throw new Error(`Sign-up failed: ${response.status}`);
      return (await response.json()) as Record<string, unknown>;
    } catch (err) {
      throw err;
    }
  }
}
