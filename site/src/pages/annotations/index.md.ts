import type { APIRoute } from "astro";

export const prerender = true;

const body = `# Stricture Annotation Guide + Index\n\nThis guide explains exactly what gets annotated and why.\n\n## Object types\n- \\\`strict-source\\\` -> field lineage annotation in code comments.\n- Source entries -> structured source refs inside annotation blocks or sidecar files.\n- \\\`systems[]\\\` -> service registry records (\\\`strict-lineage.yml\\\` or dedicated registry file) for owner/escalation/runbook/docs/flow metadata.\n- \\\`flows[]\\\` -> flow-tier catalog entries at registry top-level.\n- \\\`strict-lineage-override\\\` -> temporary lineage overrides.\n\n## Key reference\n- A-Z index of field annotation keys.\n- A-Z index of source entry fields.\n- A-Z index of override keys.\n- A-Z index of service registry keys.\n- A-Z index of flow registry keys.\n\nIncludes required vs optional status, default values, and mode labels:\n- Defaulted\n- Auto-inferred\n- Recommended\n- Required\n- Optional\n\nPolicy packs can promote recommended fields to required and override defaults.\nRepositories bind policy using \\\`stricture_policy_url\\\` (optionally pinned with \\\`stricture_policy_sha256\\\`).\n\nSystem hierarchy convention (no new key):\n- Topology system: \\\`location-tracking-service\\\`\n- Internal subsystem: \\\`location-tracking-service:tracking-api\\\`\n\nTool-assisted authoring is expected for high-churn keys like:\n- \\\`field\\\` (derived from code/schema path, with fallback from \\\`field_id\\\`)\n- \\\`from\\\` / \\\`source_system\\\` (derived from repo/service map)\n- \\\`source_version\\\` (derived from contract ref commit/tag)\n- \\\`sources\\\` (inferred from code + schemas)\n- external \\\`provider\\\` and \\\`as_of\\\`\n- service registry \\\`id\\\` (bootstrapped from sidecar or server)\n- service flow memberships from policy/server catalogs where available\n\nRead the full page at /annotations/.\n`;

export const GET: APIRoute = () =>
  new Response(body, {
    headers: {
      "Content-Type": "text/markdown; charset=utf-8",
      "Content-Signal": "ai-train=yes, search=yes, ai-input=yes"
    }
  });
