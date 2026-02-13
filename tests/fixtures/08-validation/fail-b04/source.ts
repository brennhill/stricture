// slack-client-b04.ts — Implementation handles errors, but tests don't.

export class SlackClientB04 {
  private token: string;

  constructor(token: string) {
    this.token = token;
  }

  async postMessage(channel: string, text: string): Promise<{
    channel: string;
    ts: string;
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
    if (!data.ok) throw new Error(`Slack error: ${data.error}`);
    return { channel: data.channel, ts: data.ts };
  }

  async getUserInfo(userId: string): Promise<{ id: string; name: string }> {
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
  }
}

// Tests — only happy path
import { describe, it, expect, vi } from "vitest";

describe("SlackClientB04", () => {
  it("sends a message successfully", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(
        JSON.stringify({ ok: true, channel: "C123", ts: "1234.5678" }),
        { status: 200 },
      ),
    );

    const client = new SlackClientB04("xoxb-token");
    const result = await client.postMessage("C123", "hello");
    expect(result.channel).toBe("C123");
  });

  it("fetches user info successfully", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          ok: true,
          user: { id: "U123", name: "alice" },
        }),
        { status: 200 },
      ),
    );

    const client = new SlackClientB04("xoxb-token");
    const user = await client.getUserInfo("U123");
    expect(user.name).toBe("alice");
  });

  // MISSING: No negative tests for any of these error scenarios:
  // - channel_not_found
  // - not_in_channel
  // - is_archived
  // - invalid_auth
  // - token_revoked
  // - ratelimited (429)
  // - user_not_found
  // - network failures
  // - msg_too_long
});
