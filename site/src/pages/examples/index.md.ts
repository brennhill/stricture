import type { APIRoute } from "astro";

export const prerender = true;

const body = `# Stricture Examples

Practical examples across Java, protobuf/buf, Go, Python, and Node.js.

## What this page shows
- Service-level annotations in \`strict:systems[]\` (owner, escalation, runbook, docs, flows)
- API/field-level lineage with \`stricture-source\` comments
- Reuse of existing metadata from OpenAPI and OpenTelemetry
- Inline comments showing where Stricture is "batteries included"

Read the full page at /examples/.
`;

export const GET: APIRoute = () =>
  new Response(body, {
    headers: {
      "content-type": "text/markdown; charset=utf-8",
    },
  });

