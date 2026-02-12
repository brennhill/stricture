# Cross-Service Manifest Tests

Tests for the Stricture manifest system -- cross-service contract validation, strictness parity enforcement, and manifest lifecycle. Covers sections 13.2 through 13.8 of the product spec.

> **Companion Document:** [Product Spec, Section 13](../../product-spec.md)
> **Related Rules:** CTR-strictness-parity, CTR-manifest-conformance
> **Prerequisite:** These tests require a `.stricture.yml` with a `manifest` section. Without it, all manifest-related rules are no-ops.

---

## Table of Contents

- [1. Manifest Parsing & Validation](#1-manifest-parsing--validation)
- [2. Service Declarations](#2-service-declarations)
- [3. Contract Definitions -- HTTP](#3-contract-definitions----http)
- [4. Contract Definitions -- Message Queue](#4-contract-definitions----message-queue)
- [5. Contract Definitions -- Multi-Consumer](#5-contract-definitions----multi-consumer)
- [6. Strictness Levels](#6-strictness-levels)
- [7. Strictness Rules Enforcement](#7-strictness-rules-enforcement)
- [8. Per-Service Configuration](#8-per-service-configuration)
- [9. CTR-manifest-conformance](#9-ctr-manifest-conformance)
- [10. CTR-strictness-parity -- Range Parity](#10-ctr-strictness-parity----range-parity)
- [11. CTR-strictness-parity -- Enum Parity](#11-ctr-strictness-parity----enum-parity)
- [12. CTR-strictness-parity -- Format Parity](#12-ctr-strictness-parity----format-parity)
- [13. CTR-strictness-parity -- Length Parity](#13-ctr-strictness-parity----length-parity)
- [14. CTR-strictness-parity -- Cross-Language](#14-ctr-strictness-parity----cross-language)
- [15. Stricture Audit Command](#15-stricture-audit-command)
- [16. Stricture Trace Command](#16-stricture-trace-command)
- [17. Edge Cases & Error Handling](#17-edge-cases--error-handling)
- [18. Performance & Scale](#18-performance--scale)

---

## 1. Manifest Parsing & Validation

### 1.1 Valid Manifests

**MFP-VALID-01: Minimal valid manifest (1 service, 1 contract, 1 endpoint)**

- **Input manifest** (`stricture-manifest.yml`):
```yaml
manifest_version: "1.0"
name: "minimal-platform"

services:
  api-service:
    repo: "github.com/acme/api"
    language: typescript
    role: producer
    stricture_config: ".stricture.yml"

contracts:
  - id: "health-api"
    producer: api-service
    consumers: []
    protocol: http
    endpoints:
      - path: "/healthz"
        method: GET
        response:
          type: HealthResponse
          fields:
            status: { type: string, required: true }
        status_codes: [200]

strictness:
  minimum: minimal
  rules: {}
```
- **Expected:** Parses without error. Returns valid manifest object with 1 service, 1 contract, 1 endpoint, 1 field.
- **Verification:** `stricture audit` exits 0 and reports the single contract.

**MFP-VALID-02: Full manifest (multiple services, contracts, endpoints, messages)**

- **Input manifest** (`stricture-manifest.yml`):
```yaml
manifest_version: "1.0"
name: "acme-platform"

services:
  api-gateway:
    repo: "github.com/acme/api-gateway"
    language: typescript
    role: producer
    stricture_config: ".stricture.yml"

  user-service:
    repo: "github.com/acme/user-service"
    language: go
    role: both
    stricture_config: ".stricture.yml"

  billing-service:
    repo: "github.com/acme/billing-service"
    language: go
    role: consumer
    stricture_config: ".stricture.yml"

  web-frontend:
    repo: "github.com/acme/web-app"
    language: typescript
    role: consumer

contracts:
  - id: "user-api"
    producer: user-service
    consumers: [api-gateway, web-frontend]
    protocol: http
    endpoints:
      - path: "/api/users/:id"
        method: GET
        response:
          type: User
          fields:
            id:         { type: integer, range: [1, 2147483647], required: true }
            name:       { type: string, minLength: 1, maxLength: 255, required: true }
            email:      { type: string, format: email, required: true }
            role:       { type: enum, values: ["admin", "user", "viewer"], required: true }
            created_at: { type: string, format: iso8601, required: true }
        status_codes: [200, 400, 404, 500]
        error_shape:
          code:    { type: string, required: true }
          message: { type: string, required: true }

      - path: "/api/users"
        method: POST
        request:
          type: CreateUserRequest
          fields:
            name:  { type: string, minLength: 1, maxLength: 255, required: true }
            email: { type: string, format: email, required: true }
            role:  { type: enum, values: ["admin", "user", "viewer"], required: true }
        response:
          type: User
        status_codes: [201, 400, 409, 500]

  - id: "billing-events"
    producer: billing-service
    consumers: [user-service]
    protocol: message_queue
    queue: "billing.events"
    messages:
      - event: "invoice.created"
        fields:
          invoice_id: { type: string, format: uuid, required: true }
          user_id:    { type: integer, range: [1, 2147483647], required: true }
          amount:     { type: number, range: [0.01, 999999.99], precision: 2, required: true }
          currency:   { type: enum, values: ["USD", "EUR", "GBP"], required: true }

strictness:
  minimum: strict
  rules:
    numeric-range-required: true
    string-length-required: true
    enum-exhaustive: true
    error-shape-required: true
    status-codes-exhaustive: true
```
- **Expected:** Parses without error. 4 services, 2 contracts, 2 HTTP endpoints + 1 message event, all field constraints preserved.
- **Verification:** `stricture audit` reports all contracts and their field counts. No parse errors.

**MFP-VALID-03: Manifest with all field constraint types**

- **Input manifest** (contracts section only):
```yaml
contracts:
  - id: "constraint-showcase"
    producer: api-service
    consumers: [client-service]
    protocol: http
    endpoints:
      - path: "/api/products/:id"
        method: GET
        response:
          type: Product
          fields:
            id:          { type: integer, range: [1, 9999999], required: true }
            name:        { type: string, minLength: 1, maxLength: 500, required: true }
            price:       { type: number, range: [0.01, 99999.99], precision: 2, required: true }
            currency:    { type: enum, values: ["USD", "EUR", "GBP", "JPY"], required: true }
            sku:         { type: string, format: uuid, required: true }
            email:       { type: string, format: email, required: false }
            website:     { type: string, format: url, required: false }
            created_at:  { type: string, format: iso8601, required: true }
            description: { type: string, maxLength: 10000, required: false }
        status_codes: [200, 400, 404, 500]
        error_shape:
          code:    { type: string, required: true }
          message: { type: string, required: true }
```
- **Expected:** All constraint types parsed correctly: `range`, `enum`/`values`, `format`, `minLength`, `maxLength`, `precision`, `required`.
- **Verification:** Each field's constraints are accessible in the parsed model. `stricture audit` displays the correct constraint for each field.

**MFP-VALID-04: Manifest referencing OpenAPI spec**

- **Input manifest** (contract section):
```yaml
contracts:
  - id: "user-api"
    producer: user-service
    consumers: [web-frontend]
    protocol: http
    spec: "specs/user-api.openapi.yaml"
    endpoints:
      - path: "/api/users/:id"
        method: GET
        response:
          type: User
          fields:
            id:   { type: integer, range: [1, 2147483647], required: true }
            name: { type: string, maxLength: 255, required: true }
        status_codes: [200, 404]
```
- **Expected:** `spec` field is stored as metadata. Stricture does not parse the OpenAPI file itself (that is the spec author's responsibility) but records it for reporting. No parse error.
- **Verification:** `stricture audit` output includes the spec reference. The manifest fields take precedence for parity checks.

**MFP-VALID-05: Manifest with all strictness rules enabled**

- **Input manifest** (strictness section only):
```yaml
strictness:
  minimum: exhaustive
  rules:
    numeric-range-required: true
    string-length-required: true
    enum-exhaustive: true
    error-shape-required: true
    status-codes-exhaustive: true
```
- **Expected:** All 5 rules are parsed and active. Minimum level is `exhaustive`.
- **Verification:** A field declared as `{ type: string }` (no constraints) triggers a minimum-level violation AND a `string-length-required` rule violation.

### 1.2 Invalid Manifests

**MFP-INVALID-01: Missing manifest_version**

- **Input manifest:**
```yaml
name: "acme-platform"
services:
  api:
    repo: "github.com/acme/api"
    language: typescript
    role: producer
contracts: []
strictness:
  minimum: minimal
  rules: {}
```
- **Expected:** Parse error: `manifest_version is required`.
- **Verification:** Exit code 2. Error message includes `manifest_version`.

**MFP-INVALID-02: Missing name**

- **Input manifest:**
```yaml
manifest_version: "1.0"
services:
  api:
    repo: "github.com/acme/api"
    language: typescript
    role: producer
contracts: []
strictness:
  minimum: minimal
  rules: {}
```
- **Expected:** Parse error: `name is required`.
- **Verification:** Exit code 2.

**MFP-INVALID-03: Missing services**

- **Input manifest:**
```yaml
manifest_version: "1.0"
name: "acme-platform"
contracts: []
strictness:
  minimum: minimal
  rules: {}
```
- **Expected:** Parse error: `services is required (must contain at least 1 service)`.
- **Verification:** Exit code 2.

**MFP-INVALID-04: Missing contracts**

- **Input manifest:**
```yaml
manifest_version: "1.0"
name: "acme-platform"
services:
  api:
    repo: "github.com/acme/api"
    language: typescript
    role: producer
strictness:
  minimum: minimal
  rules: {}
```
- **Expected:** Parse error: `contracts is required`.
- **Verification:** Exit code 2.

**MFP-INVALID-05: Invalid YAML syntax**

- **Input manifest:**
```yaml
manifest_version: "1.0"
name: "acme-platform"
services:
  api:
    repo: "github.com/acme/api"
    language: typescript
    role: [producer   # unclosed bracket
contracts: []
```
- **Expected:** Parse error with line number: `YAML syntax error at line 7: unclosed bracket`.
- **Verification:** Exit code 2. Error includes line number and column if available.

**MFP-INVALID-06: Unknown fields (lenient parsing)**

- **Input manifest:**
```yaml
manifest_version: "1.0"
name: "acme-platform"
description: "This field is not in the spec"
services:
  api:
    repo: "github.com/acme/api"
    language: typescript
    role: producer
    team: "backend"  # Unknown field
contracts: []
strictness:
  minimum: minimal
  rules: {}
```
- **Expected:** Warning: `Unknown field "description" at root level. Unknown field "team" in service "api".` Manifest still parses successfully.
- **Verification:** Exit code 0 (warning, not error). Unknown fields do not prevent validation.

**MFP-INVALID-07: Duplicate service names**

- **Input manifest:**
```yaml
manifest_version: "1.0"
name: "acme-platform"
services:
  api-service:
    repo: "github.com/acme/api"
    language: typescript
    role: producer
  api-service:
    repo: "github.com/acme/api-v2"
    language: go
    role: producer
contracts: []
strictness:
  minimum: minimal
  rules: {}
```
- **Expected:** Parse error: `Duplicate service name "api-service"`. Note: YAML itself may silently override duplicate keys; Stricture must detect this.
- **Verification:** Exit code 2.

**MFP-INVALID-08: Duplicate contract IDs**

- **Input manifest:**
```yaml
manifest_version: "1.0"
name: "acme-platform"
services:
  api:
    repo: "github.com/acme/api"
    language: typescript
    role: producer
contracts:
  - id: "user-api"
    producer: api
    consumers: []
    protocol: http
    endpoints: []
  - id: "user-api"
    producer: api
    consumers: []
    protocol: http
    endpoints: []
strictness:
  minimum: minimal
  rules: {}
```
- **Expected:** Parse error: `Duplicate contract ID "user-api"`.
- **Verification:** Exit code 2.

**MFP-INVALID-09: Contract references non-existent service**

- **Input manifest:**
```yaml
manifest_version: "1.0"
name: "acme-platform"
services:
  api:
    repo: "github.com/acme/api"
    language: typescript
    role: producer
contracts:
  - id: "user-api"
    producer: nonexistent-service
    consumers: [api]
    protocol: http
    endpoints:
      - path: "/api/users"
        method: GET
        response:
          type: User
          fields:
            id: { type: integer, required: true }
        status_codes: [200]
strictness:
  minimum: minimal
  rules: {}
```
- **Expected:** Parse error: `Contract "user-api" references producer "nonexistent-service" which is not declared in services`.
- **Verification:** Exit code 2. Error identifies the contract ID and the missing service name.

**MFP-INVALID-10: Contract references non-existent consumer**

- **Input manifest:**
```yaml
manifest_version: "1.0"
name: "acme-platform"
services:
  api:
    repo: "github.com/acme/api"
    language: typescript
    role: producer
contracts:
  - id: "user-api"
    producer: api
    consumers: [ghost-frontend]
    protocol: http
    endpoints:
      - path: "/api/users"
        method: GET
        response:
          type: User
          fields:
            id: { type: integer, required: true }
        status_codes: [200]
strictness:
  minimum: minimal
  rules: {}
```
- **Expected:** Parse error: `Contract "user-api" references consumer "ghost-frontend" which is not declared in services`.
- **Verification:** Exit code 2.

**MFP-INVALID-11: Unsupported manifest_version**

- **Input manifest:**
```yaml
manifest_version: "2.0"
name: "acme-platform"
services:
  api:
    repo: "github.com/acme/api"
    language: typescript
    role: producer
contracts: []
strictness:
  minimum: minimal
  rules: {}
```
- **Expected:** Parse error: `Unsupported manifest_version "2.0". Supported versions: ["1.0"]`.
- **Verification:** Exit code 2.

### 1.3 Field Constraint Validation

**MFP-FIELD-01: Range with min > max**

- **Input manifest** (field in contract):
```yaml
amount: { type: number, range: [999.99, 0.01], required: true }
```
- **Expected:** Parse error: `Field "amount": range minimum (999.99) is greater than maximum (0.01)`.
- **Verification:** Exit code 2.

**MFP-FIELD-02: Enum with empty values array**

- **Input manifest** (field):
```yaml
role: { type: enum, values: [], required: true }
```
- **Expected:** Parse error: `Field "role": enum type must have at least 1 value in "values" array`.
- **Verification:** Exit code 2.

**MFP-FIELD-03: Enum with duplicate values**

- **Input manifest** (field):
```yaml
role: { type: enum, values: ["admin", "user", "admin"], required: true }
```
- **Expected:** Warning: `Field "role": duplicate enum value "admin" in values array`.
- **Verification:** Warning emitted but manifest still parses. Deduplication applied.

**MFP-FIELD-04: Precision on non-number type**

- **Input manifest** (field):
```yaml
name: { type: string, precision: 2, required: true }
```
- **Expected:** Parse error: `Field "name": "precision" is only valid for type "number", found on type "string"`.
- **Verification:** Exit code 2.

**MFP-FIELD-05: Format on non-string type**

- **Input manifest** (field):
```yaml
count: { type: integer, format: email, required: true }
```
- **Expected:** Parse error: `Field "count": "format" is only valid for type "string", found on type "integer"`.
- **Verification:** Exit code 2.

**MFP-FIELD-06: Unknown format value**

- **Input manifest** (field):
```yaml
phone: { type: string, format: phone-number, required: true }
```
- **Expected:** Parse error: `Field "phone": unknown format "phone-number". Valid formats: [email, uuid, iso8601, url]`.
- **Verification:** Exit code 2. Error lists all valid format values.

**MFP-FIELD-07: Range on non-numeric type**

- **Input manifest** (field):
```yaml
name: { type: string, range: [1, 100], required: true }
```
- **Expected:** Parse error: `Field "name": "range" is only valid for types "integer" and "number", found on type "string"`.
- **Verification:** Exit code 2.

**MFP-FIELD-08: minLength on non-string type**

- **Input manifest** (field):
```yaml
count: { type: integer, minLength: 1, required: true }
```
- **Expected:** Parse error: `Field "count": "minLength" is only valid for type "string", found on type "integer"`.
- **Verification:** Exit code 2.

**MFP-FIELD-09: minLength > maxLength**

- **Input manifest** (field):
```yaml
name: { type: string, minLength: 256, maxLength: 255, required: true }
```
- **Expected:** Parse error: `Field "name": minLength (256) is greater than maxLength (255)`.
- **Verification:** Exit code 2.

**MFP-FIELD-10: Unknown field type**

- **Input manifest** (field):
```yaml
data: { type: binary, required: true }
```
- **Expected:** Parse error: `Field "data": unknown type "binary". Valid types: [string, integer, number, enum, boolean, object, array]`.
- **Verification:** Exit code 2.

**MFP-FIELD-11: Negative precision**

- **Input manifest** (field):
```yaml
amount: { type: number, precision: -1, required: true }
```
- **Expected:** Parse error: `Field "amount": precision must be a non-negative integer, got -1`.
- **Verification:** Exit code 2.

---

## 2. Service Declarations

**MFP-SVC-01: Producer role service**

- **Input manifest** (service):
```yaml
services:
  api-gateway:
    repo: "github.com/acme/api-gateway"
    language: typescript
    role: producer
    stricture_config: ".stricture.yml"
```
- **Expected:** Service parsed with role `producer`. Can only appear as `producer` in contracts.
- **Verification:** If this service appears as a consumer in a contract, that is a role conflict (tested in section 8).

**MFP-SVC-02: Consumer role service**

- **Input manifest** (service):
```yaml
services:
  web-frontend:
    repo: "github.com/acme/web-app"
    language: typescript
    role: consumer
```
- **Expected:** Service parsed with role `consumer`. Can only appear in `consumers` arrays.
- **Verification:** If this service appears as `producer` in a contract, that is a role conflict.

**MFP-SVC-03: "Both" role service**

- **Input manifest** (service):
```yaml
services:
  user-service:
    repo: "github.com/acme/user-service"
    language: go
    role: both
    stricture_config: ".stricture.yml"
```
- **Expected:** Service parsed with role `both`. Can appear as producer in some contracts and consumer in others.
- **Verification:** Service can be `producer` in contract A and listed in `consumers` of contract B without conflict.

**MFP-SVC-04: Missing repo field**

- **Input manifest** (service):
```yaml
services:
  local-api:
    language: typescript
    role: producer
    stricture_config: ".stricture.yml"
```
- **Expected:** Warning: `Service "local-api": missing "repo" field. Cross-repo validation disabled for this service (local-only mode).`
- **Verification:** Manifest still parses. The service is usable but limited to local validation only.

**MFP-SVC-05: Missing language field**

- **Input manifest** (service):
```yaml
services:
  mystery-service:
    repo: "github.com/acme/mystery"
    role: producer
```
- **Expected:** Parse error: `Service "mystery-service": "language" is required. Valid values: [typescript, go, python, java, ...]`.
- **Verification:** Exit code 2. Language is needed for adapter selection and parity analysis.

**MFP-SVC-06: Missing stricture_config field**

- **Input manifest** (service):
```yaml
services:
  api-gateway:
    repo: "github.com/acme/api-gateway"
    language: typescript
    role: producer
```
- **Expected:** Default applied: `stricture_config` defaults to `".stricture.yml"`. No error.
- **Verification:** When resolving per-service config, Stricture looks for `.stricture.yml` in the service repo root.

**MFP-SVC-07: Invalid role value**

- **Input manifest** (service):
```yaml
services:
  api:
    repo: "github.com/acme/api"
    language: typescript
    role: publisher
```
- **Expected:** Parse error: `Service "api": invalid role "publisher". Valid roles: [producer, consumer, both]`.
- **Verification:** Exit code 2.

**MFP-SVC-08: Invalid language value**

- **Input manifest** (service):
```yaml
services:
  api:
    repo: "github.com/acme/api"
    language: cobol
    role: producer
```
- **Expected:** Parse error: `Service "api": unsupported language "cobol". Supported: [typescript, go]`. (v0.1 supports TS + Go only.)
- **Verification:** Exit code 2.

**MFP-SVC-09: Empty services map**

- **Input manifest:**
```yaml
manifest_version: "1.0"
name: "empty"
services: {}
contracts: []
strictness:
  minimum: minimal
  rules: {}
```
- **Expected:** Parse error: `services must contain at least 1 service`.
- **Verification:** Exit code 2.

---

## 3. Contract Definitions -- HTTP

**MFP-HTTP-01: Single endpoint contract**

- **Input manifest** (contract):
```yaml
contracts:
  - id: "status-api"
    producer: api-service
    consumers: [monitoring]
    protocol: http
    endpoints:
      - path: "/status"
        method: GET
        response:
          type: StatusResponse
          fields:
            healthy: { type: boolean, required: true }
        status_codes: [200, 503]
```
- **Expected:** Contract parsed with 1 endpoint, 1 response field, 2 status codes.
- **Verification:** `stricture audit` shows the contract with the single endpoint.

**MFP-HTTP-02: Multi-endpoint contract**

- **Input manifest** (contract):
```yaml
contracts:
  - id: "user-crud"
    producer: user-service
    consumers: [frontend]
    protocol: http
    endpoints:
      - path: "/api/users"
        method: GET
        response:
          type: UserList
          fields:
            users: { type: array, required: true }
            total: { type: integer, range: [0, 2147483647], required: true }
        status_codes: [200, 500]

      - path: "/api/users/:id"
        method: GET
        response:
          type: User
          fields:
            id:   { type: integer, range: [1, 2147483647], required: true }
            name: { type: string, maxLength: 255, required: true }
        status_codes: [200, 404, 500]

      - path: "/api/users"
        method: POST
        request:
          type: CreateUserRequest
          fields:
            name:  { type: string, minLength: 1, maxLength: 255, required: true }
            email: { type: string, format: email, required: true }
        response:
          type: User
        status_codes: [201, 400, 409, 500]

      - path: "/api/users/:id"
        method: PUT
        request:
          type: UpdateUserRequest
          fields:
            name:  { type: string, minLength: 1, maxLength: 255, required: false }
            email: { type: string, format: email, required: false }
        response:
          type: User
        status_codes: [200, 400, 404, 500]

      - path: "/api/users/:id"
        method: DELETE
        status_codes: [204, 404, 500]
```
- **Expected:** 5 endpoints parsed across 3 paths. GET/PUT/DELETE on `:id`, GET/POST on collection.
- **Verification:** All HTTP methods represented. DELETE endpoint has no request/response body.

**MFP-HTTP-03: All HTTP methods**

- **Input manifest** (endpoints only -- each within a valid contract):
```yaml
endpoints:
  - path: "/api/resource"
    method: GET
    response:
      type: Resource
      fields:
        id: { type: integer, required: true }
    status_codes: [200]
  - path: "/api/resource"
    method: POST
    request:
      type: CreateResource
      fields:
        name: { type: string, required: true }
    status_codes: [201]
  - path: "/api/resource/:id"
    method: PUT
    request:
      type: UpdateResource
      fields:
        name: { type: string, required: true }
    status_codes: [200]
  - path: "/api/resource/:id"
    method: DELETE
    status_codes: [204]
  - path: "/api/resource/:id"
    method: PATCH
    request:
      type: PatchResource
      fields:
        name: { type: string, required: false }
    status_codes: [200]
```
- **Expected:** All 5 methods (GET, POST, PUT, DELETE, PATCH) parsed without error.
- **Verification:** Each endpoint's method field matches the declared value.

**MFP-HTTP-04: Path parameters**

- **Input manifest** (endpoint):
```yaml
endpoints:
  - path: "/api/orgs/:orgId/users/:userId/posts/:postId"
    method: GET
    response:
      type: Post
      fields:
        id: { type: integer, required: true }
    status_codes: [200, 404]
```
- **Expected:** Path with 3 parameters parsed. Parameters `:orgId`, `:userId`, `:postId` are extracted.
- **Verification:** Parity checker matches this path against code patterns like `/api/orgs/${orgId}/users/${userId}/posts/${postId}` (TS) or `/api/orgs/%s/users/%s/posts/%s` (Go).

**MFP-HTTP-05: Request body with field constraints**

- **Input manifest** (endpoint):
```yaml
endpoints:
  - path: "/api/orders"
    method: POST
    request:
      type: CreateOrderRequest
      fields:
        product_id: { type: string, format: uuid, required: true }
        quantity:    { type: integer, range: [1, 9999], required: true }
        notes:       { type: string, maxLength: 2000, required: false }
    status_codes: [201, 400]
```
- **Expected:** 3 request fields parsed: `product_id` (uuid format), `quantity` (range constraint), `notes` (optional with maxLength).
- **Verification:** Parity checks apply to request fields. Producer must validate incoming request bodies against these constraints. Consumer must send valid data matching these constraints.

**MFP-HTTP-06: Response body with field constraints**

- **Input manifest** (endpoint):
```yaml
endpoints:
  - path: "/api/invoices/:id"
    method: GET
    response:
      type: Invoice
      fields:
        id:         { type: string, format: uuid, required: true }
        amount:     { type: number, range: [0.01, 999999.99], precision: 2, required: true }
        currency:   { type: enum, values: ["USD", "EUR", "GBP"], required: true }
        created_at: { type: string, format: iso8601, required: true }
        paid:       { type: boolean, required: true }
    status_codes: [200, 404, 500]
```
- **Expected:** 5 response fields parsed with full constraints.
- **Verification:** Each field's constraint is available for parity comparison.

**MFP-HTTP-07: Error shape definition**

- **Input manifest** (endpoint):
```yaml
endpoints:
  - path: "/api/users/:id"
    method: GET
    response:
      type: User
      fields:
        id:   { type: integer, required: true }
        name: { type: string, required: true }
    status_codes: [200, 400, 404, 500]
    error_shape:
      code:    { type: string, required: true }
      message: { type: string, required: true }
      details: { type: string, required: false }
```
- **Expected:** `error_shape` parsed as a separate type with 3 fields. Applies to all non-2xx status codes.
- **Verification:** Consumer must handle error responses matching this shape. `stricture audit` reports whether consumer code destructures the error according to this shape.

**MFP-HTTP-08: Endpoint with no request body (GET)**

- **Input manifest** (endpoint):
```yaml
endpoints:
  - path: "/api/users"
    method: GET
    response:
      type: UserList
      fields:
        users: { type: array, required: true }
    status_codes: [200]
```
- **Expected:** No `request` field is present. This is valid -- GET endpoints typically have no request body.
- **Verification:** Parity checks apply only to response fields. No request validation expected.

**MFP-HTTP-09: Endpoint with no response body (DELETE returning 204)**

- **Input manifest** (endpoint):
```yaml
endpoints:
  - path: "/api/users/:id"
    method: DELETE
    status_codes: [204, 404]
```
- **Expected:** No `response` field is present. This is valid -- 204 means no content.
- **Verification:** Parity checks: consumer must handle 204 (no body parsing) and 404. No response field validation.

**MFP-HTTP-10: Invalid HTTP method**

- **Input manifest** (endpoint):
```yaml
endpoints:
  - path: "/api/users"
    method: YEET
    status_codes: [200]
```
- **Expected:** Parse error: `Endpoint "/api/users": invalid HTTP method "YEET". Valid methods: [GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS]`.
- **Verification:** Exit code 2.

**MFP-HTTP-11: Missing path**

- **Input manifest** (endpoint):
```yaml
endpoints:
  - method: GET
    response:
      type: Foo
      fields:
        id: { type: integer, required: true }
    status_codes: [200]
```
- **Expected:** Parse error: `Endpoint at index 0: "path" is required`.
- **Verification:** Exit code 2.

**MFP-HTTP-12: Missing method**

- **Input manifest** (endpoint):
```yaml
endpoints:
  - path: "/api/users"
    response:
      type: Foo
      fields:
        id: { type: integer, required: true }
    status_codes: [200]
```
- **Expected:** Parse error: `Endpoint "/api/users": "method" is required`.
- **Verification:** Exit code 2.

**MFP-HTTP-13: Duplicate endpoint (same path + method)**

- **Input manifest** (endpoints):
```yaml
endpoints:
  - path: "/api/users"
    method: GET
    response:
      type: UserList
      fields:
        users: { type: array, required: true }
    status_codes: [200]
  - path: "/api/users"
    method: GET
    response:
      type: UserListV2
      fields:
        data: { type: array, required: true }
    status_codes: [200]
```
- **Expected:** Parse error: `Contract "xxx": duplicate endpoint "GET /api/users"`.
- **Verification:** Exit code 2.

---

## 4. Contract Definitions -- Message Queue

**MFP-MQ-01: Single message event**

- **Input manifest** (contract):
```yaml
contracts:
  - id: "payment-events"
    producer: billing-service
    consumers: [notification-service]
    protocol: message_queue
    queue: "payments"
    messages:
      - event: "payment.completed"
        fields:
          payment_id: { type: string, format: uuid, required: true }
          amount:     { type: number, range: [0.01, 999999.99], precision: 2, required: true }
          user_id:    { type: integer, range: [1, 2147483647], required: true }
```
- **Expected:** 1 message event with 3 fields parsed.
- **Verification:** `stricture audit` reports the message queue contract with the event.

**MFP-MQ-02: Multiple events per queue**

- **Input manifest** (contract):
```yaml
contracts:
  - id: "order-events"
    producer: order-service
    consumers: [inventory-service, analytics-service]
    protocol: message_queue
    queue: "orders"
    messages:
      - event: "order.created"
        fields:
          order_id: { type: string, format: uuid, required: true }
          items:    { type: array, required: true }
      - event: "order.shipped"
        fields:
          order_id:    { type: string, format: uuid, required: true }
          tracking_id: { type: string, required: true }
      - event: "order.cancelled"
        fields:
          order_id: { type: string, format: uuid, required: true }
          reason:   { type: string, maxLength: 500, required: false }
```
- **Expected:** 3 events parsed under 1 queue. Each event has its own field set.
- **Verification:** Parity checks apply per-event. Consumer must handle all 3 events if subscribed to this queue.

**MFP-MQ-03: Event field constraints**

- **Input manifest** (message):
```yaml
messages:
  - event: "invoice.created"
    fields:
      invoice_id: { type: string, format: uuid, required: true }
      user_id:    { type: integer, range: [1, 2147483647], required: true }
      amount:     { type: number, range: [0.01, 999999.99], precision: 2, required: true }
      currency:   { type: enum, values: ["USD", "EUR", "GBP"], required: true }
      issued_at:  { type: string, format: iso8601, required: true }
```
- **Expected:** All constraint types work identically in message queue fields as in HTTP fields.
- **Verification:** `stricture audit` displays constraints per field. Parity analysis compares producer's publish logic against consumer's handling logic.

**MFP-MQ-04: Queue name validation**

- **Input manifest** (contract):
```yaml
contracts:
  - id: "events"
    producer: service-a
    consumers: [service-b]
    protocol: message_queue
    queue: "my.events.queue"
    messages:
      - event: "test"
        fields:
          id: { type: string, required: true }
```
- **Expected:** Queue name `"my.events.queue"` is valid. Stored as metadata.
- **Verification:** Parity analysis can match this queue name against code references.

**MFP-MQ-05: Missing queue field**

- **Input manifest** (contract):
```yaml
contracts:
  - id: "events"
    producer: service-a
    consumers: [service-b]
    protocol: message_queue
    messages:
      - event: "test"
        fields:
          id: { type: string, required: true }
```
- **Expected:** Parse error: `Contract "events": protocol "message_queue" requires "queue" field`.
- **Verification:** Exit code 2.

**MFP-MQ-06: Missing messages field**

- **Input manifest** (contract):
```yaml
contracts:
  - id: "events"
    producer: service-a
    consumers: [service-b]
    protocol: message_queue
    queue: "events"
```
- **Expected:** Parse error: `Contract "events": protocol "message_queue" requires "messages" with at least 1 event`.
- **Verification:** Exit code 2.

**MFP-MQ-07: Duplicate event names within same queue**

- **Input manifest** (messages):
```yaml
messages:
  - event: "order.created"
    fields:
      id: { type: string, required: true }
  - event: "order.created"
    fields:
      order_id: { type: string, required: true }
```
- **Expected:** Parse error: `Contract "xxx": duplicate event name "order.created"`.
- **Verification:** Exit code 2.

---

## 5. Contract Definitions -- Multi-Consumer

**MFP-MC-01: One producer, multiple consumers**

- **Input manifest:**
```yaml
manifest_version: "1.0"
name: "multi-consumer"

services:
  user-service:
    repo: "github.com/acme/user-service"
    language: go
    role: producer
  api-gateway:
    repo: "github.com/acme/api-gateway"
    language: typescript
    role: consumer
  web-frontend:
    repo: "github.com/acme/web-app"
    language: typescript
    role: consumer
  mobile-bff:
    repo: "github.com/acme/mobile-bff"
    language: go
    role: consumer

contracts:
  - id: "user-api"
    producer: user-service
    consumers: [api-gateway, web-frontend, mobile-bff]
    protocol: http
    endpoints:
      - path: "/api/users/:id"
        method: GET
        response:
          type: User
          fields:
            id:    { type: integer, range: [1, 2147483647], required: true }
            name:  { type: string, maxLength: 255, required: true }
            email: { type: string, format: email, required: true }
        status_codes: [200, 404, 500]

strictness:
  minimum: strict
  rules:
    numeric-range-required: true
    string-length-required: true
```
- **Expected:** Contract has 1 producer and 3 consumers. Parity checks run independently for each consumer.
- **Verification:** `stricture audit` run from each consumer's repo reports its own parity status. Consumer A can be "strict" while consumer B is "basic" -- they are evaluated independently.

**MFP-MC-02: Each consumer must satisfy all constraints independently**

- **Setup:** Manifest from MFP-MC-01. Three consumers:
  - `api-gateway` (TypeScript): validates `id` range, validates `email` format, handles all status codes.
  - `web-frontend` (TypeScript): validates `email` format, does NOT validate `id` range, does NOT handle 500.
  - `mobile-bff` (Go): validates nothing -- just `json.Decode` into a struct with no validation.
- **Expected violations per consumer:**
  - `api-gateway`: 0 parity violations.
  - `web-frontend`: 2 violations -- `id` missing range check, missing 500 status code handling.
  - `mobile-bff`: 3 violations -- `id` missing range check, `email` missing format validation, missing 404 and 500 handling.
- **Verification:** Each consumer's `stricture audit` output shows only its own violations. No cross-contamination between consumers.

**MFP-MC-03: Adding a new consumer to existing contract**

- **Setup:** Contract "user-api" initially has consumers `[api-gateway, web-frontend]`. A new service `analytics-service` is added to `consumers`.
- **Input:** `analytics-service` has zero validation code for the user-api contract.
- **Expected:** After adding `analytics-service` to the manifest consumers list, running `stricture audit` in the analytics-service repo flags all fields as missing validation.
- **Verification:** Existing consumers are unaffected. New consumer gets a full parity report.

---

## 6. Strictness Levels

### 6.1 Level Calculation

**MFP-SL-01: Minimal -- type only**

- **Input manifest** (field):
```yaml
name: { type: string }
```
- **Expected:** Strictness level: `minimal`.
- **Verification:** `stricture audit` reports this field as "minimal" strictness.

**MFP-SL-02: Basic -- type + required**

- **Input manifest** (field):
```yaml
name: { type: string, required: true }
```
- **Expected:** Strictness level: `basic`.
- **Verification:** `stricture audit` reports "basic".

**MFP-SL-03: Standard -- type + required + format**

- **Input manifest** (field):
```yaml
email: { type: string, format: email, required: true }
```
- **Expected:** Strictness level: `standard`.
- **Verification:** `stricture audit` reports "standard".

**MFP-SL-04: Strict -- type + required + format + constraint**

- **Input manifest** (field):
```yaml
email: { type: string, format: email, maxLength: 255, required: true }
```
- **Expected:** Strictness level: `strict`.
- **Verification:** `stricture audit` reports "strict".

**MFP-SL-05: Exhaustive -- all constraints + error shape**

- **Input manifest** (full endpoint):
```yaml
endpoints:
  - path: "/api/users/:id"
    method: GET
    response:
      type: User
      fields:
        id:         { type: integer, range: [1, 2147483647], required: true }
        name:       { type: string, minLength: 1, maxLength: 255, required: true }
        email:      { type: string, format: email, maxLength: 320, required: true }
        role:       { type: enum, values: ["admin", "user", "viewer"], required: true }
        created_at: { type: string, format: iso8601, required: true }
    status_codes: [200, 400, 404, 500]
    error_shape:
      code:    { type: string, required: true }
      message: { type: string, required: true }
```
- **Expected:** All fields are at "strict" or "exhaustive" level. The endpoint overall is "exhaustive" because it has full constraints plus error shape plus status codes.
- **Verification:** `stricture audit` reports "exhaustive" for the endpoint.

**MFP-SL-06: Mixed strictness within one endpoint**

- **Input manifest** (fields):
```yaml
fields:
  id:    { type: integer, range: [1, 2147483647], required: true }  # strict
  name:  { type: string, required: true }                            # basic
  email: { type: string }                                            # minimal
```
- **Expected:** Each field has its own level: `id`=strict, `name`=basic, `email`=minimal. Endpoint overall is "basic" (limited by the weakest field, or averaged -- spec should clarify).
- **Verification:** `stricture audit` shows per-field strictness. Endpoint percentage reflects the mix.

### 6.2 Minimum Enforcement

**MFP-MIN-01: minimum: minimal -- no floor violations**

- **Input manifest:**
```yaml
strictness:
  minimum: minimal
  rules: {}
```
- **Input field:** `name: { type: string }` (level: minimal)
- **Expected:** No minimum-level violation. `minimal` satisfies the `minimal` floor.
- **Verification:** `stricture audit` reports no minimum-level warnings.

**MFP-MIN-02: minimum: basic -- flag type-only fields**

- **Input manifest:**
```yaml
strictness:
  minimum: basic
  rules: {}
```
- **Input field:** `name: { type: string }` (level: minimal, below floor)
- **Expected violation:** `Manifest minimum strictness violation: field "name" is "minimal" (type only) but minimum is "basic". Add "required" constraint.`
- **Verification:** This is a manifest-level violation, not a code-level violation. The manifest author must be more precise.

**MFP-MIN-03: minimum: standard -- flag fields without format**

- **Input manifest:**
```yaml
strictness:
  minimum: standard
  rules: {}
```
- **Input field:** `email: { type: string, required: true }` (level: basic, below standard)
- **Expected violation:** `Field "email" is "basic" but minimum is "standard". Consider adding "format" constraint.`
- **Verification:** Manifest-level violation.

**MFP-MIN-04: minimum: strict -- flag fields without constraints**

- **Input manifest:**
```yaml
strictness:
  minimum: strict
  rules: {}
```
- **Input field:** `email: { type: string, format: email, required: true }` (level: standard, below strict)
- **Expected violation:** `Field "email" is "standard" but minimum is "strict". Add constraint (e.g., maxLength).`
- **Verification:** Manifest-level violation.

**MFP-MIN-05: minimum: exhaustive -- flag anything not fully specified**

- **Input manifest:**
```yaml
strictness:
  minimum: exhaustive
  rules: {}
```
- **Input endpoint:** An endpoint with all field constraints but missing `error_shape`.
- **Expected violation:** `Endpoint "GET /api/users/:id" is "strict" but minimum is "exhaustive". Add error_shape.`
- **Verification:** Manifest-level violation.

**MFP-MIN-06: Field exactly at minimum level -- no violation**

- **Input manifest:**
```yaml
strictness:
  minimum: strict
  rules: {}
```
- **Input field:** `id: { type: integer, range: [1, 2147483647], required: true }` (level: strict)
- **Expected:** No violation. The field meets the minimum.
- **Verification:** `stricture audit` reports no minimum-level warning for this field.

---

## 7. Strictness Rules Enforcement

**MFP-RULE-01: numeric-range-required -- flag numbers without range**

- **Input manifest:**
```yaml
strictness:
  minimum: minimal
  rules:
    numeric-range-required: true
```
- **Input field:** `quantity: { type: integer, required: true }` (no range)
- **Expected violation:** `Rule "numeric-range-required": field "quantity" (type integer) must declare a range.`
- **Verification:** Manifest-level violation.

**MFP-RULE-02: numeric-range-required -- number with range passes**

- **Input manifest:** Same as MFP-RULE-01.
- **Input field:** `quantity: { type: integer, range: [1, 9999], required: true }`
- **Expected:** No violation.
- **Verification:** Rule satisfied.

**MFP-RULE-03: string-length-required -- flag strings without maxLength**

- **Input manifest:**
```yaml
strictness:
  minimum: minimal
  rules:
    string-length-required: true
```
- **Input field:** `name: { type: string, required: true }` (no maxLength)
- **Expected violation:** `Rule "string-length-required": field "name" (type string) must declare maxLength.`
- **Verification:** Manifest-level violation.

**MFP-RULE-04: string-length-required -- string with maxLength passes**

- **Input field:** `name: { type: string, maxLength: 255, required: true }`
- **Expected:** No violation.

**MFP-RULE-05: enum-exhaustive -- flag enum without values list**

- **Input manifest:**
```yaml
strictness:
  minimum: minimal
  rules:
    enum-exhaustive: true
```
- **Input field:** `role: { type: enum, required: true }` (no values)
- **Expected violation:** `Rule "enum-exhaustive": field "role" (type enum) must exhaustively list values.`
- **Verification:** Manifest-level violation.

**MFP-RULE-06: enum-exhaustive -- enum with values passes**

- **Input field:** `role: { type: enum, values: ["admin", "user", "viewer"], required: true }`
- **Expected:** No violation.

**MFP-RULE-07: error-shape-required -- flag endpoints without error_shape**

- **Input manifest:**
```yaml
strictness:
  minimum: minimal
  rules:
    error-shape-required: true
```
- **Input endpoint:** Endpoint with `status_codes: [200, 400, 500]` but no `error_shape`.
- **Expected violation:** `Rule "error-shape-required": endpoint "GET /api/users" has error status codes [400, 500] but no error_shape defined.`
- **Verification:** Manifest-level violation.

**MFP-RULE-08: error-shape-required -- endpoint with only 2xx codes -- no violation**

- **Input endpoint:** Endpoint with `status_codes: [200, 201]` and no `error_shape`.
- **Expected:** No violation. All status codes are success codes, so no error shape is needed.

**MFP-RULE-09: status-codes-exhaustive -- flag endpoints without status_codes**

- **Input manifest:**
```yaml
strictness:
  minimum: minimal
  rules:
    status-codes-exhaustive: true
```
- **Input endpoint:** Endpoint with no `status_codes` field.
- **Expected violation:** `Rule "status-codes-exhaustive": endpoint "GET /api/users" must declare status_codes.`
- **Verification:** Manifest-level violation.

**MFP-RULE-10: status-codes-exhaustive -- endpoint with status_codes passes**

- **Input endpoint:** Endpoint with `status_codes: [200, 400, 404, 500]`.
- **Expected:** No violation.

**MFP-RULE-11: All rules disabled -- no enforcement**

- **Input manifest:**
```yaml
strictness:
  minimum: minimal
  rules:
    numeric-range-required: false
    string-length-required: false
    enum-exhaustive: false
    error-shape-required: false
    status-codes-exhaustive: false
```
- **Input:** Fields with no constraints, endpoints with no error shapes.
- **Expected:** No rule violations. Minimum is `minimal` so everything passes.
- **Verification:** Only code-level parity checks apply, not manifest-level rules.

---

## 8. Per-Service Configuration

### 8.1 Manifest Reference

**MFP-CFG-01: manifest.url (git URL)**

- **Input** (`.stricture.yml` in user-service repo):
```yaml
manifest:
  url: "github.com/acme/stricture-manifest"
  service: user-service
  contracts:
    - id: "user-api"
      role: producer
      handler_paths: ["cmd/server/handlers/**"]
      type_paths: ["pkg/types/**"]
```
- **Expected:** Stricture clones (or fetches) the manifest from the git URL. Parses it. Identifies this service as `user-service` within the manifest.
- **Verification:** `stricture audit` displays the manifest name and the contracts this service participates in.

**MFP-CFG-02: manifest.path (local path)**

- **Input** (`.stricture.yml`):
```yaml
manifest:
  path: "../stricture-manifest/stricture-manifest.yml"
  service: user-service
  contracts:
    - id: "user-api"
      role: producer
      handler_paths: ["cmd/server/handlers/**"]
      type_paths: ["pkg/types/**"]
```
- **Expected:** Stricture reads the manifest from the local filesystem path. Works for monorepo setups.
- **Verification:** `stricture audit` succeeds using the local manifest.

**MFP-CFG-03: Both url and path -- path takes precedence**

- **Input** (`.stricture.yml`):
```yaml
manifest:
  url: "github.com/acme/stricture-manifest"
  path: "../local-manifest/stricture-manifest.yml"
  service: user-service
  contracts:
    - id: "user-api"
      role: producer
      handler_paths: ["cmd/server/handlers/**"]
      type_paths: ["pkg/types/**"]
```
- **Expected:** `path` takes precedence over `url`. Warning emitted: `Both "url" and "path" specified in manifest config; using "path".`
- **Verification:** Manifest is loaded from the local path.

**MFP-CFG-04: Neither url nor path**

- **Input** (`.stricture.yml`):
```yaml
manifest:
  service: user-service
  contracts:
    - id: "user-api"
      role: producer
      handler_paths: ["cmd/server/handlers/**"]
      type_paths: ["pkg/types/**"]
```
- **Expected:** Parse error: `Manifest config requires either "url" or "path" to locate the manifest file.`
- **Verification:** Exit code 2.

**MFP-CFG-05: manifest.service matches service in manifest**

- **Setup:** Manifest declares `services: { user-service: ... }`. Per-service config has `service: user-service`.
- **Expected:** Match found. Stricture proceeds with validation.
- **Verification:** `stricture audit` header shows `Stricture Audit -- user-service`.

**MFP-CFG-06: manifest.service does not match any service**

- **Setup:** Manifest declares `services: { api-service: ..., billing-service: ... }`. Per-service config has `service: user-service`.
- **Expected:** Parse error: `Service "user-service" not found in manifest. Available services: [api-service, billing-service].`
- **Verification:** Exit code 2. Error lists all available service names.

### 8.2 Contract Role Mapping

**MFP-CFG-07: contracts[].id matches manifest contract**

- **Setup:** Manifest has contract `id: "user-api"`. Per-service config declares `contracts: [{ id: "user-api", role: producer, ... }]`.
- **Expected:** Match found. Stricture runs parity checks for this contract.
- **Verification:** `stricture audit` shows the contract with its endpoints.

**MFP-CFG-08: contracts[].id does not match any manifest contract**

- **Setup:** Manifest has contract `id: "user-api"`. Per-service config declares `contracts: [{ id: "billing-api", role: producer, ... }]`.
- **Expected:** Parse error: `Contract "billing-api" referenced in .stricture.yml not found in manifest. Available contracts: [user-api].`
- **Verification:** Exit code 2.

**MFP-CFG-09: contracts[].role matches service's role in manifest**

- **Setup:** Manifest declares `user-service` with `role: producer`. Contract `user-api` has `producer: user-service`. Per-service config says `role: producer`.
- **Expected:** Roles match. No conflict.
- **Verification:** `stricture audit` proceeds normally.

**MFP-CFG-10: contracts[].role conflicts with manifest declaration**

- **Setup:** Manifest declares `user-service` with `role: producer`. Per-service config says `contracts: [{ id: "user-api", role: consumer, ... }]`. But in the manifest, `user-service` is the producer of `user-api`, not a consumer.
- **Expected:** Parse error: `Role conflict for service "user-service" in contract "user-api": manifest declares this service as "producer" but .stricture.yml declares "consumer".`
- **Verification:** Exit code 2.

**MFP-CFG-11: handler_paths glob resolution**

- **Setup:** Per-service config:
```yaml
contracts:
  - id: "user-api"
    role: producer
    handler_paths: ["cmd/server/handlers/**"]
    type_paths: ["pkg/types/**"]
```
- **Filesystem:**
```
cmd/server/handlers/
  users.go
  orders.go
pkg/types/
  user.go
  order.go
```
- **Expected:** `handler_paths` resolves to `[cmd/server/handlers/users.go, cmd/server/handlers/orders.go]`. `type_paths` resolves to `[pkg/types/user.go, pkg/types/order.go]`.
- **Verification:** `stricture audit` scans exactly these files for handler and type definitions.

**MFP-CFG-12: handler_paths with no matching files**

- **Setup:** `handler_paths: ["nonexistent/dir/**"]`
- **Expected:** Warning: `handler_paths glob "nonexistent/dir/**" matched 0 files for contract "user-api".`
- **Verification:** Warning emitted. Audit proceeds but reports that no handler code was found for the contract.

**MFP-CFG-13: Empty handler_paths**

- **Setup:** `handler_paths: []`
- **Expected:** Parse error: `Contract "user-api": handler_paths must contain at least 1 glob pattern.`
- **Verification:** Exit code 2.

**MFP-CFG-14: type_paths glob resolution with nested directories**

- **Setup:** `type_paths: ["internal/**/models/**"]`
- **Filesystem:**
```
internal/
  user/models/
    user.go
  billing/models/
    invoice.go
```
- **Expected:** Resolves to `[internal/user/models/user.go, internal/billing/models/invoice.go]`.
- **Verification:** Both files are scanned for type definitions.

---

## 9. CTR-manifest-conformance

This rule verifies that actual code types and handlers match manifest declarations. No extra fields, no missing fields, no type mismatches.

### 9.1 True Positive Cases

**MFP-CONF-TP-01: Missing field in code type (Go)**

- **Manifest declares:**
```yaml
fields:
  id:    { type: integer, range: [1, 2147483647], required: true }
  name:  { type: string, maxLength: 255, required: true }
  email: { type: string, format: email, required: true }
```
- **Code** (`pkg/types/user.go`):
```go
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    // email field is MISSING
}
```
- **Expected violation:** `CTR-manifest-conformance: type "User" is missing field "email" (declared in manifest contract "user-api").`
- **Verification:** Exit code 1. Violation identifies the missing field and the manifest source.

**MFP-CONF-TP-02: Missing field in code type (TypeScript)**

- **Manifest declares:** Same as MFP-CONF-TP-01.
- **Code** (`src/types/user.ts`):
```typescript
interface User {
  id: number;
  name: string;
  // email is MISSING
}
```
- **Expected violation:** `CTR-manifest-conformance: type "User" is missing field "email".`

**MFP-CONF-TP-03: Extra field in code type not in manifest**

- **Manifest declares:** `fields: { id: ..., name: ..., email: ... }`
- **Code** (`pkg/types/user.go`):
```go
type User struct {
    ID      int    `json:"id"`
    Name    string `json:"name"`
    Email   string `json:"email"`
    Avatar  string `json:"avatar"` // NOT in manifest
}
```
- **Expected violation:** `CTR-manifest-conformance: type "User" has field "avatar" not declared in manifest contract "user-api". Either add it to the manifest or remove it from the code.`
- **Verification:** This catches drift between manifest and implementation.

**MFP-CONF-TP-04: Type mismatch between manifest and code**

- **Manifest declares:** `id: { type: integer, required: true }`
- **Code** (`src/types/user.ts`):
```typescript
interface User {
  id: string; // Manifest says integer, code says string
  name: string;
  email: string;
}
```
- **Expected violation:** `CTR-manifest-conformance: field "id" in type "User" has type "string" but manifest declares "integer".`

**MFP-CONF-TP-05: JSON tag mismatch (Go)**

- **Manifest declares:** `created_at: { type: string, format: iso8601, required: true }`
- **Code:**
```go
type User struct {
    CreatedAt time.Time `json:"createdAt"` // Manifest says "created_at"
}
```
- **Expected violation:** `CTR-manifest-conformance: field "CreatedAt" has json tag "createdAt" but manifest declares field name "created_at".`

**MFP-CONF-TP-06: Handler does not implement manifest endpoint**

- **Manifest declares:** `endpoints: [{ path: "/api/users/:id", method: GET }, { path: "/api/users", method: POST }]`
- **Code** (`cmd/server/handlers/users.go`): Only implements `GET /api/users/:id`. No handler for `POST /api/users`.
- **Expected violation:** `CTR-manifest-conformance: contract "user-api" declares endpoint "POST /api/users" but no handler found in handler_paths.`

**MFP-CONF-TP-07: Handler serves extra endpoint not in manifest**

- **Manifest declares:** Only `GET /api/users/:id`.
- **Code:** Implements both `GET /api/users/:id` and `DELETE /api/users/:id`.
- **Expected violation:** `CTR-manifest-conformance: handler "DELETE /api/users/:id" exists in code but is not declared in manifest contract "user-api".`

**MFP-CONF-TP-08: Enum values mismatch**

- **Manifest declares:** `role: { type: enum, values: ["admin", "user", "viewer"] }`
- **Code** (`src/types/user.ts`):
```typescript
type Role = "admin" | "user" | "editor"; // "editor" is not in manifest, "viewer" is missing
```
- **Expected violation:** `CTR-manifest-conformance: enum "Role" has value "editor" not in manifest. Missing manifest value "viewer".`

**MFP-CONF-TP-09: Status code mismatch**

- **Manifest declares:** `status_codes: [200, 400, 404, 500]`
- **Code** (Express handler):
```typescript
app.get("/api/users/:id", (req, res) => {
  // Returns 200, 400, 404, 500, AND 503 (not in manifest)
  if (dbDown) return res.status(503).json({ error: "service unavailable" });
});
```
- **Expected violation:** `CTR-manifest-conformance: handler "GET /api/users/:id" returns status 503 which is not declared in manifest status_codes [200, 400, 404, 500].`

**MFP-CONF-TP-10: Response shape has optional field, manifest says required**

- **Manifest declares:** `name: { type: string, required: true }`
- **Code** (`src/types/user.ts`):
```typescript
interface User {
  id: number;
  name?: string; // Optional in code, required in manifest
  email: string;
}
```
- **Expected violation:** `CTR-manifest-conformance: field "name" is optional in code but required in manifest.`

### 9.2 True Negative Cases

**MFP-CONF-TN-01:** Code type exactly matches manifest declarations -- all fields present, correct types, correct json tags. No violation.

**MFP-CONF-TN-02:** Code type has a superset of fields but all manifest-required fields are present, and the config `strictExtraFields: false`. Extra fields are allowed.

**MFP-CONF-TN-03:** Handler implements all declared endpoints with correct paths and methods.

**MFP-CONF-TN-04:** Enum in code is a superset of manifest values (code handles more than manifest requires). If `enumStrictSubset: false`, this is not a violation -- code is more robust than manifest requires.

**MFP-CONF-TN-05:** Code uses a shared type package that matches manifest exactly.

### 9.3 False Positive Risks

**MFP-CONF-FP-01:** Middleware adds fields to the response (e.g., `X-Request-ID` header or `request_id` field). These fields are not in the manifest but are added after the handler.

**MFP-CONF-FP-02:** Generated code (protobuf, gRPC) has extra serialization fields (`XXX_unrecognized`, `XXX_sizecache`). These should be excluded.

**MFP-CONF-FP-03:** Type embedding in Go adds promoted fields that do not appear in the manifest.

### 9.4 False Negative Risks

**MFP-CONF-FN-01:** Handler constructs response as a raw map (`map[string]interface{}`). No type to compare against manifest.

**MFP-CONF-FN-02:** Response built dynamically from database results. Fields depend on query results.

**MFP-CONF-FN-03:** Handler delegates to another function that builds the response. Stricture must follow the call chain to find the actual response shape.

---

## 10. CTR-strictness-parity -- Range Parity

Strictness parity means: if the manifest declares a constraint, both producer and consumer must enforce it with equal precision.

**MFP-RANGE-01: Both validate same range -- pass**

- **Manifest:** `amount: { type: number, range: [0.01, 999999.99], precision: 2 }`
- **Producer** (`billing-service`, Go):
```go
func validateAmount(amount float64) error {
    if amount < 0.01 || amount > 999999.99 {
        return fmt.Errorf("amount out of range: %f", amount)
    }
    return nil
}
```
- **Consumer** (`user-service`, Go):
```go
func processInvoice(invoice Invoice) error {
    if invoice.Amount < 0.01 || invoice.Amount > 999999.99 {
        return fmt.Errorf("invalid amount: %f", invoice.Amount)
    }
    // process...
}
```
- **Expected:** No parity violation. Both sides validate the same range.
- **Verification:** `stricture audit` shows "strict" for `amount` on both sides.

**MFP-RANGE-02: Producer validates, consumer does not -- violation**

- **Manifest:** `amount: { type: number, range: [0.01, 999999.99], precision: 2 }`
- **Producer** (Go): Same validation as MFP-RANGE-01.
- **Consumer** (Go):
```go
func processInvoice(invoice Invoice) error {
    total := invoice.Amount * 1.1 // No range check!
    return nil
}
```
- **Expected violation:** `CTR-strictness-parity: field "amount" in contract "billing-events" -- producer validates range [0.01, 999999.99] but consumer (user-service) has no range validation.`
- **Verification:** Parity gap identified. Consumer is less strict.

**MFP-RANGE-03: Consumer validates, producer does not -- violation**

- **Manifest:** `user_id: { type: integer, range: [1, 2147483647] }`
- **Producer** (Go):
```go
// Just sets the value, no validation
event := Event{UserID: getUserID()}
publishMessage(event)
```
- **Consumer** (TypeScript):
```typescript
if (event.user_id < 1 || event.user_id > 2147483647) {
  throw new Error("invalid user_id");
}
```
- **Expected violation:** `CTR-strictness-parity: field "user_id" in contract "billing-events" -- consumer validates range [1, 2147483647] but producer (billing-service) has no range validation before publishing.`

**MFP-RANGE-04: Both validate but different ranges -- violation with diff**

- **Manifest:** `amount: { type: number, range: [0.01, 999999.99] }`
- **Producer** (Go):
```go
if amount < 0.01 || amount > 999999.99 { return err }
```
- **Consumer** (TypeScript):
```typescript
if (amount < 0.00 || amount > 1000000.00) { throw new Error("invalid"); }
```
- **Expected violation:** `CTR-strictness-parity: field "amount" -- range mismatch. Manifest: [0.01, 999999.99]. Producer: [0.01, 999999.99]. Consumer: [0.00, 1000000.00]. Consumer range is wider than manifest.`

**MFP-RANGE-05: Range check in guard clause vs middleware**

- **Manifest:** `quantity: { type: integer, range: [1, 9999] }`
- **Producer** (Go, validation in middleware):
```go
func ValidationMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var body CreateOrderRequest
        json.NewDecoder(r.Body).Decode(&body)
        if body.Quantity < 1 || body.Quantity > 9999 {
            http.Error(w, "invalid quantity", 400)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```
- **Expected:** Stricture detects range validation even though it is in middleware, not the handler directly. No parity violation (validation exists).
- **Verification:** Stricture follows the middleware chain within `handler_paths`.

**MFP-RANGE-06: Range check uses constants vs literals**

- **Manifest:** `amount: { type: number, range: [0.01, 999999.99] }`
- **Code** (Go):
```go
const (
    MinAmount = 0.01
    MaxAmount = 999999.99
)

func validate(amount float64) error {
    if amount < MinAmount || amount > MaxAmount {
        return errors.New("out of range")
    }
    return nil
}
```
- **Expected:** Stricture resolves constant values and detects the range matches the manifest. No parity violation.
- **Verification:** Constants `MinAmount=0.01` and `MaxAmount=999999.99` are resolved to numeric values.

**MFP-RANGE-07: Go range check in custom Validate() method**

- **Manifest:** `amount: { type: number, range: [0.01, 999999.99] }`
- **Code** (Go):
```go
type Invoice struct {
    Amount float64 `json:"amount"`
}

func (i *Invoice) Validate() error {
    if i.Amount < 0.01 || i.Amount > 999999.99 {
        return fmt.Errorf("amount out of range")
    }
    return nil
}
```
- **Expected:** Stricture detects that `Invoice.Validate()` performs range validation on `Amount`. Must verify that `Validate()` is actually called in the handler path.
- **Verification:** If `Validate()` is called before processing, no parity violation. If `Validate()` exists but is never called in the handler chain, parity violation (dead validation code).

**MFP-RANGE-08: TypeScript range check in zod schema**

- **Manifest:** `amount: { type: number, range: [0.01, 999999.99] }`
- **Code** (TypeScript):
```typescript
import { z } from "zod";

const InvoiceSchema = z.object({
  amount: z.number().min(0.01).max(999999.99),
});

function processInvoice(raw: unknown) {
  const invoice = InvoiceSchema.parse(raw);
  // process...
}
```
- **Expected:** Stricture detects zod schema validation and extracts the range `[0.01, 999999.99]`. No parity violation.
- **Verification:** `stricture audit` reports the field as validated with matching range.

---

## 11. CTR-strictness-parity -- Enum Parity

**MFP-ENUM-01: Both handle all enum values -- pass**

- **Manifest:** `role: { type: enum, values: ["admin", "user", "viewer"] }`
- **Producer** (Go):
```go
var validRoles = map[string]bool{"admin": true, "user": true, "viewer": true}

func validateRole(role string) error {
    if !validRoles[role] {
        return fmt.Errorf("invalid role: %s", role)
    }
    return nil
}
```
- **Consumer** (TypeScript):
```typescript
switch (user.role) {
  case "admin":
    // handle admin
    break;
  case "user":
    // handle user
    break;
  case "viewer":
    // handle viewer
    break;
  default:
    throw new Error(`Unknown role: ${user.role}`);
}
```
- **Expected:** No parity violation. Both handle all 3 values.
- **Verification:** `stricture audit` reports "strict" for `role` on both sides.

**MFP-ENUM-02: Producer checks all, consumer misses one -- violation**

- **Manifest:** `role: { type: enum, values: ["admin", "user", "viewer"] }`
- **Producer** (Go): Same as MFP-ENUM-01.
- **Consumer** (TypeScript):
```typescript
switch (user.role) {
  case "admin":
    // handle admin
    break;
  case "user":
    // handle user
    break;
  // "viewer" NOT handled
  default:
    console.log("unknown role");
}
```
- **Expected violation:** `CTR-strictness-parity: field "role" in contract "user-api" -- consumer (web-frontend) does not handle enum value "viewer". Handled: [admin, user]. Missing: [viewer].`

**MFP-ENUM-03: Consumer uses switch without default -- violation**

- **Manifest:** `currency: { type: enum, values: ["USD", "EUR", "GBP"] }`
- **Consumer** (TypeScript):
```typescript
switch (invoice.currency) {
  case "USD":
    return formatUSD(amount);
  case "EUR":
    return formatEUR(amount);
  // No case for "GBP", NO default
}
```
- **Expected violation:** `CTR-strictness-parity: field "currency" -- consumer missing enum value "GBP" and has no default case. Unhandled values will fall through silently.`

**MFP-ENUM-04: Consumer uses if/else covering all values -- pass**

- **Manifest:** `role: { type: enum, values: ["admin", "user", "viewer"] }`
- **Consumer** (TypeScript):
```typescript
if (user.role === "admin") {
  // admin logic
} else if (user.role === "user") {
  // user logic
} else if (user.role === "viewer") {
  // viewer logic
} else {
  throw new Error(`Unknown role: ${user.role}`);
}
```
- **Expected:** No parity violation. All values handled. Else clause handles unknown values.

**MFP-ENUM-05: Go enum check via string set/map**

- **Manifest:** `status: { type: enum, values: ["active", "inactive", "suspended"] }`
- **Code** (Go):
```go
var validStatuses = map[string]struct{}{
    "active":    {},
    "inactive":  {},
    "suspended": {},
}

func isValidStatus(s string) bool {
    _, ok := validStatuses[s]
    return ok
}
```
- **Expected:** Stricture extracts the enum set `{active, inactive, suspended}` from the map literal. No parity violation.
- **Verification:** Map-based enum patterns are recognized.

**MFP-ENUM-06: TypeScript enum check via TypeScript enum + exhaustive switch**

- **Manifest:** `role: { type: enum, values: ["admin", "user", "viewer"] }`
- **Code** (TypeScript):
```typescript
enum Role {
  Admin = "admin",
  User = "user",
  Viewer = "viewer",
}

function handleRole(role: Role): void {
  switch (role) {
    case Role.Admin:
      break;
    case Role.User:
      break;
    case Role.Viewer:
      break;
    default:
      const _exhaustive: never = role;
      throw new Error(`Unhandled role: ${_exhaustive}`);
  }
}
```
- **Expected:** Exhaustive switch detected (uses `never` trick). All enum values handled. No parity violation.

**MFP-ENUM-07: Enum check with partial coverage treated as consumer-side gap**

- **Manifest:** `priority: { type: enum, values: ["low", "medium", "high", "critical"] }`
- **Consumer** (Go):
```go
switch event.Priority {
case "high", "critical":
    alertTeam(event)
case "medium":
    logEvent(event)
// "low" not handled at all -- silently dropped
}
```
- **Expected violation:** `CTR-strictness-parity: field "priority" -- consumer does not handle enum value "low". No default case.`

---

## 12. CTR-strictness-parity -- Format Parity

**MFP-FMT-01: Both validate email format -- pass**

- **Manifest:** `email: { type: string, format: email }`
- **Producer** (TypeScript):
```typescript
import { z } from "zod";
const schema = z.object({ email: z.string().email() });
```
- **Consumer** (Go):
```go
import "net/mail"

func validateEmail(email string) error {
    _, err := mail.ParseAddress(email)
    return err
}
```
- **Expected:** Both validate email format. No parity violation.
- **Verification:** Stricture recognizes `z.string().email()` and `mail.ParseAddress()` as email format validators.

**MFP-FMT-02: Producer validates UUID, consumer accepts any string -- violation**

- **Manifest:** `invoice_id: { type: string, format: uuid }`
- **Producer** (Go):
```go
import "github.com/google/uuid"

func createInvoice() Invoice {
    return Invoice{ID: uuid.New().String()}
}
```
- **Consumer** (TypeScript):
```typescript
function processInvoice(invoice: { invoice_id: string }) {
  // No UUID validation, just uses the string
  console.log(invoice.invoice_id);
}
```
- **Expected violation:** `CTR-strictness-parity: field "invoice_id" -- manifest declares format "uuid". Producer generates valid UUIDs. Consumer performs no format validation.`

**MFP-FMT-03: Regex-based format validation**

- **Manifest:** `email: { type: string, format: email }`
- **Code** (TypeScript):
```typescript
const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

function validateEmail(email: string): boolean {
  return EMAIL_REGEX.test(email);
}
```
- **Expected:** Stricture detects regex-based email validation. Treated as format validation present. No parity violation (regex is a reasonable email check).
- **Verification:** Stricture does not validate the precision of the regex itself (that would require email RFC compliance analysis). Presence of format validation is sufficient.

**MFP-FMT-04: Using standard library validation functions**

- **Manifest:** `created_at: { type: string, format: iso8601 }`
- **Producer** (Go):
```go
import "time"

func formatTime(t time.Time) string {
    return t.Format(time.RFC3339)
}
```
- **Consumer** (TypeScript):
```typescript
function parseDate(dateStr: string): Date {
  const d = new Date(dateStr);
  if (isNaN(d.getTime())) {
    throw new Error("Invalid date format");
  }
  return d;
}
```
- **Expected:** Both handle ISO 8601 format (Go's RFC3339 is a subset of ISO 8601, JS `Date` constructor parses ISO 8601). No parity violation.

**MFP-FMT-05: No format validation on either side -- parity maintained but manifest constraint unmet**

- **Manifest:** `email: { type: string, format: email }`
- **Producer** (Go):
```go
type User struct {
    Email string `json:"email"` // No validation
}
```
- **Consumer** (TypeScript):
```typescript
interface User {
  email: string; // No validation
}
```
- **Expected:** No parity violation (neither side validates, so they are in parity). However, CTR-manifest-conformance flags that the manifest declares `format: email` but neither side enforces it. Two separate issues: parity is symmetric (both equal), conformance is individual (both missing).

---

## 13. CTR-strictness-parity -- Length Parity

**MFP-LEN-01: Both enforce maxLength -- pass**

- **Manifest:** `name: { type: string, maxLength: 255 }`
- **Producer** (Go):
```go
func validateUser(u User) error {
    if len(u.Name) > 255 {
        return fmt.Errorf("name too long: %d characters", len(u.Name))
    }
    return nil
}
```
- **Consumer** (TypeScript):
```typescript
if (user.name.length > 255) {
  throw new Error("Name exceeds maximum length");
}
```
- **Expected:** No parity violation. Both enforce maxLength 255.

**MFP-LEN-02: Producer truncates to maxLength, consumer accepts any length -- violation**

- **Manifest:** `name: { type: string, maxLength: 255 }`
- **Producer** (Go):
```go
if len(name) > 255 {
    name = name[:255] // Silently truncate
}
```
- **Consumer** (TypeScript):
```typescript
function processUser(user: { name: string }) {
  // No length check -- assumes server sent valid data
  displayName(user.name);
}
```
- **Expected violation:** `CTR-strictness-parity: field "name" -- producer enforces maxLength 255 (truncation). Consumer performs no length validation.`
- **Verification:** Even though the consumer receives truncated data, parity requires the consumer to also validate (defensive programming).

**MFP-LEN-03: MinLength enforcement**

- **Manifest:** `name: { type: string, minLength: 1, maxLength: 255 }`
- **Producer** (Go):
```go
if len(u.Name) < 1 {
    return fmt.Errorf("name is required")
}
if len(u.Name) > 255 {
    return fmt.Errorf("name too long")
}
```
- **Consumer** (TypeScript):
```typescript
if (!user.name || user.name.length === 0) {
  throw new Error("Name is required");
}
if (user.name.length > 255) {
  throw new Error("Name too long");
}
```
- **Expected:** No parity violation. Both enforce minLength 1 and maxLength 255.

**MFP-LEN-04: Only maxLength validated, minLength ignored -- partial violation**

- **Manifest:** `name: { type: string, minLength: 1, maxLength: 255 }`
- **Producer** (Go):
```go
if len(u.Name) > 255 {
    return fmt.Errorf("name too long")
}
// minLength NOT checked
```
- **Expected violation:** `CTR-strictness-parity: field "name" -- manifest declares minLength 1 but producer does not enforce minimum length.`

---

## 14. CTR-strictness-parity -- Cross-Language

**MFP-XLANG-01: Go producer with struct validation, TS consumer without -- violation**

- **Manifest:** User API with full field constraints (range, format, maxLength, enum).
- **Producer** (`user-service`, Go):
```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=1,max=255"`
    Email string `json:"email" validate:"required,email"`
    Role  string `json:"role" validate:"required,oneof=admin user viewer"`
}

func (r *CreateUserRequest) Validate() error {
    return validator.Struct(r)
}
```
- **Consumer** (`web-frontend`, TypeScript):
```typescript
interface User {
  id: number;
  name: string;
  email: string;
  role: string;  // No validation at all
}

async function getUser(id: number): Promise<User> {
  const res = await fetch(`/api/users/${id}`);
  return res.json(); // No validation on response
}
```
- **Expected violations for consumer:**
  - `name`: missing length validation.
  - `email`: missing format validation.
  - `role`: missing enum validation (declared as `string` not union type).
  - `id`: missing range validation (if manifest declares range).
- **Verification:** Each field gets its own parity violation. The consumer has 4 violations despite the producer being fully compliant.

**MFP-XLANG-02: TS producer with zod schema, Go consumer with manual checks -- compare constraint equivalence**

- **Manifest:** `amount: { type: number, range: [0.01, 999999.99], precision: 2 }`
- **Producer** (TypeScript):
```typescript
const InvoiceSchema = z.object({
  amount: z.number().min(0.01).max(999999.99),
});
```
- **Consumer** (Go):
```go
func processInvoice(inv Invoice) error {
    if inv.Amount < 0.01 || inv.Amount > 999999.99 {
        return fmt.Errorf("amount out of range")
    }
    return nil
}
```
- **Expected:** No parity violation. Different languages, different validation approaches, but the constraints are equivalent: `[0.01, 999999.99]` on both sides.

**MFP-XLANG-03: Different validation libraries on each side (both valid but different)**

- **Manifest:** `email: { type: string, format: email }`
- **Producer** (Go): Uses `net/mail.ParseAddress()`.
- **Consumer** (TypeScript): Uses `z.string().email()`.
- **Expected:** No parity violation. Both validate email format, using idiomatic validation for their language.
- **Verification:** Stricture's cross-language analysis recognizes that different libraries can achieve the same constraint. Format presence is what matters, not implementation identity.

**MFP-XLANG-04: Go type uses int but manifest says integer -- type mapping**

- **Manifest:** `id: { type: integer, range: [1, 2147483647] }`
- **Code** (Go):
```go
type User struct {
    ID int `json:"id"`
}
```
- **Expected:** No conformance violation. Go's `int` maps to manifest's `integer`.
- **Verification:** Type mapping table: Go `int`, `int32`, `int64` all map to manifest `integer`. Go `float32`, `float64` map to manifest `number`. Go `string` maps to manifest `string`.

**MFP-XLANG-05: TypeScript type uses number but manifest says integer**

- **Manifest:** `id: { type: integer, range: [1, 2147483647] }`
- **Code** (TypeScript):
```typescript
interface User {
  id: number; // TS has no separate integer type
}
```
- **Expected:** No conformance violation. TypeScript's `number` type encompasses both integer and floating-point. Stricture accepts `number` for manifest `integer` in TypeScript contexts.
- **Verification:** TS `number` maps to both `integer` and `number` in manifest. This is a known limitation of TypeScript's type system.

---

## 15. Stricture Audit Command

**MFP-AUDIT-01: Basic audit output**

- **Setup:** Complete manifest (MFP-VALID-02), user-service repo with some validated fields and some missing.
- **Command:** `stricture audit`
- **Expected output structure:**
```
Stricture Audit -- user-service
================================

Contract: user-api (producer)
  Endpoint: GET /api/users/:id
    Response strictness: 85% (strict)
    +-------------+------------+-----------+------------------------+
    | Field       | Manifest   | Code      | Status                 |
    +-------------+------------+-----------+------------------------+
    | id          | int [1,M)  | int       | ! Missing range check  |
    | name        | str [1,255]| str       | ! Missing length check |
    | email       | str email  | str email | OK Strict              |
    | role        | enum(3)    | enum(3)   | OK Strict              |
    | created_at  | iso8601    | time.Time | OK Strict              |
    +-------------+------------+-----------+------------------------+
```
- **Verification:**
  - Header shows service name.
  - Each contract is listed with role.
  - Each endpoint shows a field-by-field comparison.
  - Strictness percentage is calculated from field-level scores.
  - Status column shows "OK" or "!" with reason.

**MFP-AUDIT-02: Audit with no violations**

- **Setup:** All fields in all contracts are fully validated.
- **Command:** `stricture audit`
- **Expected:** All fields show "OK Strict" or better. Strictness score: 100%. Exit code 0.

**MFP-AUDIT-03: Audit with no manifest configured**

- **Setup:** `.stricture.yml` has no `manifest` section.
- **Command:** `stricture audit`
- **Expected:** `No manifest configured. Add a "manifest" section to .stricture.yml to enable cross-service audit.` Exit code 0 (not an error, just no-op).

**MFP-AUDIT-04: Audit with multiple contracts**

- **Setup:** Service participates in 3 contracts.
- **Command:** `stricture audit`
- **Expected:** Output shows all 3 contracts in sequence. Each has its own field table. Summary at the bottom aggregates across all contracts.

**MFP-AUDIT-05: Audit summary line**

- **Setup:** 2 contracts, 9 fields total. 4 strict, 2 basic, 3 missing validation.
- **Expected summary:**
```
Strictness Score: 62% (needs improvement)
2 contracts, 9 fields, 4 strict, 2 basic, 3 missing validation
```

**MFP-AUDIT-06: Audit with message queue contracts**

- **Setup:** Service consumes a message queue contract with 4 fields.
- **Command:** `stricture audit`
- **Expected:** Output shows the message queue contract, event name, and field-by-field parity just like HTTP contracts but organized by event instead of endpoint.
```
Contract: billing-events (consumer)
  Message: invoice.created
    Consumer strictness: 40% (basic)
    +-------------+--------------+------------+-------------------------+
    | Field       | Manifest     | Code       | Status                  |
    +-------------+--------------+------------+-------------------------+
    | invoice_id  | uuid         | string     | X No format validation  |
    | user_id     | int [1,M)    | int        | X No range validation   |
    | amount      | num [0.01,M) | float64    | X No range validation   |
    | currency    | enum(3)      | string     | X No enum validation    |
    +-------------+--------------+------------+-------------------------+
```

**MFP-AUDIT-07: Audit JSON output**

- **Command:** `stricture audit --format json`
- **Expected:** Valid JSON output with structure:
```json
{
  "service": "user-service",
  "manifest": "acme-platform",
  "contracts": [
    {
      "id": "user-api",
      "role": "producer",
      "endpoints": [
        {
          "path": "/api/users/:id",
          "method": "GET",
          "strictness_percentage": 85,
          "strictness_level": "strict",
          "fields": [
            {
              "name": "id",
              "manifest_constraint": "integer, range [1, 2147483647], required",
              "code_constraint": "integer, no range check",
              "status": "warning",
              "message": "Missing range check"
            }
          ]
        }
      ]
    }
  ],
  "summary": {
    "total_fields": 9,
    "strict_fields": 4,
    "basic_fields": 2,
    "missing_fields": 3,
    "overall_percentage": 62
  }
}
```

---

## 16. Stricture Trace Command

**MFP-TRACE-01: Basic HAR trace validation**

- **Setup:** Manifest from MFP-VALID-02. HAR file with 10 requests to `GET /api/users/:id`, all returning valid responses.
- **Command:** `stricture trace traffic.har`
- **Expected:**
```
Trace Audit -- 10 requests matched to contracts
================================================

Contract: user-api
  GET /api/users/:id (10 requests)
    OK All responses had required fields
    OK All field values within declared ranges
    OK All enum values from allowed set
    OK All status codes from declared set
```
- **Verification:** Exit code 0.

**MFP-TRACE-02: Trace detects enum violation in response**

- **Setup:** Manifest declares `role: { type: enum, values: ["admin", "user", "viewer"] }`. HAR file has 1 response with `role: "moderator"`.
- **Command:** `stricture trace traffic.har`
- **Expected:**
```
Contract: user-api
  GET /api/users/:id (50 requests)
    OK All responses had required fields
    X 1 response had role="moderator" -- not in enum [admin, user, viewer]
```
- **Verification:** Exit code 1. Violation identifies the specific response, field, and invalid value.

**MFP-TRACE-03: Trace detects missing required field**

- **Setup:** Manifest declares `name: { type: string, required: true }`. One response in HAR file is missing the `name` field.
- **Expected:** `X 1 response missing required field "name"`.

**MFP-TRACE-04: Trace detects range violation**

- **Setup:** Manifest declares `amount: { type: number, range: [0.01, 999999.99] }`. One request has `amount: 0.00`.
- **Expected:** `X 1 request had amount=0.00 -- below range minimum 0.01`.

**MFP-TRACE-05: Trace detects undeclared status code**

- **Setup:** Manifest declares `status_codes: [200, 400, 404, 500]`. HAR file contains responses with status 503.
- **Expected:** `Warning: 12 responses had status 503 -- not in declared status_codes [200, 400, 404, 500]`.

**MFP-TRACE-06: Trace with unmatched requests**

- **Setup:** HAR file contains requests to `/healthz` and `/metrics` which have no manifest contract.
- **Expected:**
```
Unmatched requests: 57 (no manifest contract)
  GET /healthz (30)
  GET /metrics (27)
```
- **Verification:** Unmatched requests are reported as informational, not as errors.

**MFP-TRACE-07: Trace with OpenTelemetry format**

- **Command:** `stricture trace --format otel traces.json`
- **Expected:** Parses OpenTelemetry JSON export. Matches spans to manifest endpoints by HTTP method + path attributes. Same field validation as HAR.

**MFP-TRACE-08: Trace with custom JSON format**

- **Input** (`custom-traces.json`):
```json
[
  {
    "method": "GET",
    "path": "/api/users/42",
    "response_body": { "id": 42, "name": "Alice", "email": "alice@example.com", "role": "admin", "created_at": "2026-01-15T10:00:00Z" },
    "status": 200
  }
]
```
- **Command:** `stricture trace custom-traces.json`
- **Expected:** Parses custom JSON format. Validates response body fields against manifest.

**MFP-TRACE-09: Trace with no matching contracts**

- **Setup:** HAR file contains only requests to endpoints not in the manifest.
- **Expected:**
```
Trace Audit -- 0 requests matched to contracts

Unmatched requests: 100 (no manifest contract)
  GET /api/legacy (50)
  POST /api/old-endpoint (50)

Summary: 0 matched, 0 violations, 100 unmatched
```
- **Verification:** Exit code 0 (no violations found). Warning about unmatched requests.

**MFP-TRACE-10: Trace summary**

- **Setup:** 190 matched requests, 6 violations, 57 unmatched.
- **Expected:**
```
Summary: 190 matched, 6 violations, 57 unmatched
```
- **Verification:** Exit code 1 (violations found).

---

## 17. Edge Cases & Error Handling

**MFP-EDGE-01: Manifest with 0 contracts**

- **Input manifest:**
```yaml
manifest_version: "1.0"
name: "empty-contracts"
services:
  api:
    repo: "github.com/acme/api"
    language: typescript
    role: producer
contracts: []
strictness:
  minimum: minimal
  rules: {}
```
- **Expected:** Warning: `Manifest "empty-contracts" has 0 contracts. Nothing to validate.`
- **Verification:** Parses successfully. `stricture audit` produces empty output.

**MFP-EDGE-02: Contract with 0 endpoints**

- **Input manifest** (contract):
```yaml
contracts:
  - id: "empty-api"
    producer: api
    consumers: [client]
    protocol: http
    endpoints: []
```
- **Expected:** Parse error: `Contract "empty-api": must have at least 1 endpoint.`
- **Verification:** Exit code 2.

**MFP-EDGE-03: Endpoint with 0 fields (void response)**

- **Input manifest** (endpoint):
```yaml
endpoints:
  - path: "/api/users/:id"
    method: DELETE
    status_codes: [204, 404]
```
- **Expected:** Valid. DELETE with 204 has no response body. No `response` or `request` section needed.
- **Verification:** Parity checks only apply to status code handling.

**MFP-EDGE-04: Very large manifest (100 services, 500 contracts)**

- **Input:** Auto-generated manifest with 100 services, 500 contracts, 2000 endpoints, 10000 fields.
- **Expected:** Parses within 5 seconds. No out-of-memory errors. All contracts accessible.
- **Verification:** `stricture audit` completes within performance targets (see section 18). Memory usage < 200MB for manifest parsing alone.

**MFP-EDGE-05: Circular service references (A consumes B, B consumes A)**

- **Input manifest:**
```yaml
services:
  service-a:
    repo: "github.com/acme/a"
    language: go
    role: both
  service-b:
    repo: "github.com/acme/b"
    language: typescript
    role: both

contracts:
  - id: "a-to-b"
    producer: service-a
    consumers: [service-b]
    protocol: http
    endpoints:
      - path: "/api/data"
        method: GET
        response:
          type: DataResponse
          fields:
            value: { type: string, required: true }
        status_codes: [200]

  - id: "b-to-a"
    producer: service-b
    consumers: [service-a]
    protocol: http
    endpoints:
      - path: "/api/events"
        method: POST
        request:
          type: EventRequest
          fields:
            event_type: { type: string, required: true }
        status_codes: [201]
```
- **Expected:** Valid. Circular service dependencies are allowed -- service A produces contract X and consumes contract Y where service B is the reverse. This is common in microservice architectures.
- **Verification:** No circular dependency error. Both contracts are independently validated.

**MFP-EDGE-06: Service consuming its own API**

- **Input manifest:**
```yaml
services:
  api-service:
    repo: "github.com/acme/api"
    language: typescript
    role: both

contracts:
  - id: "self-api"
    producer: api-service
    consumers: [api-service]
    protocol: http
    endpoints:
      - path: "/api/internal/sync"
        method: POST
        request:
          type: SyncRequest
          fields:
            data: { type: string, required: true }
        status_codes: [200]
```
- **Expected:** Valid. Self-consumption is allowed (service calls its own API, common for internal sync endpoints). Parity checks apply: the same service is both producer and consumer, so its handler AND its client code must both be analyzed.
- **Verification:** `stricture audit` reports the contract twice: once as producer (handler validation), once as consumer (client validation).

**MFP-EDGE-07: Manifest with only message queue contracts (no HTTP)**

- **Input manifest:**
```yaml
manifest_version: "1.0"
name: "event-only"
services:
  publisher:
    repo: "github.com/acme/publisher"
    language: go
    role: producer
  subscriber:
    repo: "github.com/acme/subscriber"
    language: typescript
    role: consumer

contracts:
  - id: "events"
    producer: publisher
    consumers: [subscriber]
    protocol: message_queue
    queue: "main.events"
    messages:
      - event: "item.created"
        fields:
          id: { type: string, format: uuid, required: true }

strictness:
  minimum: basic
  rules: {}
```
- **Expected:** Valid. HTTP is not required. Message queue contracts work standalone.
- **Verification:** `stricture audit` reports the message queue contract.

**MFP-EDGE-08: Contract where producer and consumer are in different languages**

- **Setup:** Producer is Go, consumer is TypeScript. Manifest declares fields with constraints.
- **Expected:** Cross-language parity checks apply. Stricture uses the Go adapter for producer analysis and TypeScript adapter for consumer analysis. Constraints are compared at the semantic level, not the syntax level.
- **Verification:** Tested extensively in sections 10-14.

**MFP-EDGE-09: Field type "any" or "unknown"**

- **Input manifest** (field):
```yaml
metadata: { type: any, required: false }
```
- **Expected:** Parse error: `Field "metadata": unknown type "any". Valid types: [string, integer, number, enum, boolean, object, array]. Use "object" for unstructured data.`
- **Verification:** Exit code 2. The manifest must be precise -- `any` defeats the purpose of strictness.

**MFP-EDGE-10: Manifest file does not exist at path**

- **Input** (`.stricture.yml`):
```yaml
manifest:
  path: "../nonexistent/stricture-manifest.yml"
  service: api
  contracts: []
```
- **Expected:** Error: `Manifest file not found at "../nonexistent/stricture-manifest.yml". Verify the path relative to the project root.`
- **Verification:** Exit code 2.

**MFP-EDGE-11: Manifest file is not valid YAML**

- **Setup:** Manifest file exists but contains binary data.
- **Expected:** Error: `Failed to parse manifest: not valid YAML.`
- **Verification:** Exit code 2.

**MFP-EDGE-12: Manifest URL unreachable**

- **Input** (`.stricture.yml`):
```yaml
manifest:
  url: "github.com/nonexistent-org/nonexistent-repo"
  service: api
  contracts: []
```
- **Expected:** Error: `Failed to fetch manifest from "github.com/nonexistent-org/nonexistent-repo": repository not found or not accessible.`
- **Verification:** Exit code 2.

**MFP-EDGE-13: Service with role "consumer" listed as producer in contract**

- **Input manifest:**
```yaml
services:
  frontend:
    repo: "github.com/acme/frontend"
    language: typescript
    role: consumer  # Declared as consumer

contracts:
  - id: "api"
    producer: frontend  # But listed as producer!
    consumers: []
    protocol: http
    endpoints: []
```
- **Expected:** Parse error: `Contract "api" declares "frontend" as producer, but service "frontend" has role "consumer". Change the service role to "producer" or "both".`
- **Verification:** Exit code 2.

**MFP-EDGE-14: Service with role "producer" listed as consumer in contract**

- **Input manifest:**
```yaml
services:
  api:
    repo: "github.com/acme/api"
    language: go
    role: producer  # Declared as producer

contracts:
  - id: "events"
    producer: other-service
    consumers: [api]  # But "api" is listed as consumer!
    protocol: http
    endpoints: []
```
- **Expected:** Parse error: `Contract "events" lists "api" as consumer, but service "api" has role "producer". Change the service role to "consumer" or "both".`
- **Verification:** Exit code 2.

**MFP-EDGE-15: Missing strictness section**

- **Input manifest:**
```yaml
manifest_version: "1.0"
name: "no-strictness"
services:
  api:
    repo: "github.com/acme/api"
    language: typescript
    role: producer
contracts:
  - id: "api"
    producer: api
    consumers: []
    protocol: http
    endpoints:
      - path: "/health"
        method: GET
        response:
          type: Health
          fields:
            ok: { type: boolean, required: true }
        status_codes: [200]
```
- **Expected:** Default strictness applied: `minimum: minimal`, all rules disabled. No strictness-level violations for the manifest itself.
- **Verification:** Parses successfully. `stricture audit` applies code-level conformance and parity checks but not manifest-level strictness enforcement.

**MFP-EDGE-16: Contract with mixed protocol types**

- **Input manifest:**
```yaml
contracts:
  - id: "mixed"
    producer: api
    consumers: [client]
    protocol: http
    queue: "events"  # queue field on an HTTP contract
    endpoints:
      - path: "/api/test"
        method: GET
        status_codes: [200]
    messages:
      - event: "test"
        fields:
          id: { type: string, required: true }
```
- **Expected:** Warning: `Contract "mixed" has protocol "http" but includes "queue" and "messages" fields. These are ignored for HTTP contracts. Use protocol "message_queue" for queue-based contracts.`
- **Verification:** Only `endpoints` are processed. `messages` are ignored.

---

## 18. Performance & Scale

**MFP-PERF-01: Manifest parsing -- small manifest**

- **Input:** 5 services, 10 contracts, 50 endpoints, 200 fields.
- **Expected:** Parse completes in < 100ms.
- **Verification:** Time the parse step in isolation.

**MFP-PERF-02: Manifest parsing -- large manifest**

- **Input:** 100 services, 500 contracts, 2000 endpoints, 10000 fields.
- **Expected:** Parse completes in < 2s. Memory < 200MB.
- **Verification:** Time the parse step. Measure peak memory.

**MFP-PERF-03: Audit with many fields**

- **Input:** Service participates in 10 contracts, 100 endpoints, 500 fields.
- **Expected:** `stricture audit` completes in < 10s including code analysis.
- **Verification:** End-to-end timing.

**MFP-PERF-04: Trace with large HAR file**

- **Input:** HAR file with 10000 requests. Manifest has 20 endpoints.
- **Expected:** `stricture trace` completes in < 30s.
- **Verification:** End-to-end timing.

**MFP-PERF-05: Parity analysis across languages**

- **Setup:** Contract with 1 Go producer and 3 TypeScript consumers.
- **Expected:** Cross-language parity analysis adds < 50% overhead vs. single-language analysis.
- **Verification:** Compare `stricture audit` time with 1 consumer vs. 3 consumers.

**MFP-PERF-06: Caching -- repeated audit runs**

- **Setup:** Run `stricture audit` twice on unchanged code.
- **Expected:** Second run is > 2x faster than first (cache hit for parsed ASTs and constraint extraction).
- **Verification:** Compare wall-clock time of first vs. second run.

**MFP-PERF-07: Incremental audit after single file change**

- **Setup:** Run `stricture audit`, change 1 handler file, run again.
- **Expected:** Only the changed file is re-analyzed. Other files use cache.
- **Verification:** Log output shows which files were re-parsed vs. cache hit.

---

*End of cross-service manifest test plan.*
