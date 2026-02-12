# 08 — Slack Web API

**Why included:** Slack's unconventional HTTP contract (always returns 200, success/failure indicated in JSON `ok` field) is a trap for standard HTTP-based error handling. Tests cursor-based pagination, string-encoded timestamps that serve as unique message IDs, rate limiting as the sole non-200 status code, and channel/user ID format validation.

**Key contract properties:**
- HTTP 200 on ALL responses (success AND error), except 429 (rate limit)
- Success/failure determined by `ok: boolean` in response body
- Auth: Bearer token (`xoxb-*` for bot, `xoxp-*` for user)
- Message timestamp (`ts`): string decimal `"1234567890.123456"` used as message ID
- Channel ID format: `C*` (public), `G*` (private/group), `D*` (DM)
- User ID format: `U*` or `W*` (enterprise grid)
- Cursor pagination via `response_metadata.next_cursor` (empty string = no more pages)
- Rate limit tiers 1-4, `Retry-After` header on 429
- Text limit: 40,000 characters per message
- Blocks: optional array of layout blocks (`section`, `divider`, `actions`, etc.)

---

## Manifest Fragment

```yaml
contracts:
  - id: "slack-web-api"
    producer: slack
    consumers: [my-service]
    protocol: http
    auth:
      type: bearer
      format: "xoxb-*|xoxp-*"
    # CRITICAL: Slack always returns HTTP 200, even on errors.
    # The ok field in the JSON body determines success/failure.
    # 429 is the ONLY non-200 status code Slack returns.
    status_codes: [200, 429]
    error_model: body_field  # errors are in the JSON body, not HTTP status
    endpoints:

      - path: "/api/chat.postMessage"
        method: POST
        request:
          content_type: application/json
          fields:
            channel:  { type: string, format: "C*|G*|D*", required: true }
            text:     { type: string, maxLength: 40000, required: false }
            blocks:   { type: array, items: { type: object }, required: false }
            thread_ts: { type: string, format: "decimal_timestamp", required: false }
            # At least one of text or blocks is required (server-enforced)
        response:
          wrapper: { ok: boolean, error?: string }
          fields:
            ok:       { type: boolean, required: true }
            channel:  { type: string, format: "C*|G*|D*", required: true }
            ts:       { type: string, format: "decimal_timestamp", required: true }
            message:  { type: object, required: true, fields: {
              type: { type: string, value: "message" },
              text: { type: string },
              ts:   { type: string, format: "decimal_timestamp" },
              user: { type: string, format: "U*|W*" },
              blocks: { type: array, items: { type: object }, required: false }
            }}
            error:    { type: string, required: false }
          error_values: ["channel_not_found", "not_in_channel", "is_archived",
                         "msg_too_long", "no_text", "invalid_auth", "token_revoked",
                         "ratelimited", "missing_scope", "not_authed",
                         "invalid_blocks", "invalid_blocks_format"]

      - path: "/api/conversations.list"
        method: GET
        request:
          fields:
            types:    { type: string, required: false }
            limit:    { type: integer, range: [1, 1000], required: false }
            cursor:   { type: string, required: false }
        response:
          wrapper: { ok: boolean, error?: string }
          fields:
            ok:       { type: boolean, required: true }
            channels: { type: array, required: true, items: { type: object, fields: {
              id:         { type: string, format: "C*|G*|D*" },
              name:       { type: string },
              is_channel: { type: boolean },
              is_group:   { type: boolean },
              is_im:      { type: boolean },
              is_archived: { type: boolean },
              num_members: { type: integer }
            }}}
            response_metadata: { type: object, required: false, fields: {
              next_cursor: { type: string }
            }}
            error:    { type: string, required: false }

      - path: "/api/conversations.history"
        method: POST
        request:
          fields:
            channel:  { type: string, format: "C*|G*|D*", required: true }
            limit:    { type: integer, range: [1, 1000], required: false }
            cursor:   { type: string, required: false }
            oldest:   { type: string, format: "decimal_timestamp", required: false }
            latest:   { type: string, format: "decimal_timestamp", required: false }
        response:
          wrapper: { ok: boolean, error?: string }
          fields:
            ok:       { type: boolean, required: true }
            messages: { type: array, required: true, items: { type: object, fields: {
              type: { type: string, value: "message" },
              user: { type: string, format: "U*|W*" },
              text: { type: string },
              ts:   { type: string, format: "decimal_timestamp" },
              blocks: { type: array, items: { type: object }, required: false }
            }}}
            has_more: { type: boolean, required: true }
            response_metadata: { type: object, required: false, fields: {
              next_cursor: { type: string }
            }}
            error:    { type: string, required: false }

      - path: "/api/users.info"
        method: POST
        request:
          fields:
            user:     { type: string, format: "U*|W*", required: true }
        response:
          wrapper: { ok: boolean, error?: string }
          fields:
            ok:       { type: boolean, required: true }
            user:     { type: object, required: true, fields: {
              id:        { type: string, format: "U*|W*" },
              name:      { type: string },
              real_name: { type: string },
              is_bot:    { type: boolean },
              is_admin:  { type: boolean },
              profile:   { type: object, fields: {
                email:        { type: string, format: email },
                display_name: { type: string },
                image_48:     { type: string, format: url }
              }}
            }}
            error:    { type: string, required: false }
          error_values: ["user_not_found", "user_not_visible", "invalid_auth",
                         "token_revoked"]
```

---

## PERFECT — Zero Violations

All Stricture rules pass. This integration correctly handles Slack's `ok`-field error model, treats `ts` as a string, validates token format, paginates via cursor, and covers all error paths.

