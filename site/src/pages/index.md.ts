import type { APIRoute } from "astro";

export const prerender = true;

const meta = {
  title: "Stricture | Lineage-Aware API Linting",
  tagline: "Detect API drift, enforce field provenance, and block risky releases with escalation-ready evidence."
};

const heroPoints = [
  "Detect API drift before it becomes an incident.",
  "Keep AI-generated changes inside enforceable guardrails.",
  "Prevent one repo change from breaking ten systems.",
  "Ship faster without losing contract integrity.",
  "Trace each output field back to its source systems with versions and freshness." 
];

const coreCapabilities = [
  "Field-level lineage with multi-source annotations.",
  "Cross-repo drift detection on schema, enums, and provider versions.",
  "Service-level flow tier policy controls for business-critical paths.",
  "Warn/block/override policy gates wired into CI/CD.",
  "Ownership and escalation chains attached to every drifting field.",
  "Architecture invariants for dependency direction and provenance completeness.",
  "Machine-readable outputs including SARIF and JUnit formats." 
];

const workflow = [
  "Ingest service topology and field annotations from source repos and APIs.",
  "Map overlays to OpenAPI, OpenTelemetry, OpenLineage, AsyncAPI, and Buf/Protobuf schemas.",
  "Export lineage artifacts, compare against baseline snapshots, and classify drift severity.",
  "Enforce warn/block gates with flow-tier policy, override expirations, and escalation routing.",
  "Track source-system versions and as-of freshness for internal and external providers." 
];

function renderList(items: string[]): string {
  return items.map((item) => `- ${item}`).join("\n");
}

export const GET: APIRoute = () => {
  const body = `# ${meta.title}\n\n${meta.tagline}\n\n## Highlights\n${renderList(heroPoints)}\n\n## Core Capabilities\n${renderList(coreCapabilities)}\n\n## How It Works\n${workflow.map((step, idx) => `${idx + 1}. ${step}`).join("\n")}\n\n## Explore More\n- /demo/\n- /service-internals-demo/\n- /open-standard/\n- /architecture-invariants/\n- /annotations/\n- /walkthrough/\n- /what-is-stricture/\n- /with-ai/\n- /when-to-use-stricture/\n`;

  return new Response(body, {
    headers: {
      "Content-Type": "text/markdown; charset=utf-8",
      "Content-Signal": "ai-train=yes, search=yes, ai-input=yes"
    }
  });
};
