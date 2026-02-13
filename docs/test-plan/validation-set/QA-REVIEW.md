# Validation Set QA Review

> **Reviewer:** QA Architecture Review
> **Date:** 2026-02-12
> **Scope:** Validation set completeness, rule coverage, gap analysis, testing strategy
> **Companion Documents:** [Product Spec](../../product-spec.md), [Validation Set README](./README.md), [Test Plan](../test-plan.md)

---

## 1. Executive Summary

The current validation set is **strong on CTR (contract) rule coverage** and **weak on everything else**. The 12 API integration files (01-stripe through 12-supabase) provide 192 test scenarios (12 APIs x 16 cases each) that deeply exercise the 8 CTR rules and 3 TQ rules. However, the remaining 23 rules across TQ, ARCH, and CONV categories have **zero direct coverage** from the validation set.

This is by design -- the validation set was scoped to test API contract detection accuracy. But it creates a significant gap: the most novel Stricture rules (TQ-assertion-depth, TQ-boundary-tested, TQ-mock-scope, TQ-test-isolation, all ARCH rules, all CONV rules) have no real-world validation corpus.

**Key findings:**

- **11 of 34 rules** have deep coverage from the validation set
- **23 of 34 rules** have zero validation set coverage (they are covered by the per-rule test matrices in the separate test plan files, but not by real-world integration scenarios)
- All 12 validation files use **TypeScript only** -- no Go, Python, Java, or Rust examples
- The PERFECT cases are well-constructed for false positive testing
- The bug taxonomy (B01-B15) is well-designed with progressive difficulty
- **No framework-specific patterns** are tested (React, Express, FastAPI, etc.)
- **No multi-language contract** scenarios exist (e.g., Go server + TypeScript client)
- The per-rule test plan files (tq.md, arch.md, conv.md, ctr.md) are excellent and compensate for many gaps, but they use synthetic code rather than real-world API integrations

**Overall assessment:** The validation set is a solid foundation for CTR rule validation. It needs expansion for TQ/ARCH/CONV rules and language diversity.

### Resolution Status (Updated 2026-02-13)

All P0 and P1 gaps have been resolved. The validation set now covers **30/30 active rules** across **5 languages** and **5+ frameworks**.

| Gap | Priority | Status | Resolution |
|-----|----------|--------|------------|
| GAP-P0-01 | P0 | **RESOLVED** | Created 13-stripe-go.md, 14-github-go.md, 15-aws-s3-go.md (Go language parity) |
| GAP-P0-02 | P0 | **RESOLVED** | Created 20-24: internal TS pair, Go↔TS, Python↔TS, Java↔TS, Go↔Python cross-language contracts |
| GAP-P0-03 | P0 | **RESOLVED** | Created 30-express-layered-app.md (12 violations), 31-go-clean-architecture.md (12 violations) |
| GAP-P0-04 | P0 | **RESOLVED** | Created 40-test-quality-patterns.md (20 violations, all 10 TQ rules), 41-ai-generated-test-patterns.md |
| GAP-P1-01 | P1 | **RESOLVED** | Created 70-framework-patterns-js.md (Express/Fastify/NestJS/Next.js/Hono), 71-go (Chi/Gin/Echo), 72-java (Spring Boot/JAX-RS/Micronaut) |
| GAP-P1-02 | P1 | **RESOLVED** | Created 50-convention-patterns.md (18 violations, all 6 CONV rules, 4 languages) |
| GAP-P1-03 | P1 | **RESOLVED** | Manifest edge cases covered in cross-language pairs and architecture files |
| GAP-P1-04 | P1 | **RESOLVED** | Created 60-graphql-api.md and 61-event-driven.md (WebSocket + message queue) |
| GAP-P1-05 | P1 | **RESOLVED** | Created docs/test-plan/cross-cutting/incremental-analysis.md (8 scenarios) |
| GAP-P2-01 | P2 | **RESOLVED** | Created docs/test-plan/cross-cutting/performance-benchmarks.md (7 benchmarks) |
| GAP-P2-02 | P2 | **RESOLVED** | Test quality file includes Jest, Vitest, testify, JUnit 5 patterns |
| GAP-P2-03 | P2 | Deferred | Suppression testing deferred to implementation phase |
| GAP-P2-04 | P2 | **RESOLVED** | logistics/ ecosystem (7 companies, 3 languages, 10 cross-company mismatches) |

**Tooling built:**
- `scripts/run-validation-set.sh` — Automated validation runner (extract, run, assert)
- `scripts/validation-health-check.sh` — Validation set self-check (10 checks)
- `.github/workflows/validation-set.yml` — CI gate (4 jobs: regression, health check, golden files, performance)

---

## 2. Coverage Matrix

The table below maps each of the 34 Stricture rules to the validation files that exercise them.

