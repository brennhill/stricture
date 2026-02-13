# Stricture — Rule Cross-Reference Index

> **The Rosetta Stone.** For any rule, this file tells you exactly where to find its spec, error message, test plan, validation examples, fixtures, and source code.

---

## How to Use

When implementing or modifying a rule, look it up below. Each row links to the **exact line number** in every document. Read them in order: spec → error catalog → test plan → validation set → implement → verify.

---

## TQ (Test Quality) — 10 Rules

| Rule ID | Product Spec | Error Catalog | Tech Spec | Test Plan | Validation Set | Golden Fixtures | Source File | Test File |
|---------|-------------|---------------|-----------|-----------|----------------|-----------------|-------------|-----------|
| TQ-no-shallow-assertions | [§6.1 L325](product-spec.md#L325) | [L19](error-catalog.yml#L19) | [§8 L846](tech-spec.md#L846) | [tq.md §1 L7](test-plan/rules/tq.md#L7) | [01-stripe](test-plan/validation-set/01-stripe.md) B03, [40-tq](test-plan/validation-set/40-test-quality-patterns.md), [41-ai](test-plan/validation-set/41-ai-generated-test-patterns.md) | `tests/fixtures/tq-no-shallow-assertions/` | `internal/rules/tq/no_shallow.go` | `internal/rules/tq/no_shallow_test.go` |
| TQ-return-type-verified | [§6.1 L390](product-spec.md#L390) | [L34](error-catalog.yml#L34) | [§8 L846](tech-spec.md#L846) | [tq.md §2 L825](test-plan/rules/tq.md#L825) | [40-tq](test-plan/validation-set/40-test-quality-patterns.md) | `tests/fixtures/tq-return-type-verified/` | `internal/rules/tq/return_type.go` | `internal/rules/tq/return_type_test.go` |
| TQ-schema-conformance | [§6.1 L479](product-spec.md#L479) | [L49](error-catalog.yml#L49) | [§8 L846](tech-spec.md#L846) | [tq.md §3 L1290](test-plan/rules/tq.md#L1290) | [40-tq](test-plan/validation-set/40-test-quality-patterns.md) | `tests/fixtures/tq-schema-conformance/` | `internal/rules/tq/schema_conformance.go` | `internal/rules/tq/schema_conformance_test.go` |
| TQ-error-path-coverage | [§6.1 L550](product-spec.md#L550) | [L64](error-catalog.yml#L64) | [§8 L846](tech-spec.md#L846) | [tq.md §4 L1529](test-plan/rules/tq.md#L1529) | [01-stripe](test-plan/validation-set/01-stripe.md) B01, [40-tq](test-plan/validation-set/40-test-quality-patterns.md) | `tests/fixtures/tq-error-path-coverage/` | `internal/rules/tq/error_path.go` | `internal/rules/tq/error_path_test.go` |
| TQ-assertion-depth | [§6.1 L643](product-spec.md#L643) | [L79](error-catalog.yml#L79) | [§8 L846](tech-spec.md#L846) | [tq.md §5 L1821](test-plan/rules/tq.md#L1821) | [40-tq](test-plan/validation-set/40-test-quality-patterns.md) | `tests/fixtures/tq-assertion-depth/` | `internal/rules/tq/assertion_depth.go` | `internal/rules/tq/assertion_depth_test.go` |
| TQ-boundary-tested | [§6.1 L683](product-spec.md#L683) | [L94](error-catalog.yml#L94) | [§8 L846](tech-spec.md#L846) | [tq.md §6 L2092](test-plan/rules/tq.md#L2092) | [40-tq](test-plan/validation-set/40-test-quality-patterns.md) | `tests/fixtures/tq-boundary-tested/` | `internal/rules/tq/boundary.go` | `internal/rules/tq/boundary_test.go` |
| TQ-mock-scope | [§6.1 L729](product-spec.md#L729) | [L109](error-catalog.yml#L109) | [§8 L846](tech-spec.md#L846) | [tq.md §7 L2340](test-plan/rules/tq.md#L2340) | [40-tq](test-plan/validation-set/40-test-quality-patterns.md) | `tests/fixtures/tq-mock-scope/` | `internal/rules/tq/mock_scope.go` | `internal/rules/tq/mock_scope_test.go` |
| TQ-test-isolation | [§6.1 L779](product-spec.md#L779) | [L124](error-catalog.yml#L124) | [§8 L846](tech-spec.md#L846) | [tq.md §8 L2607](test-plan/rules/tq.md#L2607) | [40-tq](test-plan/validation-set/40-test-quality-patterns.md) | `tests/fixtures/tq-test-isolation/` | `internal/rules/tq/test_isolation.go` | `internal/rules/tq/test_isolation_test.go` |
| TQ-negative-cases | [§6.1 L818](product-spec.md#L818) | [L139](error-catalog.yml#L139) | [§8 L846](tech-spec.md#L846) | [tq.md §9 L2871](test-plan/rules/tq.md#L2871) | [01-stripe](test-plan/validation-set/01-stripe.md) B04, [40-tq](test-plan/validation-set/40-test-quality-patterns.md) | `tests/fixtures/tq-negative-cases/` | `internal/rules/tq/negative_cases.go` | `internal/rules/tq/negative_cases_test.go` |
| TQ-test-naming | [§6.1 L860](product-spec.md#L860) | [L154](error-catalog.yml#L154) | [§8 L846](tech-spec.md#L846) | [tq.md §10 L3123](test-plan/rules/tq.md#L3123) | [40-tq](test-plan/validation-set/40-test-quality-patterns.md) | `tests/fixtures/tq-test-naming/` | `internal/rules/tq/test_naming.go` | `internal/rules/tq/test_naming_test.go` |

## ARCH (Architecture) — 6 Rules

| Rule ID | Product Spec | Error Catalog | Tech Spec | Test Plan | Validation Set | Golden Fixtures | Source File | Test File |
|---------|-------------|---------------|-----------|-----------|----------------|-----------------|-------------|-----------|
| ARCH-dependency-direction | [§6.2 L901](product-spec.md#L901) | [L173](error-catalog.yml#L173) | [§4 Phase 2 L449](tech-spec.md#L449) | [arch.md §11 L7](test-plan/rules/arch.md#L7) | [30-express](test-plan/validation-set/30-express-layered-app.md), [31-go-clean](test-plan/validation-set/31-go-clean-architecture.md) | `tests/fixtures/arch-dependency-direction/` | `internal/rules/arch/dependency_dir.go` | `internal/rules/arch/dependency_dir_test.go` |
| ARCH-import-boundary | [§6.2 L931](product-spec.md#L931) | [L188](error-catalog.yml#L188) | [§4 Phase 2 L449](tech-spec.md#L449) | [arch.md §12 L223](test-plan/rules/arch.md#L223) | [30-express](test-plan/validation-set/30-express-layered-app.md), [31-go-clean](test-plan/validation-set/31-go-clean-architecture.md) | `tests/fixtures/arch-import-boundary/` | `internal/rules/arch/import_boundary.go` | `internal/rules/arch/import_boundary_test.go` |
| ARCH-no-circular-deps | [§6.2 L954](product-spec.md#L954) | [L203](error-catalog.yml#L203) | [§4 Phase 2 L449](tech-spec.md#L449) | [arch.md §13 L446](test-plan/rules/arch.md#L446) | [30-express](test-plan/validation-set/30-express-layered-app.md) | `tests/fixtures/arch-no-circular-deps/` | `internal/rules/arch/circular_deps.go` | `internal/rules/arch/circular_deps_test.go` |
| ARCH-max-file-lines | [§6.2 L962](product-spec.md#L962) | [L218](error-catalog.yml#L218) | [§4 Phase 2 L449](tech-spec.md#L449) | [arch.md §14 L618](test-plan/rules/arch.md#L618) | [30-express](test-plan/validation-set/30-express-layered-app.md) | `tests/fixtures/arch-max-file-lines/` | `internal/rules/arch/max_lines.go` | `internal/rules/arch/max_lines_test.go` |
| ARCH-layer-violation | [§6.2 L980](product-spec.md#L980) | [L233](error-catalog.yml#L233) | [§4 Phase 2 L449](tech-spec.md#L449) | [arch.md §15 L789](test-plan/rules/arch.md#L789) | [30-express](test-plan/validation-set/30-express-layered-app.md), [31-go-clean](test-plan/validation-set/31-go-clean-architecture.md) | `tests/fixtures/arch-layer-violation/` | `internal/rules/arch/layer_violation.go` | `internal/rules/arch/layer_violation_test.go` |
| ARCH-module-boundary | [§6.2 L1002](product-spec.md#L1002) | [L248](error-catalog.yml#L248) | [§4 Phase 2 L449](tech-spec.md#L449) | [arch.md §16 L1022](test-plan/rules/arch.md#L1022) | [30-express](test-plan/validation-set/30-express-layered-app.md) | `tests/fixtures/arch-module-boundary/` | `internal/rules/arch/module_boundary.go` | `internal/rules/arch/module_boundary_test.go` |

## CONV (Convention) — 6 Rules

| Rule ID | Product Spec | Error Catalog | Tech Spec | Test Plan | Validation Set | Golden Fixtures | Source File | Test File |
|---------|-------------|---------------|-----------|-----------|----------------|-----------------|-------------|-----------|
| CONV-file-naming | [§6.3 L1026](product-spec.md#L1026) | [L267](error-catalog.yml#L267) | [§4 Phase 1 L412](tech-spec.md#L412) | [conv.md §17 L7](test-plan/rules/conv.md#L7) | [50-convention](test-plan/validation-set/50-convention-patterns.md) | `tests/fixtures/conv-file-naming/` | `internal/rules/conv/file_naming.go` | `internal/rules/conv/file_naming_test.go` |
| CONV-file-header | [§6.3 L1042](product-spec.md#L1042) | [L282](error-catalog.yml#L282) | [§4 Phase 1 L412](tech-spec.md#L412) | [conv.md §18 L191](test-plan/rules/conv.md#L191) | [50-convention](test-plan/validation-set/50-convention-patterns.md) | `tests/fixtures/conv-file-header/` | `internal/rules/conv/file_header.go` | `internal/rules/conv/file_header_test.go` |
| CONV-error-format | [§6.3 L1060](product-spec.md#L1060) | [L297](error-catalog.yml#L297) | [§4 Phase 2 L449](tech-spec.md#L449) | [conv.md §19 L418](test-plan/rules/conv.md#L418) | [50-convention](test-plan/validation-set/50-convention-patterns.md) | `tests/fixtures/conv-error-format/` | `internal/rules/conv/error_format.go` | `internal/rules/conv/error_format_test.go` |
| CONV-export-naming | [§6.3 L1080](product-spec.md#L1080) | [L312](error-catalog.yml#L312) | [§4 Phase 2 L449](tech-spec.md#L449) | [conv.md §20 L629](test-plan/rules/conv.md#L629) | [50-convention](test-plan/validation-set/50-convention-patterns.md) | `tests/fixtures/conv-export-naming/` | `internal/rules/conv/export_naming.go` | `internal/rules/conv/export_naming_test.go` |
| CONV-test-file-location | [§6.3 L1100](product-spec.md#L1100) | [L327](error-catalog.yml#L327) | [§4 Phase 2 L449](tech-spec.md#L449) | [conv.md §21 L836](test-plan/rules/conv.md#L836) | [50-convention](test-plan/validation-set/50-convention-patterns.md) | `tests/fixtures/conv-test-file-location/` | `internal/rules/conv/test_location.go` | `internal/rules/conv/test_location_test.go` |
| CONV-required-exports | [§6.3 L1118](product-spec.md#L1118) | [L342](error-catalog.yml#L342) | [§4 Phase 2 L449](tech-spec.md#L449) | [conv.md §22 L960](test-plan/rules/conv.md#L960) | [50-convention](test-plan/validation-set/50-convention-patterns.md) | `tests/fixtures/conv-required-exports/` | `internal/rules/conv/required_exports.go` | `internal/rules/conv/required_exports_test.go` |

## CTR (Contract) — 8 Rules

| Rule ID | Product Spec | Error Catalog | Tech Spec | Test Plan | Validation Set | Golden Fixtures | Source File | Test File |
|---------|-------------|---------------|-----------|-----------|----------------|-----------------|-------------|-----------|
| CTR-request-shape | [§6.4 L1194](product-spec.md#L1194) | [L361](error-catalog.yml#L361) | [§4 Phase 4 L508](tech-spec.md#L508) | [ctr.md §23 L9](test-plan/rules/ctr.md#L9) | [01-12](test-plan/validation-set/01-stripe.md) B05, [20-24](test-plan/validation-set/20-internal-user-api.md) | `tests/fixtures/ctr-request-shape/` | `internal/rules/ctr/request_shape.go` | `internal/rules/ctr/request_shape_test.go` |
| CTR-response-shape | [§6.4 L1260](product-spec.md#L1260) | [L376](error-catalog.yml#L376) | [§4 Phase 4 L508](tech-spec.md#L508) | [ctr.md §24 L149](test-plan/rules/ctr.md#L149) | [01-12](test-plan/validation-set/01-stripe.md) B06, [20-24](test-plan/validation-set/20-internal-user-api.md) | `tests/fixtures/ctr-response-shape/` | `internal/rules/ctr/response_shape.go` | `internal/rules/ctr/response_shape_test.go` |
| CTR-status-code-handling | [§6.4 L1297](product-spec.md#L1297) | [L391](error-catalog.yml#L391) | [§4 Phase 4 L508](tech-spec.md#L508) | [ctr.md §25 L236](test-plan/rules/ctr.md#L236) | [01-12](test-plan/validation-set/01-stripe.md) B02 | `tests/fixtures/ctr-status-code-handling/` | `internal/rules/ctr/status_code.go` | `internal/rules/ctr/status_code_test.go` |
| CTR-shared-type-sync | [§6.4 L1346](product-spec.md#L1346) | [L406](error-catalog.yml#L406) | [§4 Phase 4 L508](tech-spec.md#L508) | [ctr.md §26 L345](test-plan/rules/ctr.md#L345) | [20-internal](test-plan/validation-set/20-internal-user-api.md), [21-go-ts](test-plan/validation-set/21-cross-lang-go-ts.md) | `tests/fixtures/ctr-shared-type-sync/` | `internal/rules/ctr/shared_type_sync.go` | `internal/rules/ctr/shared_type_sync_test.go` |
| CTR-json-tag-match | [§6.4 L1399](product-spec.md#L1399) | [L421](error-catalog.yml#L421) | [§4 Phase 4 L508](tech-spec.md#L508) | [ctr.md §27 L428](test-plan/rules/ctr.md#L428) | [13-stripe-go](test-plan/validation-set/13-stripe-go.md), [21-go-ts](test-plan/validation-set/21-cross-lang-go-ts.md), [24-go-py](test-plan/validation-set/24-cross-lang-go-python.md) | `tests/fixtures/ctr-json-tag-match/` | `internal/rules/ctr/json_tag_match.go` | `internal/rules/ctr/json_tag_match_test.go` |
| CTR-dual-test | [§6.4 L1437](product-spec.md#L1437) | [L436](error-catalog.yml#L436) | [§4 Phase 4 L508](tech-spec.md#L508) | [ctr.md §28 L514](test-plan/rules/ctr.md#L514) | [20-internal](test-plan/validation-set/20-internal-user-api.md), [21-go-ts](test-plan/validation-set/21-cross-lang-go-ts.md) | `tests/fixtures/ctr-dual-test/` | `internal/rules/ctr/dual_test.go` | `internal/rules/ctr/dual_test_test.go` |
| CTR-strictness-parity | [§6.4 L1482](product-spec.md#L1482) | [L451](error-catalog.yml#L451) | [§4 Phase 4 L508](tech-spec.md#L508) | [ctr.md §29 L602](test-plan/rules/ctr.md#L602) | [01-12](test-plan/validation-set/01-stripe.md) B08-B11, [20-24](test-plan/validation-set/20-internal-user-api.md) | `tests/fixtures/ctr-strictness-parity/` | `internal/rules/ctr/strictness_parity.go` | `internal/rules/ctr/strictness_parity_test.go` |
| CTR-manifest-conformance | [§6.4 L1553](product-spec.md#L1553) | [L466](error-catalog.yml#L466) | [§4 Phase 4 L508](tech-spec.md#L508) | [ctr.md §30 L1026](test-plan/rules/ctr.md#L1026) | [01-12](test-plan/validation-set/01-stripe.md) B07 | `tests/fixtures/ctr-manifest-conformance/` | `internal/rules/ctr/manifest_conformance.go` | `internal/rules/ctr/manifest_conformance_test.go` |

---

## Phase → Rule Mapping

Which rules to implement in which phase (from [tech-spec.md §4](tech-spec.md)):

| Phase | Rules | Adapters | Key Files |
|-------|-------|----------|-----------|
| **Phase 1** (Foundation) | CONV-file-naming, CONV-file-header | Go | `cmd/stricture/main.go`, config, goparser, text+json reporters |
| **Phase 2** (TS + ARCH) | CONV-error-format, CONV-export-naming, CONV-test-file-location, CONV-required-exports, all 6 ARCH rules | Go + TypeScript | TS adapter, engine/context, engine/graph |
| **Phase 3** (TQ) | All 10 TQ rules | — | assertion classifier, test-to-source mapping |
| **Phase 4** (CTR) | All 8 CTR rules | — | manifest parser, strictness calculator, audit command |
| **Phase 5** (Multi-lang) | — | Python + Java | SARIF/JUnit reporters, `--changed`/`--staged`, caching |
| **Phase 6** (Polish) | — | — | auto-fix, plugins, `init`/`inspect`/`list-rules` commands |

---

## Validation Set → Rule Mapping

Which validation files exercise which rules:

| Validation File | Primary Rules Tested |
|----------------|---------------------|
| 01-12 (API integrations) | CTR-request-shape, CTR-response-shape, CTR-status-code-handling, CTR-strictness-parity, CTR-manifest-conformance, TQ-error-path-coverage, TQ-no-shallow-assertions, TQ-negative-cases |
| 13-15 (Go language) | CTR-json-tag-match, CTR-request-shape, CTR-response-shape, TQ-* |
| 16 (Python) | CTR-request-shape, CTR-response-shape, TQ-* |
| 17 (Java) | CTR-request-shape, CTR-response-shape, TQ-* |
| 20-24 (Cross-language) | CTR-shared-type-sync, CTR-json-tag-match, CTR-dual-test, CTR-strictness-parity |
| 30 (Express layered) | All 6 ARCH rules |
| 31 (Go clean arch) | All 6 ARCH rules |
| 40 (Test quality) | All 10 TQ rules |
| 41 (AI-generated tests) | TQ-no-shallow-assertions, TQ-return-type-verified, TQ-negative-cases, TQ-test-naming |
| 50 (Conventions) | All 6 CONV rules |
| 60-61 (GraphQL/events) | CTR-request-shape, CTR-response-shape, CTR-status-code-handling |
| 70-72 (Frameworks) | CONV-file-naming, ARCH-dependency-direction, CTR-request-shape |
| logistics/ | CTR-shared-type-sync, CTR-json-tag-match, CTR-manifest-conformance |