```typescript
// slack-client.ts — Slack Web API client with full contract compliance.

import type { IncomingHttpHeaders } from "http";

// ─── Types ───────────────────────────────────────────────────────────────────

/** Slack message timestamp — string decimal, NOT a number. Used as message ID. */
type SlackTs = string;

/** Channel ID: C* (public), G* (private/group), D* (DM) */
type ChannelId = string;

/** User ID: U* or W* (enterprise grid) */
type UserId = string;

interface SlackBlock {
  type: string;
  text?: { type: string; text: string };
  elements?: unknown[];
  [key: string]: unknown;
}

interface SlackMessage {
  type: "message";
  ts: SlackTs;
  user: UserId;
  text: string;
  blocks?: SlackBlock[];
}

interface SlackChannel {
  id: ChannelId;
  name: string;
  is_channel: boolean;
  is_group: boolean;
  is_im: boolean;
  is_archived: boolean;
  num_members: number;
}

interface SlackUserProfile {
  email: string;
  display_name: string;
  image_48: string;
}

interface SlackUser {
  id: UserId;
  name: string;
  real_name: string;
  is_bot: boolean;
  is_admin: boolean;
  profile: SlackUserProfile;
}

// ─── Response types (wrapper: ok + error) ────────────────────────────────────

interface SlackBaseResponse {
  ok: boolean;
  error?: string;
  response_metadata?: { next_cursor: string };
}

interface PostMessageResponse extends SlackBaseResponse {
  channel: ChannelId;
  ts: SlackTs;
  message: SlackMessage;
}

interface ConversationsListResponse extends SlackBaseResponse {
  channels: SlackChannel[];
}

interface ConversationsHistoryResponse extends SlackBaseResponse {
  messages: SlackMessage[];
  has_more: boolean;
}

interface UsersInfoResponse extends SlackBaseResponse {
  user: SlackUser;
}

// ─── Error types ─────────────────────────────────────────────────────────────

type SlackErrorCode =
  | "channel_not_found"
  | "not_in_channel"
  | "is_archived"
  | "msg_too_long"
  | "no_text"
  | "invalid_auth"
  | "token_revoked"
  | "ratelimited"
  | "missing_scope"
  | "not_authed"
  | "invalid_blocks"
  | "invalid_blocks_format"
  | "user_not_found"
  | "user_not_visible";

class SlackApiError extends Error {
  constructor(
    public readonly code: SlackErrorCode,
    public readonly endpoint: string,
  ) {
    super(`Slack API error on ${endpoint}: ${code}`);
    this.name = "SlackApiError";
  }
}

class SlackRateLimitError extends Error {
  constructor(public readonly retryAfterSeconds: number) {
    super(`Slack rate limited. Retry after ${retryAfterSeconds}s`);
    this.name = "SlackRateLimitError";
  }
}

// ─── Validation helpers ──────────────────────────────────────────────────────

const CHANNEL_ID_RE = /^[CGD][A-Z0-9]+$/;
const USER_ID_RE = /^[UW][A-Z0-9]+$/;
const TOKEN_RE = /^xox[bp]-/;
const TS_RE = /^\d+\.\d+$/;
const MAX_TEXT_LENGTH = 40_000;

function validateChannelId(id: string): void {
  if (!CHANNEL_ID_RE.test(id)) {
    throw new Error(
      `Invalid channel ID "${id}". Must match C*, G*, or D* format.`,
    );
  }
}

function validateUserId(id: string): void {
  if (!USER_ID_RE.test(id)) {
    throw new Error(`Invalid user ID "${id}". Must match U* or W* format.`);
  }
}

function validateToken(token: string): void {
  if (!TOKEN_RE.test(token)) {
    throw new Error(
      `Invalid Slack token format. Must start with xoxb- or xoxp-.`,
    );
  }
}

function validateTs(ts: SlackTs): void {
  if (!TS_RE.test(ts)) {
    throw new Error(`Invalid timestamp "${ts}". Must be decimal string.`);
  }
}

function validateTextLength(text: string): void {
  if (text.length > MAX_TEXT_LENGTH) {
    throw new Error(
      `Message text exceeds ${MAX_TEXT_LENGTH} character limit (${text.length}).`,
    );
  }
}

// ─── Client ──────────────────────────────────────────────────────────────────

interface SlackClientConfig {
  token: string;
  baseUrl?: string;
}

export class SlackClient {
  private readonly token: string;
  private readonly baseUrl: string;

  constructor(config: SlackClientConfig) {
    validateToken(config.token);
    this.token = config.token;
    this.baseUrl = config.baseUrl ?? "https://slack.com";
  }

  // ─── chat.postMessage ────────────────────────────────────────────────────

  async postMessage(params: {
    channel: ChannelId;
    text?: string;
    blocks?: SlackBlock[];
    thread_ts?: SlackTs;
  }): Promise<{ channel: ChannelId; ts: SlackTs; message: SlackMessage }> {
    validateChannelId(params.channel);

    if (!params.text && !params.blocks) {
      throw new Error("At least one of text or blocks is required.");
    }
    if (params.text) {
      validateTextLength(params.text);
    }
    if (params.thread_ts) {
      validateTs(params.thread_ts);
    }

    const response = await this.request<PostMessageResponse>(
      "/api/chat.postMessage",
      "POST",
      {
        channel: params.channel,
        text: params.text,
        blocks: params.blocks,
        thread_ts: params.thread_ts,
      },
    );

    // Validate response ts is a string (not parsed as number)
    validateTs(response.ts);

    return {
      channel: response.channel,
      ts: response.ts,
      message: response.message,
    };
  }

  // ─── conversations.list (with full cursor pagination) ────────────────────

  async listAllChannels(params?: {
    types?: string;
    limit?: number;
  }): Promise<SlackChannel[]> {
    const allChannels: SlackChannel[] = [];
    let cursor: string | undefined;

    do {
      const response = await this.request<ConversationsListResponse>(
        "/api/conversations.list",
        "GET",
        {
          types: params?.types,
          limit: params?.limit ?? 200,
          cursor,
        },
      );

      allChannels.push(...response.channels);

      // CRITICAL: Slack returns empty string "" when no more pages.
      // Must check for empty string, not just falsy (since "" is falsy).
      const nextCursor = response.response_metadata?.next_cursor;
      cursor = nextCursor !== undefined && nextCursor !== "" ? nextCursor : undefined;
    } while (cursor !== undefined);

    return allChannels;
  }

  // ─── conversations.history (with full cursor pagination) ─────────────────

  async getChannelHistory(params: {
    channel: ChannelId;
    limit?: number;
    oldest?: SlackTs;
    latest?: SlackTs;
  }): Promise<SlackMessage[]> {
    validateChannelId(params.channel);

    if (params.oldest) {
      validateTs(params.oldest);
    }
    if (params.latest) {
      validateTs(params.latest);
    }

    const allMessages: SlackMessage[] = [];
    let cursor: string | undefined;

    do {
      const response = await this.request<ConversationsHistoryResponse>(
        "/api/conversations.history",
        "POST",
        {
          channel: params.channel,
          limit: params.limit ?? 200,
          cursor,
          oldest: params.oldest,
          latest: params.latest,
        },
      );

      allMessages.push(...response.messages);

      const nextCursor = response.response_metadata?.next_cursor;
      cursor = nextCursor !== undefined && nextCursor !== "" ? nextCursor : undefined;
    } while (cursor !== undefined);

    return allMessages;
  }

  // ─── users.info ──────────────────────────────────────────────────────────

  async getUserInfo(userId: UserId): Promise<SlackUser> {
    validateUserId(userId);

    const response = await this.request<UsersInfoResponse>(
      "/api/users.info",
      "POST",
      { user: userId },
    );

    return response.user;
  }

  // ─── Core request method ─────────────────────────────────────────────────

  private async request<T extends SlackBaseResponse>(
    path: string,
    method: "GET" | "POST",
    body?: Record<string, unknown>,
  ): Promise<T> {
    const url = new URL(path, this.baseUrl);

    const fetchOptions: RequestInit = {
      method,
      headers: {
        Authorization: `Bearer ${this.token}`,
        "Content-Type": "application/json; charset=utf-8",
      },
    };

    if (method === "POST" && body) {
      fetchOptions.body = JSON.stringify(body);
    } else if (method === "GET" && body) {
      for (const [key, value] of Object.entries(body)) {
        if (value !== undefined) {
          url.searchParams.set(key, String(value));
        }
      }
    }

    let res: Response;
    try {
      res = await fetch(url.toString(), fetchOptions);
    } catch (err) {
      throw new Error(
        `Network error calling ${path}: ${err instanceof Error ? err.message : String(err)}`,
      );
    }

    // CRITICAL: 429 is the ONLY non-200 status Slack returns.
    if (res.status === 429) {
      const retryAfter = parseInt(
        res.headers.get("Retry-After") ?? "30",
        10,
      );
      throw new SlackRateLimitError(retryAfter);
    }

    // Even non-200 is unexpected for Slack (but handle defensively)
    if (!res.ok) {
      throw new Error(`Unexpected HTTP ${res.status} from Slack API ${path}`);
    }

    let data: T;
    try {
      data = (await res.json()) as T;
    } catch (err) {
      throw new Error(
        `Failed to parse JSON from ${path}: ${err instanceof Error ? err.message : String(err)}`,
      );
    }

    // CRITICAL: Slack returns HTTP 200 even on errors.
    // The ok field in the body is the REAL success indicator.
    if (!data.ok) {
      throw new SlackApiError(
        (data.error ?? "unknown_error") as SlackErrorCode,
        path,
      );
    }

    return data;
  }
}

// ─── Tests ───────────────────────────────────────────────────────────────────

import { describe, it, expect, beforeEach, vi } from "vitest";

describe("SlackClient", () => {
  let client: SlackClient;

  beforeEach(() => {
    client = new SlackClient({ token: "xoxb-test-token-value" });
    vi.restoreAllMocks();
  });

  // ── postMessage ──────────────────────────────────────────────────────────

  describe("chat.postMessage", () => {
    it("sends a message and returns ts as string", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            ok: true,
            channel: "C1234567890",
            ts: "1678901234.567890",
            message: {
              type: "message",
              ts: "1678901234.567890",
              user: "U9876543210",
              text: "Hello, world!",
            },
          }),
          { status: 200 },
        ),
      );

      const result = await client.postMessage({
        channel: "C1234567890",
        text: "Hello, world!",
      });

      expect(result.channel).toBe("C1234567890");
      expect(typeof result.ts).toBe("string");
      expect(result.ts).toBe("1678901234.567890");
      expect(result.message.type).toBe("message");
      expect(result.message.text).toBe("Hello, world!");
      expect(result.message.user).toBe("U9876543210");
    });

    it("sends a message with blocks and optional text", async () => {
      const blocks: SlackBlock[] = [
        { type: "section", text: { type: "mrkdwn", text: "*Bold text*" } },
        { type: "divider" },
      ];

      vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            ok: true,
            channel: "C1234567890",
            ts: "1678901234.567891",
            message: {
              type: "message",
              ts: "1678901234.567891",
              user: "U9876543210",
              text: "*Bold text*",
              blocks,
            },
          }),
          { status: 200 },
        ),
      );

      const result = await client.postMessage({
        channel: "C1234567890",
        blocks,
      });

      expect(result.message.blocks).toHaveLength(2);
      expect(result.message.blocks![0].type).toBe("section");
      expect(result.message.blocks![1].type).toBe("divider");
    });

    it("rejects channel_not_found error (ok: false, HTTP 200)", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
        new Response(
          JSON.stringify({ ok: false, error: "channel_not_found" }),
          { status: 200 }, // STILL 200 — Slack's signature behavior
        ),
      );

      await expect(
        client.postMessage({ channel: "C0000000000", text: "test" }),
      ).rejects.toThrow(SlackApiError);

      await expect(
        client.postMessage({ channel: "C0000000000", text: "test" }),
      ).rejects.toMatchObject({ code: "channel_not_found" });
    });

    it("rejects not_in_channel error", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
        new Response(
          JSON.stringify({ ok: false, error: "not_in_channel" }),
          { status: 200 },
        ),
      );

      await expect(
        client.postMessage({ channel: "C1234567890", text: "test" }),
      ).rejects.toMatchObject({ code: "not_in_channel" });
    });

    it("rejects is_archived error", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
        new Response(
          JSON.stringify({ ok: false, error: "is_archived" }),
          { status: 200 },
        ),
      );

      await expect(
        client.postMessage({ channel: "C1234567890", text: "test" }),
      ).rejects.toMatchObject({ code: "is_archived" });
    });

    it("validates text length before sending", async () => {
      const longText = "x".repeat(40_001);
      await expect(
        client.postMessage({ channel: "C1234567890", text: longText }),
      ).rejects.toThrow("exceeds 40000 character limit");
    });

    it("requires at least text or blocks", async () => {
      await expect(
        client.postMessage({ channel: "C1234567890" }),
      ).rejects.toThrow("At least one of text or blocks is required");
    });

    it("validates channel ID format", async () => {
      await expect(
        client.postMessage({ channel: "invalid-channel", text: "test" }),
      ).rejects.toThrow("Invalid channel ID");
    });
  });

  // ── conversations.list (pagination) ──────────────────────────────────────

  describe("conversations.list", () => {
    it("fetches all channels across multiple pages", async () => {
      const fetchSpy = vi.spyOn(globalThis, "fetch");

      // Page 1: has next_cursor
      fetchSpy.mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            ok: true,
            channels: [
              { id: "C001", name: "general", is_channel: true, is_group: false,
                is_im: false, is_archived: false, num_members: 50 },
            ],
            response_metadata: { next_cursor: "dXNlcjpVMDYx" },
          }),
          { status: 200 },
        ),
      );

      // Page 2: empty next_cursor (no more pages)
      fetchSpy.mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            ok: true,
            channels: [
              { id: "G002", name: "private-team", is_channel: false, is_group: true,
                is_im: false, is_archived: false, num_members: 5 },
            ],
            response_metadata: { next_cursor: "" }, // Empty string = done
          }),
          { status: 200 },
        ),
      );

      const channels = await client.listAllChannels();

      expect(channels).toHaveLength(2);
      expect(channels[0].id).toBe("C001");
      expect(channels[0].name).toBe("general");
      expect(channels[1].id).toBe("G002");
      expect(channels[1].is_group).toBe(true);
      expect(fetchSpy).toHaveBeenCalledTimes(2);
    });

    it("handles single page with no cursor", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            ok: true,
            channels: [
              { id: "C001", name: "general", is_channel: true, is_group: false,
                is_im: false, is_archived: false, num_members: 50 },
            ],
            // No response_metadata at all
          }),
          { status: 200 },
        ),
      );

      const channels = await client.listAllChannels();
      expect(channels).toHaveLength(1);
    });

    it("handles invalid_auth error", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
        new Response(
          JSON.stringify({ ok: false, error: "invalid_auth" }),
          { status: 200 },
        ),
      );

      await expect(client.listAllChannels()).rejects.toThrow(SlackApiError);
      await expect(client.listAllChannels()).rejects.toMatchObject({
        code: "invalid_auth",
      });
    });
  });

  // ── conversations.history ────────────────────────────────────────────────

  describe("conversations.history", () => {
    it("fetches messages preserving ts as string", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            ok: true,
            messages: [
              { type: "message", user: "U001", text: "Hello",
                ts: "1678901234.567890" },
              { type: "message", user: "U002", text: "World",
                ts: "1678901234.567891", blocks: [{ type: "section" }] },
            ],
            has_more: false,
          }),
          { status: 200 },
        ),
      );

      const messages = await client.getChannelHistory({
        channel: "C1234567890",
      });

      expect(messages).toHaveLength(2);
      expect(typeof messages[0].ts).toBe("string");
      expect(messages[0].ts).toBe("1678901234.567890");
      expect(messages[0].type).toBe("message");
      expect(messages[1].blocks).toHaveLength(1);
    });

    it("safely accesses optional blocks field", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            ok: true,
            messages: [
              { type: "message", user: "U001", text: "No blocks", ts: "1678901234.000000" },
            ],
            has_more: false,
          }),
          { status: 200 },
        ),
      );

      const messages = await client.getChannelHistory({
        channel: "C1234567890",
      });

      // blocks is undefined — safe access with optional chaining
      expect(messages[0].blocks).toBeUndefined();
      const blockCount = messages[0].blocks?.length ?? 0;
      expect(blockCount).toBe(0);
    });
  });

  // ── users.info ───────────────────────────────────────────────────────────

  describe("users.info", () => {
    it("fetches user with full profile", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            ok: true,
            user: {
              id: "U9876543210",
              name: "alice",
              real_name: "Alice Smith",
              is_bot: false,
              is_admin: true,
              profile: {
                email: "alice@example.com",
                display_name: "Alice S",
                image_48: "https://avatars.slack.com/alice_48.png",
              },
            },
          }),
          { status: 200 },
        ),
      );

      const user = await client.getUserInfo("U9876543210");

      expect(user.id).toBe("U9876543210");
      expect(user.name).toBe("alice");
      expect(user.real_name).toBe("Alice Smith");
      expect(user.is_bot).toBe(false);
      expect(user.is_admin).toBe(true);
      expect(user.profile.email).toBe("alice@example.com");
      expect(user.profile.display_name).toBe("Alice S");
    });

    it("rejects user_not_found error", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
        new Response(
          JSON.stringify({ ok: false, error: "user_not_found" }),
          { status: 200 },
        ),
      );

      await expect(
        client.getUserInfo("U0000000000"),
      ).rejects.toMatchObject({ code: "user_not_found" });
    });

    it("validates user ID format before sending", async () => {
      await expect(
        client.getUserInfo("invalid-user-id"),
      ).rejects.toThrow("Invalid user ID");
    });
  });

  // ── Rate limiting ────────────────────────────────────────────────────────

  describe("rate limiting", () => {
    it("throws SlackRateLimitError on 429 with Retry-After", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
        new Response("", {
          status: 429,
          headers: { "Retry-After": "15" },
        }),
      );

      await expect(
        client.postMessage({ channel: "C1234567890", text: "test" }),
      ).rejects.toThrow(SlackRateLimitError);

      try {
        vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
          new Response("", {
            status: 429,
            headers: { "Retry-After": "15" },
          }),
        );
        await client.postMessage({ channel: "C1234567890", text: "test" });
      } catch (err) {
        expect(err).toBeInstanceOf(SlackRateLimitError);
        expect((err as SlackRateLimitError).retryAfterSeconds).toBe(15);
      }
    });
  });

  // ── Token validation ─────────────────────────────────────────────────────

  describe("token validation", () => {
    it("rejects tokens without xoxb-/xoxp- prefix", () => {
      expect(() => new SlackClient({ token: "bad-token" })).toThrow(
        "Invalid Slack token format",
      );
    });

    it("accepts xoxb- bot tokens", () => {
      expect(() => new SlackClient({ token: "xoxb-valid" })).not.toThrow();
    });

    it("accepts xoxp- user tokens", () => {
      expect(() => new SlackClient({ token: "xoxp-valid" })).not.toThrow();
    });
  });

  // ── Network errors ───────────────────────────────────────────────────────

  describe("network errors", () => {
    it("wraps fetch errors with context", async () => {
      vi.spyOn(globalThis, "fetch").mockRejectedValueOnce(
        new Error("ECONNREFUSED"),
      );

      await expect(
        client.postMessage({ channel: "C1234567890", text: "test" }),
      ).rejects.toThrow("Network error calling /api/chat.postMessage");
    });
  });
});
```

