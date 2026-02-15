import type { APIRoute } from "astro";
import { comparisons, type ComparisonEntry } from "../../../data/comparisons";

export const prerender = true;

function renderList(items: string[]): string {
  return items.map((item) => `- ${item}`).join("\n");
}

function renderMatrix(entry: ComparisonEntry): string {
  const header = "| Criterion | Stricture | Tool | Notes |\n|---|---|---|---|";
  const rows = entry.matrix
    .map((row) => `| ${row.criterion} | ${row.stricture} | ${row.tool} | ${row.note} |`)
    .join("\n");
  return `${header}\n${rows}`;
}

function renderReferences(entry: ComparisonEntry): string {
  return entry.references.map((ref) => `- [${ref.label}](${ref.href})`).join("\n");
}

export function getStaticPaths() {
  return comparisons.map((entry) => ({
    params: { tool: entry.slug },
    props: { entry }
  }));
}

export const GET: APIRoute = ({ props }) => {
  const entry = props.entry as ComparisonEntry;

  const body = `# When To Use Stricture With ${entry.tool}\n\n${entry.oneLineDifference}\n\n## Overlap Summary\n${entry.overlapSummary}\n\n## What ${entry.shortName} is excellent at\n${renderList(entry.bestAt)}\n\n## Overlap with Stricture\n${renderList(entry.overlapAreas)}\n\n## What Stricture adds\n${renderList(entry.strictureAdds)}\n\n## What ${entry.shortName} adds\n${renderList(entry.toolAdds)}\n\n## When to use both\n${renderList(entry.whenUseBoth)}\n\n## When to use only ${entry.shortName}\n${renderList(entry.whenUseToolOnly)}\n\n## Adoption path\n${entry.adoptionSteps.map((step, idx) => `${idx + 1}. ${step}`).join("\n")}\n\n## Overlay snippet\n\n\`\`\`yaml\n${entry.overlaySnippet}\n\`\`\`\n\n## Compare matrix\n${renderMatrix(entry)}\n\n## References\n${renderReferences(entry)}\n`;

  return new Response(body, {
    headers: {
      "Content-Type": "text/markdown; charset=utf-8",
      "Content-Signal": "ai-train=yes, search=yes, ai-input=yes"
    }
  });
};
