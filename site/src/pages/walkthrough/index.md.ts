import type { APIRoute } from "astro";

export const prerender = true;

const body = `# Stricture Walkthrough\n\nA step-by-step guide to running Stricture on a sample system map.\n\n## Steps\n1. Load topology and lineage annotations from repos.\n2. Select baseline release snapshot and compare current state.\n3. Review drift findings by severity with owning teams.\n4. Export machine-readable results (SARIF/JUnit) for CI.\n5. Apply overrides with expirations when policy exceptions are justified.\n\n## Included scenarios\n- Cross-repo schema mismatch.\n- External provider version drift.\n- Missing escalation contact metadata.\n- AI-generated change gated by invariants.\n\nVisit /walkthrough/ for the interactive view.\n`;

export const GET: APIRoute = () =>
  new Response(body, {
    headers: {
      "Content-Type": "text/markdown; charset=utf-8",
      "Content-Signal": "ai-train=yes, search=yes, ai-input=yes"
    }
  });
