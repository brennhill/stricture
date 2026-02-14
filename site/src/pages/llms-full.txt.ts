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
  "/what-is-stricture/",
  "/with-ai/",
  "/why-not/"
];

function markdownUrls(): string[] {
  const comparisonRoutes = comparisons.map((entry) => `/why-not/${entry.slug}/`);
  const routes = [...coreRoutes, ...comparisonRoutes];

  return routes.map((route) => {
    const normalized = route.endsWith("/") ? route : `${route}/`;
    return new URL(`${normalized}index.md`, siteBase).toString();
  });
}

function htmlUrls(): string[] {
  const comparisonRoutes = comparisons.map((entry) => `/why-not/${entry.slug}/`);
  const routes = [...coreRoutes, ...comparisonRoutes];

  return routes.map((route) => new URL(route, siteBase).toString());
}

function buildResponseBody() {
  return [...markdownUrls(), ...htmlUrls()].join("\n") + "\n";
}

export const GET: APIRoute = () =>
  new Response(buildResponseBody(), {
    headers: {
      "Content-Type": "text/plain; charset=utf-8",
      "Content-Signal": "ai-train=yes, search=yes, ai-input=yes"
    }
  });
