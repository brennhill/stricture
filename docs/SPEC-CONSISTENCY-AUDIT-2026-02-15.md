# Spec Consistency Audit - 2026-02-15

## Scope

This pass checked cross-doc consistency for the recent lineage model changes:

1. reference handle namespace (`strict:*`) vs parser syntax (`stricture-*`)
2. compact authoring + computed/defaulted fields
3. policy + server distribution (`strict:policy_url`, `strict:server_url`)
4. downstream-impact default finding behavior
5. service-level flow tiers (`'strict:flows'`, `systems[].flows`)
6. flow-tier policy controls (`lineage.findings.flow_criticality`)

## Updated In This Audit Pass

Primary specs/docs:

- `spec/0.1-draft.md`
- `README.md`
- `docs/data-lineage-annotations.md`
- `docs/ANNOTATION-GUIDE.md`
- `docs/LINEAGE-AUTOMATION-SPEC.md`
- `docs/POLICY-PACK-SPEC.md`
- `docs/POLICY-CLI-CONTRACT.md`
- `docs/LINEAGE-PLAIN-LANGUAGE-LIBRARY.md`
- `docs/ROADMAP.md`
- `docs/helper/SPEC.md`
- `docs/server/SPEC.md`
- `docs/server/STORAGE.md`
- `docs/server/README.md`
- `docs/lineage-escalation.md`
- `docs/visual-demo-storyboard.md`

Schemas/examples:

- `docs/schemas/lineage-policy-pack.schema.json`
- `docs/schemas/lineage-system-registry.schema.json`
- `docs/schemas/lineage-artifact.schema.json`
- `docs/config-examples/lineage-systems.yml`
- `docs/config-examples/lineage-policy-strict.yml`
- `docs/config-examples/lineage-policy-binding.yml`

Website docs/pages:

- `site/src/pages/annotations.astro`
- `site/src/pages/annotations/index.md.ts`
- `site/src/pages/getting-started.astro`
- `site/src/pages/index.astro`
- `site/src/pages/index.md.ts`
- `site/src/pages/open-standard/index.md.ts`
- `site/src/pages/demo/index.md.ts`
- `site/src/pages/walkthrough.astro`
- `site/src/pages/walkthrough/index.md.ts`
- `site/src/pages/what-is-stricture.astro`
- `site/src/pages/what-is-stricture/index.md.ts`
- `site/src/pages/with-ai/index.md.ts`
- `site/src/pages/architecture-invariants.astro`
- `site/src/pages/architecture-invariants/index.md.ts`

## Findings

1. Core spec/policy/automation/server surfaces are aligned to flow-tier and
   downstream-impact defaults.
2. Reference naming is aligned to `strict:*` across docs/tooling UX, with
   explicit compatibility notes for existing `stricture:*` aliases.
3. High-traffic website docs are aligned to the same model.
4. Artifact schema now includes optional structures for `systems`, `edges`,
   `findings`, and `flow_context` to match documented behavior.

## Remaining Intentional Differences

1. Source comment examples intentionally still use parser tokens
   (`stricture-source`, `stricture-lineage-override`).
2. Some long-tail test-plan and research docs are rule-test fixtures rather than
   lineage product docs; they were not rewritten unless directly in scope.
3. Override semantics remain documented because they are still part of policy,
   even though the current public demo UI de-emphasizes override controls.

## Verification

- `python3 -m json.tool docs/schemas/lineage-policy-pack.schema.json`
- `python3 -m json.tool docs/schemas/lineage-system-registry.schema.json`
- `python3 -m json.tool docs/schemas/lineage-artifact.schema.json`
- `GOCACHE=/tmp/go-build npm --prefix site run build`
