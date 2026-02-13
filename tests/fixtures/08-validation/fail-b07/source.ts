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
