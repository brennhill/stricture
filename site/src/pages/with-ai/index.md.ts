import type { APIRoute } from "astro";

export const prerender = true;

const body = `# Stricture With AI\n\nStricture keeps AI-generated changes inside enforceable guardrails.\n\n## Controls\n- Require provenance annotations on AI-touched fields.\n- Enforce source-version compatibility after AI edits.\n- Apply downstream-impact gating and flow-tier policy for business-critical paths.\n- Compute defaults (owner/escalation/merge metadata) from policy/server catalogs when available.\n- Resolve service metadata (including optional runbook/docs links) from \`strict:systems[]\`.\n- Emit machine-readable findings for downstream AI safety reviews.\n\n## Example flow\n1. AI proposes schema change.\n2. Stricture validates lineage, freshness, flow context, and invariants.\n3. Drift surfaced in CI with cause, impact, owners, and policy rationale.\n4. Temporary override allowed only with expiration and reviewer trace.\n\nMore details at /with-ai/.\n`;

export const GET: APIRoute = () =>
  new Response(body, {
    headers: {
      "Content-Type": "text/markdown; charset=utf-8",
      "Content-Signal": "ai-train=yes, search=yes, ai-input=yes"
    }
  });
