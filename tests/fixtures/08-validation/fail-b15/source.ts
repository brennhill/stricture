// slack-client-b15.ts — Race condition between list and post.

export class SlackClientB15 {
  private token: string;

  constructor(token: string) {
    this.token = token;
  }

  async findChannelByName(name: string): Promise<string | null> {
    try {
      const res = await fetch("https://slack.com/api/conversations.list", {
        headers: { Authorization: `Bearer ${this.token}` },
      });
      const data = await res.json();
      if (!data.ok) throw new Error(`Slack error: ${data.error}`);

      const channel = data.channels.find(
        (ch: { name: string }) => ch.name === name,
      );
      return channel?.id ?? null;
    } catch (err) {
      throw new Error(`findChannel failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }

  async postToChannelByName(
    channelName: string,
    text: string,
  ): Promise<{ channel: string; ts: string }> {
    // Step 1: Find channel ID by name
    const channelId = await this.findChannelByName(channelName);
    if (!channelId) {
      throw new Error(`Channel "${channelName}" not found`);
    }

    // BUG: RACE CONDITION — Between findChannelByName and postMessage,
    // the channel can be:
    // - Archived by an admin
    // - Deleted
    // - Made private (bot loses access)
    // - Bot removed from channel
    //
    // There is no protection against this time-of-check-to-time-of-use
    // (TOCTOU) gap.

    // Step 2: Post to channel (using stale channel ID)
    try {
      const res = await fetch("https://slack.com/api/chat.postMessage", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ channel: channelId, text }),
      });
      const data = await res.json();
      if (!data.ok) {
        // BUG: Does not handle is_archived or not_in_channel errors
        // that indicate the race condition occurred.
        // A retry-after-rejoin or fallback channel strategy is needed.
        throw new Error(`Slack error: ${data.error}`);
      }
      return { channel: data.channel, ts: data.ts };
    } catch (err) {
      throw new Error(`postMessage failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }
}

// Tests — no race condition coverage
import { describe, it, expect, vi } from "vitest";

describe("SlackClientB15", () => {
  it("posts to channel by name", async () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch");

    // Step 1: conversations.list returns the channel
    fetchSpy.mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          ok: true,
          channels: [
            { id: "C123", name: "general", is_archived: false },
            { id: "C456", name: "random", is_archived: false },
          ],
        }),
        { status: 200 },
      ),
    );

    // Step 2: chat.postMessage succeeds
    fetchSpy.mockResolvedValueOnce(
      new Response(
        JSON.stringify({ ok: true, channel: "C123", ts: "1234.5678" }),
        { status: 200 },
      ),
    );

    const client = new SlackClientB15("xoxb-token");
    const result = await client.postToChannelByName("general", "hello");
    expect(result.channel).toBe("C123");
  });

  // MISSING: Test where channel is archived between list and post
  // MISSING: Test where bot is removed from channel between list and post
  // MISSING: Test where channel is deleted between list and post
  // MISSING: Test where channel visibility changes between list and post
  // MISSING: Test for retry/recovery strategy on race condition errors
});
