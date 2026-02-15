import type { APIRoute } from "astro";

export const prerender = true;

const body = `# Stricture Annotation Guide + Index\n\nThis guide explains exactly what gets annotated and why.\n\n## Object types\n- \\\`strict:source\\\` -> field lineage annotation (\\\`stricture-source\\\`) in code comments.\n- \\\`strict:source-edge\\\` -> source edge refs inside \\\`sources=\\\` values.\n- \\\`strict:systems[]\\\` -> service registry records (\\\`lineage-systems.yml\\\`) for owner/escalation/flow memberships.\n- \\\`strict:flows[]\\\` -> flow-tier catalog entries at registry top-level.\n- \\\`strict:lineage-override\\\` -> temporary lineage overrides (\\\`stricture-lineage-override\\\`).\n\n## Key reference\n- A-Z index of field annotation keys.\n- A-Z index of source edge query keys.\n- A-Z index of override keys.\n- A-Z index of service registry keys.\n- A-Z index of flow registry keys.\n\nIncludes required vs optional status, default values, and mode labels:\n- Defaulted\n- Auto-gen\n- Strongly recommended\n- Manual + required\n- Manual + optional\n\nPolicy packs can promote recommended fields to required and override defaults.\nRepositories bind policy using \\\`strict:policy_url\\\` (optionally pinned with \\\`strict:policy_sha256\\\`).\n\nSystem hierarchy convention (no new key):\n- Topology system: \\\`location-tracking-service\\\`\n- Internal subsystem: \\\`location-tracking-service:tracking-api\\\`\n\nTool-assisted authoring is expected for high-churn keys like:\n- \\\`field\\\` (derived from code/schema path, with fallback from \\\`field_id\\\`)\n- \\\`source_system\\\` (derived from repo/service map, preferably via \\\`strict:server_url\\\`)\n- \\\`source_version\\\` (derived from contract ref commit/tag)\n- \\\`sources\\\` (inferred from code + schemas)\n- external \\\`provider_id\\\` and \\\`as_of\\\`\n- service registry \\\`id\\\` (bootstrapped when \\\`strict:server_url\\\` is configured)\n- service flow memberships from policy/server catalogs where available\n\nRead the full page at /annotations/.\n`;

export const GET: APIRoute = () =>
  new Response(body, {
    headers: {
      "Content-Type": "text/markdown; charset=utf-8",
      "Content-Signal": "ai-train=yes, search=yes, ai-input=yes"
    }
  });
