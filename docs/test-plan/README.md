# Test Plan Directory

> **Navigation index for the split test plan files.** The monolith test-plan.md has been retired. All content lives in the directories below.

---

## Per-Rule Test Matrices

Detailed test cases (positive, negative, edge cases) for every rule. Each file contains a numbered section per rule with exact inputs and expected outputs.

| File | Rules | Lines |
|------|-------|-------|
| [`rules/tq.md`](rules/tq.md) | TQ-no-shallow-assertions through TQ-test-naming (10 rules) | ~3,300 |
| [`rules/arch.md`](rules/arch.md) | ARCH-dependency-direction through ARCH-module-boundary (6 rules) | ~1,250 |
| [`rules/conv.md`](rules/conv.md) | CONV-file-naming through CONV-required-exports (6 rules) | ~1,100 |
| [`rules/ctr.md`](rules/ctr.md) | CTR-request-shape through CTR-manifest-conformance (8 rules) | ~1,400 |

## Cross-Cutting Tests

Tests that span multiple rules or validate infrastructure.

| File | Scope |
|------|-------|
| [`cross-cutting/infrastructure.md`](cross-cutting/infrastructure.md) | Language adapter parity, config resolution, plugin system, CLI behavior, output formats, auto-fix, contract detection, rule interaction, error handling |
| [`cross-cutting/incremental-analysis.md`](cross-cutting/incremental-analysis.md) | `--changed` / `--staged` mode scenarios |
| [`cross-cutting/performance-benchmarks.md`](cross-cutting/performance-benchmarks.md) | BM-01 through BM-07 benchmark specifications |

## Tool Tests

Tests for specific CLI commands.

| File | Scope |
|------|-------|
| [`tools/cli.md`](tools/cli.md) | `stricture lint`, `stricture fix`, `stricture audit`, `stricture init`, `stricture inspect`, `stricture list-rules` |
| [`tools/trace.md`](tools/trace.md) | `stricture trace` (HAR, OpenTelemetry, custom JSON) |

## Manifest Tests

| File | Scope |
|------|-------|
| [`manifest/manifest-tests.md`](manifest/manifest-tests.md) | Manifest parsing, contract declarations, strictness levels, validation rules |

## Validation Set

Real-world code examples for testing detection accuracy. **34 files** across 5 languages, 5 frameworks, 3 protocols.

| Directory | Description |
|-----------|-------------|
| [`validation-set/README.md`](validation-set/README.md) | Full index with file descriptions and bug taxonomy |
| [`validation-set/01-12`](validation-set/) | 12 API integrations (Stripe, GitHub, Twilio, etc.) |
| [`validation-set/13-17`](validation-set/) | Language parity (Go, Python, Java) |
| [`validation-set/20-24`](validation-set/) | Cross-language contract pairs |
| [`validation-set/30-31`](validation-set/) | Architecture validation (Express, Go) |
| [`validation-set/40-41`](validation-set/) | Test quality + AI-generated anti-patterns |
| [`validation-set/50`](validation-set/) | Convention patterns |
| [`validation-set/60-61`](validation-set/) | GraphQL + event-driven protocols |
| [`validation-set/70-72`](validation-set/) | Framework patterns (JS, Go, Java) |
| [`validation-set/logistics/`](validation-set/logistics/) | Multi-company ecosystem (7 companies, 3 languages) |

## Spec Gap Analysis

| File | Description |
|------|-------------|
| [`spec-gaps.md`](spec-gaps.md) | Ambiguities, missing specs, contradictions, scalability concerns, recommendations |

---

## Reading Order

- **Implementing a rule?** Read the rule's section in `rules/{category}.md`, then its validation set file. See [`RULE-INDEX.md`](../RULE-INDEX.md) for cross-references.
- **Building infrastructure?** Read `cross-cutting/infrastructure.md`.
- **Adding a CLI command?** Read `tools/cli.md`.
- **Working on performance?** Read `cross-cutting/performance-benchmarks.md`.
