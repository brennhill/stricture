// slack-client-b14.ts — Pagination bug: stops after first page.

export class SlackClientB14 {
  private token: string;

  constructor(token: string) {
    this.token = token;
  }

  async listAllChannels(): Promise<Array<{ id: string; name: string }>> {
    const allChannels: Array<{ id: string; name: string }> = [];
    let cursor: string | undefined;

    try {
      do {
        const url = new URL("https://slack.com/api/conversations.list");
        url.searchParams.set("limit", "200");
        if (cursor) {
          url.searchParams.set("cursor", cursor);
        }

        const res = await fetch(url.toString(), {
          headers: { Authorization: `Bearer ${this.token}` },
        });
        const data = await res.json();
        if (!data.ok) throw new Error(`Slack error: ${data.error}`);

        allChannels.push(
          ...data.channels.map((ch: { id: string; name: string }) => ({
            id: ch.id,
            name: ch.name,
          })),
        );

        // BUG: Treats cursor as truthy/falsy.
        // Slack returns next_cursor: "" when there are no more pages.
        // But "" is falsy in JavaScript, so this ALSO stops when
        // next_cursor is a real cursor value that happens to be
        // the first page's response_metadata.next_cursor.
        //
        // The real bug: code only fetches first page and stops
        // because it assigns cursor from response_metadata?.next_cursor
        // and then checks `while (cursor)` — but if next_cursor
        // is present and non-empty, it should continue.
        cursor = data.response_metadata?.next_cursor;
        // When next_cursor is "" (empty string), cursor becomes ""
        // which is falsy — loop correctly stops.
        // BUT the loop below uses `while (cursor)` which also
        // means: if response_metadata is undefined (no pagination
        // info), cursor is undefined, loop stops after first page
        // even if there IS more data (has_more: true).
      } while (cursor);
      //       ^^^^^^ BUG: "" is falsy, undefined is falsy.
      //       Should check: cursor !== undefined && cursor !== ""
      //       to distinguish "no more pages" from "no pagination metadata"
    } catch (err) {
      throw new Error(`listChannels failed: ${err instanceof Error ? err.message : String(err)}`);
    }

    return allChannels;
  }
}

// Tests — only test single page
import { describe, it, expect, vi } from "vitest";

describe("SlackClientB14", () => {
  it("fetches channels — but only first page", async () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch");

    // Page 1 with cursor
    fetchSpy.mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          ok: true,
          channels: [{ id: "C001", name: "general" }],
          response_metadata: { next_cursor: "cursor_page2" },
        }),
        { status: 200 },
      ),
    );

    // Page 2 — end of pagination
    fetchSpy.mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          ok: true,
          channels: [{ id: "C002", name: "random" }],
          response_metadata: { next_cursor: "" },
        }),
        { status: 200 },
      ),
    );

    const client = new SlackClientB14("xoxb-token");
    const channels = await client.listAllChannels();

    // NOTE: This test actually passes because while("cursor_page2") is truthy.
    // The real scenario that fails is when response_metadata is missing entirely
    // on page 1 (some older Slack API responses omit it).
    // But the test only covers the standard case.
    expect(channels).toHaveLength(2);

    // MISSING test: response without response_metadata field at all
    // In that case, cursor = undefined, loop stops at 1 page
  });

  // MISSING: Test with 3+ pages
  // MISSING: Test where response_metadata is completely absent
  // MISSING: Test boundary: exactly at cursor "" vs missing metadata
});
