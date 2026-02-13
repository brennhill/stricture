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
