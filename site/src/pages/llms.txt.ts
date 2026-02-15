import type { APIRoute } from "astro";
import { comparisons } from "../data/comparisons";

export const prerender = true;

const siteBase = process.env.SITE_BASE_URL ?? "https://stricture-lint.com";

const coreRoutes = [
  "/",
  "/demo/",
  "/open-standard/",
  "/architecture-invariants/",
  "/walkthrough/",
  "/service-internals-demo/",
  "/what-is-stricture/",
  "/with-ai/",
  "/when-to-use-stricture/"
];

function markdownUrls(): string[] {
  const comparisonRoutes = comparisons.map((entry) => `/when-to-use-stricture/${entry.slug}/`);
  const routes = [...coreRoutes, ...comparisonRoutes];

  return routes.map((route) => {
    const normalized = route.endsWith("/") ? route : `${route}/`;
    return new URL(`${normalized}index.md`, siteBase).toString();
  });
}

function buildResponseBody() {
  return markdownUrls().join("\n") + "\n";
}

export const GET: APIRoute = () =>
  new Response(buildResponseBody(), {
    headers: {
      "Content-Type": "text/plain; charset=utf-8",
      "Content-Signal": "ai-train=yes, search=yes, ai-input=yes"
    }
  });