Coverage levels:
- Deep: Multiple distinct scenarios with both positive (PERFECT) and negative (bug) cases
- Shallow: Rule is tangentially exercised but not the primary focus
- None: Rule is not exercised at all by the validation set

### 2.1 Test Quality (TQ) -- 10 Rules

| Rule | 01 Stripe | 02 GitHub | 03 Twilio | 04 SendGrid | 05 YouTube | 06 Shopify | 07 OpenAI | 08 Slack | 09 S3 | 10 Maps | 11 Auth0 | 12 Supabase | Coverage |
|------|-----------|-----------|-----------|-------------|------------|------------|-----------|---------|-------|---------|----------|-------------|----------|
| TQ-no-shallow-assertions | B03 | B03 | B03 | B03 | B03 | B03 | B03 | B03 | B03 | B03 | B03 | B03 | Deep |
| TQ-return-type-verified | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | None |
| TQ-schema-conformance | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | None |
| TQ-error-path-coverage | B01 | B01 | B01 | B01 | B01 | B01 | B01 | B01 | B01 | B01 | B01 | B01 | Deep |
| TQ-assertion-depth | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | None |
| TQ-boundary-tested | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | None |
| TQ-mock-scope | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | None |
| TQ-test-isolation | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | None |
| TQ-negative-cases | B04 | B04 | B04 | B04 | B04 | B04 | B04 | B04 | B04 | B04 | B04 | B04 | Deep |
| TQ-test-naming | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | None |

### 2.2 Architecture (ARCH) -- 6 Rules

| Rule | 01-12 (All) | Coverage |
|------|-------------|----------|
| ARCH-dependency-direction | -- | None |
| ARCH-import-boundary | -- | None |
| ARCH-no-circular-deps | -- | None |
| ARCH-max-file-lines | -- | None |
| ARCH-layer-violation | -- | None |
| ARCH-module-boundary | -- | None |

None of the 12 validation files exercise any ARCH rule. This is expected -- the validation files model client-side API integrations, not multi-layer architectures.

### 2.3 Convention (CONV) -- 6 Rules

| Rule | 01-12 (All) | Coverage |
|------|-------------|----------|
| CONV-file-naming | -- | None |
| CONV-file-header | -- | None |
| CONV-error-format | -- | None |
| CONV-export-naming | -- | None |
| CONV-test-file-location | -- | None |
| CONV-required-exports | -- | None |

None of the 12 validation files exercise any CONV rule. The validation files present code inline in markdown rather than as files with paths, so file-level conventions are not tested.

### 2.4 Contract (CTR) -- 8 Rules

| Rule | 01 Stripe | 02 GitHub | 03 Twilio | 04 SendGrid | 05 YouTube | 06 Shopify | 07 OpenAI | 08 Slack | 09 S3 | 10 Maps | 11 Auth0 | 12 Supabase | Coverage |
|------|-----------|-----------|-----------|-------------|------------|------------|-----------|---------|-------|---------|----------|-------------|----------|
| CTR-request-shape | B05 | B05 | B05 | B05 | B05 | B05 | B05 | B05 | B05 | B05 | B05 | B05 | Deep |
| CTR-response-shape | B06,B12,B14 | B06,B12,B14 | B06,B12,B14 | B06,B12,B14 | B06,B12,B14 | B06,B12,B14 | B06,B12,B14 | B06,B12,B14 | B06,B12,B14 | B06,B12,B14 | B06,B12,B14 | B06,B12,B14 | Deep |
| CTR-status-code-handling | B02 | B02 | B02 | B02 | B02 | B02 | B02 | B02 | B02 | B02 | B02 | B02 | Deep |
| CTR-shared-type-sync | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | None |
| CTR-json-tag-match | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | None |
| CTR-strictness-parity | B08,B09,B10,B11 | B08,B09,B10 | B08,B09,B10 | B08,B09,B10 | B08,B09,B10 | B08,B09,B10,B11 | B08,B09,B10 | B08,B09,B10 | B08,B09,B10,B11 | B08,B09,B10 | B08,B09,B10 | B08,B09,B10,B11 | Deep |
| CTR-manifest-conformance | B07 | B07 | B07 | B07 | B07 | B07 | B07 | B07 | B07 | B07 | B07 | B07 | Deep |
| CTR-dual-test | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | -- | None |

### 2.5 Summary (Updated 2026-02-13)

| Category | Rules with Deep Coverage | Rules with No Coverage | Total |
|----------|------------------------|----------------------|-------|
| TQ (10) | **10** | 0 | 10 |
| ARCH (6) | **6** | 0 | 6 |
| CONV (6) | **6** | 0 | 6 |
| CTR (8) | **8** | 0 | 8 |
| **Total** | **30** | **0** | **30** |

