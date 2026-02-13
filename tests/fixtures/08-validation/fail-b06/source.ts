// slack-client-b06.ts — Response type missing optional blocks field.

// BUG: Message type is incomplete — missing blocks field
interface Message {
  type: string;
  ts: string;
  user: string;
  text: string;
  // MISSING: blocks?: Block[]
}

interface PostMessageResult {
  channel: string;
  ts: string;
  message: Message;
}

export class SlackClientB06 {
  private token: string;

  constructor(token: string) {
    this.token = token;
  }

  async postMessage(channel: string, text: string): Promise<PostMessageResult> {
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
        ts: data.ts,
        message: data.message as Message,
      };
    } catch (err) {
      throw new Error(`postMessage failed: ${err instanceof Error ? err.message : String(err)}`);
    }
  }
}

// Tests
import { describe, it, expect, vi } from "vitest";

describe("SlackClientB06", () => {
  it("sends a message", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          ok: true,
          channel: "C123",
          ts: "1234.5678",
          message: {
            type: "message",
            ts: "1234.5678",
            user: "U001",
            text: "hello",
            blocks: [{ type: "section", text: { type: "mrkdwn", text: "hello" } }],
          },
        }),
        { status: 200 },
      ),
    );

    const client = new SlackClientB06("xoxb-token");
    const result = await client.postMessage("C123", "hello");
    expect(result.message.text).toBe("hello");
    expect(result.message.type).toBe("message");
    // Cannot test blocks — type does not include them
  });
});