---

## B01 — No Error Handling

**Violated rule:** `TQ-error-path-coverage`

No try/catch around fetch calls. Network failures, JSON parse errors, and Slack API errors all propagate as unhandled exceptions.

```typescript
// slack-client-b01.ts — No error handling at all.

export class SlackClientB01 {
  private token: string;

  constructor(token: string) {
    this.token = token;
  }

  // NO try/catch. NO error handling. Crashes on any failure.
  async postMessage(channel: string, text: string): Promise<unknown> {
    const res = await fetch("https://slack.com/api/chat.postMessage", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${this.token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ channel, text }),
    });
    return res.json();
  }

  async listChannels(): Promise<unknown> {
    const res = await fetch("https://slack.com/api/conversations.list", {
      headers: { Authorization: `Bearer ${this.token}` },
    });
    return res.json();
  }
}

// Tests — no error path tests whatsoever
import { describe, it, expect, vi } from "vitest";

describe("SlackClientB01", () => {
  it("sends a message", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(
        JSON.stringify({ ok: true, ts: "1234.5678", channel: "C123" }),
        { status: 200 },
      ),
    );
    const client = new SlackClientB01("xoxb-token");
    const result = await client.postMessage("C123", "hello");
    expect(result).toBeDefined();
  });

  // NO tests for:
  // - Network failure (fetch rejects)
  // - JSON parse failure
  // - Slack ok: false
  // - 429 rate limit
  // - invalid_auth
  // - channel_not_found
});
```

