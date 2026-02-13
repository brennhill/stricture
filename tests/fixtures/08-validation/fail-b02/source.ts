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
