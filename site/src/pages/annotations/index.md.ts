import type { APIRoute } from "astro";

export const prerender = true;

const body = `# Stricture Annotation Guide + Index\n\nThis guide explains exactly what gets annotated and why.\n\n## Object types\n- Field lineage annotation (\\\`stricture-source\\\`) in code comments.\n- Source edge refs inside \\\`sources=\\\` values.\n- Service registry records (\\\`lineage-systems.yml\\\`) for owner/escalation defaults.\n- Temporary lineage overrides (\\\`stricture-lineage-override\\\`).\n\n## Key reference\n- A-Z index of field annotation keys.\n- A-Z index of source edge query keys.\n- A-Z index of override keys.\n- A-Z index of service registry keys.\n\nIncludes required vs optional status and default values.\n\nRead the full page at /annotations/.\n`;

export const GET: APIRoute = () =>
  new Response(body, {
    headers: {
      "Content-Type": "text/markdown; charset=utf-8",
      "Content-Signal": "ai-train=yes, search=yes, ai-input=yes"
    }
  });
