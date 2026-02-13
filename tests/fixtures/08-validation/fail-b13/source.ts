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
