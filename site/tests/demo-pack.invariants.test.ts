import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

interface ServiceNode {
  id: string;
  name: string;
}

interface Edge {
  from: string;
  to: string;
  fieldId: string;
}

function normalizeToken(value: string): string {
  return String(value || "").toLowerCase().replace(/[^a-z0-9]/g, "");
}

function loadPack() {
  const packPath = resolve(process.cwd(), "public/demo/demo-pack.json");
  return JSON.parse(readFileSync(packPath, "utf8"));
}

function resolveServiceID(token: string, services: ServiceNode[]): string {
  const want = normalizeToken(token);
  if (!want) return "";
  for (const service of services) {
    if (normalizeToken(service.id) === want) return service.id;
    if (normalizeToken(service.name) === want) return service.id;
  }
  return "";
}

describe("demo pack invariants", () => {
  it("edges only reference known services", () => {
    const pack = loadPack();
    const ids = new Set<string>((pack.services || []).map((service: ServiceNode) => service.id));
    (pack.edges || []).forEach((edge: Edge) => {
      expect(ids.has(edge.from), `edge.from ${edge.from} should exist in services`).toBe(true);
      expect(ids.has(edge.to), `edge.to ${edge.to} should exist in services`).toBe(true);
    });
  });

  it("mutation fields are present in scenarios and edges", () => {
    const pack = loadPack();
    const edgeFields = new Set<string>((pack.edges || []).map((edge: Edge) => edge.fieldId));
    Object.entries(pack.fields_by_mutation || {}).forEach(([mutationType, fields]) => {
      (fields as string[]).forEach((fieldId) => {
        expect(edgeFields.has(fieldId), `${mutationType}:${fieldId} should appear in topology edges`).toBe(true);
        expect(pack.mutation_scenarios?.[fieldId], `${mutationType}:${fieldId} should have scenario entries`).toBeTruthy();
      });
    });
  });

  it("scenario source/impact services are resolvable", () => {
    const pack = loadPack();
    const services = pack.services as ServiceNode[];
    Object.values(pack.mutation_scenarios || {}).forEach((byMutation: any) => {
      Object.values(byMutation || {}).forEach((scenario: any) => {
        (scenario.changes || []).forEach((change: any) => {
          if (change.source?.service) {
            expect(
              resolveServiceID(change.source.service, services),
              `source service ${change.source.service} should resolve`,
            ).not.toBe("");
          }
          if (change.impact?.service) {
            expect(
              resolveServiceID(change.impact.service, services),
              `impact service ${change.impact.service} should resolve`,
            ).not.toBe("");
          }
        });
      });
    });
  });
});