**Expected violation:** `TQ-error-path-coverage` -- fetch() calls have no try/catch. No error paths exercised in tests.

**Production impact:** Any network timeout, DNS failure, or Slack outage crashes the calling service with an unhandled promise rejection. Since Slack returns 200 on errors, the client silently treats `{ ok: false, error: "channel_not_found" }` as a successful response, storing invalid data.

---

## B02 — No `ok` Field Check (CRITICAL Slack-Specific)

**Violated rule:** `CTR-status-code-handling`

This is the most common Slack integration bug. The client checks HTTP status (which is always 200) but never checks the `ok` field in the response body. Every Slack error is silently treated as success.

```typescript
// slack-client-b02.ts — Checks HTTP 200 but not response.ok.

export class SlackClientB02 {
  private token: string;

  constructor(token: string) {
    this.token = token;
  }

  async postMessage(
    channel: string,
    text: string,
  ): Promise<{ channel: string; ts: string }> {
    const res = await fetch("https://slack.com/api/chat.postMessage", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${this.token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ channel, text }),
    });

    // BUG: Checks HTTP status, but Slack ALWAYS returns 200.
    // This check passes even when Slack returns { ok: false, error: "channel_not_found" }
    if (!res.ok) {
      throw new Error(`HTTP error: ${res.status}`);
    }

    const data = await res.json();
    // MISSING: if (!data.ok) throw ...
    // data.channel and data.ts are undefined when ok is false
    return { channel: data.channel, ts: data.ts };
  }
}

// Tests — only happy path
import { describe, it, expect, vi } from "vitest";

describe("SlackClientB02", () => {
  it("sends a message", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          ok: true,
          channel: "C123",
          ts: "1234.5678",
          message: { type: "message", ts: "1234.5678", user: "U123", text: "hi" },
        }),
        { status: 200 },
      ),
    );

    const client = new SlackClientB02("xoxb-token");
    const result = await client.postMessage("C123", "hello");
    expect(result.ts).toBe("1234.5678");
  });

  // NO test for Slack returning { ok: false, error: "channel_not_found" } with HTTP 200
});
```

**Expected violation:** `CTR-status-code-handling` -- Client checks HTTP status but Slack API uses body-level `ok` field for error signaling. HTTP 200 does not mean success. The `ok` field is never checked.

**Production impact:** Every Slack error (invalid token, channel not found, not in channel, archived channel, rate limited body response) is silently swallowed. The client returns `{ channel: undefined, ts: undefined }` as if the message was sent. Messages appear to succeed but are never delivered. Downstream logic stores undefined values, corrupting state.

---

## B03 — Shallow Assertions

**Violated rule:** `TQ-no-shallow-assertions`

Tests exist but only check `toBeDefined()` or `toBeTruthy()` on return values. Never verifies the actual shape, field values, or types.

```typescript
// slack-client-b03.ts — Implementation is OK, tests are shallow.

export class SlackClientB03 {
  private token: string;

  constructor(token: string) {
    this.token = token;
  }

  async postMessage(channel: string, text: string): Promise<{
    channel: string;
    ts: string;
    message: { type: string; ts: string; user: string; text: string };
  }> {
    const res = await fetch("https://slack.com/api/chat.postMessage", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${this.token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ channel, text }),
    });
    const data = await res.json();
    if (!data.ok) throw new Error(data.error);
    return { channel: data.channel, ts: data.ts, message: data.message };
  }
}

// Tests — all assertions are shallow
import { describe, it, expect, vi } from "vitest";

describe("SlackClientB03", () => {
  it("sends a message", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          ok: true,
          channel: "C123",
          ts: "1234.5678",
          message: { type: "message", ts: "1234.5678", user: "U001", text: "hi" },
        }),
        { status: 200 },
      ),
    );

    const client = new SlackClientB03("xoxb-token");
    const result = await client.postMessage("C123", "hello");

    // BUG: All shallow — none of these verify actual contract shape
    expect(result).toBeDefined();
    expect(result).toBeTruthy();
    expect(result.channel).toBeDefined();
    expect(result.ts).toBeDefined();
    expect(result.message).toBeDefined();
    expect(typeof result).toBe("object");
  });

  it("handles error", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(
        JSON.stringify({ ok: false, error: "channel_not_found" }),
        { status: 200 },
      ),
    );

    const client = new SlackClientB03("xoxb-token");
    try {
      await client.postMessage("C123", "hello");
    } catch (err) {
      // BUG: Only checks that error exists, not what it contains
      expect(err).toBeDefined();
      expect(err).toBeTruthy();
    }
  });
});
```

**Expected violation:** `TQ-no-shallow-assertions` -- 8 assertions use `.toBeDefined()`, `.toBeTruthy()`, or `typeof === "object"`. None verify actual values like `ts === "1234.5678"`, `message.type === "message"`, or error code matching.

**Production impact:** Tests pass even if the implementation returns completely wrong data. A refactor that breaks the response shape (e.g., renaming `ts` to `timestamp`) would not be caught. The test suite provides false confidence.

---

## B04 — Missing Negative Tests

**Violated rule:** `TQ-negative-cases`

Only tests the happy path. No tests for `channel_not_found`, `not_in_channel`, `invalid_auth`, `token_revoked`, rate limiting, or any Slack-specific error codes.

