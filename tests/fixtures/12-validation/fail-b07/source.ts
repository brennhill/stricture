// B07: count from Content-Range parsed as string, balance as number, UUID as number.

interface UserB07 {
  id: number;            // BUG: UUID is a string, not a number
  email: string;
  user_role: string;
  profile: { avatar_url: string; bio: string } | null;  // BUG: inner fields not nullable
  balance: number;       // BUG: PostgreSQL bigint can exceed Number.MAX_SAFE_INTEGER
  metadata: Record<string, unknown> | null;
  created_at: string;
  updated_at: string;
}

class SupabaseClientB07 {
  private readonly baseUrl: string;
  private readonly anonKey: string;
  private readonly jwt: string;

  constructor(baseUrl: string, anonKey: string, jwt: string) {
    this.baseUrl = baseUrl;
    this.anonKey = anonKey;
    this.jwt = jwt;
  }

  async queryUsers(): Promise<{ users: UserB07[]; count: string }> {
    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/rest/v1/users?select=*`, {
        headers: {
          apikey: this.anonKey,
          Authorization: `Bearer ${this.jwt}`,
          Range: "0-24",
          Prefer: "count=exact",
        },
      });
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) {
      throw new Error(`Query failed: ${response.status}`);
    }

    const rows = (await response.json()) as Record<string, unknown>[];

    // BUG: Content-Range header value "0-24/100" is kept as a raw string.
    // Downstream code does count + 1 which produces "0-24/1001" (string concat).
    const contentRange = response.headers.get("Content-Range") ?? "0-0/0";
    const count = contentRange.split("/")[1]; // "100" as string, not number 100

    const users = rows.map((row) => ({
      // BUG: parseInt on UUID "550e8400-e29b-..." returns 550 (stops at first dash)
      id: parseInt(String(row.id), 10),
      email: String(row.email),
      user_role: String(row.user_role),
      // BUG: profile inner fields typed as non-null string but can be null
      profile: row.profile as { avatar_url: string; bio: string } | null,
      // BUG: Number() silently loses precision for values > 2^53
      balance: Number(row.balance),
      metadata: row.metadata as Record<string, unknown> | null,
      created_at: String(row.created_at),
      updated_at: String(row.updated_at),
    }));

    return { users, count };
  }
}
