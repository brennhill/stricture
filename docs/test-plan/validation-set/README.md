# Stricture Validation Set

Real-world code examples for validating Stricture's detection accuracy across all 30 active rules, 5 languages, and 5 frameworks.

## Purpose

This validation set provides concrete code examples with:
- A **perfect** integration for each (Stricture must pass — false positive detection)
- **15 increasingly subtle bugs** for each (Stricture must catch — false negative detection)
- **Cross-language contract pairs** (Go↔TS, Python↔TS, Java↔TS, Go↔Python)
- **Framework-specific patterns** (Express, Fastify, NestJS, Chi, Gin, Spring Boot)
- **Architecture and convention validation** scenarios

**Coverage:** 30/30 active rules | 5 languages | 5 frameworks | 3 protocols

---

## API Integrations (TypeScript — 12 APIs x 16 cases = 192 scenarios)

| File | API | Why included |
|------|-----|-------------|
| [01-stripe.md](01-stripe.md) | Stripe Payments | Monetary precision, enum states, webhooks, idempotency |
| [02-github.md](02-github.md) | GitHub REST API | Pagination, nested objects, optional fields, status codes |
| [03-twilio.md](03-twilio.md) | Twilio SMS/Voice | Callback URLs, phone format validation, status callbacks |
| [04-sendgrid.md](04-sendgrid.md) | SendGrid Email | Batch operations, template variables, bounce handling |
| [05-youtube.md](05-youtube.md) | YouTube Data API | Quota limits, pagination tokens, nested resources |
| [06-shopify.md](06-shopify.md) | Shopify Admin API | Monetary values, inventory, webhook HMAC |
| [07-openai.md](07-openai.md) | OpenAI Chat API | Streaming, token limits, model enum validation |
| [08-slack.md](08-slack.md) | Slack Web API | Rate limits, cursor pagination, message blocks |
| [09-aws-s3.md](09-aws-s3.md) | AWS S3 (presigned URLs) | Expiration, content-type, ACL validation |
| [10-google-maps.md](10-google-maps.md) | Google Maps Geocoding | Coordinate ranges, status enums, component parsing |
| [11-auth0.md](11-auth0.md) | Auth0 Management API | JWT validation, token expiration, scopes |
| [12-supabase.md](12-supabase.md) | Supabase REST API | Row-level security, type coercion, range queries |

## Language Parity (Go, Python, Java — 5 x 16 = 80 scenarios)

| File | Language | API | Why included |
|------|----------|-----|-------------|
| [13-stripe-go.md](13-stripe-go.md) | Go | Stripe | Struct json tags, `if err != nil`, table-driven tests, CTR-json-tag-match |
| [14-github-go.md](14-github-go.md) | Go | GitHub | `encoding/json`, struct embedding, `*string` nullable, Link header pagination |
| [15-aws-s3-go.md](15-aws-s3-go.md) | Go | AWS S3 | AWS SDK v2, presigned URLs, interface-based testing |
| [16-stripe-python.md](16-stripe-python.md) | Python | Stripe | Pydantic models, httpx, pytest fixtures, type hints |
| [17-stripe-java.md](17-stripe-java.md) | Java | Stripe | Records, @JsonProperty, HttpClient, JUnit 5 @ParameterizedTest |

## Cross-Language Contract Pairs (5 pairs — CTR-shared-type-sync, CTR-json-tag-match, CTR-dual-test)

| File | Languages | Why included |
|------|-----------|-------------|
| [20-internal-user-api.md](20-internal-user-api.md) | TS server + TS client | Same-repo contract pair, shared types, dual-test compliance |
| [21-cross-lang-go-ts.md](21-cross-lang-go-ts.md) | Go server + TS client | Hardest case: `json:"created_at"` vs `createdAt` |
| [22-cross-lang-python-ts.md](22-cross-lang-python-ts.md) | Python (FastAPI) + TS client | `snake_case` ↔ `camelCase` automatic mismatch |
| [23-cross-lang-java-ts.md](23-cross-lang-java-ts.md) | Java (Spring Boot) + TS client | `@JsonProperty` vs TS interfaces |
| [24-cross-lang-go-python.md](24-cross-lang-go-python.md) | Go server + Python client | Go json tags vs Pydantic models |

