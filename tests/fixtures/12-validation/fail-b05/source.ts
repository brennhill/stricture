// B05: Omits required `email` field from sign-up request.

class SupabaseClientB05 {
  private readonly baseUrl: string;
  private readonly anonKey: string;

  constructor(baseUrl: string, anonKey: string) {
    this.baseUrl = baseUrl;
    this.anonKey = anonKey;
  }

  async signUp(password: string): Promise<Record<string, unknown>> {
    // BUG: `email` is required by the manifest but is never sent.
    // GoTrue will return a 400 or 422: "Signup requires a valid email"
    const body = { password };
    // email is completely omitted from the request body

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/auth/v1/signup`, {
        method: "POST",
        headers: {
          apikey: this.anonKey,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(body),
      });
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const error = await response.json();
      throw new Error(`Sign-up failed: ${JSON.stringify(error)}`);
    }

    return (await response.json()) as Record<string, unknown>;
  }

  async insertUser(role: string): Promise<Record<string, unknown>> {
    // BUG: `email` is required by the manifest (NOT NULL column) but omitted.
    // PostgREST returns: { code: "23502", message: "null value in column \"email\"..." }

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/rest/v1/users`, {
        method: "POST",
        headers: {
          apikey: this.anonKey,
          Authorization: `Bearer fake-jwt`,
          "Content-Type": "application/json",
          Prefer: "return=representation",
        },
        body: JSON.stringify({ user_role: role }),
      });
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const error = await response.json();
      throw new Error(`Insert failed: ${JSON.stringify(error)}`);
    }

    const rows = (await response.json()) as Record<string, unknown>[];
    return rows[0];
  }
}
