import type { APIRoute } from "astro";

export const prerender = true;

const body = `# Stricture Live Demo\n\nThe interactive demo runs the Stricture reference engine against a synthetic multi-service topology.\n\n## What you can try\n- Toggle warn vs block policy gates.\n- Test flow-tier criticality behavior (service-level flow memberships).\n- Rerun Stricture to see drift classifications update.\n- Inspect lineage for API fields and their upstream sources.\n- View escalation contacts and optional runbook/docs links attached to failing fields.\n\n## Demo contents\n- Logistics, fintech, media, and ecommerce services.\n- Cross-repo schema drift scenarios.\n- Source-version and as-of freshness checks.\n- Downstream-impact-gated findings and self-only change tracking.\n\nOpen the demo at /demo/.\n`;

export const GET: APIRoute = () =>
  new Response(body, {
    headers: {
      "Content-Type": "text/markdown; charset=utf-8",
      "Content-Signal": "ai-train=yes, search=yes, ai-input=yes"
    }
  });