## Architecture Validation (All 6 ARCH rules)

| File | Scope | Why included |
|------|-------|-------------|
| [30-express-layered-app.md](30-express-layered-app.md) | TypeScript/Express | 4-layer app: routes→services→repos→models, 12 violation scenarios |
| [31-go-clean-architecture.md](31-go-clean-architecture.md) | Go | cmd/internal/pkg structure, interface-driven design, 12 violation scenarios |

## Test Quality Validation (All 10 TQ rules)

| File | Focus | Why included |
|------|-------|-------------|
| [40-test-quality-patterns.md](40-test-quality-patterns.md) | TQ rule coverage | 7 previously uncovered TQ rules with realistic source + test patterns |
| [41-ai-generated-test-patterns.md](41-ai-generated-test-patterns.md) | AI anti-patterns | 10 AI-generated test anti-patterns Stricture must catch (marketing showcase) |

## Convention Validation (All 6 CONV rules)

| File | Focus | Why included |
|------|-------|-------------|
| [50-convention-patterns.md](50-convention-patterns.md) | CONV rule coverage | File naming, headers, error format, export naming across TS/Go/Python/Java |

## Protocol Diversity

| File | Protocol | Why included |
|------|----------|-------------|
| [60-graphql-api.md](60-graphql-api.md) | GraphQL | Single endpoint, varying query shapes, connection pagination, union types |
| [61-event-driven.md](61-event-driven.md) | WebSocket + Message Queue | Bidirectional contracts, event schemas, consumer validation |

## Framework Patterns

| File | Frameworks | Why included |
|------|------------|-------------|
| [70-framework-patterns-js.md](70-framework-patterns-js.md) | Express, Fastify, NestJS, Next.js, Hono | JS/TS route detection, middleware, request parsing patterns |
| [71-framework-patterns-go.md](71-framework-patterns-go.md) | net/http, Chi, Gin, Echo | Go route detection, handler patterns, middleware |
| [72-framework-patterns-java.md](72-framework-patterns-java.md) | Spring Boot, JAX-RS, Micronaut | Java annotation-based routing, DI, validation |

## Multi-Company Integration

| Directory | Description |
|-----------|-------------|
| [logistics/](logistics/README.md) | 7-company dropshipping ecosystem (3 languages, 8 contracts, 10 cross-company mismatches) |

## QA Review

| File | Description |
|------|-------------|
| [QA-REVIEW.md](QA-REVIEW.md) | Coverage analysis, gap identification, and testing strategy recommendations |

---

## Bug Taxonomy (15 levels, most obvious to most subtle)

Each API file contains examples of all 15 bug levels. Bugs are ordered from "any linter catches this" to "only Stricture catches this":

### Level 1 — Obvious (any tool catches)
| # | Bug | Stricture Rule | Description |
|---|-----|---------------|-------------|
| B01 | No error handling | TQ-error-path-coverage | No try/catch, no `.catch()`, errors crash caller |
| B02 | No status code check | CTR-status-code-handling | Treats all responses as success |

### Level 2 — Moderate (test quality issues)
| # | Bug | Stricture Rule | Description |
|---|-----|---------------|-------------|
| B03 | Shallow test assertions | TQ-no-shallow-assertions | `expect(result).toBeDefined()` instead of shape checks |
| B04 | Missing negative tests | TQ-negative-cases | Only tests happy path, no error/edge tests |

### Level 3 — Structural (contract shape issues)
| # | Bug | Stricture Rule | Description |
|---|-----|---------------|-------------|
| B05 | Request missing required fields | CTR-request-shape | Client omits fields server requires |
| B06 | Response type mismatch | CTR-response-shape | Client type has extra/missing fields vs API |
| B07 | Wrong field types | CTR-manifest-conformance | String ID stored as number, date as string |

