import type { APIRoute } from "astro";

export const prerender = true;

const body = `# Architecture Invariants\n\nStricture enforces non-negotiable system rules in CI/CD so risky deploys are blocked early.\n\n## Invariant themes\n- Dependency direction safeguards between services.\n- Provenance completeness for external and internal sources.\n- Source-version compatibility and freshness windows.\n- Escalation routing present for every critical field.\n- Flow-tier controls for business-critical paths (for example checkout level 1).\n\n## Adoption path\n1. Start with high-risk response fields.\n2. Encode dependency and provenance invariants.\n3. Run in warn mode for one release to collect evidence.\n4. Move to block mode with severity + flow-tier policy and tightly scoped overrides.\n\nSee /architecture-invariants/ for the full walkthrough and examples.\n`;

export const GET: APIRoute = () =>
  new Response(body, {
    headers: {
      "Content-Type": "text/markdown; charset=utf-8",
      "Content-Signal": "ai-train=yes, search=yes, ai-input=yes"
    }
  });