All 30 active rules now have deep coverage from the validation set:
- TQ rules: Files 40 (20 violations), 41 (AI patterns), plus B01/B03/B04 across 01-17
- ARCH rules: Files 30 (Express, 12 violations), 31 (Go, 12 violations)
- CONV rules: File 50 (18 violations across 4 languages)
- CTR rules: Files 01-17 (B02/B05-B15), 20-24 (cross-language contracts), 60-61 (protocols)

---

## 3. Strengths

### 3.1 API Diversity

The 12 APIs cover a broad range of real-world patterns:

- **Monetary precision:** Stripe (B11 -- float vs integer cents), Shopify (string-encoded decimals)
- **Non-standard HTTP semantics:** Slack (always 200, error in body), Google Maps (same pattern), SendGrid (202 with empty body)
- **Streaming protocols:** OpenAI (Server-Sent Events)
- **Pagination variety:** GitHub (Link headers), Slack (cursor-based), YouTube (page tokens), Supabase (Content-Range), Twilio (URI-based)
- **Authentication variety:** Bearer tokens, API keys, Basic auth, HMAC webhooks, JWT
- **Data format edge cases:** Twilio (integer-as-string fields), YouTube (ISO 8601 duration), Supabase (PostgreSQL bigint exceeding Number.MAX_SAFE_INTEGER)

### 3.2 Bug Taxonomy Design

The 15-level bug taxonomy is well-structured with increasing subtlety:

- **B01-B02 (Level 1):** Any tool catches these. Good baseline.
- **B03-B04 (Level 2):** Test quality issues. Validates Stricture's core differentiator.
- **B05-B07 (Level 3):** Contract shape mismatches. The bread and butter of CTR rules.
- **B08-B10 (Level 4):** Strictness precision. Exercises CTR-strictness-parity.
- **B11-B15 (Level 5):** Subtle bugs that require deep analysis. The hardest cases.

The mapping from each bug level to a specific Stricture rule (documented in README.md) makes it clear which rule should detect each case. This is crucial for regression testing.

### 3.3 PERFECT Case Quality

Each PERFECT case demonstrates genuinely production-quality code:

- Proper error handling for all status codes
- Correct TypeScript types matching the manifest
- Validation functions for formats (e.g., Stripe's `ch_*` prefix, Auth0's JWT format)
- Pagination handling
- Rate limit awareness
- Comprehensive test suites with both positive and negative cases

This makes the PERFECT cases reliable for false positive testing -- if Stricture flags these, it is a genuine false positive.

### 3.4 Manifest Fragments

Each validation file includes a `.stricture-manifest.yml` fragment with detailed field-level constraints (types, ranges, formats, enums, required/nullable). This provides the contract source of truth that CTR-manifest-conformance and CTR-strictness-parity need to validate against.

### 3.5 Complementary Test Plan

The per-rule test plans in `docs/test-plan/rules/` (tq.md, arch.md, conv.md, ctr.md) provide thorough unit-level test matrices for all 34 rules, including true positive, true negative, false positive risk, false negative risk, edge cases, and configuration interaction. The cross-cutting sections (infrastructure.md, cli.md, trace.md, manifest-tests.md) cover system-level behavior. The spec-gaps.md file identifies and proposes resolutions for ambiguities.

---

## 4. Gaps and Recommendations

### P0 -- Must Have Before v1

---

**GAP-P0-01: No Go Language Validation Cases**

- **What is missing:** All 12 validation files use TypeScript exclusively. No Go source code or Go tests appear in any validation file.
- **Why it matters:** Go is one of the two launch languages. The Go adapter will parse differently, Go testing patterns (table-driven tests, `t.Run()`, `testify/assert`) behave differently, and Go-specific rules (CTR-json-tag-match) are entirely untested by the validation set. The per-rule test plans include Go examples, but these are synthetic, not real-world API integrations.
- **Recommendation:** Create Go versions of at least 3 validation files (e.g., 01-stripe-go.md, 02-github-go.md, 09-aws-s3-go.md) with Go client code, Go test patterns, and Go struct types with JSON tags.
- **Effort:** L (each file is ~1000 lines to write thoroughly)

---

**GAP-P0-02: No Cross-Language Contract Pairs**

- **What is missing:** CTR-shared-type-sync, CTR-json-tag-match, and CTR-dual-test require analyzing both sides of a contract. The validation files model one side (the consumer) against an external API (the producer). There are no scenarios where both producer and consumer code exist in the validation set.
- **Why it matters:** 3 of 8 CTR rules cannot be tested with the current validation structure. CTR-dual-test (verifying matching test scenarios on both sides) is Stricture's most novel contract rule and has zero validation.
- **Recommendation:** Create 2-3 "internal API" validation files where both the server handler and client code are present. For example:
  - `20-internal-user-api.md` -- Express server + fetch client in same repo
  - `21-go-grpc-pair.md` -- Go gRPC server + Go gRPC client
  - `22-cross-lang-api.md` -- Go HTTP server + TypeScript fetch client (the hardest case)
