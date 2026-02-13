# Stricture

Higher-order linter for test quality, architecture, conventions, and API contracts. Written in Go. Uses tree-sitter for TS/Python/Java parsing.

---

## Commands

```bash
make build            # Build binary to bin/stricture
make test             # All tests
make test-phase1      # Phase 1 tests only (CLI, config, Go adapter, CONV rules)
make test-phase2      # Phase 1+2 (add TS adapter, ARCH rules)
make test-phase3      # Phase 1+2+3 (add TQ rules)
make test-phase4      # Phase 1+2+3+4 (add CTR rules, manifest)
make lint             # golangci-lint
make benchmark        # Performance benchmarks
make ci               # Full CI: lint + test + benchmark + validate
```

## Rules

1. **TDD** — Write failing tests first, then implement. Never the reverse.
2. **Zero production deps** — Only Go stdlib + tree-sitter C bindings. `testify` and `yaml.v3` are allowed.
3. **File headers** — Every `.go` file starts with `// filename.go — Purpose.`
4. **Max 800 LOC** — No source file exceeds 800 lines. Split if needed.
5. **Error format** — `{OPERATION}: {ROOT_CAUSE}. {RECOVERY_ACTION}`
6. **File naming** — `snake_case.go` for all Go files.
7. **Gate compliance** — Read and follow [DEVELOPMENT-GATES.md](docs/DEVELOPMENT-GATES.md) before ANY implementation.

## Development Workflow

```
1. Read DEVELOPMENT-GATES.md → identify which gate applies
2. Check gate prerequisites → create missing artifacts FIRST
3. Write tests (TDD red phase)
4. Implement (TDD green phase)
5. Verify all exit criteria
6. Report gate compliance
```

---

## Phase-Scoped Reading Guides

**Read ONLY the sections listed for your current phase.** Do not read the entire product spec.

### Phase 1: Foundation (CLI + Config + Go Adapter + 2 CONV Rules)

Read in this order:

| # | Document | Section | Lines | What you learn |
|---|----------|---------|-------|----------------|
| 1 | [DEVELOPMENT-GATES.md](docs/DEVELOPMENT-GATES.md) | Gate 1 (Rules) | Full file | What artifacts must exist before coding |
| 2 | [product-spec.md](docs/product-spec.md) | §3 Design Principles | L89-115 | Strict defaults, deep analysis, explainable violations |
| 3 | [product-spec.md](docs/product-spec.md) | §5 Configuration | L191-306 | .stricture.yml format, resolution order |
| 4 | [product-spec.md](docs/product-spec.md) | §6.3 CONV-file-naming | L1026-1041 | Rule spec |
| 5 | [product-spec.md](docs/product-spec.md) | §6.3 CONV-file-header | L1042-1059 | Rule spec |
| 6 | [product-spec.md](docs/product-spec.md) | §9 CLI Interface | L1912-1985 | Commands, flags, exit codes |
| 7 | [product-spec.md](docs/product-spec.md) | §10 Output | L2023-2082 | Output formats |
| 8 | [tech-spec.md](docs/tech-spec.md) | §1-3 Stack + Structure + Interfaces | L1-405 | Project layout, all Go interfaces |
| 9 | [tech-spec.md](docs/tech-spec.md) | §4 Phase 1 | L412-447 | Deliverables, test gate, files to create |
| 10 | [tech-spec.md](docs/tech-spec.md) | §6.1 Go Adapter | L688-711 | How to parse Go files |
| 11 | [tech-spec.md](docs/tech-spec.md) | §7 Rule Pattern | L744-839 | Rule template + test template |
| 12 | [error-catalog.yml](docs/error-catalog.yml) | CONV-file-naming, CONV-file-header | L267-296 | Exact error messages |
| 13 | [RULE-INDEX.md](docs/RULE-INDEX.md) | CONV rows | Full file | Cross-references to all docs |
| 14 | [test-plan/rules/conv.md](docs/test-plan/rules/conv.md) | §17-18 | L7-417 | Test matrices for both rules |
| 15 | [validation-set/50-convention-patterns.md](docs/test-plan/validation-set/50-convention-patterns.md) | PERFECT + bugs | Skim | Real-world examples |

