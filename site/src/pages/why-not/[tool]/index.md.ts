import type { APIRoute } from "astro";
import { comparisons } from "../../../data/comparisons";

export const prerender = true;

export function getStaticPaths() {
  return comparisons.map((entry) => ({
    params: { tool: entry.slug },
    props: { target: `/when-to-use-stricture/${entry.slug}/` },
  }));
}

export const GET: APIRoute = ({ props }) => {
  const target = String((props as { target?: string }).target || "/when-to-use-stricture/");
  return new Response(`# Moved\n\nThis page moved to ${target}\n`, {
    headers: {
      "Content-Type": "text/markdown; charset=utf-8",
      "Content-Signal": "ai-train=yes, search=yes, ai-input=yes",
    },
  });
};
