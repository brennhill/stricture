# 12 -- Supabase REST API

**Why included:** Row-Level Security filtering, PostgREST bare-array responses (no wrapper object), Content-Range pagination, PostgreSQL type coercion (bigint, UUID, JSONB), Prefer header semantics, Auth JWT lifecycle, Phoenix Channels realtime over WebSocket.

**Base URL:** `https://<project-ref>.supabase.co`
**Auth pattern:** Dual-header: `apikey` header (anon/service key) + `Authorization: Bearer <jwt>` (user JWT from auth flow).

---

## Manifest Fragment

```yaml
# .stricture-manifest.yml
contracts:
  # ── PostgREST Query ────────────────────────────────────────
  - id: "supabase-postgrest-query"
    producer: supabase-postgrest
    consumers: [my-service]
    protocol: http
    endpoints:
      - path: "/rest/v1/users"
        method: GET
        request:
          headers:
            apikey:        { type: string, required: true, format: "eyJ*" }
            Authorization: { type: string, required: true, format: "Bearer eyJ*" }
            Range:         { type: string, required: false, format: "^\\d+-\\d+$" }
            Prefer:        { type: string, required: false, values: ["count=exact", "count=planned", "count=estimated"] }
          query_params:
            select:   { type: string, required: false }
            id:       { type: string, required: false, format: "^(eq|neq|gt|gte|lt|lte|in|is)\\..*$" }
            email:    { type: string, required: false, format: "^(eq|neq|like|ilike)\\..*$" }
            order:    { type: string, required: false, format: "^\\w+\\.(asc|desc)(\\.nullsfirst|\\.nullslast)?$" }
        response:
          type: array
          note: "PostgREST returns bare arrays, NOT { data: [...] } wrapper objects"
          items:
            fields:
              id:         { type: string, format: uuid_v4, required: true }
              email:      { type: string, format: email, required: true }
              user_role:  { type: enum, values: ["admin", "editor", "viewer", "guest"], required: true }
              profile:    { type: object, nullable: true, fields: { avatar_url: { type: string, format: url, nullable: true }, bio: { type: string, nullable: true } } }
              balance:    { type: integer, range: [0, 9223372036854775807], required: true, note: "PostgreSQL bigint -- exceeds Number.MAX_SAFE_INTEGER" }
              metadata:   { type: object, nullable: true, note: "Arbitrary JSONB column" }
              created_at: { type: string, format: iso8601_timestamptz, required: true }
              updated_at: { type: string, format: iso8601_timestamptz, required: true }
          response_headers:
            Content-Range: { type: string, format: "^\\d+-\\d+/\\d+$", required: false, note: "Present when Prefer: count=exact is set" }
        status_codes: [200, 206, 400, 401, 403, 406, 416]

  # ── Row Insert ─────────────────────────────────────────────
  - id: "supabase-postgrest-insert"
    producer: supabase-postgrest
    consumers: [my-service]
    protocol: http
    endpoints:
      - path: "/rest/v1/users"
        method: POST
        request:
          headers:
            apikey:        { type: string, required: true }
            Authorization: { type: string, required: true, format: "Bearer eyJ*" }
            Content-Type:  { type: string, required: true, values: ["application/json"] }
            Prefer:        { type: string, required: false, values: ["return=minimal", "return=headers-only", "return=representation"] }
          body:
            fields:
              email:     { type: string, format: email, required: true }
              user_role: { type: enum, values: ["admin", "editor", "viewer", "guest"], required: true }
              profile:   { type: object, nullable: true, required: false }
              balance:   { type: integer, range: [0, 9223372036854775807], required: false, default: 0 }
              metadata:  { type: object, nullable: true, required: false }
        response:
          note: "Empty body with 201 unless Prefer: return=representation, then array of inserted rows"
          type: conditional
          conditions:
            - when: { header: "Prefer", value: "return=representation" }
              type: array
              items: { ref: "#/contracts/0/endpoints/0/response/items" }
            - when: { header: "Prefer", absent: true }
              type: empty
        status_codes: [201, 400, 401, 403, 409, 500]

  # ── Auth Sign-Up ───────────────────────────────────────────
  - id: "supabase-auth-signup"
    producer: supabase-gotrue
    consumers: [my-service]
    protocol: http
    endpoints:
      - path: "/auth/v1/signup"
        method: POST
        request:
          headers:
            apikey:       { type: string, required: true, format: "eyJ*" }
            Content-Type: { type: string, required: true, values: ["application/json"] }
          body:
            fields:
              email:    { type: string, format: email, required: true }
              password: { type: string, minLength: 6, required: true }
              data:     { type: object, required: false, note: "Custom user_metadata" }
              phone:    { type: string, format: "^\\+[1-9]\\d{1,14}$", required: false }
        response:
          fields:
            id:                { type: string, format: uuid_v4, required: true }
            aud:               { type: string, required: true }
            role:              { type: string, required: true }
            email:             { type: string, format: email, required: true }
            phone:             { type: string, required: false, nullable: true }
            confirmation_sent_at: { type: string, format: iso8601_timestamptz, required: false, nullable: true }
            email_confirmed_at:   { type: string, format: iso8601_timestamptz, required: false, nullable: true }
            created_at:        { type: string, format: iso8601_timestamptz, required: true }
            updated_at:        { type: string, format: iso8601_timestamptz, required: true }
            user_metadata:     { type: object, required: true, nullable: true }
            app_metadata:      { type: object, required: true }
            identities:        { type: array, items: { type: object }, required: true }
            access_token:      { type: string, format: jwt, required: true, note: "Returned only on auto-confirm or after email confirmation" }
            token_type:        { type: literal, value: "bearer", required: true }
            expires_in:        { type: integer, required: true }
            expires_at:        { type: integer, required: true, note: "Unix epoch seconds" }
            refresh_token:     { type: string, required: true }
          error:
            shape:
              error:             { type: string, required: false }
              error_description: { type: string, required: false }
              msg:               { type: string, required: false }
              code:              { type: integer, required: false }
        status_codes: [200, 400, 422, 429, 500]

  # ── Realtime Subscription ──────────────────────────────────
  - id: "supabase-realtime"
    producer: supabase-realtime
    consumers: [my-service]
    protocol: websocket
    endpoints:
      - path: "/realtime/v1/websocket"
        query_params:
          apikey: { type: string, required: true }
          vsn:    { type: string, required: false, values: ["1.0.0"] }
        note: "Phoenix Channels protocol over WebSocket"
        messages:
          join:
            topic:   { type: string, format: "realtime:<schema>:<table>", required: true }
            event:   { type: literal, value: "phx_join", required: true }
            payload: { type: object, fields: { config: { type: object, fields: { broadcast: { type: object }, presence: { type: object }, postgres_changes: { type: array, items: { type: object, fields: { event: { type: enum, values: ["INSERT", "UPDATE", "DELETE", "*"] }, schema: { type: string, default: "public" }, table: { type: string }, filter: { type: string, required: false } } } } } } }
            ref:     { type: string, required: true }
          heartbeat:
            topic:   { type: literal, value: "phoenix", required: true }
            event:   { type: literal, value: "heartbeat", required: true }
            payload: { type: object, required: true }
            ref:     { type: string, required: true }
          postgres_changes:
            topic:   { type: string, required: true }
            event:   { type: literal, value: "postgres_changes", required: true }
            payload: { type: object, fields: { data: { type: object, fields: { type: { type: enum, values: ["INSERT", "UPDATE", "DELETE"] }, table: { type: string }, schema: { type: string }, record: { type: object, nullable: true }, old_record: { type: object, nullable: true }, commit_timestamp: { type: string, format: iso8601_timestamptz } } } } }

    error_shape:
      fields:
        message: { type: string, required: true }
        details: { type: string, nullable: true, required: true }
        hint:    { type: string, nullable: true, required: true }
        code:    { type: string, required: true, note: "PostgreSQL error code, e.g. 23505 for unique_violation" }
```

---

## PERFECT -- Correct Integration

### Implementation (TypeScript)