**Deliverables:** CLI entry point, config loader, Go adapter, CONV-file-naming, CONV-file-header, text + JSON reporters. All with tests.

**Test gate:** `make test-phase1` passes.

---

### Phase 2: TypeScript Adapter + Remaining CONV + ARCH Rules

| # | Document | Section | Lines | What you learn |
|---|----------|---------|-------|----------------|
| 1 | [DEVELOPMENT-GATES.md](docs/DEVELOPMENT-GATES.md) | Gate 1 + Gate 2 | Full file | Rule gate + adapter gate |
| 2 | [product-spec.md](docs/product-spec.md) | §6.2 Architecture | L895-1019 | All 6 ARCH rule specs |
| 3 | [product-spec.md](docs/product-spec.md) | §6.3 CONV remaining | L1060-1134 | CONV-error-format through CONV-required-exports |
| 4 | [product-spec.md](docs/product-spec.md) | §7.3 TS Adapter | L1751-1764 | TypeScript adapter spec |
| 5 | [tech-spec.md](docs/tech-spec.md) | §4 Phase 2 | L449-477 | Deliverables, test gate |
| 6 | [tech-spec.md](docs/tech-spec.md) | §6.2 tree-sitter | L713-738 | tree-sitter query patterns |
| 7 | [error-catalog.yml](docs/error-catalog.yml) | ARCH + remaining CONV | L173-357 | Error messages |
| 8 | [test-plan/rules/arch.md](docs/test-plan/rules/arch.md) | All 6 sections | Full file | ARCH test matrices |
| 9 | [test-plan/rules/conv.md](docs/test-plan/rules/conv.md) | §19-22 | L418-end | Remaining CONV test matrices |
| 10 | [validation-set/30-express-layered-app.md](docs/test-plan/validation-set/30-express-layered-app.md) | PERFECT + violations | Skim | ARCH validation examples |
| 11 | [validation-set/31-go-clean-architecture.md](docs/test-plan/validation-set/31-go-clean-architecture.md) | PERFECT + violations | Skim | Go ARCH validation examples |

