// B06: Client type expects { data: User[] } wrapper but PostgREST returns
// bare arrays. Also uses "role" instead of "user_role".

interface UserB06 {
  id: string;
  email: string;
  role: string;          // BUG: Database column is "user_role", not "role"
  profile: { avatar_url: string | null; bio: string | null } | null;
  balance: number;
  metadata: Record<string, unknown> | null;
  created_at: string;
  updated_at: string;
}

interface PostgRESTResponseB06 {
  // BUG: PostgREST does NOT wrap results in a { data: [...] } object.
  // It returns bare arrays directly: [{ id, email, ... }, ...]
  data: UserB06[];
  count: number;
}

class SupabaseClientB06 {
  private readonly baseUrl: string;
  private readonly anonKey: string;
  private readonly jwt: string;

  constructor(baseUrl: string, anonKey: string, jwt: string) {
    this.baseUrl = baseUrl;
    this.anonKey = anonKey;
    this.jwt = jwt;
  }

  async queryUsers(): Promise<UserB06[]> {
    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/rest/v1/users?select=*`, {
        headers: {
          apikey: this.anonKey,
          Authorization: `Bearer ${this.jwt}`,
          Prefer: "count=exact",
        },
      });
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) {
      throw new Error(`Query failed: ${response.status}`);
    }

    // BUG: Casts response to { data: [...], count: ... } but PostgREST
    // returns a bare array. result.data is undefined.
    const result = (await response.json()) as PostgRESTResponseB06;
    return result.data; // returns undefined
  }

  async getUserRole(userId: string): Promise<string> {
    const users = await this.queryUsers();
    // BUG: users is undefined (from the wrapper mismatch), so this crashes
    // with TypeError: Cannot read properties of undefined (reading 'find')
    const user = users.find((u) => u.id === userId);
    if (!user) throw new Error("User not found");

    // BUG: Even if we fix the wrapper issue, user.role is undefined
    // because the actual field is user_role.
    return user.role;
  }
}