```typescript
// supabase-client.ts -- Production-quality Supabase integration.

// ── Types ──────────────────────────────────────────────────

interface SupabaseError {
  message: string;
  details: string | null;
  hint: string | null;
  code: string;
}

interface AuthError {
  error?: string;
  error_description?: string;
  msg?: string;
  code?: number;
}

interface UserProfile {
  avatar_url: string | null;
  bio: string | null;
}

interface UserRow {
  id: string;                                    // UUID v4
  email: string;
  user_role: "admin" | "editor" | "viewer" | "guest";
  profile: UserProfile | null;                   // nullable JSONB
  balance: bigint;                               // PostgreSQL bigint
  metadata: Record<string, unknown> | null;      // nullable JSONB
  created_at: string;                            // ISO 8601 timestamptz
  updated_at: string;                            // ISO 8601 timestamptz
}

interface AuthUser {
  id: string;                                    // UUID v4
  aud: string;
  role: string;
  email: string;
  phone: string | null;
  confirmation_sent_at: string | null;
  email_confirmed_at: string | null;
  created_at: string;
  updated_at: string;
  user_metadata: Record<string, unknown> | null;
  app_metadata: Record<string, unknown>;
  identities: Record<string, unknown>[];
}

interface AuthSession {
  access_token: string;
  token_type: "bearer";
  expires_in: number;
  expires_at: number;                            // Unix epoch seconds
  refresh_token: string;
  user: AuthUser;
}

interface PaginatedResult<T> {
  data: T[];
  totalCount: number;
  rangeStart: number;
  rangeEnd: number;
}

type UserRole = "admin" | "editor" | "viewer" | "guest";

interface InsertUserParams {
  email: string;
  user_role: UserRole;
  profile?: UserProfile | null;
  balance?: bigint;
  metadata?: Record<string, unknown> | null;
}

interface SignUpParams {
  email: string;
  password: string;
  data?: Record<string, unknown>;
}

interface RealtimeMessage {
  type: "INSERT" | "UPDATE" | "DELETE";
  table: string;
  schema: string;
  record: Record<string, unknown> | null;
  old_record: Record<string, unknown> | null;
  commit_timestamp: string;
}

type SupabaseResult<T> =
  | { ok: true; data: T }
  | { ok: false; status: number; error: SupabaseError };

type AuthResult =
  | { ok: true; session: AuthSession }
  | { ok: false; status: number; error: AuthError };

// ── Validation Helpers ──────────────────────────────────────

const UUID_V4_PATTERN =
  /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

const JWT_PATTERN = /^eyJ[A-Za-z0-9_-]+\.eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+$/;

const ISO8601_PATTERN =
  /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$/;

const VALID_ROLES: ReadonlySet<UserRole> = new Set([
  "admin", "editor", "viewer", "guest",
]);

const VALID_CHANGE_EVENTS: ReadonlySet<string> = new Set([
  "INSERT", "UPDATE", "DELETE", "*",
]);

const MAX_PAGE_SIZE = 1000;
const MIN_PASSWORD_LENGTH = 6;

function validateUUIDv4(id: string): void {
  if (!UUID_V4_PATTERN.test(id)) {
    throw new Error(`Invalid UUID v4 format: ${id}`);
  }
}

function validateEmail(email: string): void {
  if (!EMAIL_PATTERN.test(email)) {
    throw new Error(`Invalid email format: ${email}`);
  }
}

function validateJWT(token: string): void {
  if (!JWT_PATTERN.test(token)) {
    throw new Error(`Invalid JWT format`);
  }
}

function validateISO8601(ts: string): void {
  if (!ISO8601_PATTERN.test(ts)) {
    throw new Error(`Invalid ISO 8601 timestamp: ${ts}`);
  }
}

function validateRole(role: string): asserts role is UserRole {
  if (!VALID_ROLES.has(role as UserRole)) {
    throw new Error(`Invalid user_role: ${role}. Must be one of: ${[...VALID_ROLES].join(", ")}`);
  }
}

function validatePageSize(size: number): void {
  if (!Number.isInteger(size) || size < 1 || size > MAX_PAGE_SIZE) {
    throw new Error(`Page size must be between 1 and ${MAX_PAGE_SIZE}, got ${size}`);
  }
}

function validatePassword(password: string): void {
  if (typeof password !== "string" || password.length < MIN_PASSWORD_LENGTH) {
    throw new Error(`Password must be at least ${MIN_PASSWORD_LENGTH} characters`);
  }
}

// ── Content-Range Parser ────────────────────────────────────

interface ContentRange {
  start: number;
  end: number;
  total: number;
}

function parseContentRange(header: string | null): ContentRange | null {
  if (!header) return null;

  const match = header.match(/^(\d+)-(\d+)\/(\d+|\*)$/);
  if (!match) return null;

  const start = parseInt(match[1], 10);
  const end = parseInt(match[2], 10);
  const total = match[3] === "*" ? -1 : parseInt(match[3], 10);

  return { start, end, total };
}

// ── Row Parser (validates response shape) ───────────────────

function parseUserRow(raw: Record<string, unknown>): UserRow {
  const id = String(raw.id);
  validateUUIDv4(id);

  const email = String(raw.email);
  const userRole = String(raw.user_role);
  validateRole(userRole);

  const createdAt = String(raw.created_at);
  const updatedAt = String(raw.updated_at);
  validateISO8601(createdAt);
  validateISO8601(updatedAt);

  // Balance: PostgreSQL bigint can exceed Number.MAX_SAFE_INTEGER.
  // PostgREST returns it as a string in JSON. Use BigInt for safe handling.
  const balance = BigInt(String(raw.balance));

  // Profile: nullable JSONB -- guard before accessing nested fields
  const rawProfile = raw.profile as Record<string, unknown> | null;
  const profile: UserProfile | null =
    rawProfile !== null && rawProfile !== undefined
      ? {
          avatar_url:
            rawProfile.avatar_url != null ? String(rawProfile.avatar_url) : null,
          bio: rawProfile.bio != null ? String(rawProfile.bio) : null,
        }
      : null;

  // Metadata: nullable JSONB
  const metadata = (raw.metadata as Record<string, unknown> | null) ?? null;

  return {
    id,
    email,
    user_role: userRole,
    profile,
    balance,
    metadata,
    created_at: createdAt,
    updated_at: updatedAt,
  };
}

// ── JWT Token Refresh ───────────────────────────────────────

function isTokenExpired(expiresAt: number): boolean {
  const nowSeconds = Math.floor(Date.now() / 1000);
  // Refresh 60 seconds before actual expiry to avoid edge cases
  return nowSeconds >= expiresAt - 60;
}

// ── Supabase Client ─────────────────────────────────────────

class SupabaseClient {
  private readonly baseUrl: string;
  private readonly anonKey: string;
  private accessToken: string;
  private refreshToken: string;
  private expiresAt: number;

  constructor(
    baseUrl: string,
    anonKey: string,
    accessToken: string,
    refreshToken: string,
    expiresAt: number
  ) {
    this.baseUrl = baseUrl;
    this.anonKey = anonKey;
    this.accessToken = accessToken;
    this.refreshToken = refreshToken;
    this.expiresAt = expiresAt;
  }

  // ── Token Refresh ──────────────────────────────────────

  private async ensureFreshToken(): Promise<void> {
    if (!isTokenExpired(this.expiresAt)) return;

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/auth/v1/token?grant_type=refresh_token`, {
        method: "POST",
        headers: {
          apikey: this.anonKey,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ refresh_token: this.refreshToken }),
      });
    } catch (err) {
      throw new Error(`Network error refreshing token: ${(err as Error).message}`);
    }

    if (!response.ok) {
      throw new Error(`Token refresh failed with status ${response.status}`);
    }

    const body = (await response.json()) as {
      access_token: string;
      refresh_token: string;
      expires_at: number;
    };

    this.accessToken = body.access_token;
    this.refreshToken = body.refresh_token;
    this.expiresAt = body.expires_at;
  }

  private authHeaders(): Record<string, string> {
    return {
      apikey: this.anonKey,
      Authorization: `Bearer ${this.accessToken}`,
    };
  }

  // ── PostgREST Query (GET with pagination) ──────────────

  async queryUsers(
    filters?: { email?: string; role?: UserRole },
    pageSize: number = 25
  ): Promise<PaginatedResult<UserRow>> {
    validatePageSize(pageSize);
    if (filters?.role) validateRole(filters.role);

    await this.ensureFreshToken();

    const allUsers: UserRow[] = [];
    let offset = 0;
    let totalCount = 0;

    while (true) {
      const params = new URLSearchParams();
      params.set("select", "*");
      if (filters?.email) {
        params.set("email", `ilike.%${filters.email}%`);
      }
      if (filters?.role) {
        params.set("user_role", `eq.${filters.role}`);
      }
      params.set("order", "created_at.desc");

      let response: Response;
      try {
        response = await fetch(
          `${this.baseUrl}/rest/v1/users?${params.toString()}`,
          {
            method: "GET",
            headers: {
              ...this.authHeaders(),
              Range: `${offset}-${offset + pageSize - 1}`,
              Prefer: "count=exact",
            },
          }
        );
      } catch (err) {
        throw new Error(`Network error querying users: ${(err as Error).message}`);
      }

      // PostgREST returns 200 for full results, 206 for partial (paginated)
      if (response.status !== 200 && response.status !== 206) {
        const errorBody = (await response.json()) as SupabaseError;

        switch (response.status) {
          case 400:
            throw new Error(`Bad request: ${errorBody.message}. Hint: ${errorBody.hint}`);
          case 401:
            throw new Error(`Unauthorized: ${errorBody.message}`);
          case 403:
            throw new Error(`Forbidden (RLS): ${errorBody.message}`);
          case 406:
            throw new Error(`Not acceptable: ${errorBody.message}`);
          case 416:
            throw new Error(`Range not satisfiable: ${errorBody.message}`);
          default:
            throw new Error(`Query failed (${response.status}): ${errorBody.message}`);
        }
      }

      // PostgREST returns bare arrays, not { data: [...] }
      const rows = (await response.json()) as Record<string, unknown>[];

      if (!Array.isArray(rows)) {
        throw new Error("Expected array response from PostgREST, got non-array");
      }

      for (const row of rows) {
        allUsers.push(parseUserRow(row));
      }

      // Parse Content-Range header for total count
      const contentRange = parseContentRange(
        response.headers.get("Content-Range")
      );
      if (contentRange && contentRange.total >= 0) {
        totalCount = contentRange.total;
      }

      // Check if there are more pages
      if (rows.length < pageSize) {
        break; // Last page -- fewer rows than requested
      }

      offset += pageSize;

      // Safety: stop if we have fetched all rows
      if (totalCount > 0 && offset >= totalCount) {
        break;
      }
    }

    return {
      data: allUsers,
      totalCount,
      rangeStart: 0,
      rangeEnd: allUsers.length - 1,
    };
  }

  // ── Row Insert (POST with Prefer: return=representation) ──

  async insertUser(params: InsertUserParams): Promise<SupabaseResult<UserRow>> {
    validateEmail(params.email);
    validateRole(params.user_role);

    await this.ensureFreshToken();

    const body: Record<string, unknown> = {
      email: params.email,
      user_role: params.user_role,
      profile: params.profile ?? null,
      balance: params.balance !== undefined ? String(params.balance) : "0",
      metadata: params.metadata ?? null,
    };

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/rest/v1/users`, {
        method: "POST",
        headers: {
          ...this.authHeaders(),
          "Content-Type": "application/json",
          Prefer: "return=representation",
        },
        body: JSON.stringify(body),
      });
    } catch (err) {
      throw new Error(`Network error inserting user: ${(err as Error).message}`);
    }

    if (response.status !== 201) {
      const errorBody = (await response.json()) as SupabaseError;

      switch (response.status) {
        case 400:
          return { ok: false, status: 400, error: errorBody };
        case 401:
          return { ok: false, status: 401, error: errorBody };
        case 403:
          return { ok: false, status: 403, error: errorBody };
        case 409:
          return {
            ok: false,
            status: 409,
            error: {
              ...errorBody,
              message: `Unique constraint violated: ${errorBody.message}`,
            },
          };
        default:
          return { ok: false, status: response.status, error: errorBody };
      }
    }

    // Prefer: return=representation returns an array of inserted rows
    const rows = (await response.json()) as Record<string, unknown>[];
    if (!Array.isArray(rows) || rows.length === 0) {
      throw new Error("Expected non-empty array from insert with return=representation");
    }

    const user = parseUserRow(rows[0]);
    return { ok: true, data: user };
  }

  // ── Auth Sign-Up ───────────────────────────────────────

  async signUp(params: SignUpParams): Promise<AuthResult> {
    validateEmail(params.email);
    validatePassword(params.password);

    const body: Record<string, unknown> = {
      email: params.email,
      password: params.password,
    };
    if (params.data !== undefined) {
      body.data = params.data;
    }

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
      throw new Error(`Network error during sign-up: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = (await response.json()) as AuthError;

      switch (response.status) {
        case 400:
          return { ok: false, status: 400, error: errorBody };
        case 422:
          return {
            ok: false,
            status: 422,
            error: {
              ...errorBody,
              msg: errorBody.msg ?? "Validation failed",
            },
          };
        case 429:
          return {
            ok: false,
            status: 429,
            error: {
              ...errorBody,
              msg: "Rate limit exceeded. Try again later.",
            },
          };
        default:
          return { ok: false, status: response.status, error: errorBody };
      }
    }

    const result = (await response.json()) as Record<string, unknown>;

    // GoTrue returns user object at top level with session fields
    const user: AuthUser = {
      id: String(result.id),
      aud: String(result.aud),
      role: String(result.role),
      email: String(result.email),
      phone: result.phone != null ? String(result.phone) : null,
      confirmation_sent_at: result.confirmation_sent_at != null
        ? String(result.confirmation_sent_at) : null,
      email_confirmed_at: result.email_confirmed_at != null
        ? String(result.email_confirmed_at) : null,
      created_at: String(result.created_at),
      updated_at: String(result.updated_at),
      user_metadata: (result.user_metadata as Record<string, unknown> | null) ?? null,
      app_metadata: (result.app_metadata as Record<string, unknown>) ?? {},
      identities: (result.identities as Record<string, unknown>[]) ?? [],
    };

    // Validate the returned user ID format
    validateUUIDv4(user.id);

    const session: AuthSession = {
      access_token: String(result.access_token),
      token_type: "bearer",
      expires_in: Number(result.expires_in),
      expires_at: Number(result.expires_at),
      refresh_token: String(result.refresh_token),
      user,
    };

    // Store tokens for subsequent requests
    this.accessToken = session.access_token;
    this.refreshToken = session.refresh_token;
    this.expiresAt = session.expires_at;

    return { ok: true, session };
  }

  // ── Realtime Subscription ──────────────────────────────

  subscribeToChanges(
    table: string,
    events: Array<"INSERT" | "UPDATE" | "DELETE" | "*">,
    callback: (message: RealtimeMessage) => void,
    onError?: (error: Error) => void
  ): { unsubscribe: () => void } {
    // Validate event types
    for (const event of events) {
      if (!VALID_CHANGE_EVENTS.has(event)) {
        throw new Error(`Invalid change event: ${event}`);
      }
    }

    const wsUrl = this.baseUrl
      .replace("https://", "wss://")
      .replace("http://", "ws://");

    const url = `${wsUrl}/realtime/v1/websocket?apikey=${this.anonKey}&vsn=1.0.0`;

    let ws: WebSocket;
    try {
      ws = new WebSocket(url);
    } catch (err) {
      throw new Error(`Failed to create WebSocket: ${(err as Error).message}`);
    }

    const topic = `realtime:public:${table}`;
    let heartbeatInterval: ReturnType<typeof setInterval> | null = null;
    let refCounter = 0;
    let joined = false;
    const pendingMessages: RealtimeMessage[] = [];

    const nextRef = (): string => {
      refCounter += 1;
      return String(refCounter);
    };

    // Register message handler BEFORE sending join to avoid race condition
    // where messages arrive before the handler is set up.
    ws.onmessage = (event: MessageEvent) => {
      let msg: Record<string, unknown>;
      try {
        msg = JSON.parse(String(event.data)) as Record<string, unknown>;
      } catch {
        return; // Ignore non-JSON frames
      }

      const msgEvent = String(msg.event);

      if (msgEvent === "phx_reply" && msg.ref === "1") {
        // Join acknowledgement
        joined = true;

        // Flush any messages that arrived during join
        for (const pending of pendingMessages) {
          callback(pending);
        }
        pendingMessages.length = 0;
        return;
      }

      if (msgEvent === "postgres_changes") {
        const payload = msg.payload as Record<string, unknown> | undefined;
        const data = payload?.data as Record<string, unknown> | undefined;
        if (!data) return;

        const changeType = String(data.type);
        if (
          changeType !== "INSERT" &&
          changeType !== "UPDATE" &&
          changeType !== "DELETE"
        ) {
          return;
        }

        const message: RealtimeMessage = {
          type: changeType,
          table: String(data.table),
          schema: String(data.schema),
          record: (data.record as Record<string, unknown> | null) ?? null,
          old_record: (data.old_record as Record<string, unknown> | null) ?? null,
          commit_timestamp: String(data.commit_timestamp),
        };

        // If join is not yet acknowledged, buffer the message
        if (!joined) {
          pendingMessages.push(message);
        } else {
          callback(message);
        }
      }
    };

    ws.onerror = (event: Event) => {
      if (onError) {
        onError(new Error("WebSocket error"));
      }
    };

    ws.onclose = () => {
      if (heartbeatInterval) {
        clearInterval(heartbeatInterval);
        heartbeatInterval = null;
      }
    };

    ws.onopen = () => {
      // Send join message
      const joinPayload = {
        topic,
        event: "phx_join",
        payload: {
          config: {
            broadcast: { self: false },
            presence: {},
            postgres_changes: events.map((e) => ({
              event: e,
              schema: "public",
              table,
            })),
          },
        },
        ref: nextRef(),
      };
      ws.send(JSON.stringify(joinPayload));

      // Start heartbeat (every 30 seconds)
      heartbeatInterval = setInterval(() => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(
            JSON.stringify({
              topic: "phoenix",
              event: "heartbeat",
              payload: {},
              ref: nextRef(),
            })
          );
        }
      }, 30_000);
    };

    return {
      unsubscribe: () => {
        if (heartbeatInterval) {
          clearInterval(heartbeatInterval);
        }
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(
            JSON.stringify({
              topic,
              event: "phx_leave",
              payload: {},
              ref: nextRef(),
            })
          );
          ws.close(1000, "Client unsubscribe");
        }
      },
    };
  }
}