**Deliverables:** TypeScript adapter, ProjectContext builder, dependency graph (Tarjan's SCC), 4 CONV rules, 6 ARCH rules. All with tests.

**Test gate:** `make test-phase2` passes.

---

### Phase 3: TQ Rules (Test Quality) — The Core Differentiator

| # | Document | Section | Lines | What you learn |
|---|----------|---------|-------|----------------|
| 1 | [DEVELOPMENT-GATES.md](docs/DEVELOPMENT-GATES.md) | Gate 1 | — | Rule gate |
| 2 | [product-spec.md](docs/product-spec.md) | §6.1 Test Quality | L308-894 | All 10 TQ rule specs (detailed) |
| 3 | [tech-spec.md](docs/tech-spec.md) | §4 Phase 3 | L480-505 | Deliverables, test gate |
| 4 | [tech-spec.md](docs/tech-spec.md) | §8 Assertion Classification | L843-880 | Classification algorithm |
| 5 | [error-catalog.yml](docs/error-catalog.yml) | TQ rules | L19-170 | Error messages |
| 6 | [test-plan/rules/tq.md](docs/test-plan/rules/tq.md) | All 10 sections | Full file | TQ test matrices (3,300 lines) |
| 7 | [validation-set/40-test-quality-patterns.md](docs/test-plan/validation-set/40-test-quality-patterns.md) | All | Full file | TQ validation examples |
| 8 | [validation-set/41-ai-generated-test-patterns.md](docs/test-plan/validation-set/41-ai-generated-test-patterns.md) | All | Full file | AI anti-patterns to catch |

**Deliverables:** Assertion classifier, test-to-source mapping, all 10 TQ rules. All with tests.

**Test gate:** `make test-phase3` passes.

---

### Phase 4: CTR Rules + Manifest Parser

| # | Document | Section | Lines | What you learn |
|---|----------|---------|-------|----------------|
| 1 | [DEVELOPMENT-GATES.md](docs/DEVELOPMENT-GATES.md) | Gate 1 + Gate 3 | — | Rule gate + CTR-specific gate |
| 2 | [product-spec.md](docs/product-spec.md) | §6.4 Contract | L1135-1657 | All 8 CTR rule specs |
| 3 | [product-spec.md](docs/product-spec.md) | §13 Manifest | L2177-2527 | Manifest format, strictness, audit, trace |
| 4 | [tech-spec.md](docs/tech-spec.md) | §4 Phase 4 | L508-535 | Deliverables, test gate |
| 5 | [error-catalog.yml](docs/error-catalog.yml) | CTR rules | L361-479 | Error messages |
| 6 | [test-plan/rules/ctr.md](docs/test-plan/rules/ctr.md) | All 8 sections | Full file | CTR test matrices |
| 7 | [test-plan/manifest/manifest-tests.md](docs/test-plan/manifest/manifest-tests.md) | All | Skim | Manifest parsing tests |
| 8 | Cross-language validation files | [20](docs/test-plan/validation-set/20-internal-user-api.md), [21](docs/test-plan/validation-set/21-cross-lang-go-ts.md), [22](docs/test-plan/validation-set/22-cross-lang-python-ts.md), [23](docs/test-plan/validation-set/23-cross-lang-java-ts.md), [24](docs/test-plan/validation-set/24-cross-lang-go-python.md) | Skim | Cross-language mismatch examples |

**Deliverables:** Manifest parser, strictness calculator, all 8 CTR rules, `stricture audit` command. All with tests.

**Test gate:** `make test-phase4` passes.

---

### Phase 5: Python + Java Adapters + Advanced Output

| # | Document | Section | Lines | What you learn |
|---|----------|---------|-------|----------------|
| 1 | [DEVELOPMENT-GATES.md](docs/DEVELOPMENT-GATES.md) | Gate 2 + Gate 4 | — | Adapter + reporter gates |
| 2 | [product-spec.md](docs/product-spec.md) | §7.4-7.5 | L1765-1788 | Python + Java adapter specs |
| 3 | [product-spec.md](docs/product-spec.md) | §10.3-10.4 | L2069-2082 | SARIF + JUnit output specs |
| 4 | [tech-spec.md](docs/tech-spec.md) | §4 Phase 5 | L538-563 | Deliverables, test gate |
| 5 | Python/Java validation files | [16](docs/test-plan/validation-set/16-stripe-python.md), [17](docs/test-plan/validation-set/17-stripe-java.md) | Skim | Language-specific examples |
| 6 | [cross-cutting/performance-benchmarks.md](docs/test-plan/cross-cutting/performance-benchmarks.md) | All | Full file | Benchmark targets |

**Deliverables:** Python adapter, Java adapter, SARIF reporter, JUnit reporter, `--changed`/`--staged`, caching. All with tests.

**Test gate:** `make test-phase5` passes.

---

### Phase 6: Polish + Distribution

| # | Document | Section | Lines | What you learn |
|---|----------|---------|-------|----------------|
| 1 | [DEVELOPMENT-GATES.md](docs/DEVELOPMENT-GATES.md) | Gate 7 | — | Release gate |
| 2 | [product-spec.md](docs/product-spec.md) | §8 Plugins | L1815-1910 | Plugin system spec |
| 3 | [product-spec.md](docs/product-spec.md) | §11 Auto-Fix | L2085-2130 | Fix system spec |
| 4 | [product-spec.md](docs/product-spec.md) | §12 CI Integration | L2132-2175 | CI integration patterns |
| 5 | [tech-spec.md](docs/tech-spec.md) | §4 Phase 6 | L566-589 | Deliverables, test gate |
| 6 | [tech-spec.md](docs/tech-spec.md) | §10 Distribution | L901-935 | Cross-platform builds |

**Deliverables:** Auto-fix engine, suppression parser, plugin system, GoReleaser, `init`/`list-rules`/`inspect` commands.

**Test gate:** `make ci` passes. `stricture .` passes on own codebase.

---

## Finding Things

| Need | Location |
|------|----------|
| Rule cross-reference (THE lookup table) | [`docs/RULE-INDEX.md`](docs/RULE-INDEX.md) |
| Product spec (WHAT to build) | [`docs/product-spec.md`](docs/product-spec.md) |
| Tech spec (HOW to build) | [`docs/tech-spec.md`](docs/tech-spec.md) |
| Development gates (WHEN to proceed) | [`docs/DEVELOPMENT-GATES.md`](docs/DEVELOPMENT-GATES.md) |
| Error messages (exact templates) | [`docs/error-catalog.yml`](docs/error-catalog.yml) |
| Error handling contract | [`docs/GO-ERROR-HANDLING.md`](docs/GO-ERROR-HANDLING.md) |
| API stability contract | [`docs/API-STABILITY.md`](docs/API-STABILITY.md) |
| Regression protocol | [`docs/REGRESSION-PROTOCOL.md`](docs/REGRESSION-PROTOCOL.md) |
| Test plans (per-rule test matrices) | [`docs/test-plan/rules/`](docs/test-plan/rules/) |
| Test plan index | [`docs/test-plan/README.md`](docs/test-plan/README.md) |
| Validation set (real-world examples) | [`docs/test-plan/validation-set/`](docs/test-plan/validation-set/) |
| Golden file fixtures | [`tests/golden/`](tests/golden/) |
| JSON output schema | [`tests/golden/schema.json`](tests/golden/schema.json) |
| Test fixture directory | [`tests/fixtures/`](tests/fixtures/) |
| Integration tests | [`tests/integration/`](tests/integration/) |
| Invariants (MUST always hold) | [`docs/INVARIANTS.md`](docs/INVARIANTS.md) |
| Tree-sitter version pinning | [`docs/TREE-SITTER-PINNING.md`](docs/TREE-SITTER-PINNING.md) |
| Sentinel errors | [`internal/model/errors.go`](internal/model/errors.go) |
| Adapter parity fixtures | [`tests/fixtures/adapter-parity/`](tests/fixtures/adapter-parity/) |
| Stress test corpus | [`tests/fixtures/stress/`](tests/fixtures/stress/) |
| Fuzz tests | `internal/adapter/goparser/fuzz_test.go`, `internal/config/fuzz_test.go` |
| Linter config | [`.golangci.yml`](.golangci.yml) |
| CI pipeline | [`.github/workflows/ci.yml`](.github/workflows/ci.yml) |
| Pre-commit hooks | [`.pre-commit-config.yaml`](.pre-commit-config.yaml) |
| Research / analysis (NOT instructions) | [`docs/research/`](docs/research/) |

## Important Constraints

- **Never start Phase N+1 until all Phase N tests pass.**
- **Never implement a rule without reading its error catalog entry first.**
- **Never skip a development gate prerequisite.**
- **All Go interfaces are defined in `internal/model/` — start there.**
- **All errors must follow [`docs/GO-ERROR-HANDLING.md`](docs/GO-ERROR-HANDLING.md)** — wrap with `%w`, use sentinels, never panic.
- **All bugs get a regression fixture BEFORE the fix** — see [`docs/REGRESSION-PROTOCOL.md`](docs/REGRESSION-PROTOCOL.md).
- **JSON output is additive-only** — never remove fields, see [`docs/API-STABILITY.md`](docs/API-STABILITY.md).
- The product spec is 2,674 lines. **Do NOT read it all.** Use the phase reading guides above.
- The test plan is split across `docs/test-plan/rules/{tq,arch,conv,ctr}.md`. There is no monolith file.
- Run `go test -race ./...` before any PR — race conditions in goroutine pools are real bugs.
