import type { APIRoute } from "astro";

export const prerender = true;

const body = `# Stricture Open Standard\n\nStricture Open Standard (SOS) is a contract for field-level lineage, drift classes, overlays, and policy semantics. It is implementation-neutral and licensed CC BY 4.0.\n\n## Key documents\n- Charter: https://github.com/brennhill/stricture/blob/main/SPEC-CHARTER.md\n- v0.1 Draft: https://github.com/brennhill/stricture/blob/main/spec/0.1-draft.md\n- License split: https://github.com/brennhill/stricture/blob/main/LICENSES.md\n\n## Principles\n- Separate interoperability contract (SOS) from engine implementation.\n- Support overlays for OpenAPI, OpenTelemetry, OpenLineage, AsyncAPI, and Buf/Protobuf.\n- Keep metadata authored once and reused across profiles.\n\n## Implementation\n- Reference CLI: https://github.com/brennhill/stricture (AGPL-3.0).\n- Conformance fixtures and profile overlays define behavior.\n`;

export const GET: APIRoute = () =>
  new Response(body, {
    headers: {
      "Content-Type": "text/markdown; charset=utf-8",
      "Content-Signal": "ai-train=yes, search=yes, ai-input=yes"
    }
  });
