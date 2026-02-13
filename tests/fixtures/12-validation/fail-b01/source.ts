// B01: No try/catch on any Supabase API call.

class SupabaseClientB01 {
  private readonly baseUrl: string;
  private readonly anonKey: string;
  private readonly jwt: string;

  constructor(baseUrl: string, anonKey: string, jwt: string) {
    this.baseUrl = baseUrl;
    this.anonKey = anonKey;
    this.jwt = jwt;
  }

  async queryUsers(): Promise<Record<string, unknown>[]> {
    // BUG: No try/catch. If fetch() throws (DNS failure, timeout, network
    // disconnect), the error propagates as an unhandled rejection.
    const response = await fetch(`${this.baseUrl}/rest/v1/users?select=*`, {
      headers: {
        apikey: this.anonKey,
        Authorization: `Bearer ${this.jwt}`,
      },
    });
    return (await response.json()) as Record<string, unknown>[];
  }

  async signUp(email: string, password: string): Promise<Record<string, unknown>> {
    // BUG: Same issue -- no error handling at all.
    const response = await fetch(`${this.baseUrl}/auth/v1/signup`, {
      method: "POST",
      headers: {
        apikey: this.anonKey,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ email, password }),
    });
    return (await response.json()) as Record<string, unknown>;
  }
}