export {
  SupabaseClient,
  parseContentRange,
  parseUserRow,
  isTokenExpired,
  validateUUIDv4,
  validateEmail,
  validateJWT,
  validateISO8601,
  validateRole,
  validatePageSize,
  validatePassword,
};
export type {
  UserRow,
  UserProfile,
  AuthUser,
  AuthSession,
  PaginatedResult,
  SupabaseError,
  AuthError,
  SupabaseResult,
  AuthResult,
  RealtimeMessage,
  InsertUserParams,
  SignUpParams,
  UserRole,
  ContentRange,
};
```

### Tests

```typescript
// supabase-client.test.ts -- Comprehensive tests for Supabase integration.

import { describe, it, expect, beforeEach, vi, afterEach } from "vitest";
import {
  SupabaseClient,
  parseContentRange,
  parseUserRow,
  isTokenExpired,
  validateUUIDv4,
  validateEmail,
  validateJWT,
  validateISO8601,
  validateRole,
  validatePageSize,
  validatePassword,
} from "./supabase-client";
import type {
  UserRow,
  AuthSession,
  SupabaseError,
  AuthError,
  RealtimeMessage,
} from "./supabase-client";

// ── Fixtures ────────────────────────────────────────────────

const VALID_UUID = "550e8400-e29b-41d4-a716-446655440000";

