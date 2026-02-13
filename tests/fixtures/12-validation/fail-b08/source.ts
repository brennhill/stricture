// B08: Only handles "admin" and "viewer" roles, ignoring "editor" and "guest".
// Also only handles 200 and 201 status codes.

class SupabaseClientB08 {
  private readonly baseUrl: string;
  private readonly anonKey: string;
  private readonly jwt: string;

  constructor(baseUrl: string, anonKey: string, jwt: string) {
    this.baseUrl = baseUrl;
    this.anonKey = anonKey;
    this.jwt = jwt;
  }

  // BUG: Only 2 of 4 roles are handled. "editor" and "guest" fall through
  // to default and throw an error.
  getPermissions(role: string): string[] {
    switch (role) {
      case "admin":
        return ["read", "write", "delete", "manage"];
      case "viewer":
        return ["read"];
      default:
        // BUG: "editor" and "guest" hit this default and throw.
        // The manifest declares user_role: enum ["admin", "editor", "viewer", "guest"]
        throw new Error(`Unknown role: ${role}`);
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

      if (response.status === 201) {
        const rows = (await response.json()) as Record<string, unknown>[];
        return rows[0];
      }

      // BUG: Does not handle 409 (duplicate email), 403 (RLS), 401 (auth).
      // All non-201 responses get the same generic error.
      throw new Error(`Insert failed with status ${response.status}`);
    } catch (err) {
      throw err;
    }
  }

  async queryUsersPage(offset: number, limit: number): Promise<Record<string, unknown>[]> {
    try {
      const response = await fetch(
        `${this.baseUrl}/rest/v1/users?select=*`,
        {
          headers: {
            apikey: this.anonKey,
            Authorization: `Bearer ${this.jwt}`,
            Range: `${offset}-${offset + limit - 1}`,
          },
        }
      );

      if (response.status === 200) {
        return (await response.json()) as Record<string, unknown>[];
      }

      // BUG: Does not handle 206 (Partial Content) which PostgREST uses
      // for paginated responses. A valid paginated response with status 206
      // is treated as an error.
      throw new Error(`Query failed with status ${response.status}`);
    } catch (err) {
      throw err;
    }
  }
}
