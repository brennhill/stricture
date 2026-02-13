# Stricture — API Stability Contract

> **What users can depend on not breaking between releases.**
>
> Stricture is used in CI pipelines. Breaking changes to CLI flags, exit codes, or output formats break users' builds. This document defines what's stable, what's unstable, and how we handle breaking changes.

---

## Stability Tiers

### Tier 1: Stable (breaking change = major version bump)

These are promises to users. Changing them requires a major version bump and migration guide.

| API Surface | Guarantee |
|-------------|-----------|
| Exit codes | 0 = clean, 1 = violations, 2 = operational error |
| `--version` flag | Prints version, exits 0 |
| `--help` flag | Prints help, exits 0 |
| `--format json` output | JSON schema is additive-only (new fields OK, removed/renamed fields = breaking) |
| `--format sarif` output | Conforms to SARIF 2.1.0 schema |
| Rule IDs | `TQ-no-shallow-assertions` etc. — never renamed |
| Rule default severity | Never changes without major version |
| `.stricture.yml` schema | Additive-only (new keys OK, removed/renamed keys = breaking) |
| `.stricture-manifest.yml` schema | Additive-only |
| Suppression syntax | `// stricture-disable-next-line RULE-ID` — never changes |

### Tier 2: Semi-stable (breaking change = minor version bump + changelog)

| API Surface | Guarantee |
|-------------|-----------|
| `--format text` output | Human-readable, may change format between minor versions |
| Violation messages | Wording may improve, template structure stays the same |
| `--format junit` output | Conforms to JUnit XML schema, field mapping may change |
| New CLI flags | May be added in minor versions |
| New rules | May be added in minor versions (default: `off` for first release, then `error`) |

### Tier 3: Unstable (may change at any time)

| API Surface | Guarantee |
|-------------|-----------|
| `--verbose` / debug output | No stability guarantee |
| Cache format (`.stricture-cache/`) | May change between any release |
| Internal Go API | Not a public API, may change without notice |
| Benchmark numbers | May vary by environment |
| `stricture inspect` output | Debugging tool, format may change |

---

## Breaking Change Protocol

When a breaking change is necessary:

1. **Document in CHANGELOG.md** with `BREAKING:` prefix
2. **Deprecation period:** If possible, support both old and new behavior for one minor version with a deprecation warning
3. **Migration guide:** Provide exact steps to update
4. **Major version bump:** Increment major version

### Example deprecation:

```
# v0.2.0 — Deprecation warning
$ stricture --output json .
WARNING: --output is deprecated, use --format instead. --output will be removed in v0.3.0.

# v0.3.0 — Removed
$ stricture --output json .
ERROR: unknown flag --output. Use --format instead.
```

---

## JSON Output Schema

The JSON output schema is the most critical stability surface. CI tools parse it.

### Additive-only rule

New fields may be added to any object. Existing fields are never removed or renamed.

```json
{
  "version": "1",
  "violations": [
    {
      "ruleId": "string (stable)",
      "category": "string (stable)",
      "severity": "string (stable)",
      "file": "string (stable)",
      "line": "number (stable)",
      "column": "number (stable)",
      "message": "string (stable — template structure, wording may improve)",
      "suggestion": "string (stable)",
      "newFieldInV02": "any (additive — OK)"
    }
  ],
  "summary": {
    "filesChecked": "number (stable)",
    "errors": "number (stable)",
    "warnings": "number (stable)",
    "elapsedMs": "number (stable)"
  }
}
```

### Validation

Every release runs the JSON output through a JSON Schema validator. The schema file lives at `tests/golden/schema.json` and is updated only when new fields are added.

---

## Version Numbering

Follows [Semantic Versioning 2.0.0](https://semver.org/):

- **MAJOR** (1.0.0): Breaking changes to Tier 1 APIs
- **MINOR** (0.2.0): New features, new rules, Tier 2 changes
- **PATCH** (0.1.1): Bug fixes, performance improvements, no API changes

Pre-1.0: All Tier 1 APIs are still forming. Breaking changes may happen in minor versions with changelog documentation. After 1.0: full semver stability.
