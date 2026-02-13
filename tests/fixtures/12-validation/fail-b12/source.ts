// B12: Accesses user_metadata and profile without null checks.

class SupabaseClientB12 {
  private readonly baseUrl: string;
  private readonly anonKey: string;

  constructor(baseUrl: string, anonKey: string) {
    this.baseUrl = baseUrl;
    this.anonKey = anonKey;
  }

  async signUp(
    email: string,
    password: string,
    displayName: string
  ): Promise<{ userId: string; displayName: string }> {
    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/auth/v1/signup`, {
        method: "POST",
        headers: {
          apikey: this.anonKey,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          email,
          password,
          data: { display_name: displayName },
        }),
      });
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) throw new Error(`Sign-up failed: ${response.status}`);

    const result = (await response.json()) as Record<string, unknown>;

    // BUG: user_metadata can be null (manifest declares nullable: true).
    // When email confirmation is required, the response may not include
    // the user_metadata field or it may be null.
    // Accessing .display_name on null throws:
    //   TypeError: Cannot read properties of null (reading 'display_name')
    const metadata = result.user_metadata as Record<string, unknown>;
    const name = metadata.display_name as string; // CRASH when metadata is null

    return { userId: String(result.id), displayName: name };
  }

  async getUserAvatar(userId: string): Promise<string> {
    let response: Response;
    try {
      response = await fetch(
        `${this.baseUrl}/rest/v1/users?id=eq.${userId}&select=profile`,
        {
          headers: {
            apikey: this.anonKey,
            Authorization: `Bearer fake-jwt`,
          },
        }
      );
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) throw new Error(`Query failed: ${response.status}`);

    const rows = (await response.json()) as Record<string, unknown>[];
    const row = rows[0];

    // BUG: profile is nullable JSONB. When profile column is NULL:
    // row.profile is null, accessing .avatar_url on null throws TypeError.
    const profile = row.profile as { avatar_url: string; bio: string };
    return profile.avatar_url ?? "https://default-avatar.example.com/default.png";
    // ^ The ?? never executes because the TypeError crashes first
  }

  async getUserMetadataKeys(userId: string): Promise<string[]> {
    let response: Response;
    try {
      response = await fetch(
        `${this.baseUrl}/rest/v1/users?id=eq.${userId}&select=metadata`,
        {
          headers: {
            apikey: this.anonKey,
            Authorization: `Bearer fake-jwt`,
          },
        }
      );
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) throw new Error(`Query failed: ${response.status}`);

    const rows = (await response.json()) as Record<string, unknown>[];
    const row = rows[0];

    // BUG: metadata is nullable JSONB. Object.keys(null) throws TypeError.
    const metadata = row.metadata as Record<string, unknown>;
    return Object.keys(metadata);
  }
}
