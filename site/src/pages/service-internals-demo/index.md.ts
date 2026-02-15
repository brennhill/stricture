import type { APIRoute } from "astro";

export const prerender = true;

const body = `# Service Internals Demo\n\nThis focused page preloads the live demo in Service Internals mode.\n\n- URL: /service-internals-demo/\n- Preloaded view: /demo?view=service&service=logisticsgateway\n- Use this when you want to show subsystem topology without manually selecting a node first.\n`;

export const GET: APIRoute = () =>
  new Response(body, {
    headers: {
      "Content-Type": "text/markdown; charset=utf-8",
      "Content-Signal": "ai-train=yes, search=yes, ai-input=yes",
    },
  });
