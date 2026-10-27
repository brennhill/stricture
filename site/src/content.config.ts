import { defineCollection, z } from "astro:content";
import { glob } from "astro/loaders";

const docs = defineCollection({
  // Only user-facing guides are published; specs and internal docs stay out of the site.
  loader: glob({
    pattern: [
      "ANNOTATION-GUIDE.md",
      "data-lineage-annotations.md",
      "lineage-escalation.md",
      "LINEAGE-PLAIN-LANGUAGE-LIBRARY.md",
      "POLICY-GUIDE.md",
    ],
    base: "../docs",
  }),
  schema: z.object({
    title: z.string().optional(),
    description: z.string().optional(),
    order: z.number().optional(),
    sidebar: z.string().optional(),
  }),
});

export const collections = { docs };