const MOCK_USER_ROW: Record<string, unknown> = {
  id: VALID_UUID,
  email: "alice@example.com",
  user_role: "editor",
  profile: { avatar_url: "https://example.com/avatar.jpg", bio: "Hello" },
  balance: "1000",
  metadata: { theme: "dark" },
  created_at: "2026-01-15T10:30:00.000Z",
  updated_at: "2026-02-01T14:00:00.000Z",
};

const MOCK_USER_NULL_PROFILE: Record<string, unknown> = {
  ...MOCK_USER_ROW,
  id: "660e8400-e29b-41d4-a716-446655440001",
  profile: null,
  metadata: null,
};

const MOCK_USER_BIGINT: Record<string, unknown> = {
  ...MOCK_USER_ROW,
  id: "770e8400-e29b-41d4-a716-446655440002",
  balance: "9007199254740993", // MAX_SAFE_INTEGER + 2
};

const MOCK_AUTH_RESPONSE = {
  id: VALID_UUID,
  aud: "authenticated",
  role: "authenticated",
  email: "new@example.com",
  phone: null,
  confirmation_sent_at: "2026-02-12T10:00:00.000Z",
  email_confirmed_at: null,
  created_at: "2026-02-12T10:00:00.000Z",
  updated_at: "2026-02-12T10:00:00.000Z",
  user_metadata: { display_name: "New User" },
  app_metadata: { provider: "email" },
  identities: [{ id: VALID_UUID, provider: "email" }],
  access_token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
  token_type: "bearer",
  expires_in: 3600,
  expires_at: Math.floor(Date.now() / 1000) + 3600,
  refresh_token: "refresh_abc123",
};

const MOCK_POSTGREST_ERROR_401: SupabaseError = {
  message: "JWT expired",
  details: null,
  hint: "Provide a valid JWT",
  code: "PGRST301",
};

const MOCK_POSTGREST_ERROR_403: SupabaseError = {
  message: "permission denied for table users",
  details: null,
  hint: null,
  code: "42501",
};

const MOCK_POSTGREST_ERROR_409: SupabaseError = {
  message: "duplicate key value violates unique constraint",
  details: "Key (email)=(alice@example.com) already exists.",
  hint: null,
  code: "23505",
};

const MOCK_POSTGREST_ERROR_416: SupabaseError = {
  message: "Requested range not satisfiable",
  details: null,
  hint: null,
  code: "PGRST103",
};

const MOCK_AUTH_ERROR_422: AuthError = {
  error: "invalid_grant",
  error_description: "Email address is not valid",
  msg: "Unable to validate email address: invalid format",
  code: 422,
};

const MOCK_AUTH_ERROR_429: AuthError = {
  msg: "For security purposes, you can only request this after 60 seconds.",
  code: 429,
};

// ── Helpers ─────────────────────────────────────────────────

function mockFetch(status: number, body: unknown, headers?: Record<string, string>): void {
  global.fetch = vi.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    json: vi.fn().mockResolvedValue(body),
    headers: new Map(Object.entries(headers ?? {})),
  } as unknown as Response);
}

function mockFetchSequence(
  ...responses: Array<{ status: number; body: unknown; headers?: Record<string, string> }>
): void {
  const fn = vi.fn();
  for (const [i, resp] of responses.entries()) {
    fn.mockResolvedValueOnce({
      ok: resp.status >= 200 && resp.status < 300,
      status: resp.status,
      json: vi.fn().mockResolvedValue(resp.body),
      headers: new Map(Object.entries(resp.headers ?? {})),
    } as unknown as Response);
  }
  global.fetch = fn;
}

function mockFetchNetworkError(message: string): void {
  global.fetch = vi.fn().mockRejectedValue(new Error(message));
}

function createClient(): SupabaseClient {
  return new SupabaseClient(
    "https://test-project.supabase.co",
    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSJ9.fake",
    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
    "refresh_token_abc",
    Math.floor(Date.now() / 1000) + 3600 // expires 1 hour from now
  );
}

// ── Tests ───────────────────────────────────────────────────

