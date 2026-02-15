import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

import { __test as demo } from "../src/scripts/demo-app.js";

interface SnapshotShape {
  services: Array<{ id: string; name: string }>;
  edges: Array<{ id: string; from: string; to: string; fieldId: string; status: string; label: string }>;
  fieldsByMutation: Record<string, string[]>;
}

function loadSnapshot(): SnapshotShape {
  const packPath = resolve(process.cwd(), "public/demo/demo-pack.json");
  const pack = JSON.parse(readFileSync(packPath, "utf8"));
  return {
    services: pack.services,
    edges: (pack.edges || []).map((edge: any) => ({ ...edge, status: edge.status || "healthy" })),
    fieldsByMutation: pack.fields_by_mutation,
  };
}

describe("demo controls", () => {
  it("every service root has selectable mutation fields", () => {
    const snapshot = loadSnapshot();
    const fieldMutationMap = demo.buildFieldMutationMap(snapshot);
    const roots = [...new Set(snapshot.services.map((service) => demo.topologyRootId(service.id)).filter(Boolean))];
    roots.forEach((root) => {
      const fields = demo.fieldsForService(snapshot, root, fieldMutationMap);
      expect(fields.length, `root ${root} should have at least one selectable field`).toBeGreaterThan(0);
    });
  });

  it("every mutated field can be selected from its source root", () => {
    const snapshot = loadSnapshot();
    const fieldMutationMap = demo.buildFieldMutationMap(snapshot);
    [...fieldMutationMap.keys()].forEach((fieldId) => {
      const edge = snapshot.edges.find((row) => row.fieldId === fieldId);
      expect(edge, `field ${fieldId} should have at least one edge`).toBeTruthy();
      const sourceRoot = demo.topologyRootId(edge!.from);
      const fields = demo.fieldsForService(snapshot, sourceRoot, fieldMutationMap);
      expect(fields.includes(fieldId), `field ${fieldId} should be selectable for ${sourceRoot}`).toBe(true);
    });
  });

  it("classifies escalation lanes as primary path vs secondary context", () => {
    const snapshot = {
      services: [
        { id: "a", name: "A" },
        { id: "b", name: "B" },
        { id: "c", name: "C" },
        { id: "d", name: "D" },
      ],
      edges: [
        { id: "e1", from: "a", to: "b", fieldId: "response_x", status: "healthy", label: "" },
        { id: "e2", from: "b", to: "c", fieldId: "response_x", status: "healthy", label: "" },
        { id: "e3", from: "a", to: "d", fieldId: "response_x", status: "healthy", label: "" },
      ],
      fieldsByMutation: {},
      findings: [],
      mutations: [],
    } as any;

    const finding = {
      id: "f1",
      fieldId: "response_x",
      serviceId: "c",
      source: { serviceId: "a" },
      impact: { serviceId: "c" },
    } as any;

    const lanes = demo.classifyEscalationServiceIds(snapshot, finding);
    expect(lanes.primary).toEqual(expect.arrayContaining(["a", "b", "c"]));
    expect(lanes.secondary).toEqual(expect.arrayContaining(["d"]));
  });
});