### Level 4 — Precision (strictness parity)
| # | Bug | Stricture Rule | Description |
|---|-----|---------------|-------------|
| B08 | Incomplete enum handling | CTR-strictness-parity | Handles 3 of 5 possible status values |
| B09 | Missing range validation | CTR-strictness-parity | No bounds check on amounts/quantities |
| B10 | Format not validated | CTR-strictness-parity | UUID accepted as plain string, no regex |

### Level 5 — Subtle (only deep analysis catches)
| # | Bug | Stricture Rule | Description |
|---|-----|---------------|-------------|
| B11 | Precision loss on currency | CTR-strictness-parity | Float math on monetary values (0.1 + 0.2 != 0.3) |
| B12 | Nullable field crashes | CTR-response-shape | Assumes optional field always present, crashes on null |
| B13 | Missing webhook verification | CTR-request-shape | Accepts POST body without signature verification |
| B14 | Pagination terminated early | CTR-response-shape | Ignores `has_more`/`next_page_token`, returns partial data |
| B15 | Race condition (read-modify-write) | CTR-request-shape | No version/etag check, overwrites concurrent changes |

---

## Rule Coverage Summary

| Category | Rules | Validation Files |
|----------|-------|-----------------|
| TQ (10) | All 10 covered | 01-12 (B01,B03,B04), 40 (all 10), 41 (AI patterns) |
| ARCH (6) | All 6 covered | 30 (Express), 31 (Go) |
| CONV (6) | All 6 covered | 50 (multi-language) |
| CTR (8) | All 8 covered | 01-12 (B02,B05-B15), 20-24 (cross-contract) |
| **Total** | **30/30** | **34 files + logistics ecosystem** |

## Language Coverage

| Language | Files | Patterns |
|----------|-------|----------|
| TypeScript | 01-12, 20-21, 22-23, 30, 40-41, 50, 60-61, 70 | Jest, Vitest, fetch, Express, NestJS, Next.js |
| Go | 13-15, 21, 24, 31, 71 | testify, table-driven tests, chi, gin, echo |
| Python | 16, 22, 24, 50 | pytest, httpx, pydantic, FastAPI |
| Java | 17, 23, 50, 72 | JUnit 5, HttpClient, records, Spring Boot |
| Multi-language | logistics/ | TS + Go + Python (7-company ecosystem) |

---

## Automation

| Tool | Purpose |
|------|---------|
| `scripts/run-validation-set.sh` | Extracts code blocks, runs Stricture, asserts violations |
| `scripts/validation-health-check.sh` | Validates the validation set structure itself |
| `.github/workflows/validation-set.yml` | CI gate: regression on rule changes, health check, benchmarks |

## How to Use

1. **False positive test**: Run Stricture against each "PERFECT" example. Zero violations expected.
2. **Detection test**: Run Stricture against each "B01"-"B15" example. Specific violation expected.
3. **Scoring**: Track detection rate per bug level. Goal: 100% for B01-B10, >90% for B11-B15.
4. **Regression**: Run `./scripts/run-validation-set.sh` after any rule change.
5. **Health check**: Run `./scripts/validation-health-check.sh` to validate file structure.

## Manifest Template

Each API file includes a `.stricture-manifest.yml` fragment that declares the contract. Example:

```yaml
contracts:
  - id: "stripe-charges"
    producer: stripe           # Stripe is the producer
    consumers: [my-service]    # Our service consumes
    protocol: http
    endpoints:
      - path: "/v1/charges"
        method: POST
        request:
          fields:
            amount:   { type: integer, range: [50, 99999999], required: true }
            currency: { type: enum, values: ["usd", "eur", "gbp", ...], required: true }
            source:   { type: string, format: "tok_*|card_*", required: true }
        response:
          fields:
            id:       { type: string, format: "ch_*", required: true }
            status:   { type: enum, values: ["succeeded", "pending", "failed"], required: true }
            amount:   { type: integer, range: [50, 99999999], required: true }
        status_codes: [200, 400, 402, 429, 500]
```