describe("SupabaseClient", () => {
  let client: SupabaseClient;

  beforeEach(() => {
    client = createClient();
    vi.restoreAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  // ── Validation Helpers ──────────────────────────────

  describe("validateUUIDv4", () => {
    it("accepts valid UUID v4", () => {
      expect(() => validateUUIDv4(VALID_UUID)).not.toThrow();
    });

    it("rejects UUID v1", () => {
      expect(() => validateUUIDv4("6ba7b810-9dad-11d1-80b4-00c04fd430c8")).toThrow("Invalid UUID v4");
    });

    it("rejects non-UUID strings", () => {
      expect(() => validateUUIDv4("not-a-uuid")).toThrow("Invalid UUID v4");
      expect(() => validateUUIDv4("")).toThrow("Invalid UUID v4");
      expect(() => validateUUIDv4("123")).toThrow("Invalid UUID v4");
    });
  });

  describe("validateEmail", () => {
    it("accepts valid emails", () => {
      expect(() => validateEmail("user@example.com")).not.toThrow();
    });

    it("rejects invalid emails", () => {
      expect(() => validateEmail("not-an-email")).toThrow("Invalid email");
      expect(() => validateEmail("")).toThrow("Invalid email");
    });
  });

  describe("validateRole", () => {
    it("accepts all valid roles", () => {
      for (const role of ["admin", "editor", "viewer", "guest"]) {
        expect(() => validateRole(role)).not.toThrow();
      }
    });

    it("rejects invalid roles", () => {
      expect(() => validateRole("superadmin")).toThrow("Invalid user_role");
      expect(() => validateRole("")).toThrow("Invalid user_role");
    });
  });

  describe("validatePageSize", () => {
    it("accepts valid page sizes", () => {
      expect(() => validatePageSize(1)).not.toThrow();
      expect(() => validatePageSize(25)).not.toThrow();
      expect(() => validatePageSize(1000)).not.toThrow();
    });

    it("rejects out-of-range sizes", () => {
      expect(() => validatePageSize(0)).toThrow("Page size must be between");
      expect(() => validatePageSize(1001)).toThrow("Page size must be between");
      expect(() => validatePageSize(-1)).toThrow("Page size must be between");
    });
  });

  describe("validatePassword", () => {
    it("accepts valid passwords", () => {
      expect(() => validatePassword("secure123")).not.toThrow();
      expect(() => validatePassword("123456")).not.toThrow();
    });

    it("rejects short passwords", () => {
      expect(() => validatePassword("12345")).toThrow("at least 6 characters");
      expect(() => validatePassword("")).toThrow("at least 6 characters");
    });
  });

  // ── Content-Range Parser ────────────────────────────

  describe("parseContentRange", () => {
    it("parses valid Content-Range header", () => {
      const result = parseContentRange("0-9/100");
      expect(result).toEqual({ start: 0, end: 9, total: 100 });
    });

    it("handles unknown total (*)", () => {
      const result = parseContentRange("0-24/*");
      expect(result).toEqual({ start: 0, end: 24, total: -1 });
    });

    it("returns null for missing header", () => {
      expect(parseContentRange(null)).toBeNull();
    });

    it("returns null for malformed header", () => {
      expect(parseContentRange("invalid")).toBeNull();
      expect(parseContentRange("abc-def/ghi")).toBeNull();
    });
  });

  // ── Row Parser ──────────────────────────────────────

  describe("parseUserRow", () => {
    it("parses a complete row with all fields", () => {
      const user = parseUserRow(MOCK_USER_ROW);
      expect(user.id).toBe(VALID_UUID);
      expect(user.email).toBe("alice@example.com");
      expect(user.user_role).toBe("editor");
      expect(user.profile).not.toBeNull();
      expect(user.profile!.avatar_url).toBe("https://example.com/avatar.jpg");
      expect(user.profile!.bio).toBe("Hello");
      expect(typeof user.balance).toBe("bigint");
      expect(user.balance).toBe(BigInt(1000));
      expect(user.metadata).toEqual({ theme: "dark" });
      expect(user.created_at).toBe("2026-01-15T10:30:00.000Z");
      expect(user.updated_at).toBe("2026-02-01T14:00:00.000Z");
    });

    it("handles null profile and metadata", () => {
      const user = parseUserRow(MOCK_USER_NULL_PROFILE);
      expect(user.profile).toBeNull();
      expect(user.metadata).toBeNull();
    });

    it("preserves bigint precision beyond MAX_SAFE_INTEGER", () => {
      const user = parseUserRow(MOCK_USER_BIGINT);
      expect(typeof user.balance).toBe("bigint");
      expect(user.balance).toBe(BigInt("9007199254740993"));
    });

    it("rejects invalid UUID", () => {
      expect(() => parseUserRow({ ...MOCK_USER_ROW, id: "not-uuid" })).toThrow(
        "Invalid UUID v4"
      );
    });

    it("rejects invalid user_role", () => {
      expect(() =>
        parseUserRow({ ...MOCK_USER_ROW, user_role: "superadmin" })
      ).toThrow("Invalid user_role");
    });

    it("rejects invalid timestamp", () => {
      expect(() =>
        parseUserRow({ ...MOCK_USER_ROW, created_at: "not-a-date" })
      ).toThrow("Invalid ISO 8601");
    });
  });

  // ── queryUsers (PostgREST GET) ──────────────────────

  describe("queryUsers", () => {
    it("returns parsed users on 200", async () => {
      mockFetch(200, [MOCK_USER_ROW], { "Content-Range": "0-0/1" });
      const result = await client.queryUsers();

      expect(result.data).toHaveLength(1);
      expect(result.data[0].id).toBe(VALID_UUID);
      expect(result.data[0].email).toBe("alice@example.com");
      expect(result.data[0].user_role).toBe("editor");
      expect(typeof result.data[0].balance).toBe("bigint");
      expect(result.totalCount).toBe(1);
    });

    it("handles 206 Partial Content for paginated responses", async () => {
      mockFetchSequence(
        {
          status: 206,
          body: [MOCK_USER_ROW],
          headers: { "Content-Range": "0-0/2" },
        },
        {
          status: 200,
          body: [MOCK_USER_NULL_PROFILE],
          headers: { "Content-Range": "1-1/2" },
        }
      );

      const result = await client.queryUsers(undefined, 1);
      expect(result.data).toHaveLength(2);
      expect(result.totalCount).toBe(2);
    });

    it("throws on 401 unauthorized", async () => {
      mockFetch(401, MOCK_POSTGREST_ERROR_401);
      await expect(client.queryUsers()).rejects.toThrow("Unauthorized");
    });

    it("throws on 403 RLS violation", async () => {
      mockFetch(403, MOCK_POSTGREST_ERROR_403);
      await expect(client.queryUsers()).rejects.toThrow("Forbidden (RLS)");
    });

    it("throws on 416 range not satisfiable", async () => {
      mockFetch(416, MOCK_POSTGREST_ERROR_416);
      await expect(client.queryUsers()).rejects.toThrow("Range not satisfiable");
    });

    it("throws on network error", async () => {
      mockFetchNetworkError("ECONNREFUSED");
      await expect(client.queryUsers()).rejects.toThrow(
        "Network error querying users: ECONNREFUSED"
      );
    });

    it("validates page size before calling API", async () => {
      await expect(client.queryUsers(undefined, 0)).rejects.toThrow(
        "Page size must be between"
      );
      await expect(client.queryUsers(undefined, 1001)).rejects.toThrow(
        "Page size must be between"
      );
    });

    it("validates role filter", async () => {
      await expect(
        client.queryUsers({ role: "superadmin" as never })
      ).rejects.toThrow("Invalid user_role");
    });

    it("handles users with null profile without crashing", async () => {
      mockFetch(200, [MOCK_USER_NULL_PROFILE], { "Content-Range": "0-0/1" });
      const result = await client.queryUsers();

      expect(result.data[0].profile).toBeNull();
      // Accessing nested fields does NOT crash
    });
  });

  // ── insertUser (PostgREST POST) ─────────────────────

  describe("insertUser", () => {
    it("inserts user and returns parsed row on 201", async () => {
      mockFetch(201, [MOCK_USER_ROW]);
      const result = await client.insertUser({
        email: "alice@example.com",
        user_role: "editor",
      });

      expect(result.ok).toBe(true);
      if (result.ok) {
        expect(result.data.id).toBe(VALID_UUID);
        expect(result.data.email).toBe("alice@example.com");
        expect(result.data.user_role).toBe("editor");
      }
    });

    it("returns error on 409 unique constraint", async () => {
      mockFetch(409, MOCK_POSTGREST_ERROR_409);
      const result = await client.insertUser({
        email: "alice@example.com",
        user_role: "editor",
      });

      expect(result.ok).toBe(false);
      if (!result.ok) {
        expect(result.status).toBe(409);
        expect(result.error.code).toBe("23505");
        expect(result.error.message).toContain("Unique constraint");
      }
    });

    it("returns error on 401", async () => {
      mockFetch(401, MOCK_POSTGREST_ERROR_401);
      const result = await client.insertUser({
        email: "test@example.com",
        user_role: "viewer",
      });

      expect(result.ok).toBe(false);
      if (!result.ok) {
        expect(result.status).toBe(401);
      }
    });

    it("returns error on 403 RLS violation", async () => {
      mockFetch(403, MOCK_POSTGREST_ERROR_403);
      const result = await client.insertUser({
        email: "test@example.com",
        user_role: "viewer",
      });

      expect(result.ok).toBe(false);
      if (!result.ok) {
        expect(result.status).toBe(403);
      }
    });

    it("validates email before calling API", async () => {
      await expect(
        client.insertUser({ email: "not-email", user_role: "viewer" })
      ).rejects.toThrow("Invalid email");
    });

    it("validates role before calling API", async () => {
      await expect(
        client.insertUser({
          email: "test@example.com",
          user_role: "superadmin" as never,
        })
      ).rejects.toThrow("Invalid user_role");
    });

    it("throws on network failure", async () => {
      mockFetchNetworkError("ETIMEDOUT");
      await expect(
        client.insertUser({ email: "test@example.com", user_role: "viewer" })
      ).rejects.toThrow("Network error inserting user: ETIMEDOUT");
    });
  });

  // ── signUp (Auth) ───────────────────────────────────

  describe("signUp", () => {
    it("returns session on successful sign-up", async () => {
      mockFetch(200, MOCK_AUTH_RESPONSE);
      const result = await client.signUp({
        email: "new@example.com",
        password: "secure123",
      });

      expect(result.ok).toBe(true);
      if (result.ok) {
        expect(result.session.user.id).toBe(VALID_UUID);
        expect(result.session.user.email).toBe("new@example.com");
        expect(result.session.token_type).toBe("bearer");
        expect(typeof result.session.access_token).toBe("string");
        expect(typeof result.session.expires_in).toBe("number");
        expect(typeof result.session.expires_at).toBe("number");
        expect(typeof result.session.refresh_token).toBe("string");
        expect(result.session.user.user_metadata).toEqual({ display_name: "New User" });
        expect(result.session.user.app_metadata).toEqual({ provider: "email" });
        expect(Array.isArray(result.session.user.identities)).toBe(true);
      }
    });

    it("handles null user_metadata", async () => {
      mockFetch(200, { ...MOCK_AUTH_RESPONSE, user_metadata: null });
      const result = await client.signUp({
        email: "new@example.com",
        password: "secure123",
      });

      expect(result.ok).toBe(true);
      if (result.ok) {
        expect(result.session.user.user_metadata).toBeNull();
      }
    });

    it("returns error on 422 (invalid email)", async () => {
      mockFetch(422, MOCK_AUTH_ERROR_422);
      const result = await client.signUp({
        email: "bad-email@",
        password: "secure123",
      });

      // Note: local validation may also catch this, but server can reject too
      expect(result.ok).toBe(false);
      if (!result.ok) {
        expect(result.status).toBe(422);
      }
    });

    it("returns error on 429 (rate limit)", async () => {
      mockFetch(429, MOCK_AUTH_ERROR_429);
      const result = await client.signUp({
        email: "new@example.com",
        password: "secure123",
      });

      expect(result.ok).toBe(false);
      if (!result.ok) {
        expect(result.status).toBe(429);
        expect(result.error.msg).toContain("Rate limit");
      }
    });

    it("validates email before calling API", async () => {
      await expect(
        client.signUp({ email: "invalid", password: "secure123" })
      ).rejects.toThrow("Invalid email");
    });

    it("validates password length before calling API", async () => {
      await expect(
        client.signUp({ email: "test@example.com", password: "12345" })
      ).rejects.toThrow("at least 6 characters");
    });

    it("throws on network failure", async () => {
      mockFetchNetworkError("DNS_RESOLUTION_FAILED");
      await expect(
        client.signUp({ email: "test@example.com", password: "secure123" })
      ).rejects.toThrow("Network error during sign-up");
    });
  });

  // ── Token Refresh ───────────────────────────────────

  describe("token refresh", () => {
    it("refreshes token before API call when expired", async () => {
      // Create client with expired token
      const expiredClient = new SupabaseClient(
        "https://test-project.supabase.co",
        "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSJ9.fake",
        "expired-access-token",
        "valid-refresh-token",
        Math.floor(Date.now() / 1000) - 100 // expired 100 seconds ago
      );

      mockFetchSequence(
        {
          status: 200,
          body: {
            access_token: "new-access-token",
            refresh_token: "new-refresh-token",
            expires_at: Math.floor(Date.now() / 1000) + 3600,
          },
        },
        {
          status: 200,
          body: [MOCK_USER_ROW],
          headers: { "Content-Range": "0-0/1" },
        }
      );

      const result = await expiredClient.queryUsers();
      expect(result.data).toHaveLength(1);

      // Verify refresh was called first
      const calls = (global.fetch as ReturnType<typeof vi.fn>).mock.calls;
      expect(calls[0][0]).toContain("/auth/v1/token?grant_type=refresh_token");
      expect(calls[1][0]).toContain("/rest/v1/users");
    });
  });

  // ── isTokenExpired ──────────────────────────────────

  describe("isTokenExpired", () => {
    it("returns false for future expiry", () => {
      const future = Math.floor(Date.now() / 1000) + 3600;
      expect(isTokenExpired(future)).toBe(false);
    });

    it("returns true for past expiry", () => {
      const past = Math.floor(Date.now() / 1000) - 100;
      expect(isTokenExpired(past)).toBe(true);
    });

    it("returns true within 60-second grace period", () => {
      const almostExpired = Math.floor(Date.now() / 1000) + 30;
      expect(isTokenExpired(almostExpired)).toBe(true);
    });
  });
});
```

### Expected Stricture Result

```
PASS  0 violations
```

Stricture must produce zero violations. Every status code (200, 201, 206, 400, 401, 403, 409, 416, 422, 429, 500) has test coverage, all error paths are exercised, all enum values are handled, all assertions check specific field values, nullable fields are tested for null, UUIDs are validated against v4 regex, Content-Range headers are parsed for pagination, BigInt is used for PostgreSQL bigint columns, JWT tokens are refreshed before expiry, and realtime subscription handlers are registered before the join message to prevent message loss.

---

## B01 -- No Error Handling

### Implementation

```typescript
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
```

### Expected Violation

```
TQ-error-path-coverage: fetch() call at queryUsers has no error handling.
  Network errors (ECONNREFUSED, ETIMEDOUT, DNS failures) will crash the
  caller with an unhandled promise rejection. All external HTTP calls must
  be wrapped in try/catch or .catch() with meaningful error propagation.
  Locations: queryUsers, signUp
```

### What Makes This Obvious

Any linter or code reviewer catches bare `await fetch()` without error handling. In production, transient network failures are common -- a DNS blip, a load balancer timeout, or Supabase itself being briefly unreachable -- and without a try/catch boundary, the error propagates as an unhandled rejection, crashing the Node.js process or leaving the caller with an opaque stack trace instead of a recoverable error.

---

## B02 -- No Status Code Check

### Implementation

```typescript
// B02: Ignores HTTP status codes; treats all responses as success.

class SupabaseClientB02 {
  private readonly baseUrl: string;
  private readonly anonKey: string;
  private readonly jwt: string;

  constructor(baseUrl: string, anonKey: string, jwt: string) {
    this.baseUrl = baseUrl;
    this.anonKey = anonKey;
    this.jwt = jwt;
  }

  async queryUsers(): Promise<Record<string, unknown>[]> {
    try {
      const response = await fetch(`${this.baseUrl}/rest/v1/users?select=*`, {
        headers: {
          apikey: this.anonKey,
          Authorization: `Bearer ${this.jwt}`,
        },
      });

      // BUG: No check on response.ok or response.status.
      // A 401 (expired JWT), 403 (RLS violation), or 416 (bad range)
      // response is parsed as if it were a user array.
      // Supabase returns { message, details, hint, code } on errors,
      // which gets cast to Record<string, unknown>[] -- a single-element
      // "array" that is actually an error object.
      return (await response.json()) as Record<string, unknown>[];
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
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

      // BUG: Does not check for 201 vs 409 (duplicate) vs 400.
      // On 409, the response body is an error object, not a user row array.
      const rows = (await response.json()) as Record<string, unknown>[];
      return rows[0];
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }
  }
}
```

### Expected Violation

```
CTR-status-code-handling: Client does not check response status.
  Manifest declares status_codes [200, 206, 400, 401, 403, 406, 416] for
  GET and [201, 400, 401, 403, 409, 500] for POST but client treats all
  responses as success. A 401 error response { message, details, hint, code }
  will be misinterpreted as a user row array.
  Locations: queryUsers, insertUser
```

### What Makes This Obvious

The manifest declares multiple error status codes for each endpoint. Without checking `response.ok` or `response.status`, the client casts the error JSON `{ message, details, hint, code }` into a user array. Downstream code that iterates over users will process the error object as if it were a user row, with `user.email` being `undefined` and `user.id` being `undefined`, producing corrupted results instead of surfacing the actual error.

---

## B03 -- Shallow Test Assertions

### Implementation

```typescript
// B03: Implementation is correct (same as PERFECT).
// The bug is entirely in the tests.

// (Same correct implementation as PERFECT -- omitted for brevity)
```

### Tests (THE BUG IS HERE)

```typescript
// B03: Tests only check existence, never shape or value.

describe("SupabaseClientB03", () => {
  it("queries users", async () => {
    mockFetch(200, [MOCK_USER_ROW], { "Content-Range": "0-0/1" });
    const client = createClient();
    const result = await client.queryUsers();

    // BUG: These assertions prove nothing about correctness.
    // They pass even if result is { data: [{ id: 999, email: null }] }.
    expect(result).toBeDefined();
    expect(result.data).toBeDefined();
    expect(result.data.length).toBeTruthy();
    if (result.data.length > 0) {
      expect(result.data[0]).toBeDefined();
      expect(result.data[0].id).toBeDefined();
      expect(result.data[0].email).toBeDefined();
      expect(result.data[0].balance).toBeDefined();
      expect(result.data[0].user_role).toBeDefined();
    }
  });

  it("signs up user", async () => {
    mockFetch(200, MOCK_AUTH_RESPONSE);
    const client = createClient();
    const result = await client.signUp({
      email: "new@example.com",
      password: "secure123",
    });

    // BUG: Only checks that session exists, not its shape or values.
    expect(result).toBeDefined();
    expect(result.ok).toBeTruthy();
    if (result.ok) {
      expect(result.session).toBeDefined();
      expect(result.session.user).toBeDefined();
      expect(result.session.access_token).toBeDefined();
    }
  });

  it("handles errors", async () => {
    mockFetch(401, MOCK_POSTGREST_ERROR_401);
    const client = createClient();
    // BUG: Only checks that it throws, not what it throws.
    await expect(client.queryUsers()).rejects.toBeDefined();
  });
});
```

### Expected Violation

```
TQ-no-shallow-assertions: 11 assertions use toBeDefined()/toBeTruthy()
  instead of checking specific values or types. Assertions like
  expect(result.data[0].id).toBeDefined() pass for id=0, id="", id=false.
  Must use toBe(), toEqual(), toMatch(), or toStrictEqual() to verify
  the actual contract shape.
  Locations: test:9-15, test:25-28, test:33
```

### What Makes This Obvious

`expect(x).toBeDefined()` only asserts `x !== undefined`. If the API response has `{ id: 0 }` instead of `{ id: "550e8400-..." }`, these tests still pass. They provide zero contract validation: no UUID format check, no bigint type check, no role enum check, no timestamp format check. The tests give false confidence while the implementation could return any shape at all.

---

## B04 -- Missing Negative Tests

### Implementation

```typescript
// B04: Implementation is correct (same as PERFECT).
// The bug is in the test file: only happy-path tests.

// (Same correct implementation as PERFECT -- omitted for brevity)
```

### Tests (THE BUG IS HERE)

```typescript
// B04: Only tests the successful path. No error, edge, or failure tests.

describe("SupabaseClientB04", () => {
  it("queries users successfully", async () => {
    mockFetch(200, [MOCK_USER_ROW], { "Content-Range": "0-0/1" });
    const client = createClient();
    const result = await client.queryUsers();
    expect(result.data).toHaveLength(1);
    expect(result.data[0].id).toBe(VALID_UUID);
  });

  it("inserts a user successfully", async () => {
    mockFetch(201, [MOCK_USER_ROW]);
    const client = createClient();
    const result = await client.insertUser({
      email: "alice@example.com",
      user_role: "editor",
    });
    expect(result.ok).toBe(true);
  });

  it("signs up successfully", async () => {
    mockFetch(200, MOCK_AUTH_RESPONSE);
    const client = createClient();
    const result = await client.signUp({
      email: "new@example.com",
      password: "secure123",
    });
    expect(result.ok).toBe(true);
  });

  // BUG: No tests for:
  //   - 401 (expired JWT)
  //   - 403 (RLS violation)
  //   - 409 (unique constraint violation)
  //   - 416 (range not satisfiable)
  //   - 422 (invalid email in auth)
  //   - 429 (rate limit)
  //   - Network failures (ECONNREFUSED, ETIMEDOUT)
  //   - Invalid UUID format input
  //   - Invalid user_role enum value
  //   - Null profile JSONB handling
  //   - Null user_metadata from auth
  //   - BigInt precision for balance
  //   - Content-Range parsing edge cases
  //   - Token refresh on expired JWT
  //   - Realtime subscription errors
});
```

### Expected Violation

```
TQ-negative-cases: All 3 methods have 0 negative tests.
  Manifest declares 7 non-success status codes across endpoints and
  multiple validation constraints. Functions with external dependencies
  require tests for each failure mode.
  Missing: 401 test, 403 test, 409 test, 416 test, 422 test, 429 test,
           network error test, invalid UUID test, invalid email test,
           invalid role test, null JSONB test, bigint precision test.
  Functions without negative tests: queryUsers, insertUser, signUp,
  subscribeToChanges
```

### What Makes This Obvious

A test suite that only verifies success says nothing about what happens when things go wrong. Supabase calls can fail in many distinct ways (network error, expired JWT, RLS violation, unique constraint, rate limit), and without negative tests, regressions in error handling go undetected. If someone removes the status check or the Content-Range parser, these tests still pass.

---

## B05 -- Request Missing Required Fields

### Implementation

```typescript
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
```

### Expected Violation

```
CTR-request-shape: POST /auth/v1/signup request is missing required field
  `email`. Manifest declares email as { type: string, format: email,
  required: true } but the client never includes it in the request body.
  POST /rest/v1/users request is also missing required field `email`.
  Every call will receive a 400 or 422 response.
  Locations: signUp (missing: email), insertUser (missing: email)
```

### What Makes This Obvious

The manifest explicitly declares `email: { required: true }` on both the auth sign-up and PostgREST insert endpoints. The implementation's function signatures omit email entirely. Every single call will fail -- sign-up requires an email to send the confirmation link, and the database column has a NOT NULL constraint. This is a structural mismatch between what the client sends and what the server requires.

---

## B06 -- Response Shape Mismatch

### Implementation

```typescript
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
```

### Expected Violation

```
CTR-response-shape: Client type PostgRESTResponseB06 expects wrapped
  object { data: UserB06[], count: number } but manifest declares
  response type as bare array. PostgREST returns [...] not { data: [...] }.
  Additionally, client type UserB06 has field "role" but manifest and
  PostgREST response use "user_role". Two shape mismatches:
  1. Wrapper mismatch: { data: [...] } vs bare [...]
  2. Field name mismatch: "role" vs "user_role"
  Locations: PostgRESTResponseB06, UserB06.role
```

### What Makes This Obvious

PostgREST's defining characteristic is returning bare arrays, not wrapped objects. This is documented extensively and is a common migration pitfall for developers coming from REST APIs like Stripe that use `{ data: [...] }` wrappers. The double bug (wrapper + field name) means the code returns `undefined` from `queryUsers()` and then crashes on the next operation. The `user_role` vs `role` mismatch is a second, independent contract violation that would cause role-based access control to fail silently.

---

## B07 -- Wrong Field Types

### Implementation

```typescript
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
```

### Expected Violation

```
CTR-manifest-conformance: Field `id` declared as { type: string, format:
  uuid_v4 } in manifest but typed as `number` in client. parseInt on
  UUID produces corrupted IDs (550 instead of "550e8400-...").
  Field `balance` declared as { type: integer, range: [0, 9223372036854775807] }
  but typed as `number` -- exceeds Number.MAX_SAFE_INTEGER.
  Field `count` from Content-Range should be parsed as integer but kept
  as string, causing string concatenation in arithmetic.
  Field `profile.avatar_url` and `profile.bio` declared nullable but
  typed as non-null string.
  Locations: UserB07.id, UserB07.balance, UserB07.profile, count return type
```

### What Makes This Obvious

The manifest says `id: { type: string, format: uuid_v4 }` and `balance: { type: integer, range: [0, 9223372036854775807] }`. The client types `id` as `number` and uses `parseInt`, which stops parsing at the first non-numeric character in a UUID, producing `550` instead of `"550e8400-e29b-41d4-a716-446655440000"`. The `balance` as `number` silently truncates values above `Number.MAX_SAFE_INTEGER`. The Content-Range count kept as a string causes `count + 1` to produce `"1001"` instead of `101`.

---

## B08 -- Incomplete Enum Handling

### Implementation

```typescript
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
```

### Expected Violation

```
CTR-strictness-parity: user_role enum has 4 values in manifest
  ["admin", "editor", "viewer", "guest"] but getPermissions only
  handles 2 of 4. Missing: "editor", "guest".
CTR-status-code-handling: Incomplete status code handling.
  GET response: manifest declares 200 and 206 but only 200 is handled.
  206 (Partial Content) is treated as an error.
  POST response: only 201 handled, missing 409, 403, 401 specific handling.
  Locations: getPermissions (missing cases: "editor", "guest"),
             queryUsersPage (missing: 206), insertUser (missing: 409, 403, 401)
```

### What Makes This Obvious

The manifest declares four valid roles. The switch statement only handles two. The "editor" role is likely the most common non-admin role, and it throws an unexpected error. Similarly, PostgREST returns 206 Partial Content for paginated results (when the Range header selects a subset), which is a success status code. Treating 206 as an error breaks all pagination.

---

## B09 -- Missing Range Validation

### Implementation

```typescript
// B09: No validation on Range header bounds or page size.

class SupabaseClientB09 {
  private readonly baseUrl: string;
  private readonly anonKey: string;
  private readonly jwt: string;

  constructor(baseUrl: string, anonKey: string, jwt: string) {
    this.baseUrl = baseUrl;
    this.anonKey = anonKey;
    this.jwt = jwt;
  }

  // BUG: No validation that pageSize does not exceed PostgREST max-rows
  // (default 1000). Also no validation on offset being non-negative.
  async queryUsers(pageSize: number = 10000): Promise<Record<string, unknown>[]> {
    try {
      const response = await fetch(
        `${this.baseUrl}/rest/v1/users?select=*`,
        {
          headers: {
            apikey: this.anonKey,
            Authorization: `Bearer ${this.jwt}`,
            Range: `0-${pageSize - 1}`, // BUG: range(0, 9999) -- way beyond server max
          },
        }
      );

      if (!response.ok) throw new Error(`Query failed: ${response.status}`);
      return (await response.json()) as Record<string, unknown>[];
    } catch (err) {
      throw err;
    }
  }

  // BUG: Negative offset not validated. range(-1, 10) is nonsensical.
  async queryUsersOffset(offset: number, limit: number): Promise<Record<string, unknown>[]> {
    try {
      const response = await fetch(
        `${this.baseUrl}/rest/v1/users?select=*`,
        {
          headers: {
            apikey: this.anonKey,
            Authorization: `Bearer ${this.jwt}`,
            Range: `${offset}-${offset + limit - 1}`, // No bounds check
          },
        }
      );

      if (!response.ok) throw new Error(`Query failed: ${response.status}`);
      return (await response.json()) as Record<string, unknown>[];
    } catch (err) {
      throw err;
    }
  }

  // BUG: Uses Number.MAX_SAFE_INTEGER as page size.
  async getAllUsers(): Promise<Record<string, unknown>[]> {
    return this.queryUsers(Number.MAX_SAFE_INTEGER);
  }
}
```

### Expected Violation

```
CTR-strictness-parity: No range validation on page size or offset parameters.
  PostgREST max-rows default is 1000. The default pageSize of 10000 exceeds
  this limit, causing silent data truncation (server caps to 1000 without error).
  Negative offsets produce undefined behavior. Number.MAX_SAFE_INTEGER as page
  size creates a nonsensical Range header.
  Locations: queryUsers (default pageSize: 10000 > max 1000),
             queryUsersOffset (no bounds check on offset),
             getAllUsers (Number.MAX_SAFE_INTEGER)
```

### What Makes This Obvious

The manifest and PostgREST configuration enforce a maximum row limit (default 1000). The code sends `Range: 0-9999` which exceeds this limit. PostgREST silently caps the response to 1000 rows, meaning the caller believes it requested 10000 rows but received only 1000 -- and has no way to know the truncation occurred. This is a data loss bug disguised as a default parameter.

---

## B10 -- Format Not Validated

### Implementation

```typescript
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
```

### Expected Violation

```
CTR-strictness-parity: Manifest declares format constraints that client
  does not enforce:
  - getUser param `userId`: format uuid_v4 not validated
  - signUp param `email`: format email not validated
  - constructor param `anonKey`: format "eyJ*" not validated
  - constructor param `jwt`: format JWT not validated
  Accepts arbitrary strings, which cause PostgreSQL cast errors (22P02)
  or GoTrue validation errors instead of clear local validation messages.
  Locations: getUser:userId, signUp:email, constructor:anonKey, constructor:jwt
```

### What Makes This Obvious

Supabase uses PostgreSQL's strict UUID type for IDs. Passing "hello" as a user ID generates a database cast error (`22P02: invalid input syntax for type uuid`) that surfaces as a cryptic server-side error. Local format validation catches these immediately with a clear message. The JWT format validation prevents spending a network round trip only to get a 401 back because the token is malformed.

---

## B11 -- Precision/Conversion Issue

### Implementation

```typescript
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
```

### Expected Violation

```
CTR-strictness-parity: Field `balance` declared as { type: integer,
  range: [0, 9223372036854775807] } but converted via Number(), causing
  precision loss above 2^53.
  Field `created_at` and `updated_at` declared as { format: iso8601_timestamptz }
  but converted to Date objects, discarding the original timezone offset.
  The manifest format is ISO 8601 with timezone; Date objects normalize to
  UTC and lose the original offset context.
  Locations: getUser:balance (Number()), getUser:created_at (new Date()),
             getUser:updated_at (new Date())
```

### What Makes This Obvious

The manifest declares timestamps as `iso8601_timestamptz` -- the "tz" suffix means timezone information is significant. Converting to `Date` discards the original timezone offset. For audit logging, regulatory compliance, or multi-timezone applications, knowing that an event occurred at "10:00 Tokyo time" versus "01:00 UTC" matters, even though they represent the same instant. The balance precision loss is the same BigInt issue as B07 but manifested differently -- here the code uses `Number()` explicitly during response parsing.

---

## B12 -- Nullable Field Crash

### Implementation

```typescript
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
```

### Expected Violation

```
CTR-response-shape: Accessing properties on nullable fields without
  null check. Manifest declares these fields as nullable:
  - user_metadata: { nullable: true } -- .display_name crashes when null
  - profile: { nullable: true } -- .avatar_url crashes when null
  - metadata: { nullable: true } -- Object.keys() crashes when null
  On sign-up with email confirmation pending, user_metadata can be null.
  On a user without a profile, profile column is NULL.
  Locations: signUp (metadata.display_name), getUserAvatar (profile.avatar_url),
             getUserMetadataKeys (Object.keys(metadata))
```

### What Makes This Obvious

The manifest marks `user_metadata`, `profile`, and `metadata` as `nullable: true`. The code casts them to non-nullable objects and immediately accesses properties. This crashes on any user who has not set a profile (profile is NULL in the database), any sign-up that requires email confirmation (user_metadata may be null before confirmation), or any user without custom metadata. These are common states, not edge cases -- a brand-new user has null profile and null metadata by default.

---

## B13 -- Missing RLS Policy Verification

### Implementation

```typescript
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
```

### Expected Violation

```
CTR-request-shape: Manifest declares Authorization header as
  { type: string, format: "Bearer <jwt>", required: true } but the
  constructor does not accept or set a JWT token. All requests use
  only the anon key, bypassing user-scoped RLS policies.
  The code cannot distinguish between "row does not exist" and
  "row exists but is invisible under current RLS policy."
  Locations: constructor (missing: jwt parameter),
             countAllUsers (missing: Authorization header),
             getUser (missing: Authorization header)
```

### What Makes This Obvious

Row-Level Security is Supabase's core security model. Every table query returns different results depending on the JWT claims of the authenticated user. Without a JWT in the Authorization header, all queries execute as the `anon` role, which typically has severely restricted access. An admin dashboard calling `countAllUsers()` with only the anon key would show "0 users" when there are thousands, because the anon role cannot see them. This is a security-critical bug: the application appears to work but operates on a fundamentally restricted view of the data.

---

## B14 -- Pagination via Range Headers -- Missing Content-Range Parsing

### Implementation

```typescript
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
```

### Expected Violation

```
CTR-response-shape: Response pagination is incomplete. The manifest
  declares Content-Range as a response header with format "^\\d+-\\d+/\\d+$"
  indicating paginated results. The code:
  1. listAllUsers: Does not read Content-Range at all. Returns max 25 rows.
  2. listUsersWithCount: Reads total from Content-Range but does not
     fetch subsequent pages. Also does not handle "*" total.
  3. exportAllUserEmails: Exports 25 of 10,000 users silently.
  Locations: listAllUsers (no Content-Range parsing, no pagination loop),
             listUsersWithCount (no pagination loop, NaN on "*" total)
```

### What Makes This Obvious

PostgREST uses Range headers for pagination (like HTTP byte ranges but for rows). The Content-Range response header `0-24/500` tells the client: "You received rows 0-24 of 500 total." The PERFECT implementation uses this to drive a pagination loop. This buggy version either ignores the header entirely or reads the total without acting on it. The function named `listAllUsers` that returns 25 of 500 rows is a data integrity violation. In a CSV export scenario, the file would be 95% incomplete.

---

## B15 -- Race Condition in Realtime Subscription

### Implementation

```typescript
// B15: Realtime subscription handler registered AFTER join message sent.
// Messages that arrive between WebSocket open and handler registration are lost.

class SupabaseClientB15 {
  private readonly baseUrl: string;
  private readonly anonKey: string;

  constructor(baseUrl: string, anonKey: string) {
    this.baseUrl = baseUrl;
    this.anonKey = anonKey;
  }

  subscribeToChanges(
    table: string,
    callback: (message: {
      type: string;
      record: Record<string, unknown> | null;
    }) => void
  ): { unsubscribe: () => void } {
    const wsUrl = this.baseUrl
      .replace("https://", "wss://")
      .replace("http://", "ws://");

    const ws = new WebSocket(
      `${wsUrl}/realtime/v1/websocket?apikey=${this.anonKey}&vsn=1.0.0`
    );

    const topic = `realtime:public:${table}`;

    ws.onopen = () => {
      // Step 1: Send join message
      ws.send(JSON.stringify({
        topic,
        event: "phx_join",
        payload: {
          config: {
            broadcast: { self: false },
            presence: {},
            postgres_changes: [
              { event: "*", schema: "public", table },
            ],
          },
        },
        ref: "1",
      }));

      // BUG: Message handler is registered AFTER the join is sent.
      // In a fast server scenario, the phx_reply (join ack) and
      // postgres_changes events can arrive BEFORE onmessage is set.
      //
      // Timeline:
      //   T0: ws.send(join) -- join request sent
      //   T1: Server processes join, queues phx_reply + buffered changes
      //   T2: Server sends phx_reply
      //   T3: Server sends INSERT notification (another client inserted a row)
      //   T4: ws.onmessage = handler -- TOO LATE, T2 and T3 are lost
      //
      // This is especially bad when:
      // (a) The table has a trigger that fires on subscription
      // (b) A batch insert happens concurrently with subscription setup
      // (c) The server has low latency and responds before JS event loop yields
    };

    // BUG: Handler registered inside onopen, after send().
    // Should be registered BEFORE open or at minimum before send().
    ws.onopen = () => {
      // This overwrites the previous onopen handler!
      // BUG: The join message from the first onopen is now lost because
      // this second assignment replaces it entirely.
    };

    // Even if we fix the double-assignment, the handler is still too late:
    setTimeout(() => {
      ws.onmessage = (event: MessageEvent) => {
        let msg: Record<string, unknown>;
        try {
          msg = JSON.parse(String(event.data)) as Record<string, unknown>;
        } catch {
          return;
        }

        if (String(msg.event) === "postgres_changes") {
          const payload = msg.payload as Record<string, unknown>;
          const data = payload?.data as Record<string, unknown>;
          if (!data) return;

          callback({
            type: String(data.type),
            record: (data.record as Record<string, unknown> | null) ?? null,
          });
        }
      };
    }, 0); // setTimeout(fn, 0) defers to next microtask -- too late

    // BUG: No heartbeat. Phoenix Channels disconnect after missing
    // heartbeats (default timeout ~30-60 seconds). Without a heartbeat
    // interval, the connection silently drops and no more events arrive.
    // The caller has no way to know the subscription is dead.

    // BUG: No error handler on ws. WebSocket errors are silently swallowed.

    return {
      unsubscribe: () => {
        // BUG: Does not send phx_leave before closing.
        // The server does not know the client unsubscribed and may
        // continue buffering events.
        ws.close();
      },
    };
  }
}

// Test -- does not test race condition
import { describe, it, expect, vi } from "vitest";

describe("subscribeToChanges", () => {
  it("subscribes to a table", () => {
    const client = new SupabaseClientB15("https://test.supabase.co", "anon-key");
    const callback = vi.fn();

    // BUG: This test creates the subscription but never simulates
    // messages arriving before the handler is registered.
    // It only verifies the subscription object is returned.
    const sub = client.subscribeToChanges("users", callback);
    expect(sub).toBeDefined();
    expect(sub.unsubscribe).toBeDefined();

    // BUG: Never tests:
    // - Messages arriving during join handshake
    // - Connection errors
    // - Heartbeat timeout / silent disconnect
    // - phx_leave on unsubscribe
    // - Multiple concurrent subscriptions
  });
});
```

### Expected Violation

```
CTR-request-shape: Realtime subscription has multiple race conditions
  and protocol violations:
  1. Message handler (onmessage) registered after join message sent via
     setTimeout(fn, 0). Messages arriving between send() and handler
     registration are permanently lost.
  2. No heartbeat sent to Phoenix Channels. Connection will silently
     disconnect after server timeout (~30-60 seconds).
  3. onopen handler overwritten by second assignment, losing the join message.
  4. unsubscribe() does not send phx_leave, causing server-side resource leak.
  5. No onerror handler -- WebSocket errors are silently swallowed.
  Additionally, TQ-negative-cases: No test for messages arriving during
  join handshake, connection errors, or heartbeat timeout.
  Locations: subscribeToChanges (race: onmessage after send,
             missing: heartbeat, missing: phx_leave, missing: onerror)
```

### What Makes This Obvious

The Phoenix Channels protocol requires specific message ordering: the client must be ready to receive messages before sending the join request, because the server can respond immediately. Registering the message handler after `send()` -- especially in a `setTimeout` -- creates a window where messages are delivered to the default (no-op) handler and silently dropped. The PERFECT implementation registers `ws.onmessage` before `ws.onopen` fires, and buffers messages received before the join is acknowledged. The missing heartbeat is a second critical issue: Phoenix Channels require periodic heartbeat messages to keep the connection alive. Without them, the server closes the connection and the client has no notification that events are no longer being delivered.

---

## Summary Table

| Case | Rule Violated | Bug Description | Severity |
|------|--------------|-----------------|----------|
| PERFECT | (none) | Full RLS, pagination, UUID, null JSONB, BigInt, JWT refresh, realtime | N/A |
| B01 | TQ-error-path-coverage | No try/catch on any API call | Obvious |
| B02 | CTR-status-code-handling | Ignores response status, treats errors as success | Obvious |
| B03 | TQ-no-shallow-assertions | toBeDefined/toBeTruthy only, no value checks | Moderate |
| B04 | TQ-negative-cases | No 401/403/409/416/422/429 tests | Moderate |
| B05 | CTR-request-shape | Sign-up and insert missing required email field | Structural |
| B06 | CTR-response-shape | Expects { data: [...] } wrapper; field name "role" vs "user_role" | Structural |
| B07 | CTR-manifest-conformance | UUID as number, bigint as number, Content-Range count as string | Structural |
| B08 | CTR-strictness-parity | Only 2 of 4 roles handled; 206 treated as error | Precision |
| B09 | CTR-strictness-parity | Page size exceeds PostgREST max-rows, no bounds validation | Precision |
| B10 | CTR-strictness-parity | UUID, email, JWT format not validated client-side | Precision |
| B11 | CTR-strictness-parity | BigInt precision loss via Number(); timezone loss via Date() | Subtle |
| B12 | CTR-response-shape | Nullable user_metadata, profile, metadata accessed without null check | Subtle |
| B13 | CTR-request-shape | No JWT/RLS awareness, anon key only, cannot distinguish row absence from RLS filter | Subtle |
| B14 | CTR-response-shape | Content-Range header not parsed; single page returned as "all" | Subtle |
| B15 | CTR-request-shape | Realtime handler registered after join; no heartbeat; messages lost in race window | Subtle |