```typescript
// slack-client-b04.ts — Implementation handles errors, but tests don't.

export class SlackClientB04 {
  private token: string;

  constructor(token: string) {
    this.token = token;
  }

  async postMessage(channel: string, text: string): Promise<{
    channel: string;
    ts: string;
  }> {
    const res = await fetch("https://slack.com/api/chat.postMessage", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${this.token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ channel, text }),
    });
    const data = await res.json();
    if (!data.ok) throw new Error(`Slack error: ${data.error}`);
    return { channel: data.channel, ts: data.ts };
  }

  async getUserInfo(userId: string): Promise<{ id: string; name: string }> {
    const res = await fetch("https://slack.com/api/users.info", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${this.token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ user: userId }),
    });
    const data = await res.json();
    if (!data.ok) throw new Error(`Slack error: ${data.error}`);
    return { id: data.user.id, name: data.user.name };
  }
}

// Tests — only happy path
import { describe, it, expect, vi } from "vitest";

describe("SlackClientB04", () => {
  it("sends a message successfully", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(
        JSON.stringify({ ok: true, channel: "C123", ts: "1234.5678" }),
        { status: 200 },
      ),
    );

    const client = new SlackClientB04("xoxb-token");
    const result = await client.postMessage("C123", "hello");
    expect(result.channel).toBe("C123");
  });

  it("fetches user info successfully", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          ok: true,
          user: { id: "U123", name: "alice" },
        }),
        { status: 200 },
      ),
    );

    const client = new SlackClientB04("xoxb-token");
    const user = await client.getUserInfo("U123");
    expect(user.name).toBe("alice");
  });

  // MISSING: No negative tests for any of these error scenarios:
  // - channel_not_found
  // - not_in_channel
  // - is_archived
  // - invalid_auth
  // - token_revoked
  // - ratelimited (429)
  // - user_not_found
  // - network failures
  // - msg_too_long
});
```

**Expected violation:** `TQ-negative-cases` -- Only success paths tested. No tests for Slack error codes (`channel_not_found`, `not_in_channel`, `invalid_auth`, `token_revoked`, `ratelimited`), no 429 test, no network error test.

**Production impact:** Error handling code is untested. A refactor that accidentally removes the `if (!data.ok)` check would not be caught. When the first `channel_not_found` error occurs in production, there is no test proving the application handles it correctly.

---

## B05 — Request Missing Required Fields

**Violated rule:** `CTR-request-shape`

`chat.postMessage` is called without the required `channel` field. The request body shape does not match the API contract.

```typescript
// slack-client-b05.ts — Request missing required fields.

interface PostMessageParams {
  text: string;
  // BUG: "channel" field is missing from the interface
  thread_ts?: string;
}

export class SlackClientB05 {
  private token: string;

  constructor(token: string) {
    this.token = token;
  }

  async postMessage(params: PostMessageParams): Promise<unknown> {
    try {
      const res = await fetch("https://slack.com/api/chat.postMessage", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.token}`,
          "Content-Type": "application/json",
        },
        // BUG: params has no channel field — Slack will return ok: false
        body: JSON.stringify(params),
      });
      const data = await res.json();
      if (!data.ok) throw new Error(data.error);
      return data;
    } catch (err) {
      throw new Error(`postMessage failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }
}

// Tests
import { describe, it, expect, vi } from "vitest";

describe("SlackClientB05", () => {
  it("sends a message", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(
        JSON.stringify({ ok: true, channel: "C123", ts: "1234.5678" }),
        { status: 200 },
      ),
    );

    const client = new SlackClientB05("xoxb-token");
    // BUG: no channel field in the call — mock hides the contract violation
    const result = await client.postMessage({ text: "hello" });
    expect(result).toMatchObject({ ok: true });
  });
});
```

**Expected violation:** `CTR-request-shape` -- `PostMessageParams` is missing the required `channel` field. Manifest declares `channel` as `required: true` for `/api/chat.postMessage`.

**Production impact:** Every call to `postMessage` sends a request without a channel. Slack returns `{ ok: false, error: "channel_not_found" }` (HTTP 200). The mock in tests hides this entirely. In production, no messages are ever delivered.

---

## B06 — Response Type Mismatch

**Violated rule:** `CTR-response-shape`

The client-side `Message` type is missing the optional `blocks` field. Code that receives messages with blocks cannot type-check them.

```typescript
// slack-client-b06.ts — Response type missing optional blocks field.

// BUG: Message type is incomplete — missing blocks field
interface Message {
  type: string;
  ts: string;
  user: string;
  text: string;
  // MISSING: blocks?: Block[]
}

interface PostMessageResult {
  channel: string;
  ts: string;
  message: Message;
}

export class SlackClientB06 {
  private token: string;

  constructor(token: string) {
    this.token = token;
  }

  async postMessage(channel: string, text: string): Promise<PostMessageResult> {
    try {
      const res = await fetch("https://slack.com/api/chat.postMessage", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ channel, text }),
      });
      const data = await res.json();
      if (!data.ok) throw new Error(data.error);
      return {
        channel: data.channel,
        ts: data.ts,
        message: data.message as Message,
      };
    } catch (err) {
      throw new Error(`postMessage failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }
}

// Tests
import { describe, it, expect, vi } from "vitest";

describe("SlackClientB06", () => {
  it("sends a message", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          ok: true,
          channel: "C123",
          ts: "1234.5678",
          message: {
            type: "message",
            ts: "1234.5678",
            user: "U001",
            text: "hello",
            blocks: [{ type: "section", text: { type: "mrkdwn", text: "hello" } }],
          },
        }),
        { status: 200 },
      ),
    );

    const client = new SlackClientB06("xoxb-token");
    const result = await client.postMessage("C123", "hello");
    expect(result.message.text).toBe("hello");
    expect(result.message.type).toBe("message");
    // Cannot test blocks — type does not include them
  });
});
```

**Expected violation:** `CTR-response-shape` -- Client `Message` type is missing the `blocks` field. Manifest declares `message.blocks` as an optional array of objects. The response may contain blocks that the client type cannot represent.

**Production impact:** Downstream code that tries to render rich message content (blocks with buttons, images, or interactive elements) has no type information. Any access to `message.blocks` requires an unsafe cast. Feature requests to display block content require a type change that could break existing consumers.

---

## B07 — Wrong Field Types (ts as Number)

**Violated rule:** `CTR-manifest-conformance`

Message `ts` (timestamp) is stored as a `number` instead of a `string`. Slack timestamps like `"1234567890.123456"` are string decimals that serve as unique message IDs. Parsing as float loses precision.

```typescript
// slack-client-b07.ts — ts stored as number instead of string.

interface SlackMessageB07 {
  type: string;
  // BUG: ts should be string, not number.
  // Slack ts "1234567890.123456" is a string decimal used as message ID.
  ts: number;
  user: string;
  text: string;
}

export class SlackClientB07 {
  private token: string;

  constructor(token: string) {
    this.token = token;
  }

