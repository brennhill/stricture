# Stricture Server Docs

- Spec: `docs/server/SPEC.md`
- Storage + auth design: `docs/server/STORAGE.md`
- Contribution terms: `docs/server/CONTRIBUTING.md`
- Day-1 runtime requirement: local process + cloud-function support with the
  same API contract (see `docs/server/SPEC.md`).
- Default drift model: warn/block findings require downstream impact; self-only
  drift is still tracked for publication/audit.
- Flow-tier policy context is supported via `'strict:flows'` and
  `systems[].flows` metadata carried with artifacts/registries.
