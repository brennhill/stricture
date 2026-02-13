// B14: Fetches first page but does not parse Content-Range header for total count
// and does not paginate through subsequent pages.

class SupabaseClientB14 {
  private readonly baseUrl: string;
  private readonly anonKey: string;
  private readonly jwt: string;

  constructor(baseUrl: string, anonKey: string, jwt: string) {
    this.baseUrl = baseUrl;
    this.anonKey = anonKey;
    this.jwt = jwt;
  }

  // BUG: Named "listAllUsers" but only fetches first 25 rows.
  async listAllUsers(): Promise<Record<string, unknown>[]> {
    let response: Response;
    try {
      response = await fetch(
        `${this.baseUrl}/rest/v1/users?select=*`,
        {
          headers: {
            apikey: this.anonKey,
            Authorization: `Bearer ${this.jwt}`,
            Range: "0-24",
            Prefer: "count=exact",
          },
        }
      );
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok && response.status !== 206) {
      throw new Error(`Query failed: ${response.status}`);
    }

    const rows = (await response.json()) as Record<string, unknown>[];

    // BUG: Content-Range header "0-24/500" contains the total count (500)
    // but it is completely ignored. The function returns at most 25 rows
    // and the caller has no way to know 475 rows were not returned.
    // The Content-Range header is not even read.

    return rows;
  }

  // BUG: This variant reads Content-Range but does not paginate.
  async listUsersWithCount(): Promise<{
    users: Record<string, unknown>[];
    total: number;
  }> {
    let response: Response;
    try {
      response = await fetch(
        `${this.baseUrl}/rest/v1/users?select=*`,
        {
          headers: {
            apikey: this.anonKey,
            Authorization: `Bearer ${this.jwt}`,
            Range: "0-24",
            Prefer: "count=exact",
          },
        }
      );
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok && response.status !== 206) {
      throw new Error(`Query failed: ${response.status}`);
    }

    const rows = (await response.json()) as Record<string, unknown>[];

    // BUG: Reads Content-Range but only extracts total. Does not use it
    // to determine if more pages exist. The caller receives { total: 500 }
    // but only 25 users. No loop to fetch pages 1, 2, 3, ... 19.
    const contentRange = response.headers.get("Content-Range") ?? "";
    const totalStr = contentRange.split("/")[1];
    // BUG: Does not handle "*" total (unknown count).
    // parseInt("*") returns NaN.
    const total = parseInt(totalStr ?? "0", 10);

    return { users: rows, total };
  }

  // Example of how this causes real damage:
  async exportAllUserEmails(): Promise<string[]> {
    const users = await this.listAllUsers();
    return users.map((u) => String(u.email));
    // BUG: If there are 10,000 users, the export contains 25 email addresses.
  }
}
