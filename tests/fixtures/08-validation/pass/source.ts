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