  async postMessage(channel: string, text: string): Promise<{
    channel: string;
    ts: number;  // BUG: should be string
  }> {
    try {
      const res = await fetch("https://slack.com/api/chat.postMessage", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ channel, text }),
      });
      const data = await res.json();
      if (!data.ok) throw new Error(data.error);
      return {
        channel: data.channel,
        // BUG: Converts string ts to number — loses trailing zeros
        ts: Number(data.ts),
      };
    } catch (err) {
      throw new Error(`postMessage failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }
}

// Tests
import { describe, it, expect, vi } from "vitest";

describe("SlackClientB07", () => {
  it("sends a message", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          ok: true,
          channel: "C123",
          ts: "1678901234.567890",
          message: { type: "message", ts: "1678901234.567890", user: "U001", text: "hello" },
        }),
        { status: 200 },
      ),
    );

    const client = new SlackClientB07("xoxb-token");
    const result = await client.postMessage("C123", "hello");
    // BUG: ts is now 1678901234.56789 (number) — trailing zero lost
    expect(result.ts).toBe(1678901234.56789);
    // The string "1678901234.567890" became number 1678901234.56789
    // These are DIFFERENT message IDs in Slack!
  });
});
```

**Expected violation:** `CTR-manifest-conformance` -- Field `ts` type mismatch. Manifest declares `ts` as `string` (format: `decimal_timestamp`). Code uses `number`. The `Number()` conversion changes the value.

**Production impact:** Slack uses `ts` as a unique message identifier. The string `"1678901234.567890"` and `"1678901234.56789"` refer to different messages. Converting to a JavaScript number loses trailing zeros and potentially introduces floating-point rounding. Replying to threads, deleting messages, or fetching reactions by `ts` will fail silently or target the wrong message.

---

## B08 — Incomplete Enum Handling

**Violated rule:** `CTR-strictness-parity`

Error handling only covers `invalid_auth` but not `token_revoked`, `ratelimited`, `channel_not_found`, `not_in_channel`, or any other Slack error codes.

```typescript
// slack-client-b08.ts — Handles only one error code.

export class SlackClientB08 {
  private token: string;

  constructor(token: string) {
    this.token = token;
  }

  async postMessage(channel: string, text: string): Promise<{
    channel: string;
    ts: string;
  }> {
    try {
      const res = await fetch("https://slack.com/api/chat.postMessage", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ channel, text }),
      });
      const data = await res.json();

      if (!data.ok) {
        // BUG: Only handles one error code out of 12+
        switch (data.error) {
          case "invalid_auth":
            throw new Error("Authentication failed. Check your token.");
          // MISSING: "token_revoked" — token was revoked by admin
          // MISSING: "ratelimited" — rate limit exceeded (body-level, not 429)
          // MISSING: "channel_not_found" — channel does not exist
          // MISSING: "not_in_channel" — bot not in channel
          // MISSING: "is_archived" — channel is archived
          // MISSING: "msg_too_long" — text exceeds limit
          // MISSING: "no_text" — neither text nor blocks provided
          // MISSING: "missing_scope" — token lacks required scope
          default:
            // Silent fallthrough — unknown errors are silently ignored
            return { channel: "", ts: "" };
        }
      }

      return { channel: data.channel, ts: data.ts };
    } catch (err) {
      throw err;
    }
  }
}

// Tests
import { describe, it, expect, vi } from "vitest";

describe("SlackClientB08", () => {
  it("sends a message", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(
        JSON.stringify({ ok: true, channel: "C123", ts: "1234.5678" }),
        { status: 200 },
      ),
    );

    const client = new SlackClientB08("xoxb-token");
    const result = await client.postMessage("C123", "hello");
    expect(result.channel).toBe("C123");
  });

  it("handles invalid_auth", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(
        JSON.stringify({ ok: false, error: "invalid_auth" }),
        { status: 200 },
      ),
    );

    const client = new SlackClientB08("xoxb-bad");
    await expect(client.postMessage("C123", "hello")).rejects.toThrow(
      "Authentication failed",
    );
  });

  // NO tests for token_revoked, ratelimited, channel_not_found, etc.
});
```

**Expected violation:** `CTR-strictness-parity` -- Consumer handles 1 of 12+ error values for the `error` field. Manifest declares: `channel_not_found`, `not_in_channel`, `is_archived`, `msg_too_long`, `no_text`, `invalid_auth`, `token_revoked`, `ratelimited`, `missing_scope`, `not_authed`, `invalid_blocks`, `invalid_blocks_format`. Only `invalid_auth` is handled. The `default` case silently returns empty data instead of throwing.

**Production impact:** When a token is revoked (e.g., by a workspace admin), the client silently returns `{ channel: "", ts: "" }` instead of surfacing the `token_revoked` error. The system continues trying to send messages, all silently failing. When a channel is archived, same behavior. The application has no way to distinguish "sent successfully" from "channel archived" or "rate limited."

---

## B09 — Missing Range Validation

**Violated rule:** `CTR-strictness-parity`

No validation that message text stays under Slack's 40,000 character limit. Oversized messages are sent to the API where they fail with `msg_too_long`.

```typescript
// slack-client-b09.ts — No text length validation.

export class SlackClientB09 {
  private token: string;

  constructor(token: string) {
    this.token = token;
  }

  async postMessage(channel: string, text: string): Promise<{
    channel: string;
    ts: string;
  }> {
    // BUG: No validation that text.length <= 40,000
    // Slack will reject with { ok: false, error: "msg_too_long" }
    // but the client should validate before making the network call
    try {
      const res = await fetch("https://slack.com/api/chat.postMessage", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ channel, text }),
      });
      const data = await res.json();
      if (!data.ok) throw new Error(`Slack error: ${data.error}`);
      return { channel: data.channel, ts: data.ts };
    } catch (err) {
      throw new Error(`postMessage failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }

  async getChannelHistory(channel: string, limit?: number): Promise<unknown[]> {
    // BUG: No validation that limit is in [1, 1000]
    try {
      const res = await fetch("https://slack.com/api/conversations.history", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ channel, limit }),
      });
      const data = await res.json();
      if (!data.ok) throw new Error(`Slack error: ${data.error}`);
      return data.messages;
    } catch (err) {
      throw new Error(`getHistory failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }
}

// Tests
import { describe, it, expect, vi } from "vitest";

describe("SlackClientB09", () => {
  it("sends a message", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(
        JSON.stringify({ ok: true, channel: "C123", ts: "1234.5678" }),
        { status: 200 },
      ),
    );

    const client = new SlackClientB09("xoxb-token");
    const result = await client.postMessage("C123", "hello");
    expect(result.ts).toBe("1234.5678");
  });

  // NO test for:
  // - text with 50,000 characters
  // - limit of 0 or 5000
  // - boundary: exactly 40,000 characters
  // - boundary: 40,001 characters
});
```

**Expected violation:** `CTR-strictness-parity` -- Manifest declares `text.maxLength: 40000` and `limit.range: [1, 1000]`. Client performs no range validation on either field. Producer (Slack) enforces these; consumer does not.

**Production impact:** A user pasting a large log dump or document into a message field causes an unnecessary network round-trip to Slack, which rejects it. The error message from Slack (`msg_too_long`) is less helpful than a client-side validation message. For `limit`, passing 0 or a negative number produces unexpected results from the API.

---

## B10 — Format Not Validated

**Violated rule:** `CTR-strictness-parity`

Channel IDs are accepted as any arbitrary string. No validation that they match the `C*`/`G*`/`D*` prefix format.

```typescript
// slack-client-b10.ts — No channel/user ID format validation.

export class SlackClientB10 {
  private token: string;

  constructor(token: string) {
    this.token = token;
  }

  // BUG: channel accepts any string, no C*/G*/D* format validation
  async postMessage(channel: string, text: string): Promise<{
    channel: string;
    ts: string;
  }> {
    try {
      const res = await fetch("https://slack.com/api/chat.postMessage", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ channel, text }),
      });
      const data = await res.json();
      if (!data.ok) throw new Error(`Slack error: ${data.error}`);
      return { channel: data.channel, ts: data.ts };
    } catch (err) {
      throw new Error(`postMessage failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }

  // BUG: userId accepts any string, no U*/W* format validation
  async getUserInfo(userId: string): Promise<{ id: string; name: string }> {
    try {
      const res = await fetch("https://slack.com/api/users.info", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ user: userId }),
      });
      const data = await res.json();
      if (!data.ok) throw new Error(`Slack error: ${data.error}`);
      return { id: data.user.id, name: data.user.name };
    } catch (err) {
      throw new Error(`getUserInfo failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }
}

// Tests
import { describe, it, expect, vi } from "vitest";

describe("SlackClientB10", () => {
  it("accepts any string as channel", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(
        JSON.stringify({ ok: true, channel: "C123", ts: "1234.5678" }),
        { status: 200 },
      ),
    );

    const client = new SlackClientB10("xoxb-token");
    // BUG: "#general" is a channel name, not an ID. Slack needs "C1234567890".
    // This passes because the mock doesn't validate the input.
    const result = await client.postMessage("#general", "hello");
    expect(result.channel).toBe("C123");
  });

  it("accepts any string as user ID", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(
        JSON.stringify({ ok: true, user: { id: "U123", name: "alice" } }),
        { status: 200 },
      ),
    );

    const client = new SlackClientB10("xoxb-token");
    // BUG: "alice" is a username, not a user ID. Slack needs "U1234567890".
    const user = await client.getUserInfo("alice");
    expect(user.name).toBe("alice");
  });
});
```

**Expected violation:** `CTR-strictness-parity` -- Manifest declares `channel` format as `C*|G*|D*` and `user` format as `U*|W*`. Client accepts any string for both fields. No format validation performed.

**Production impact:** Developers pass channel names (`#general`) or user handles (`@alice`) instead of IDs (`C1234567890`, `U9876543210`). These requests hit the Slack API and fail with `channel_not_found` or `user_not_found`. The error message from Slack does not explain that the issue is a name vs. ID confusion. This is the most common Slack integration mistake reported in developer forums.

---

## B11 — Precision Loss (ts Parsed as Float)

**Violated rule:** `CTR-strictness-parity`

Slack's `ts` field `"1234567890.123456"` is parsed as a JavaScript floating-point number, losing microsecond precision needed for message uniqueness.

```typescript
// slack-client-b11.ts — ts precision loss from float parsing.

export class SlackClientB11 {
  private token: string;

  constructor(token: string) {
    this.token = token;
  }

  async postMessage(channel: string, text: string): Promise<{
    channel: string;
    ts: string;
  }> {
    try {
      const res = await fetch("https://slack.com/api/chat.postMessage", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ channel, text }),
      });
      const data = await res.json();
      if (!data.ok) throw new Error(`Slack error: ${data.error}`);

      // BUG: Parsing ts through parseFloat and back loses precision.
      // "1678901234.567890" -> parseFloat -> 1678901234.56789 -> toString -> "1678901234.56789"
      // The trailing zero is lost. These are DIFFERENT message IDs in Slack.
      const messageTs = parseFloat(data.ts).toString();

      return { channel: data.channel, ts: messageTs };
    } catch (err) {
      throw new Error(`postMessage failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }

  async replyToMessage(
    channel: string,
    threadTs: string,
    text: string,
  ): Promise<{ ts: string }> {
    try {
      // BUG: threadTs was already precision-damaged by postMessage
      const res = await fetch("https://slack.com/api/chat.postMessage", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ channel, text, thread_ts: threadTs }),
      });
      const data = await res.json();
      if (!data.ok) throw new Error(`Slack error: ${data.error}`);
      return { ts: parseFloat(data.ts).toString() };
    } catch (err) {
      throw new Error(`replyToMessage failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }
}

// Tests
import { describe, it, expect, vi } from "vitest";

describe("SlackClientB11", () => {
  it("sends a message and loses ts precision", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          ok: true,
          channel: "C123",
          ts: "1678901234.567890",  // 6 decimal digits
        }),
        { status: 200 },
      ),
    );

    const client = new SlackClientB11("xoxb-token");
    const result = await client.postMessage("C123", "hello");

    // BUG: ts changed from "1678901234.567890" to "1678901234.56789"
    // The test asserts the WRONG value (the precision-damaged one)
    expect(result.ts).toBe("1678901234.56789");
    // Should be: expect(result.ts).toBe("1678901234.567890");
  });
});
```

**Expected violation:** `CTR-strictness-parity` -- Manifest declares `ts` as format `decimal_timestamp` (string). The client parses it through `parseFloat()` which causes precision loss. The value `"1678901234.567890"` becomes `"1678901234.56789"`.

**Production impact:** Thread replies target the wrong parent message because `thread_ts` no longer matches the original `ts`. Message deletion requests (`chat.delete`) fail silently with `message_not_found`. Reaction lookups (`reactions.get`) return empty results. In high-volume channels, two messages posted in the same second may have timestamps that differ only in the last digit — precision loss makes them appear identical, causing data corruption in message indexing.

---

## B12 — Nullable Field Crash

**Violated rule:** `CTR-response-shape`

Code directly accesses `message.blocks[0].text` without checking if `blocks` is undefined. Slack messages without rich content have no `blocks` field.

```typescript
// slack-client-b12.ts — Crashes on missing blocks field.

interface SlackMessageB12 {
  type: string;
  ts: string;
  user: string;
  text: string;
  blocks?: Array<{
    type: string;
    text?: { type: string; text: string };
  }>;
}

export class SlackClientB12 {
  private token: string;

  constructor(token: string) {
    this.token = token;
  }

  async getChannelHistory(channel: string): Promise<{
    messages: SlackMessageB12[];
    firstBlockText: string;
  }> {
    try {
      const res = await fetch("https://slack.com/api/conversations.history", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ channel }),
      });
      const data = await res.json();
      if (!data.ok) throw new Error(`Slack error: ${data.error}`);

      const messages: SlackMessageB12[] = data.messages;

      // BUG: Directly accesses blocks[0].text without null checks.
      // blocks is optional — plain text messages have no blocks.
      // blocks[0].text is also optional — divider blocks have no text.
      const firstMessage = messages[0];
      const firstBlockText = firstMessage.blocks[0].text.text;
      //                      ^^^^^^^^^^^^^^^^^^^ TypeError: Cannot read properties of undefined

      return { messages, firstBlockText };
    } catch (err) {
      throw new Error(`getHistory failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }
}

