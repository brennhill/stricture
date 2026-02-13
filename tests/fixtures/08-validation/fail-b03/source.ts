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
