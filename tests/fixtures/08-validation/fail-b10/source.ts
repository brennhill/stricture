// slack-client-b10.ts — No channel/user ID format validation.

export class SlackClientB10 {
  private token: string;

  constructor(token: string) {
    this.token = token;
  }

  // BUG: channel accepts any string, no C*/G*/D* format validation
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

  // BUG: userId accepts any string, no U*/W* format validation
  async getUserInfo(userId: string): Promise<{ id: string; name: string }> {
    try {
      const res = await fetch("https://slack.com/api/users.info", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ user: userId }),
      });
      const data = await res.json();
      if (!data.ok) throw new Error(`Slack error: ${data.error}`);
      return { id: data.user.id, name: data.user.name };
    } catch (err) {
      throw new Error(`getUserInfo failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }
}

// Tests
import { describe, it, expect, vi } from "vitest";

describe("SlackClientB10", () => {
  it("accepts any string as channel", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(
        JSON.stringify({ ok: true, channel: "C123", ts: "1234.5678" }),
        { status: 200 },
      ),
    );

    const client = new SlackClientB10("xoxb-token");
    // BUG: "#general" is a channel name, not an ID. Slack needs "C1234567890".
    // This passes because the mock doesn't validate the input.
    const result = await client.postMessage("#general", "hello");
    expect(result.channel).toBe("C123");
  });

  it("accepts any string as user ID", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(
        JSON.stringify({ ok: true, user: { id: "U123", name: "alice" } }),
        { status: 200 },
      ),
    );

    const client = new SlackClientB10("xoxb-token");
    // BUG: "alice" is a username, not a user ID. Slack needs "U1234567890".
    const user = await client.getUserInfo("alice");
    expect(user.name).toBe("alice");
  });
});
