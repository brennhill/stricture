// slack-client-b12.ts — Crashes on missing blocks field.

interface SlackMessageB12 {
  type: string;
  ts: string;
  user: string;
  text: string;
  blocks?: Array<{
    type: string;
    text?: { type: string; text: string };
  }>;
}

export class SlackClientB12 {
  private token: string;

  constructor(token: string) {
    this.token = token;
  }

  async getChannelHistory(channel: string): Promise<{
    messages: SlackMessageB12[];
    firstBlockText: string;
  }> {
    try {
      const res = await fetch("https://slack.com/api/conversations.history", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ channel }),
      });
      const data = await res.json();
      if (!data.ok) throw new Error(`Slack error: ${data.error}`);

      const messages: SlackMessageB12[] = data.messages;

      // BUG: Directly accesses blocks[0].text without null checks.
      // blocks is optional — plain text messages have no blocks.
      // blocks[0].text is also optional — divider blocks have no text.
      const firstMessage = messages[0];
      const firstBlockText = firstMessage.blocks[0].text.text;
      //                      ^^^^^^^^^^^^^^^^^^^ TypeError: Cannot read properties of undefined

      return { messages, firstBlockText };
    } catch (err) {
      throw new Error(`getHistory failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }
}

// Tests — only test with blocks present
import { describe, it, expect, vi } from "vitest";

describe("SlackClientB12", () => {
  it("fetches history with blocks", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          ok: true,
          messages: [
            {
              type: "message",
              ts: "1234.5678",
              user: "U001",
              text: "hello",
              blocks: [
                { type: "section", text: { type: "mrkdwn", text: "hello" } },
              ],
            },
          ],
          has_more: false,
        }),
        { status: 200 },
      ),
    );

    const client = new SlackClientB12("xoxb-token");
    const result = await client.getChannelHistory("C123");
    expect(result.firstBlockText).toBe("hello");
  });

  // MISSING: Test with message that has NO blocks (plain text message)
  // MISSING: Test with message that has blocks but first block has no text (divider)
  // MISSING: Test with empty messages array
});
