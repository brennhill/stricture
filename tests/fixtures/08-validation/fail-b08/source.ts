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