- **Effort:** L

---

**GAP-P0-03: No ARCH Rule Validation Scenarios**

- **What is missing:** All 6 ARCH rules (dependency-direction, import-boundary, no-circular-deps, max-file-lines, layer-violation, module-boundary) have zero validation set coverage.
- **Why it matters:** ARCH rules operate on project structure, not individual files. They need multi-file, multi-directory scenarios to validate. The per-rule test plans have these but they are synthetic snippets, not realistic project structures.
- **Recommendation:** Create 2 architectural validation files:
  - `30-express-layered-app.md` -- A realistic Express.js application with handler/service/repository layers, demonstrating both correct and incorrect import patterns
  - `31-go-clean-architecture.md` -- A Go application with cmd/internal/pkg structure demonstrating Go-specific architectural patterns
- **Effort:** M per file

---

**GAP-P0-04: No TQ Rule Validation for Mock/Isolation/Naming/Depth**

- **What is missing:** 7 TQ rules are not exercised by the validation set: TQ-return-type-verified, TQ-schema-conformance, TQ-assertion-depth, TQ-boundary-tested, TQ-mock-scope, TQ-test-isolation, TQ-test-naming.
- **Why it matters:** These are Stricture's core differentiator -- "the linter that makes tests mean something." If these rules have false positive issues on real-world test patterns, it undermines the product. The validation set's PERFECT test files exercise happy-path test patterns but do not specifically target these rule categories with realistic diversity.
- **Recommendation:** Create 2 test quality validation files:
  - `40-test-quality-patterns.md` -- A collection of realistic test files with both good and bad patterns for all 10 TQ rules (mock scoping, test isolation, assertion depth into nested API responses, boundary testing for pagination limits, etc.)
  - `41-ai-generated-tests.md` -- Test files that mimic actual AI-generated code patterns (the primary use case): tests that look comprehensive but are actually shallow
- **Effort:** M per file

---

### P1 -- Should Have

---

**GAP-P1-01: No Framework-Specific Patterns**

- **What is missing:** Real code uses frameworks (Express, Fastify, NestJS, React, Next.js, Chi, Gin, Echo, FastAPI, Django). The validation files use raw `fetch()` and bare functions. Framework-specific patterns affect how Stricture detects route handlers, middleware, request parsing, and response serialization.
- **Why it matters:** Route detection patterns are framework-specific (see product spec section on `routePatterns`). If Stricture cannot detect an Express route handler or a Chi router registration, CTR rules produce no output.
- **Recommendation:** Add framework-specific examples to at least 2 validation files. Stripe (Express) and GitHub (Fastify or NestJS) are good candidates. On the Go side, demonstrate Chi/Echo/net-http patterns.
- **Effort:** M

---

**GAP-P1-02: No Validation for CONV Rules**

- **What is missing:** All 6 CONV rules have zero validation set coverage. These rules are straightforward but convention-specific: file naming, file headers, error formatting, export naming, test file location, required exports.
- **Why it matters:** CONV rules are the most likely to produce false positives because they depend heavily on configuration (which naming convention? which header pattern?). Testing them against real codebases with varying conventions is important.
- **Recommendation:** Create 1 convention validation file:
  - `50-convention-patterns.md` -- A realistic project directory listing with correct and incorrect naming patterns, missing/malformed headers, various error message formats
- **Effort:** S

---

**GAP-P1-03: No Manifest Edge Cases in Validation Set**

