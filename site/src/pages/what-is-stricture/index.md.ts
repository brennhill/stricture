import type { APIRoute } from "astro";

export const prerender = true;

const body = `# What Is Stricture?\n\nStricture is a lineage-aware API linting engine that enforces contract integrity, provenance, and escalation readiness before deploy.\n\n## Problems it solves\n- Silent drift across multiple repos and providers.\n- Missing provenance or freshness on critical response fields.\n- AI-generated changes that bypass review.\n- Lack of clear owners and escalation paths when data quality breaks.\n\n## How it works\n- Annotate fields with lineage, source versions, and contacts.\n- Compare current state to a baseline snapshot.\n- Classify drift by severity; gate deploys in warn or block mode.\n- Apply service-level flow-tier policy (for example checkout level 1 hard-blocks).\n- Export evidence for CI, governance, and observability overlays.\n\n## Profiles and overlays\n- OpenAPI, OpenTelemetry, OpenLineage, AsyncAPI, Buf/Protobuf.\n- Single annotation set emits to multiple ecosystems.\n\nLearn more at /what-is-stricture/.\n`;

export const GET: APIRoute = () =>
  new Response(body, {
    headers: {
      "Content-Type": "text/markdown; charset=utf-8",
      "Content-Signal": "ai-train=yes, search=yes, ai-input=yes"
    }
  });
