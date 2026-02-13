// B13: Uses anon key without JWT, ignoring Row-Level Security.
// No verification that the correct rows are returned per user identity.

class SupabaseClientB13 {
  private readonly baseUrl: string;
  private readonly anonKey: string;

  constructor(baseUrl: string, anonKey: string) {
    // BUG: No JWT token provided. Uses only the anon key.
    // All queries execute with the anon role's RLS policies,
    // which may return zero rows or a restricted subset.
    this.baseUrl = baseUrl;
    this.anonKey = anonKey;
  }

  async countAllUsers(): Promise<number> {
    let response: Response;
    try {
      response = await fetch(
        `${this.baseUrl}/rest/v1/users?select=*`,
        {
          method: "HEAD",
          headers: {
            apikey: this.anonKey,
            // BUG: No Authorization header with user JWT.
            // The request authenticates as anon, which may see 0 rows.
            Prefer: "count=exact",
          },
        }
      );
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    // BUG: This count is RLS-filtered. With anon key only,
    // the RLS policy may restrict visibility to 0 rows.
    // The caller assumes this is the total user count.
    const contentRange = response.headers.get("Content-Range");
    if (!contentRange) return 0;

    const total = contentRange.split("/")[1];
    return parseInt(total ?? "0", 10);
  }

  async getUser(userId: string): Promise<Record<string, unknown> | null> {
    let response: Response;
    try {
      response = await fetch(
        `${this.baseUrl}/rest/v1/users?id=eq.${userId}&select=*`,
        {
          headers: {
            apikey: this.anonKey,
            // BUG: No Authorization header. RLS may filter this row out.
          },
        }
      );
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) throw new Error(`Query failed: ${response.status}`);

    const rows = (await response.json()) as Record<string, unknown>[];

    // BUG: When RLS blocks access, PostgREST returns an empty array,
    // not a 403. The code returns null, which is indistinguishable
    // from "user does not exist."
    return rows[0] ?? null;
  }

  // BUG: This function is used to verify "user exists before sending email."
  // But with anon RLS, a user that exists may return null (invisible to anon role).
  // The system incorrectly concludes the user does not exist and skips the email.
  async userExists(userId: string): Promise<boolean> {
    const user = await this.getUser(userId);
    return user !== null;
  }
}
