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
