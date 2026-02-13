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
