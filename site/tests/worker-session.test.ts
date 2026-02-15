import { describe, expect, it } from "vitest";

import { DemoSession, __test as worker } from "../worker/src/index";

class MemoryStorage {
  private readonly data = new Map<string, unknown>();

  async get<T>(key: string): Promise<T | undefined> {
    return this.data.get(key) as T | undefined;
  }

  async put(key: string, value: unknown): Promise<void> {
    this.data.set(key, value);
  }
}

function createSession(): DemoSession {
  const ctx = { storage: new MemoryStorage() } as any;
  return new DemoSession(ctx, {} as any);
}

function rootID(value: string): string {
  const idx = String(value || "").indexOf(":");
  return idx >= 0 ? value.slice(0, idx) : value;
}

describe("demo worker session", () => {
  it("accepts valid mutation payloads across all mutation types", async () => {
    const session = createSession();
    const init = await session.fetch(new Request("https://demo/init", { method: "POST" }));
    expect(init.status).toBe(200);
    const initPayload = await init.json() as any;
    const snapshot = initPayload.snapshot;

    for (const mutationType of snapshot.mutationTypes || []) {
      const fieldId = (snapshot.fieldsByMutation?.[mutationType] || [])[0];
      expect(fieldId, `mutation type ${mutationType} should have at least one field`).toBeTruthy();
      const sourceEdge = (snapshot.edges || []).find((edge: any) => edge.fieldId === fieldId);
      expect(sourceEdge, `field ${fieldId} should have a source edge`).toBeTruthy();
      const serviceId = rootID(sourceEdge.from);

      const response = await session.fetch(new Request("https://demo/mutations", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify({ type: mutationType, serviceId, fieldId }),
      }));
      expect(response.status, `mutation ${mutationType}:${fieldId} should be accepted`).toBe(200);
    }
  });

  it("rejects invalid mutation type/field combinations", async () => {
    const session = createSession();
    await session.fetch(new Request("https://demo/init", { method: "POST" }));
    const response = await session.fetch(new Request("https://demo/mutations", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ type: "enum_changed", serviceId: "commercegateway", fieldId: "response_unknown_field" }),
    }));
    expect(response.status).toBe(400);
    expect(await response.text()).toContain("mutation requires valid type and fieldId");
  });

  it("reports zero findings in default snapshot before mutations", () => {
    const snapshot = worker.defaultSnapshot();
    const ran = worker.runEngine(snapshot);
    expect(ran.findings.length).toBe(0);
    expect(ran.runSummary.findingCount).toBe(0);
  });
});
