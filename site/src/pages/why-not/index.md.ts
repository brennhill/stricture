import type { APIRoute } from "astro";

export const prerender = true;

const target = "/when-to-use-stricture/";

export const GET: APIRoute = () =>
  new Response(`# Moved\n\nThis page moved to ${target}\n`, {
    headers: {
      "Content-Type": "text/markdown; charset=utf-8",
      "Content-Signal": "ai-train=yes, search=yes, ai-input=yes",
    },
  });