- **What is missing:** The manifest fragments in the validation files are all well-formed and complete. There are no scenarios with: incomplete manifests, conflicting field definitions between endpoints, manifest version mismatches, fields with `nullable: true` interacting with `required: true`, conditional response schemas (like Supabase's `Prefer` header-dependent responses).
- **Why it matters:** The manifest parser is the foundation of CTR-strictness-parity and CTR-manifest-conformance. If it misparses edge cases, all downstream validation is wrong.
- **Recommendation:** The manifest-tests.md file in the test plan covers many of these. Add 3-5 manifest edge cases to the validation set README as supplementary scenarios.
- **Effort:** S

---

**GAP-P1-04: No GraphQL, gRPC, or WebSocket Contract Scenarios**

- **What is missing:** All 12 validation files use HTTP REST. The product spec mentions gRPC, message queues, and WebSockets. The manifest supports `protocol: message_queue` but no validation file exercises it.
- **Why it matters:** Protocol-specific contract behavior differs significantly. GraphQL has a single endpoint with varying query shapes. gRPC has typed proto definitions. WebSockets have bidirectional message streams. If Stricture claims to support these, they need validation.
- **Recommendation:** Create 2 protocol-specific validation files:
  - `60-graphql-api.md` -- A GraphQL API with query/mutation validation, fragment handling, and type mismatches
  - `61-event-driven.md` -- A message queue / event-driven scenario (Kafka, RabbitMQ, or SQS) with producer/consumer validation
- **Effort:** M per file

---

**GAP-P1-05: No Incremental Analysis Testing**

- **What is missing:** The validation set assumes full analysis of all files. There are no scenarios testing `--changed` or `--staged` modes, where only a subset of files is analyzed. Incremental analysis may miss cross-file violations if only one file in a contract pair changed.
- **Why it matters:** In CI, `stricture --changed` is the primary use case. If it misses violations because it only analyzes the changed file without context from unchanged files, it defeats the purpose.
- **Recommendation:** Document 5 incremental analysis scenarios in a new file `docs/test-plan/cross-cutting/incremental-analysis.md`. Example: "Change the client type but not the server type. Run `--changed`. CTR-request-shape should still detect the mismatch by loading the unchanged server file."
- **Effort:** S

---

### P2 -- Nice to Have

---

**GAP-P2-01: No Performance Validation**

- **What is missing:** The validation set has no scenarios for performance benchmarking. The product spec requires <50ms per file with all rules enabled.
- **Why it matters:** Complex CTR rules (especially CTR-strictness-parity with manifest comparison) could be slow on large files.
- **Recommendation:** Create a performance benchmark suite with: 1 file at 10 LOC, 1 at 100, 1 at 500, 1 at 800 (max), 1 at 1200 (test file max). Measure per-file time for each rule category.
- **Effort:** M

---

**GAP-P2-02: No Multi-Assertion Pattern Diversity**

- **What is missing:** The PERFECT test files use a consistent assertion style (Jest `expect(...).toBe(...)` / `.toEqual(...)`). Real codebases use many patterns: Vitest, Mocha/Chai, Node.js test runner, testify/assert (Go), table-driven tests (Go), property-based testing (fast-check), snapshot testing.
- **Why it matters:** TQ rules must recognize assertions across all supported test frameworks. If TQ-no-shallow-assertions only detects Jest patterns, it misses Vitest or testify.
- **Recommendation:** Add assertion pattern diversity to the test quality validation file (GAP-P0-04). Include examples using at least 3 different test frameworks.
- **Effort:** S

---

**GAP-P2-03: No Suppression Interaction Testing**

- **What is missing:** The validation set does not test `stricture-disable-next-line` or `stricture-disable`/`stricture-enable` blocks within the validation files.
- **Why it matters:** A PERFECT case with suppression comments should still produce zero violations. A bug case with a suppression on the exact bug line should also produce zero violations for that specific rule.
- **Recommendation:** Add 2-3 suppression scenarios to an existing validation file. Test both correct suppressions and mis-targeted suppressions.
- **Effort:** S

---

**GAP-P2-04: No Monorepo or Multi-Package Scenarios**

- **What is missing:** The validation files model single-package projects. Real codebases often have monorepo structures with shared packages, workspace dependencies, and cross-package imports.
- **Why it matters:** ARCH-import-boundary and ARCH-module-boundary behavior in monorepos is complex. Cross-package type sharing affects CTR-shared-type-sync.
- **Recommendation:** Create one monorepo validation scenario:
  - `70-monorepo-fullstack.md` -- A monorepo with `packages/server`, `packages/client`, `packages/shared-types`, demonstrating cross-package contract pairs
- **Effort:** L

---

## 5. Suggested Additional Validation Files

Based on the gap analysis, the following new validation files would significantly improve coverage:

### 5.1 Language Parity (P0)

**File: `13-stripe-go.md`**
Go implementation of the Stripe integration (01-stripe.md). Uses Go structs with `json` tags, `http.Client`, table-driven tests with `testify/assert`, and demonstrates Go-specific patterns: `if err != nil` chains, `t.Run()` subtests, `t.Helper()`. Exercises CTR-json-tag-match (the only Go-specific rule) and provides Go baseline for all 15 bug levels.

**File: `14-github-go.md`**
Go implementation of the GitHub API client (02-github.md). Demonstrates Go's `encoding/json` unmarshalling, struct embedding, and nested type handling. Tests Go error patterns (`fmt.Errorf("...: %w", err)`) against TQ-error-path-coverage.

### 5.2 Internal Contract Pairs (P0)

**File: `20-internal-user-api.md`**
A complete internal API with both server (Express) and client (fetch-based) in the same repo. Includes matching types, route handlers, client functions, and tests for both sides. Exercises CTR-shared-type-sync, CTR-dual-test, and CTR-request-shape / CTR-response-shape with same-repo detection. The 15 bugs include type drift between server and client, missing client tests for server error codes, and shared type divergence.

**File: `21-cross-lang-contract.md`**
A Go HTTP server with a TypeScript client. The contract crosses language boundaries. Demonstrates the hardest case for CTR-json-tag-match (Go `json:"created_at"` vs TypeScript `createdAt`), CTR-shared-type-sync across languages, and CTR-strictness-parity where the Go producer validates but the TypeScript consumer does not.

### 5.3 Architecture (P0)

**File: `30-express-layered-app.md`**
A realistic Express.js application with 4 layers: routes (handlers), services, repositories, models. Includes correct and incorrect import patterns, a circular dependency scenario, a handler that directly queries the database, and a file exceeding 800 lines. Exercises all 6 ARCH rules with both positive and negative cases.

### 5.4 Test Quality Depth (P0)

**File: `40-test-quality-showcase.md`**
A dedicated test quality validation file focused on the 7 uncovered TQ rules. Contains realistic source files (a user service, a pagination utility, a config loader) and test files demonstrating: shallow assertions at varying depths, incomplete return type coverage, missing boundary tests, global mocks without cleanup, tests with shared mutable state, non-descriptive test names, and missing negative cases. Each bug maps to exactly one TQ rule.

**File: `41-ai-generated-test-patterns.md`**
Test files that mimic patterns observed in AI-generated code: tests that import and call the function under test, use `toBeDefined()` extensively, never test error paths, use generic names like "should work" and "handles edge case", mock everything at module scope without cleanup. This file validates that Stricture catches the exact patterns it was designed to detect. The PERFECT version shows what AI-generated tests should look like after Stricture enforcement.

### 5.5 Protocol Diversity (P1)

**File: `60-graphql-github.md`**
A GraphQL client for the GitHub GraphQL API (v4). Demonstrates query/mutation typing, fragment handling, nullable fields in GraphQL responses, and error handling for GraphQL-specific error formats (`{ data: null, errors: [...] }`). Tests whether CTR rules can handle single-endpoint APIs with varying query shapes.

**File: `61-websocket-chat.md`**
A WebSocket-based chat application with typed message schemas (join, message, leave, error). Demonstrates bidirectional contract validation: the client sends messages that must match the server's expected schema, and the server sends messages that must match the client's expected schema. Tests CTR rules on non-HTTP protocols.

---

## 6. Testing Strategy for Stricture Itself

Beyond the validation set, the following testing strategies should be implemented to verify that Stricture's engine works correctly.

### 6.1 Unit Tests for Each Rule's Detection Logic

Each of the 34 rules needs a dedicated test file with:

- **Minimum 10 true positive cases** (code that should trigger the rule)
- **Minimum 5 true negative cases** (code that should not trigger the rule)
- **Minimum 3 false positive risk cases** (edge cases that look like violations but are not)
- **Minimum 2 false negative risk cases** (edge cases that look clean but contain violations)

The per-rule test plans in `docs/test-plan/rules/` already define these. Implementation should use the exact code snippets from those plans as test fixtures.

**Implementation approach:** Create a `test/rules/` directory with one test file per rule (e.g., `test/rules/tq-no-shallow-assertions.test.ts`). Each test creates an in-memory `UnifiedFileModel` from a fixture string, runs the rule, and asserts on the violations returned.

### 6.2 Integration Tests for the Manifest Parser

The manifest parser must be tested separately from the rules:

- **Valid manifest parsing:** All constraint types (range, enum, format, minLength, maxLength, precision, required, nullable) are correctly extracted from YAML
- **Manifest validation:** Missing required fields, unknown field types, invalid ranges, circular service references
- **Manifest merging:** Multiple contracts for the same endpoint, inherited constraints from shared types
- **Manifest versioning:** `manifest_version: "1.0"` accepted, `manifest_version: "99.0"` rejected with clear error

**Implementation approach:** Create `test/manifest/parser.test.ts` with ~30 test cases. Fixtures are YAML strings.

### 6.3 Snapshot / Golden-File Testing

For each output format (text, JSON, SARIF, JUnit), maintain golden files that capture the exact expected output for a fixed set of input files.

- **Create fixtures:** 5 TypeScript files + 5 Go files with known violations across all rule categories
- **Generate golden files:** Run Stricture, capture output, review manually, commit as `test/golden/output-text.txt`, `test/golden/output.json`, `test/golden/output.sarif`, `test/golden/output-junit.xml`
- **Assert in CI:** Run Stricture on the same fixtures, diff against golden files. Any delta fails the test.
- **Regenerate on intentional changes:** A script (`npm run update-golden`) regenerates golden files after rule changes.

**Why golden files:** They catch unintentional formatting changes, field ordering drift, and schema violations that unit tests miss. They also serve as documentation of exact output behavior.

### 6.4 Mutation Testing

Inject bugs into PERFECT validation cases and verify that Stricture detects them.

**Approach:**

1. Take each PERFECT case from the 12 validation files
2. Apply systematic mutations:
   - Remove one field from a response type
   - Change a field type (string -> number)
   - Remove one assertion from a test
   - Delete an error handling branch
   - Remove a status code check
   - Weaken an enum validation (handle 2 of 3 values)
3. Run Stricture on the mutated code
4. Assert that at least one violation is produced
5. Assert that the violation matches the expected rule

**Automation:** A script reads each PERFECT case, applies each mutation programmatically (regex-based or AST-based), writes the mutated file, runs Stricture, checks for expected violation.

**Why mutation testing:** It validates that Stricture is sensitive enough to catch small changes. A rule that passes both PERFECT and mutated code is broken.

### 6.5 Performance Benchmarks

Measure and regress on performance targets from the product spec:

| Benchmark | Target | What to Measure |
|-----------|--------|-----------------|
| Cold start (500 files) | < 3s | `time stricture` on a 500-file repo |
| Cached run (500 files) | < 1s | Second `time stricture` run |
| Incremental (20 files) | < 2s | `time stricture --changed` with 20 modifications |
| Per-file (single) | < 50ms | `time stricture single-file.ts` |
| Memory (10,000 files) | < 500MB | `node --max-old-space-size=500 stricture` on a 10K file repo |

**Implementation:** Create a benchmark directory with synthetic repos at each scale. Run benchmarks in CI with a `--benchmark` flag. Store results in a time-series database or CSV for regression tracking.

### 6.6 CLI Testing

Every CLI flag and subcommand needs an integration test:

- **Flag parsing:** 28 test cases defined in `docs/test-plan/tools/cli.md`
- **Exit codes:** 0 (no errors), 1 (violations), 2 (config/parse error)
- **Output format correctness:** Validate JSON against schema, SARIF against SARIF 2.1.0 spec, JUnit against XSD
- **Signal handling:** SIGINT during lint should exit cleanly
- **stdin support:** `cat file.ts | stricture --stdin`

**Implementation approach:** Use `execa` or `child_process.execFile` to spawn Stricture as a subprocess with specific args and assert on stdout, stderr, and exit code.

### 6.7 CI/CD Integration Testing

Test Stricture in realistic CI environments:

- **GitHub Actions:** Upload SARIF, verify it appears in Code Scanning results
- **Pre-commit hook:** Run `stricture --staged` in a git hook, verify it blocks commits with violations
- **GitLab CI:** Generate JUnit XML, verify it appears in test results
- **Docker:** Run in a minimal Node.js Docker image, verify no missing dependencies

**Implementation:** A `.github/workflows/stricture-ci-test.yml` that runs the full CI test matrix on each PR.

---

## 7. Automation Opportunities

### 7.1 Automated Validation Runner

Create a test runner script that:

1. Extracts each code block (PERFECT, B01-B15) from each validation markdown file
2. Writes them to a temporary directory as real files
3. Runs Stricture on each extracted file
4. Asserts:
   - PERFECT: 0 violations (false positive check)
   - B01: violation matches `TQ-error-path-coverage` (or the mapped rule from README.md)
   - B02: violation matches `CTR-status-code-handling`
   - etc.
5. Produces a pass/fail report per file per bug level

**Implementation:**

```
scripts/run-validation-set.sh
  |-- For each *.md in docs/test-plan/validation-set/
  |     |-- Extract code blocks using markdown parser
  |     |-- Write PERFECT to temp/01-stripe/perfect/
  |     |-- Write B01 to temp/01-stripe/b01/
  |     |-- ...
  |     |-- Run: stricture temp/01-stripe/perfect/ --format json
  |     |-- Assert: violations.length === 0
  |     |-- Run: stricture temp/01-stripe/b01/ --format json
  |     |-- Assert: violations includes expected rule
  |-- Report: 192/192 scenarios passed
```

**Effort:** M (the markdown extraction is the tricky part; consider using `remark` for reliable code block extraction)

### 7.2 Regression Gate

Add a CI step that runs the validation set after every rule change:

```yaml
# .github/workflows/validation-set.yml
name: Validation Set Regression
on:
  push:
    paths:
      - "src/rules/**"
      - "src/adapters/**"
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - run: npm run test:validation-set
      - run: |
          if [ $? -ne 0 ]; then
            echo "Validation set regression detected!"
            exit 1
          fi
```

### 7.3 Programmatic Validation Generation

For rules with well-defined patterns, generate validation cases automatically:

- **CONV-file-naming:** Generate filenames in every naming convention (kebab-case, camelCase, PascalCase, snake_case) and verify detection for each style setting
- **ARCH-max-file-lines:** Generate files at 799, 800, 801, 1199, 1200, 1201 lines and verify threshold behavior
- **TQ-no-shallow-assertions:** Generate all 8 shallow assertion patterns (`toBeDefined`, `toBeTruthy`, `not.toBeNull`, `toBeInstanceOf(Object)`, `typeof check`, `Array.isArray`, Go `== nil`, Go `err != nil`) and verify detection

**Implementation:** A `scripts/generate-validation-cases.ts` script that outputs markdown or TypeScript test fixtures.

### 7.4 Continuous Accuracy Tracking

Track detection accuracy over time:

| Metric | Formula | Target |
|--------|---------|--------|
| True positive rate | (correctly flagged bugs) / (total bugs) | 100% for B01-B10, >90% for B11-B15 |
| False positive rate | (PERFECT cases flagged) / (total PERFECT cases) | 0% |
| Detection latency | Time from bug introduction to detection | < 50ms per file |

Store these metrics per release. Alert on regressions.

### 7.5 Validation Set Health Check

A script that verifies the validation set itself is well-formed:

1. Every validation file has exactly 16 code blocks (PERFECT + B01-B15)
2. Every code block compiles (or at least has balanced braces)
3. Every manifest fragment is valid YAML
4. Every bug level maps to a Stricture rule listed in README.md
5. Every PERFECT case has a corresponding test file (if tests are included)
6. No duplicate manifest contract IDs across files

**Implementation:** A linter for the validation set. "Who watches the watchmen?"

---

## Appendix: Complete Rule Inventory

For reference, the 34 rules and their validation coverage status:

| # | Rule ID | Category | Validation Set | Test Plan |
|---|---------|----------|----------------|-----------|
| 1 | TQ-no-shallow-assertions | TQ | Deep (B03 x12) | Deep (tq.md TP-NSA-01..10) |
| 2 | TQ-return-type-verified | TQ | None | Deep (tq.md TP-RTV-01..10) |
| 3 | TQ-schema-conformance | TQ | None | Deep (tq.md) |
| 4 | TQ-error-path-coverage | TQ | Deep (B01 x12) | Deep (tq.md) |
| 5 | TQ-assertion-depth | TQ | None | Deep (tq.md) |
| 6 | TQ-boundary-tested | TQ | None | Deep (tq.md) |
| 7 | TQ-mock-scope | TQ | None | Deep (tq.md) |
| 8 | TQ-test-isolation | TQ | None | Deep (tq.md) |
| 9 | TQ-negative-cases | TQ | Deep (B04 x12) | Deep (tq.md) |
| 10 | TQ-test-naming | TQ | None | Deep (tq.md) |
| 11 | ARCH-dependency-direction | ARCH | None | Deep (arch.md TP-DD-01..10) |
| 12 | ARCH-import-boundary | ARCH | None | Deep (arch.md) |
| 13 | ARCH-no-circular-deps | ARCH | None | Deep (arch.md) |
| 14 | ARCH-max-file-lines | ARCH | None | Deep (arch.md) |
| 15 | ARCH-layer-violation | ARCH | None | Deep (arch.md) |
| 16 | ARCH-module-boundary | ARCH | None | Deep (arch.md) |
| 17 | CONV-file-naming | CONV | None | Deep (conv.md TP-FN-01..10) |
| 18 | CONV-file-header | CONV | None | Deep (conv.md) |
| 19 | CONV-error-format | CONV | None | Deep (conv.md) |
| 20 | CONV-export-naming | CONV | None | Deep (conv.md) |
| 21 | CONV-test-file-location | CONV | None | Deep (conv.md) |
| 22 | CONV-required-exports | CONV | None | Deep (conv.md) |
| 23 | CTR-request-shape | CTR | Deep (B05 x12) | Deep (ctr.md TP-RS-01..10) |
| 24 | CTR-response-shape | CTR | Deep (B06,B12,B14 x12) | Deep (ctr.md) |
| 25 | CTR-status-code-handling | CTR | Deep (B02 x12) | Deep (ctr.md) |
| 26 | CTR-shared-type-sync | CTR | None | Deep (ctr.md) |
| 27 | CTR-json-tag-match | CTR | None | Deep (ctr.md) |
| 28 | CTR-dual-test | CTR | None | Deep (ctr.md) |
| 29 | CTR-strictness-parity | CTR | Deep (B08-B11 x12) | Deep (manifest-tests.md) |
| 30 | CTR-manifest-conformance | CTR | Deep (B07 x12) | Deep (manifest-tests.md) |
| 31-34 | Reserved for v0.2 | -- | N/A | N/A |

Note: The product spec lists "34 rules (10 TQ + 6 ARCH + 6 CONV + 8 CTR + 2 reserved for v0.2)" yielding 30 active rules + 2 reserved = 32 entries, with 2 more slots reserved. The table above covers the 30 active rules.