// Tests — only test with blocks present
import { describe, it, expect, vi } from "vitest";

describe("SlackClientB12", () => {
  it("fetches history with blocks", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          ok: true,
          messages: [
            {
              type: "message",
              ts: "1234.5678",
              user: "U001",
              text: "hello",
              blocks: [
                { type: "section", text: { type: "mrkdwn", text: "hello" } },
              ],
            },
          ],
          has_more: false,
        }),
        { status: 200 },
      ),
    );

    const client = new SlackClientB12("xoxb-token");
    const result = await client.getChannelHistory("C123");
    expect(result.firstBlockText).toBe("hello");
  });

  // MISSING: Test with message that has NO blocks (plain text message)
  // MISSING: Test with message that has blocks but first block has no text (divider)
  // MISSING: Test with empty messages array
});
```

**Expected violation:** `CTR-response-shape` -- Field `blocks` is declared optional in the manifest (`required: false`). Code accesses `message.blocks[0].text.text` without checking for undefined at each level. `blocks` may be absent, `blocks[0]` may have no `text` field (e.g., `divider` blocks).

**Production impact:** The first time a user sends a plain text message (no blocks), the entire history fetch crashes with `TypeError: Cannot read properties of undefined (reading '0')`. This is a runtime crash, not a compile-time error. Since most messages are plain text, this crashes immediately in production.

---

## B13 — Missing Token Format Validation

**Violated rule:** `CTR-request-shape`

The client accepts any string as a Slack token. No validation that the token starts with `xoxb-` (bot) or `xoxp-` (user). Invalid tokens produce confusing `invalid_auth` errors at request time instead of failing fast at construction.

```typescript
// slack-client-b13.ts — No token format validation.

export class SlackClientB13 {
  private token: string;

  constructor(token: string) {
    // BUG: Accepts any string as token. No xoxb-*/xoxp-* validation.
    // An OpenAI API key, an AWS secret, or an empty string are all accepted.
    this.token = token;
  }

