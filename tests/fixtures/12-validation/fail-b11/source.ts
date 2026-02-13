// B11: Timestamps parsed with timezone loss. Balance converted via Number().

class SupabaseClientB11 {
  private readonly baseUrl: string;
  private readonly anonKey: string;
  private readonly jwt: string;

  constructor(baseUrl: string, anonKey: string, jwt: string) {
    this.baseUrl = baseUrl;
    this.anonKey = anonKey;
    this.jwt = jwt;
  }

  async getUser(userId: string): Promise<{
    id: string;
    email: string;
    balance: number;           // BUG: Should be bigint
    created_at: Date;          // BUG: Date object loses timezone info
    updated_at: Date;
  }> {
    let response: Response;
    try {
      response = await fetch(
        `${this.baseUrl}/rest/v1/users?id=eq.${userId}&select=*`,
        {
          headers: {
            apikey: this.anonKey,
            Authorization: `Bearer ${this.jwt}`,
          },
        }
      );
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) throw new Error(`Query failed: ${response.status}`);

    const rows = (await response.json()) as Record<string, unknown>[];
    const row = rows[0];

    return {
      id: String(row.id),
      email: String(row.email),

      // BUG: Number() truncates values > Number.MAX_SAFE_INTEGER
      // PostgreSQL bigint "9007199254740993" becomes 9007199254740992
      balance: Number(row.balance),

      // BUG: new Date() converts to local timezone, losing the original
      // timezone offset. PostgreSQL "2026-02-12T10:00:00+09:00" (Tokyo time)
      // becomes Date in the server's local timezone. If the server is in UTC,
      // the Date object is correct, but toString() shows UTC. If the server
      // is in US/Pacific, the Date stores the same instant but the timezone
      // context is lost. Comparing two timestamps from different timezones
      // requires the original offset, which is discarded.
      created_at: new Date(String(row.created_at)),
      updated_at: new Date(String(row.updated_at)),
    };
  }

  // BUG: Duration calculation using Date objects loses sub-second precision.
  // PostgreSQL timestamptz has microsecond precision. JavaScript Date has
  // millisecond precision. A 500-microsecond difference is rounded to 0ms.
  calculateAccountAge(createdAt: Date): number {
    const now = new Date();
    // Returns days, but with timezone and precision loss compounding
    return Math.floor((now.getTime() - createdAt.getTime()) / (1000 * 60 * 60 * 24));
  }
}
