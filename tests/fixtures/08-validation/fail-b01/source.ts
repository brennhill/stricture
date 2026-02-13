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
