# Stricture Validation Set

Real-world API integration examples for validating Stricture's detection accuracy.

## Purpose

This validation set provides concrete code examples against **12 well-known APIs** with:
- A **perfect** integration for each (Stricture must pass — false positive detection)
- **15 increasingly subtle bugs** for each (Stricture must catch — false negative detection)

Total: **12 APIs x 16 cases = 192 test scenarios**, plus a multi-company logistics ecosystem with 10 cross-company mismatch scenarios.

## Structure

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

### Multi-Company Integration

| Directory | Description |
|-----------|-------------|
| [logistics/](logistics/README.md) | 7-company dropshipping ecosystem (3 languages, 8 contracts, 10 cross-company mismatches) |

### QA Review

| File | Description |
|------|-------------|
| [QA-REVIEW.md](QA-REVIEW.md) | Coverage analysis, gap identification, and testing strategy recommendations |

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

## How to Use This Validation Set

1. **False positive test**: Run Stricture against each "PERFECT" example. Zero violations expected.
2. **Detection test**: Run Stricture against each "B01"-"B15" example. Specific violation expected.
3. **Scoring**: Track detection rate per bug level. Goal: 100% for B01-B10, >90% for B11-B15.
4. **Regression**: If Stricture changes, re-run all 192 scenarios. No regressions allowed.
