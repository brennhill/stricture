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
