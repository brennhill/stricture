// B15: Realtime subscription handler registered AFTER join message sent.
// Messages that arrive between WebSocket open and handler registration are lost.

class SupabaseClientB15 {
  private readonly baseUrl: string;
  private readonly anonKey: string;

  constructor(baseUrl: string, anonKey: string) {
    this.baseUrl = baseUrl;
    this.anonKey = anonKey;
  }

  subscribeToChanges(
    table: string,
    callback: (message: {
      type: string;
      record: Record<string, unknown> | null;
    }) => void
  ): { unsubscribe: () => void } {
    const wsUrl = this.baseUrl
      .replace("https://", "wss://")
      .replace("http://", "ws://");

    const ws = new WebSocket(
      `${wsUrl}/realtime/v1/websocket?apikey=${this.anonKey}&vsn=1.0.0`
    );

    const topic = `realtime:public:${table}`;

    ws.onopen = () => {
      // Step 1: Send join message
      ws.send(JSON.stringify({
        topic,
        event: "phx_join",
        payload: {
          config: {
            broadcast: { self: false },
            presence: {},
            postgres_changes: [
              { event: "*", schema: "public", table },
            ],
          },
        },
        ref: "1",
      }));

      // BUG: Message handler is registered AFTER the join is sent.
      // In a fast server scenario, the phx_reply (join ack) and
      // postgres_changes events can arrive BEFORE onmessage is set.
      //
      // Timeline:
      //   T0: ws.send(join) -- join request sent
      //   T1: Server processes join, queues phx_reply + buffered changes
      //   T2: Server sends phx_reply
      //   T3: Server sends INSERT notification (another client inserted a row)
      //   T4: ws.onmessage = handler -- TOO LATE, T2 and T3 are lost
      //
      // This is especially bad when:
      // (a) The table has a trigger that fires on subscription
      // (b) A batch insert happens concurrently with subscription setup
      // (c) The server has low latency and responds before JS event loop yields
    };

    // BUG: Handler registered inside onopen, after send().
    // Should be registered BEFORE open or at minimum before send().
    ws.onopen = () => {
      // This overwrites the previous onopen handler!
      // BUG: The join message from the first onopen is now lost because
      // this second assignment replaces it entirely.
    };

    // Even if we fix the double-assignment, the handler is still too late:
    setTimeout(() => {
      ws.onmessage = (event: MessageEvent) => {
        let msg: Record<string, unknown>;
        try {
          msg = JSON.parse(String(event.data)) as Record<string, unknown>;
        } catch {
          return;
        }

        if (String(msg.event) === "postgres_changes") {
          const payload = msg.payload as Record<string, unknown>;
          const data = payload?.data as Record<string, unknown>;
          if (!data) return;

          callback({
            type: String(data.type),
            record: (data.record as Record<string, unknown> | null) ?? null,
          });
        }
      };
    }, 0); // setTimeout(fn, 0) defers to next microtask -- too late

    // BUG: No heartbeat. Phoenix Channels disconnect after missing
    // heartbeats (default timeout ~30-60 seconds). Without a heartbeat
    // interval, the connection silently drops and no more events arrive.
    // The caller has no way to know the subscription is dead.

    // BUG: No error handler on ws. WebSocket errors are silently swallowed.

    return {
      unsubscribe: () => {
        // BUG: Does not send phx_leave before closing.
        // The server does not know the client unsubscribed and may
        // continue buffering events.
        ws.close();
      },
    };
  }
}

// Test -- does not test race condition
import { describe, it, expect, vi } from "vitest";

describe("subscribeToChanges", () => {
  it("subscribes to a table", () => {
    const client = new SupabaseClientB15("https://test.supabase.co", "anon-key");
    const callback = vi.fn();

    // BUG: This test creates the subscription but never simulates
    // messages arriving before the handler is registered.
    // It only verifies the subscription object is returned.
    const sub = client.subscribeToChanges("users", callback);
    expect(sub).toBeDefined();
    expect(sub.unsubscribe).toBeDefined();

    // BUG: Never tests:
    // - Messages arriving during join handshake
    // - Connection errors
    // - Heartbeat timeout / silent disconnect
    // - phx_leave on unsubscribe
    // - Multiple concurrent subscriptions
  });
});
