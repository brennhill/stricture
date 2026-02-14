import type { APIRoute } from "astro";
import { comparisons } from "../../data/comparisons";

export const prerender = true;

function renderComparisonList() {
  return comparisons
    .map((entry) => `- ${entry.tool}: ${entry.oneLineDifference} (see /why-not/${entry.slug}/)`)
    .join("\n");
}

const body = `# Why Not Another Tool?\n\nStricture overlays with popular standards and tools instead of replacing them. This page links to detailed comparisons.\n\n## Comparisons\n${renderComparisonList()}\n`;

export const GET: APIRoute = () =>
  new Response(body, {
    headers: {
      "Content-Type": "text/markdown; charset=utf-8",
      "Content-Signal": "ai-train=yes, search=yes, ai-input=yes"
    }
  });