  async postMessage(channel: string, text: string): Promise<{
    channel: string;
    ts: string;
  }> {
    try {
      const res = await fetch("https://slack.com/api/chat.postMessage", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ channel, text }),
      });
      const data = await res.json();
      if (!data.ok) throw new Error(`Slack error: ${data.error}`);
      return { channel: data.channel, ts: data.ts };
    } catch (err) {
      throw new Error(`postMessage failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }
}

// Tests
import { describe, it, expect, vi } from "vitest";

describe("SlackClientB13", () => {
  it("creates client with any string", () => {
    // BUG: All of these should be rejected
    expect(() => new SlackClientB13("")).not.toThrow();
    expect(() => new SlackClientB13("sk-openai-key")).not.toThrow();
    expect(() => new SlackClientB13("AKIA-aws-key")).not.toThrow();
    expect(() => new SlackClientB13("random-garbage")).not.toThrow();
  });

  it("sends a message", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(
        JSON.stringify({ ok: true, channel: "C123", ts: "1234.5678" }),
        { status: 200 },
      ),
    );

    // BUG: Using a clearly invalid token. Mock hides the problem.
    const client = new SlackClientB13("not-a-slack-token");
    const result = await client.postMessage("C123", "hello");
    expect(result.ts).toBe("1234.5678");
  });
});
```

**Expected violation:** `CTR-request-shape` -- Manifest declares auth format as `xoxb-*|xoxp-*`. Client constructor accepts any string without format validation. The `Authorization: Bearer` header is set with an unvalidated token.

**Production impact:** Configuration errors (wrong environment variable, copy-paste of wrong token) are not caught until the first API call. The error from Slack is `invalid_auth` which is ambiguous (could mean token expired, revoked, or wrong format). A developer debugging this issue wastes time checking token scopes and workspace settings when the real problem is that they pasted an OpenAI key into the Slack token config. Fail-fast validation at construction time would immediately tell them the format is wrong.

---

## B14 — Pagination Terminated Early

**Violated rule:** `CTR-response-shape`

The cursor pagination loop treats the empty string `""` as falsy and stops after the first page. In JavaScript, `""` is falsy, so `if (cursor)` evaluates to false when Slack returns `next_cursor: ""`. The bug is that the code uses this same pattern to also check non-empty cursors, causing early termination when the cursor check is inverted.

```typescript
// slack-client-b14.ts — Pagination bug: stops after first page.

export class SlackClientB14 {
  private token: string;

  constructor(token: string) {
    this.token = token;
  }

  async listAllChannels(): Promise<Array<{ id: string; name: string }>> {
    const allChannels: Array<{ id: string; name: string }> = [];
    let cursor: string | undefined;

    try {
      do {
        const url = new URL("https://slack.com/api/conversations.list");
        url.searchParams.set("limit", "200");
        if (cursor) {
          url.searchParams.set("cursor", cursor);
        }

        const res = await fetch(url.toString(), {
          headers: { Authorization: `Bearer ${this.token}` },
        });
        const data = await res.json();
        if (!data.ok) throw new Error(`Slack error: ${data.error}`);

        allChannels.push(
          ...data.channels.map((ch: { id: string; name: string }) => ({
            id: ch.id,
            name: ch.name,
          })),
        );

        // BUG: Treats cursor as truthy/falsy.
        // Slack returns next_cursor: "" when there are no more pages.
        // But "" is falsy in JavaScript, so this ALSO stops when
        // next_cursor is a real cursor value that happens to be
        // the first page's response_metadata.next_cursor.
        //
        // The real bug: code only fetches first page and stops
        // because it assigns cursor from response_metadata?.next_cursor
        // and then checks `while (cursor)` — but if next_cursor
        // is present and non-empty, it should continue.
        cursor = data.response_metadata?.next_cursor;
        // When next_cursor is "" (empty string), cursor becomes ""
        // which is falsy — loop correctly stops.
        // BUT the loop below uses `while (cursor)` which also
        // means: if response_metadata is undefined (no pagination
        // info), cursor is undefined, loop stops after first page
        // even if there IS more data (has_more: true).
      } while (cursor);
      //       ^^^^^^ BUG: "" is falsy, undefined is falsy.
      //       Should check: cursor !== undefined && cursor !== ""
      //       to distinguish "no more pages" from "no pagination metadata"
    } catch (err) {
      throw new Error(`listChannels failed: ${err instanceof Error ? err.message : String(err)}`);
    }

    return allChannels;
  }
}

// Tests — only test single page
import { describe, it, expect, vi } from "vitest";

describe("SlackClientB14", () => {
  it("fetches channels — but only first page", async () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch");

    // Page 1 with cursor
    fetchSpy.mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          ok: true,
          channels: [{ id: "C001", name: "general" }],
          response_metadata: { next_cursor: "cursor_page2" },
        }),
        { status: 200 },
      ),
    );

    // Page 2 — end of pagination
    fetchSpy.mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          ok: true,
          channels: [{ id: "C002", name: "random" }],
          response_metadata: { next_cursor: "" },
        }),
        { status: 200 },
      ),
    );

    const client = new SlackClientB14("xoxb-token");
    const channels = await client.listAllChannels();

    // NOTE: This test actually passes because while("cursor_page2") is truthy.
    // The real scenario that fails is when response_metadata is missing entirely
    // on page 1 (some older Slack API responses omit it).
    // But the test only covers the standard case.
    expect(channels).toHaveLength(2);

    // MISSING test: response without response_metadata field at all
    // In that case, cursor = undefined, loop stops at 1 page
  });

  // MISSING: Test with 3+ pages
  // MISSING: Test where response_metadata is completely absent
  // MISSING: Test boundary: exactly at cursor "" vs missing metadata
});
```

**Expected violation:** `CTR-response-shape` -- Pagination logic uses JavaScript truthiness (`while (cursor)`) instead of explicit empty-string check (`cursor !== ""`). The empty string `""` and `undefined` are both falsy. When `response_metadata` is absent from the response (which happens with some Slack API edge cases), pagination stops after one page even if more data exists. Manifest declares `response_metadata.next_cursor` as the pagination control.

**Production impact:** Workspaces with more than 200 channels (one page) appear to have fewer channels than they actually do. Features like "search all channels" or "list all channels for admin" show incomplete results. The bug is intermittent because it depends on whether Slack includes `response_metadata` in the response (which varies by API version and response size).

---

## B15 — Race Condition (Read-then-Post)

**Violated rule:** `CTR-request-shape`

The client reads the channel list to find a channel by name, then posts to that channel. But between the read and the post, the channel can be archived by another user. The code does not handle the `is_archived` error that results.

```typescript
// slack-client-b15.ts — Race condition between list and post.

export class SlackClientB15 {
  private token: string;

  constructor(token: string) {
    this.token = token;
  }

  async findChannelByName(name: string): Promise<string | null> {
    try {
      const res = await fetch("https://slack.com/api/conversations.list", {
        headers: { Authorization: `Bearer ${this.token}` },
      });
      const data = await res.json();
      if (!data.ok) throw new Error(`Slack error: ${data.error}`);

      const channel = data.channels.find(
        (ch: { name: string }) => ch.name === name,
      );
      return channel?.id ?? null;
    } catch (err) {
      throw new Error(`findChannel failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }

  async postToChannelByName(
    channelName: string,
    text: string,
  ): Promise<{ channel: string; ts: string }> {
    // Step 1: Find channel ID by name
    const channelId = await this.findChannelByName(channelName);
    if (!channelId) {
      throw new Error(`Channel "${channelName}" not found`);
    }

    // BUG: RACE CONDITION — Between findChannelByName and postMessage,
    // the channel can be:
    // - Archived by an admin
    // - Deleted
    // - Made private (bot loses access)
    // - Bot removed from channel
    //
    // There is no protection against this time-of-check-to-time-of-use
    // (TOCTOU) gap.

    // Step 2: Post to channel (using stale channel ID)
    try {
      const res = await fetch("https://slack.com/api/chat.postMessage", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ channel: channelId, text }),
      });
      const data = await res.json();
      if (!data.ok) {
        // BUG: Does not handle is_archived or not_in_channel errors
        // that indicate the race condition occurred.
        // A retry-after-rejoin or fallback channel strategy is needed.
        throw new Error(`Slack error: ${data.error}`);
      }
      return { channel: data.channel, ts: data.ts };
    } catch (err) {
      throw new Error(`postMessage failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }
}

// Tests — no race condition coverage
import { describe, it, expect, vi } from "vitest";

describe("SlackClientB15", () => {
  it("posts to channel by name", async () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch");

    // Step 1: conversations.list returns the channel
    fetchSpy.mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          ok: true,
          channels: [
            { id: "C123", name: "general", is_archived: false },
            { id: "C456", name: "random", is_archived: false },
          ],
        }),
        { status: 200 },
      ),
    );

    // Step 2: chat.postMessage succeeds
    fetchSpy.mockResolvedValueOnce(
      new Response(
        JSON.stringify({ ok: true, channel: "C123", ts: "1234.5678" }),
        { status: 200 },
      ),
    );

    const client = new SlackClientB15("xoxb-token");
    const result = await client.postToChannelByName("general", "hello");
    expect(result.channel).toBe("C123");
  });

  // MISSING: Test where channel is archived between list and post
  // MISSING: Test where bot is removed from channel between list and post
  // MISSING: Test where channel is deleted between list and post
  // MISSING: Test where channel visibility changes between list and post
  // MISSING: Test for retry/recovery strategy on race condition errors
});
```

**Expected violation:** `CTR-request-shape` -- Time-of-check-to-time-of-use (TOCTOU) gap between `conversations.list` (read) and `chat.postMessage` (write). The channel state can change between these two calls. The client does not handle `is_archived`, `not_in_channel`, or `channel_not_found` errors that indicate the race condition occurred. No retry or fallback strategy is implemented.

**Production impact:** In active workspaces, channels are regularly archived, renamed, or modified. A scheduled message system that looks up channels by name before posting is vulnerable: an admin archives a channel seconds before the message is sent. The error `is_archived` propagates up as an unhandled `"Slack error: is_archived"` string (not a typed error), making it impossible for callers to implement recovery logic. In high-throughput systems (e.g., alert routing), this race condition causes alert loss during channel reorganizations.
