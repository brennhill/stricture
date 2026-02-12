# Contract (CTR) Rules — Test Cases

8 rules covering dual-contract testing and cross-service validation.

Sections 23-28 cover the original 6 rules. Sections 29-30 cover the 2 new cross-service rules (CTR-strictness-parity, CTR-manifest-conformance) — these are added in a separate step.

---

## 23. CTR-request-shape

**Rule purpose:** Verify that the request body type the client sends matches what the server expects.

### 23.1 True Positive Cases

**TP-RS-01: Client sends wrong field name**

- **Server** (`src/routes/users.ts`):
```typescript
interface CreateUserRequest { name: string; email: string; role: "admin" | "user"; }
app.post("/api/users", (req, res) => { const body: CreateUserRequest = req.body; });
```
- **Client** (`src/services/api-client.ts`):
```typescript
async function createUser(data: { name: string; email: string; type: string }) {
  return fetch("/api/users", { method: "POST", body: JSON.stringify(data) });
}
```
- **Expected violation:** `CTR-request-shape`. Client sends "type" but server expects "role". Missing field "role", extra field "type".

**TP-RS-02: Client missing required field**

- **Server expects:** `{ name: string; email: string; role: string; }`
- **Client sends:** `{ name: string; email: string; }`
- **Expected violation:** Missing "role" field.

**TP-RS-03: Type mismatch on field**

- **Server expects:** `{ id: number; }`
- **Client sends:** `{ id: string; }`
- **Expected violation:** Type mismatch: "id" is number on server, string on client.

**TP-RS-04: Go struct missing json tag field**

- **Server** (`handlers.go`):
```go
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Role  string `json:"role"`
}
```
- **Client** (`client.go`):
```go
type CreateUserPayload struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}
```
- **Expected violation:** Client missing "role" field.

**TP-RS-05: Fuzzy name mismatch (userId vs user_id)**

- **Config:** `fuzzyNameMatch: true`
- **Server expects:** `{ user_id: string; }`
- **Client sends:** `{ userId: string; }`
- **Expected violation:** Potential name mismatch: `userId` vs `user_id`.

**TP-RS-06: Nested object shape mismatch**

- **Server expects:** `{ address: { street: string; city: string; zip: string; } }`
- **Client sends:** `{ address: { street: string; city: string; } }`
- **Expected violation:** Missing "zip" in nested address.

**TP-RS-07: Array element type mismatch**

- **Server expects:** `{ items: { id: number; qty: number; }[] }`
- **Client sends:** `{ items: { id: string; quantity: number; }[] }`
- **Expected violation:** "id" type mismatch (number vs string), field name mismatch ("qty" vs "quantity").

**TP-RS-08: Client sends extra fields (strictExtraFields=true)**

- **Config:** `strictExtraFields: true`
- **Client sends:** `{ name: string; email: string; role: string; avatar: string; }`
- **Server expects:** `{ name: string; email: string; role: string; }`
- **Expected violation:** Extra field "avatar" on client.

**TP-RS-09: Annotation-based pair with mismatch**

- Both files annotated with `// stricture-contract: server=... client=...`
- Request shapes differ.
- **Expected violation:** Same detection as auto-detected pairs.

**TP-RS-10: Optional field on server, client sends wrong type**

- **Server:** `{ name: string; nickname?: string; }` (nickname optional)
- **Client:** `{ name: string; nickname: number; }` (nickname wrong type)
- **Expected violation:** Type mismatch on nickname (string vs number).

### 23.2 True Negative Cases

**TN-RS-01:** Client and server have matching request types.
**TN-RS-02:** Client sends extra optional fields (strictExtraFields=false).
**TN-RS-03:** Client imports type from shared package.
**TN-RS-04:** Go structs with matching json tags.
**TN-RS-05:** Optional field missing from client (ignoreOptionalFields=true).

### 23.3 False Positive Risks

**FP-RS-01:** Server uses middleware to transform request before handler sees it.
**FP-RS-02:** Client sends to external API (not in repo) -- should not pair.
**FP-RS-03:** Server accepts `any` type or untyped body.
**FP-RS-04:** REST convention differences (query params vs body).
**FP-RS-05:** Multipart form data vs JSON body.

### 23.4 False Negative Risks

**FN-RS-01:** Request type is `any` or `interface{}` -- no type info to compare.
**FN-RS-02:** Request body built dynamically (`body[key] = value`).
**FN-RS-03:** Middleware adds fields to request before handler.
**FN-RS-04:** Different endpoint versions (v1 vs v2) with different shapes.
**FN-RS-05:** GraphQL mutation vs REST endpoint (different paradigms).

### 23.5 Edge Cases

**EC-RS-01:** Path parameters vs body fields.
**EC-RS-02:** Query string parameters.
**EC-RS-03:** Headers as part of contract (Authorization, Content-Type).
**EC-RS-04:** Multipart form data fields.
**EC-RS-05:** File upload endpoints.
**EC-RS-06:** Server with multiple request types per endpoint (content negotiation).
**EC-RS-07:** Go request parsed with json.Decoder vs json.Unmarshal.
**EC-RS-08:** TypeScript request with Zod schema validation.
**EC-RS-09:** Request through API gateway (path rewriting).
**EC-RS-10:** WebSocket message shapes (not HTTP).

### 23.6 Configuration Interaction

**CI-RS-01:** `strictExtraFields: true` -- extra client fields are errors.
**CI-RS-02:** `strictExtraFields: false` -- extra client fields are warnings.
**CI-RS-03:** `fuzzyNameMatch: true` -- flag potential casing mismatches.
**CI-RS-04:** `ignoreOptionalFields: true` -- optional server fields not required from client.

### 23.7 Inline Suppression Testing

Same structure as 1.7 with rule ID `CTR-request-shape`.

---

## 24. CTR-response-shape

**Rule purpose:** Verify that the response body type the server sends matches what the client expects.

### 24.1 True Positive Cases

**TP-RESP-01: Client expects field server does not send**

- **Server sends:** `{ id: number; name: string; email: string; created_at: string; }`
- **Client expects:** `{ id: number; name: string; email: string; createdAt: string; avatar: string; }`
- **Expected violation:** Name mismatch: "created_at" vs "createdAt". Missing from server: "avatar".

**TP-RESP-02: Go json tag mismatch in response**

- **Server:**
```go
type UserResponse struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}
```
- **Client:**
```go
type UserResponse struct {
    ID   int    `json:"user_id"` // MISMATCH
    Name string `json:"name"`
}
```
- **Expected violation:** Server sends "id", client expects "user_id".

**TP-RESP-03: Server response has extra fields client ignores**

- Not a violation by default (server can send more than client needs).
- But if configured strictly, could warn.

**TP-RESP-04: Type mismatch on response field**

- Server sends `{ count: number }`, client expects `{ count: string }`.
- **Expected violation:** Type mismatch.

**TP-RESP-05: Nested response object mismatch**

- Server sends `{ data: { items: { id: number }[] } }`
- Client expects `{ data: { items: { id: string }[] } }`
- **Expected violation:** Nested type mismatch on items[].id.

**TP-RESP-06: Server sends literal object, client expects typed interface**

- Server: `res.json({ id: user.id, name: user.name })` (no email)
- Client: `interface UserResponse { id: number; name: string; email: string; }`
- **Expected violation:** Client expects "email" but server does not send it.

**TP-RESP-07: Date field serialization mismatch**

- Server sends `createdAt` as ISO string, client expects `Date` object.
- **Expected:** This is a type mismatch (string vs Date). Whether to flag depends on serialization analysis.

**TP-RESP-08: Enum value mismatch**

- Server can send `role: "admin" | "user" | "moderator"`
- Client expects `role: "admin" | "user"`
- **Expected violation:** Server can send "moderator" which client does not handle.

**TP-RESP-09: Annotation-based pair with response mismatch**

- Same as annotation-based request detection but for response.

**TP-RESP-10: OpenAPI spec says one thing, code does another**

- OpenAPI spec defines response as `{ id: number; name: string; }`.
- Server actually sends `{ id: number; fullName: string; }`.
- **Expected violation:** Server implementation does not match spec.

### 24.2 True Negative Cases

**TN-RESP-01:** Matching response types on both sides.
**TN-RESP-02:** Client uses shared type imported from common package.
**TN-RESP-03:** Server sends superset of what client expects (extra fields OK).
**TN-RESP-04:** Go structs with matching json tags.
**TN-RESP-05:** Response type is `any`/`unknown` on client (no type to check).

### 24.3-24.7: Same structure as CTR-request-shape (False Positives, False Negatives, Edge Cases, Config Interaction, Inline Suppression)

Key differences: direction is reversed (server sends, client receives). Go `json.Marshal` tag resolution is critical. Object literal analysis on server side needed.

---

## 25. CTR-status-code-handling

**Rule purpose:** Verify that the client handles all status codes the server can return.

### 25.1 True Positive Cases

**TP-SCH-01: Client does not check status at all**

- **Server returns:** 200, 400, 404, 500
- **Client:**
```typescript
async function getUser(id: string): Promise<User> {
  const res = await fetch(`/api/users/${id}`);
  return res.json(); // No status check!
}
```
- **Expected violation:** Client handles 0/4 status codes.

**TP-SCH-02: Client only checks res.ok**

- **Config:** `requireExplicit: true`
- **Client:**
```typescript
if (!res.ok) throw new Error("request failed");
return res.json();
```
- **Expected violation:** `res.ok` is not explicit. Must handle 400, 404, 500 individually.

**TP-SCH-03: Client handles 200 and 404 but not 400 and 500**

- **Server can return:** 200, 400, 404, 500
- **Client handles:** 200, 404
- **Expected violation:** Unhandled: 400, 500.

**TP-SCH-04: Go client ignoring non-200 status**

- **Client:**
```go
resp, err := http.Get(url)
if err != nil { return nil, err }
defer resp.Body.Close()
var user User
json.NewDecoder(resp.Body).Decode(&user)
return &user, nil // No status code check!
```
- **Expected violation:** No status code handling.

**TP-SCH-05: Server returns 201 (Created), client only checks 200**

- **Client:** `if (res.status === 200)` -- misses 201.
- **Expected violation:** Unhandled 201.

**TP-SCH-06: Server returns 204 (No Content), client tries to parse body**

- **Client:**
```typescript
const data = await res.json(); // 204 has no body -- this will throw
```
- **Expected violation:** 204 not handled (attempting to parse empty body).

**TP-SCH-07: Client ignores 5xx errors (ignore5xx=false)**

- **Config:** `ignore5xx: false`
- **Client:** Handles 200, 400, 404 but not 500.
- **Expected violation:** 500 unhandled.

**TP-SCH-08: Server uses implicit 200 and explicit error codes**

- **Server:**
```typescript
app.get("/users", (req, res) => {
  // Implicit 200 on res.json(data)
  // Explicit 404 on res.status(404).json({error: "not found"})
  // Implicit 500 on unhandled exception
});
```
- **Client handles:** Only 200.
- **Expected violation:** 404 and 500 not handled.

**TP-SCH-09: Multiple endpoints with different status codes**

- Endpoint A returns 200, 400. Endpoint B returns 200, 404, 500. Client for B only handles 200.
- **Expected violation:** Only on client B's unhandled codes.

**TP-SCH-10: Go server returning custom error codes**

- Server: `w.WriteHeader(http.StatusConflict)` (409)
- Client: Does not handle 409.
- **Expected violation:** 409 unhandled.

### 25.2 True Negative Cases

**TN-SCH-01:** Client handles all server status codes explicitly.
**TN-SCH-02:** Client uses `res.ok` check (with `requireExplicit: false`).
**TN-SCH-03:** Client has catch-all `if (!res.ok)` plus specific handlers.
**TN-SCH-04:** Go client checks `resp.StatusCode` for all possible values.
**TN-SCH-05:** Client handles 5xx with `ignore5xx: true`.

### 25.3-25.7: Same structure as previous CTR rules (FP risks include: server status codes in unreachable code, generic error middleware; FN risks include: status codes from middleware not in handler; Edge cases include redirect status codes 301/302, HEAD requests, CORS preflight 204).

### 25.6 Configuration Interaction

**CI-SCH-01:** `requireExplicit: true` -- `res.ok` alone is not enough.
**CI-SCH-02:** `requireExplicit: false` -- `res.ok` counts as handling non-2xx.
**CI-SCH-03:** `ignore5xx: true` -- 5xx codes do not need explicit handling.
**CI-SCH-04:** `ignore5xx: false` -- 5xx codes must be handled.

---

## 26. CTR-shared-type-sync

**Rule purpose:** When client and server reference the same type name, verify the definitions are identical or from a shared package.

### 26.1 True Positive Cases

**TP-STS-01: Same name, different fields (TS)**

- **Server** (`src/routes/types.ts`):
```typescript
interface User { id: number; name: string; email: string; }
```
- **Client** (`src/services/types.ts`):
```typescript
interface User { id: number; name: string; } // Missing email
```
- **Expected violation:** `CTR-shared-type-sync`. Type "User" defined in both files with different fields. Server has "email", client does not.

**TP-STS-02: Same name, different types on field (Go)**

- **Server:**
```go
type User struct { ID int; Name string; }
```
- **Client:**
```go
type User struct { ID string; Name string; } // ID is string, not int
```
- **Expected violation:** Type mismatch on "ID": int vs string.

**TP-STS-03: Extra field on one side**

- Server User has 5 fields, client User has 4.
- **Expected violation:** Field count mismatch.

**TP-STS-04: Same fields, different json tags (Go)**

- **Server:** `Name string \`json:"name"\``
- **Client:** `Name string \`json:"user_name"\``
- **Expected violation:** JSON tag mismatch.

**TP-STS-05: Type redefined locally instead of importing shared**

- **Config:** `requireSharedPackage: true`
- Both server and client define their own `UserResponse` type.
- **Expected violation:** Type should be imported from shared package.

**TP-STS-06: Deeply nested type with difference at leaf**

- Both sides define `OrderResponse` with nested `items[].product.name`. Server has `name: string`, client has `name: string | null`.
- **Expected violation:** Type mismatch at nested level.

**TP-STS-07: Enum/union type with different members**

- Server: `type Role = "admin" | "user" | "moderator"`
- Client: `type Role = "admin" | "user"`
- **Expected violation:** Client missing "moderator" variant.

**TP-STS-08: Go embedded struct difference**

- Server embeds `BaseModel` with 3 fields, client does not embed.
- **Expected violation:** Field count mismatch after flattening embedded struct.

**TP-STS-09: Optional vs required field difference**

- Server: `email: string` (required)
- Client: `email?: string` (optional)
- **Expected violation:** Optionality mismatch.

**TP-STS-10: Array vs single value**

- Server: `items: Item[]`
- Client: `items: Item`
- **Expected violation:** Type mismatch (array vs single).

### 26.2-26.7: Standard structure for TN, FP, FN, Edge Cases, Config, Inline Suppression.

Key config options:
- `requireSharedPackage: false/true` -- whether duplicate type names always error.
- `ignoreTestFiles: true` -- test-local type redefinitions are OK.

---

## 27. CTR-json-tag-match

**Rule purpose:** (Go-specific) Verify that JSON struct tags match across contract boundaries.

### 27.1 True Positive Cases

**TP-JTM-01: Snake_case vs camelCase tag mismatch**

- **Server:**
```go
type UserResponse struct {
    CreatedAt time.Time `json:"created_at"`
}
```
- **Client:**
```go
type UserResponse struct {
    CreatedAt time.Time `json:"createdAt"` // MISMATCH
}
```
- **Expected violation:** JSON tag mismatch: server uses "created_at", client uses "createdAt".

**TP-JTM-02: Missing json tag on one side**

- **Server:** `Name string \`json:"name"\``
- **Client:** `Name string` (no json tag -- defaults to "Name")
- **Expected violation:** Server sends "name", client expects "Name".

**TP-JTM-03: Omitempty on one side only**

- **Server:** `Email string \`json:"email,omitempty"\``
- **Client:** `Email string \`json:"email"\``
- **Expected:** Warning. `omitempty` means server might not send the field.

**TP-JTM-04: json:"-" on one side**

- **Server:** `Password string \`json:"-"\``
- **Client:** `Password string \`json:"password"\``
- **Expected violation:** Server never sends "password" (json:"-"), but client expects it.

**TP-JTM-05: Nested struct tag mismatch**

- Server's nested struct has `json:"user_id"`, client's nested struct has `json:"userId"`.
- **Expected violation:** Nested tag mismatch.

**TP-JTM-06: All tags use different conventions**

- Server consistently uses snake_case, client consistently uses camelCase.
- **Expected violation:** Multiple mismatches flagged, with suggestion to standardize.

**TP-JTM-07: Embedded struct with conflicting tags**

- Embedded struct promotes fields. Tag on promoted field differs.
- **Expected violation:** After flattening embedded structs, compare promoted field tags.

**TP-JTM-08: Slice element struct tag mismatch**

- Server: `Items []struct { ID int \`json:"id"\` }`
- Client: `Items []struct { ID int \`json:"item_id"\` }`
- **Expected violation:** Array element struct tag mismatch.

**TP-JTM-09: Convention enforcement**

- **Config:** `convention: snake_case`
- **Input:** Struct with `json:"userId"` (camelCase)
- **Expected violation:** Tag "userId" does not match enforced snake_case convention.

**TP-JTM-10: Map key struct tag**

- `map[string]struct{ Value int \`json:"val"\` }` on one side, `json:"value"` on other.
- **Expected violation:** Tag mismatch.

### 27.2 True Negative Cases

**TN-JTM-01:** Matching tags on both sides.
**TN-JTM-02:** Both sides import type from shared package.
**TN-JTM-03:** Tags match convention (all snake_case).
**TN-JTM-04:** Non-contract struct (internal only) -- not checked.
**TN-JTM-05:** TypeScript file -- rule is Go-only, skipped.

### 27.3-27.7: Standard structure for FP, FN, Edge Cases, Config, Inline Suppression.

Key edge cases: protobuf-generated structs, custom JSON marshalers, struct tags with multiple keys (`json:"name" xml:"name"`).

---

## 28. CTR-dual-test

**Rule purpose:** Verify that contract pairs have tests on both sides that exercise the same scenarios.

### 28.1 True Positive Cases

**TP-DT-01: Server tests 404, client has no 404 test**

- **Server test:**
```typescript
it("returns 404 for unknown user", async () => {
  const res = await request(app).get("/api/users/unknown");
  expect(res.status).toBe(404);
});
```
- **Client test:** Only tests 200 success case.
- **Expected violation:** `CTR-dual-test`. Server tests 404 for /api/users/:id, but client has no matching 404 test.

**TP-DT-02: Client tests error, server has no error test**

- **Client test:**
```typescript
it("handles network error", async () => {
  mock("/api/users/123", { status: 500 });
  await expect(getUser("123")).rejects.toThrow(ServerError);
});
```
- **Server test:** No test that returns 500.
- **Expected violation:** Client tests 500 from /api/users/:id, server has no matching 500 test.

**TP-DT-03: Server tests validation (400), client ignores validation**

- Server has test for 400 on invalid input. Client has no test for validation error handling.
- **Expected violation:** Unmatched scenario.

**TP-DT-04: Multiple endpoints, one without dual tests**

- `/api/users` has dual tests. `/api/orders` has server tests but no client tests.
- **Expected violation:** Only /api/orders flagged.

**TP-DT-05: Go server test with no matching Go client test**

- Server tests return different status codes. Client tests only cover success.
- **Expected violation:** Per unmatched scenario.

**TP-DT-06: Server test uses different path format**

- Server registers `/api/users/:id`, test uses `/api/users/123`. Client fetches `/api/users/${id}`.
- **Expected:** Should still match (both reference same endpoint pattern).

**TP-DT-07: Only success scenarios tested on both sides**

- Both sides test 200. Neither tests 400/404/500.
- **Expected violation:** If `requireBothDirections: true` and server can return 400/404/500, the missing error scenarios should be flagged.

**TP-DT-08: Client test uses mock, server test is integration**

- Both exist but test different things. Client mocks the server. Server tests real logic.
- **Expected:** This is acceptable -- both sides have tests. Scenario fingerprint matching should still work.

**TP-DT-09: Confidence below threshold**

- **Config:** `minConfidence: 80`
- Match is at 70% confidence.
- **Expected:** No violation (below threshold).

**TP-DT-10: Confidence above threshold**

- Match is at 90% confidence.
- **Expected violation:** Flagged.

### 28.2 True Negative Cases

**TN-DT-01:** Both sides have matching test scenarios for all status codes.
**TN-DT-02:** Client has error handling tests that match server error tests.
**TN-DT-03:** Both sides test the same endpoint with same scenarios.
**TN-DT-04:** Internal endpoint with `ignoreInternalEndpoints: true`.
**TN-DT-05:** No contract pairs detected -- rule does not apply.

### 28.3-28.7: Standard structure for FP, FN, Edge Cases, Config, Inline Suppression.

Key configuration:
- `minConfidence: 80` -- only flag above this confidence.
- `requireBothDirections: true` -- both server-to-client and client-to-server checked.
- `ignoreInternalEndpoints: false` -- check all endpoints.

---

## 29. CTR-strictness-parity

**Rule purpose:** When a cross-service manifest exists, verify that both producer and consumer enforce the same constraints for every field in a contract. If the manifest says `amount` has range `[0.01, 999999.99]`, both the emitting service and the consuming service must validate that range.

### 29.1 True Positive Cases

**TP-SP-01: Range parity — producer validates, consumer does not**

- **Manifest:**
```yaml
fields:
  amount: { type: number, range: [0.01, 999999.99] }
```
- **Producer** (`billing-service/handlers.go`):
```go
func createInvoice(amount float64) error {
    if amount < 0.01 || amount > 999999.99 {
        return ErrInvalidAmount
    }
    return publishEvent("invoice.created", Invoice{Amount: amount})
}
```
- **Consumer** (`user-service/handlers.go`):
```go
func handleInvoice(event InvoiceEvent) error {
    // No range validation — uses amount directly
    db.SaveInvoice(event.InvoiceID, event.Amount)
    return nil
}
```
- **Expected violation:** `CTR-strictness-parity`. Consumer does not validate range `[0.01, 999999.99]` on field "amount". Producer validates; consumer does not.

**TP-SP-02: Enum parity — producer checks all values, consumer handles only some**

- **Manifest:**
```yaml
fields:
  role: { type: enum, values: ["admin", "user", "viewer"] }
```
- **Producer** (`src/serializers/user.ts`):
```typescript
function serializeUser(user: User) {
  if (!["admin", "user", "viewer"].includes(user.role)) {
    throw new ValidationError("invalid role");
  }
  return JSON.stringify(user);
}
```
- **Consumer** (`src/handlers/user-handler.ts`):
```typescript
function handleUser(data: UserResponse) {
  switch (data.role) {
    case "admin": return renderAdmin(data);
    case "user": return renderUser(data);
    // MISSING: "viewer" — no default case
  }
}
```
- **Expected violation:** Consumer handles 2 of 3 enum values for field "role". Missing: "viewer".

**TP-SP-03: Format parity — producer validates email format, consumer accepts any string**

- **Manifest:**
```yaml
fields:
  email: { type: string, format: email }
```
- **Producer** (`src/routes/users.ts`):
```typescript
app.post("/api/users", (req, res) => {
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(req.body.email)) {
    return res.status(400).json({ error: "invalid email" });
  }
  // ...
});
```
- **Consumer** (`src/services/user-service.ts`):
```typescript
function processUser(data: UserResponse) {
  // Accepts email as plain string, no format validation
  sendWelcomeEmail(data.email);
}
```
- **Expected violation:** Producer validates email format on field "email"; consumer does not.

**TP-SP-04: Length parity — producer enforces maxLength, consumer does not**

- **Manifest:**
```yaml
fields:
  name: { type: string, maxLength: 255 }
```
- **Producer** (`handlers.go`):
```go
func createUser(req CreateUserRequest) error {
    if len(req.Name) > 255 {
        return ErrNameTooLong
    }
    return repo.Save(req)
}
```
- **Consumer** (`client.go`):
```go
func handleUserCreated(event UserCreatedEvent) error {
    // No length check — stores name directly
    return db.InsertUser(event.Name, event.Email)
}
```
- **Expected violation:** Producer validates maxLength(255) on field "name"; consumer does not.

**TP-SP-05: Required field parity — producer marks required, consumer treats as optional**

- **Manifest:**
```yaml
fields:
  email: { type: string, required: true }
```
- **Producer** (`src/routes/users.ts`):
```typescript
app.post("/api/users", (req, res) => {
  if (!req.body.email) {
    return res.status(400).json({ error: "email is required" });
  }
  // ...
});
```
- **Consumer** (`src/services/user-service.ts`):
```typescript
interface UserEvent {
  name: string;
  email?: string; // Optional — but manifest says required
}
function handleUserEvent(event: UserEvent) {
  if (event.email) {
    sendEmail(event.email);
  }
  // Silently skips email if missing — no error raised
}
```
- **Expected violation:** Manifest declares "email" as required. Producer enforces; consumer treats as optional.

**TP-SP-06: One side uses custom validation library (zod), other has no validation**

- **Manifest:**
```yaml
fields:
  age: { type: integer, range: [0, 150] }
```
- **Producer** (`src/schemas/user.ts`):
```typescript
import { z } from "zod";
const UserSchema = z.object({
  age: z.number().int().min(0).max(150),
});
export function validateUser(data: unknown) {
  return UserSchema.parse(data);
}
```
- **Consumer** (`src/handlers/user-handler.ts`):
```typescript
function processUser(data: UserResponse) {
  // No zod or any validation — uses age directly
  renderAgeGroup(data.age);
}
```
- **Expected violation:** Producer validates range [0, 150] on field "age" via zod schema; consumer has no validation.

**TP-SP-07: Validation in middleware vs handler — consumer has middleware validation but producer validates in handler**

- **Manifest:**
```yaml
fields:
  amount: { type: number, range: [1, 10000] }
```
- **Producer** (`src/routes/payments.ts`):
```typescript
app.post("/api/payments", (req, res) => {
  if (req.body.amount < 1 || req.body.amount > 10000) {
    return res.status(400).json({ error: "invalid amount" });
  }
  // ...
});
```
- **Consumer** (`src/middleware/validate.ts`):
```typescript
function validatePaymentEvent(event: PaymentEvent): boolean {
  // Only checks presence, not range
  return event.amount !== undefined;
}
```
- **Consumer handler** (`src/handlers/payment-handler.ts`):
```typescript
function handlePayment(event: PaymentEvent) {
  // Relies on middleware for validation — but middleware only checks presence
  processPayment(event.amount);
}
```
- **Expected violation:** Producer validates range [1, 10000] on field "amount". Consumer only checks presence, not range.

**TP-SP-08: Both sides validate but with different ranges**

- **Manifest:**
```yaml
fields:
  quantity: { type: integer, range: [1, 999] }
```
- **Producer** (`handlers.go`):
```go
func createOrder(req OrderRequest) error {
    if req.Quantity < 1 || req.Quantity > 999 {
        return ErrInvalidQuantity
    }
    return repo.Save(req)
}
```
- **Consumer** (`consumer.go`):
```go
func handleOrder(event OrderEvent) error {
    if event.Quantity < 0 || event.Quantity > 9999 {
        return fmt.Errorf("invalid quantity")
    }
    return processOrder(event)
}
```
- **Expected violation:** Range mismatch on field "quantity". Manifest: [1, 999]. Producer: [1, 999]. Consumer: [0, 9999]. Consumer range is wider than manifest.

**TP-SP-09: Producer has no validation — relies on database constraints**

- **Manifest:**
```yaml
fields:
  email: { type: string, format: email, maxLength: 320 }
```
- **Producer** (`handlers.go`):
```go
func createUser(req CreateUserRequest) error {
    // No application-level validation — relies on DB constraint
    return db.Exec("INSERT INTO users (email) VALUES ($1)", req.Email)
}
```
- **Consumer** (`consumer.go`):
```go
func handleUserCreated(event UserCreatedEvent) error {
    if !isValidEmail(event.Email) {
        return fmt.Errorf("invalid email: %s", event.Email)
    }
    return processUser(event)
}
```
- **Expected violation:** Producer does not validate format or maxLength on field "email" at the application level. Database constraints are not recognized as equivalent validation.

**TP-SP-10: Consumer validates with switch/case but missing default**

- **Manifest:**
```yaml
fields:
  status: { type: enum, values: ["pending", "active", "suspended", "deleted"] }
```
- **Producer** (`src/services/account.ts`):
```typescript
function emitAccountStatus(status: string) {
  const valid = ["pending", "active", "suspended", "deleted"];
  if (!valid.includes(status)) {
    throw new Error("invalid status");
  }
  emit("account.status", { status });
}
```
- **Consumer** (`src/handlers/account-handler.ts`):
```typescript
function handleAccountStatus(event: AccountEvent) {
  switch (event.status) {
    case "pending": return markPending(event);
    case "active": return activate(event);
    case "suspended": return suspend(event);
    // MISSING: "deleted" and no default case
  }
}
```
- **Expected violation:** Consumer handles 3 of 4 enum values for field "status". Missing: "deleted". No default/fallback case.

**TP-SP-11: Go producer uses go-playground/validator tags, consumer has no validation**

- **Manifest:**
```yaml
fields:
  price: { type: number, range: [0.01, 99999.99] }
  sku: { type: string, minLength: 3, maxLength: 20 }
```
- **Producer** (`models.go`):
```go
type Product struct {
    Price float64 `json:"price" validate:"required,min=0.01,max=99999.99"`
    SKU   string  `json:"sku" validate:"required,min=3,max=20"`
}
```
- **Consumer** (`consumer.go`):
```go
type ProductEvent struct {
    Price float64 `json:"price"`
    SKU   string  `json:"sku"`
}
func handleProduct(event ProductEvent) error {
    // No validation at all
    return db.SaveProduct(event.Price, event.SKU)
}
```
- **Expected violation:** Producer validates range and length via struct tags; consumer has no equivalent validation on fields "price" and "sku".

**TP-SP-12: TypeScript producer uses joi schema, consumer uses manual partial checks**

- **Manifest:**
```yaml
fields:
  username: { type: string, minLength: 3, maxLength: 30, pattern: "^[a-zA-Z0-9_]+$" }
```
- **Producer** (`src/schemas/user.ts`):
```typescript
import Joi from "joi";
const userSchema = Joi.object({
  username: Joi.string().min(3).max(30).pattern(/^[a-zA-Z0-9_]+$/),
});
```
- **Consumer** (`src/handlers/user-handler.ts`):
```typescript
function handleUser(data: UserResponse) {
  if (data.username.length < 3) {
    throw new Error("username too short");
  }
  // Missing: maxLength and pattern checks
  processUser(data);
}
```
- **Expected violation:** Consumer partially validates field "username" (minLength only). Missing: maxLength and pattern constraints.

### 29.2 True Negative Cases

**TN-SP-01:** Both producer and consumer validate the same range on a numeric field. No violation.
**TN-SP-02:** Both sides use the same shared validation library (e.g., both import from a shared `validators` package) with identical constraints. No violation.
**TN-SP-03:** Consumer handles all enum values including a default/else fallback. No violation.
**TN-SP-04:** Field has type-only constraint in manifest (`{ type: string }`) with no range/format/length. Neither side validates beyond type. No violation.
**TN-SP-05:** Both sides use go-playground/validator tags with matching constraints. No violation.
**TN-SP-06:** `allowTrustedInternal: true` and both services are in the trusted-internal list. No violation even if consumer skips validation.

### 29.3 False Positive Risks

**FP-SP-01:** Consumer validates the constraint in a middleware layer that the rule engine cannot trace. The validation exists but is not in the handler function.
**FP-SP-02:** Consumer uses a database-level CHECK constraint that effectively enforces the range. Application code has no explicit check, but the constraint is enforced.
**FP-SP-03:** Consumer validates via a type system constraint (e.g., a branded type in TypeScript: `type PositiveInt = number & { __brand: "positive" }`) that is not recognized as runtime validation.
**FP-SP-04:** Consumer is a read-only aggregator that only displays data and never persists or acts on invalid values. Strict validation is arguably unnecessary.
**FP-SP-05:** Producer validates in a pre-serialization hook or interceptor that is not co-located with the handler. Rule engine cannot trace the validation path.
**FP-SP-06:** Both services share the same database, and the DB schema enforces the constraint. Neither application needs redundant validation.

### 29.4 False Negative Risks

**FN-SP-01:** Consumer validates the field, but with a wider range than the manifest specifies (e.g., manifest says [1, 100], consumer checks [0, 200]). Rule might see "has validation" without comparing bounds.
**FN-SP-02:** Consumer checks the field in a dynamically constructed conditional (e.g., `if (field < config.MIN)`) where `config.MIN` is loaded at runtime from environment variables.
**FN-SP-03:** Producer constructs the field value without validation but guarantees correctness via business logic (e.g., auto-incrementing IDs). The manifest constraint is trivially satisfied but not explicitly validated.
**FN-SP-04:** Consumer uses `try/catch` around the entire handler, which might implicitly catch invalid values at a lower layer. The rule engine sees no explicit field-level validation.
**FN-SP-05:** Both sides use a shared protobuf definition that has implicit type constraints, but neither side validates the values beyond deserialization. The constraint checking happens in generated code that the rule engine does not analyze.

### 29.5 Edge Cases

**EC-SP-01:** Manifest declares a field with multiple constraints (range + format + maxLength). Producer validates all three; consumer validates only one. Should flag the two missing constraints individually.
**EC-SP-02:** Field constraint is `{ type: enum, values: [] }` (empty enum). Both sides should reject all values. If neither validates, is that parity? (Yes — both are equally non-strict.)
**EC-SP-03:** Manifest declares a nested object field with per-sub-field constraints. Producer validates the top-level object; consumer validates only some sub-fields.
**EC-SP-04:** Producer validates via a `switch` with a `default` that returns an error. Consumer has the same `switch` but `default` is a no-op. This is a parity gap even though both have `default`.
**EC-SP-05:** Field is an array. Manifest says `{ type: array, items: { type: number, range: [0, 100] } }`. Producer validates each element; consumer only checks array length.
**EC-SP-06:** Manifest has a `nullable: true` constraint. Producer allows null; consumer does not check for null and crashes on `null.toString()`.
**EC-SP-07:** Both sides validate but use different error handling strategies (producer returns 400, consumer throws and retries). Parity is about constraint enforcement, not error handling. Should not flag.
**EC-SP-08:** Go producer uses `json:",string"` tag to serialize a number as a string. Consumer expects a number. The constraint is on the serialized form, not the Go type.
**EC-SP-09:** Manifest constraint references another field (e.g., `endDate > startDate`). Cross-field validation parity is harder to detect statically.
**EC-SP-10:** Service acts as both producer and consumer for the same field (relay/proxy). Should be validated in both roles.
**EC-SP-11:** Manifest declares `deprecated: true` on a field. Both sides still use the field. Parity check should still apply to deprecated fields unless configured otherwise.
**EC-SP-12:** Multiple producers for the same manifest field (e.g., two microservices both emit `user.created`). Parity must hold for each producer-consumer pair independently.

### 29.6 Configuration Interaction

**CI-SP-01:** `requireBothSides: true` -- both producer and consumer must validate. This is the default. When false, only flag if one side explicitly validates and the other does not.
**CI-SP-02:** `allowTrustedInternal: true` -- skip parity checks for services listed in the manifest's `trustedInternal` list.
**CI-SP-03:** `allowTrustedInternal: false` -- even internal services must validate.
**CI-SP-04:** `manifestPath: "custom/path/manifest.yml"` -- override the default manifest location. Rule is a no-op if the manifest file does not exist.
**CI-SP-05:** Rule is a no-op when no `manifest` section exists in `.stricture.yml`. Enabling the rule without a manifest produces no violations.

### 29.7 Inline Suppression Testing

**IS-SP-01: Next-line suppression in TypeScript**
```typescript
// stricture-disable-next-line CTR-strictness-parity
function handleUser(data: UserResponse) {
  // No validation — suppressed
  processUser(data);
}
```
- **Expected:** No violation reported for this function.

**IS-SP-02: Block suppression in Go**
```go
// stricture-disable CTR-strictness-parity
func handleInvoice(event InvoiceEvent) error {
    db.SaveInvoice(event.InvoiceID, event.Amount)
    return nil
}
// stricture-enable CTR-strictness-parity
```
- **Expected:** No violation reported within the block.

**IS-SP-03: File-level suppression**
- File begins with `// stricture-disable-file CTR-strictness-parity`.
- **Expected:** No violations for the entire file.

**IS-SP-04: Suppression on producer side only**
- Producer suppresses the rule, consumer does not.
- **Expected:** Consumer is still checked. If consumer lacks validation, violation is reported on consumer side only.

**IS-SP-05: Suppression with reason**
```typescript
// stricture-disable-next-line CTR-strictness-parity -- trusted internal service, validated at API gateway
function handleEvent(data: EventPayload) { /* ... */ }
```
- **Expected:** Suppression accepted. Reason is recorded in lint output.

---

## 30. CTR-manifest-conformance

**Rule purpose:** Verify that the actual code (types, handlers, clients) conforms to the cross-service manifest declarations. The manifest is the source of truth -- code must match it. Catches extra fields (data leak risk), missing fields (contract violation), and type mismatches.

### 30.1 True Positive Cases

**TP-MC-01: Extra field in code not in manifest (data leak risk)**

- **Manifest:**
```yaml
endpoints:
  GET /api/users/:id:
    response:
      fields:
        id: { type: integer }
        name: { type: string }
        email: { type: string }
```
- **Server** (`handlers.go`):
```go
type UserResponse struct {
    ID            int       `json:"id"`
    Name          string    `json:"name"`
    Email         string    `json:"email"`
    InternalNotes string    `json:"internal_notes"` // NOT in manifest
}
func getUser(w http.ResponseWriter, r *http.Request) {
    user := repo.FindUser(r.PathValue("id"))
    json.NewEncoder(w).Encode(UserResponse{
        ID: user.ID, Name: user.Name, Email: user.Email,
        InternalNotes: user.InternalNotes,
    })
}
```
- **Expected violation:** `CTR-manifest-conformance`. Field "internal_notes" exists in code response but not in manifest. Potential data leak.

**TP-MC-02: Missing field in code that manifest requires**

- **Manifest:**
```yaml
endpoints:
  GET /api/users/:id:
    response:
      fields:
        id: { type: integer, required: true }
        name: { type: string, required: true }
        email: { type: string, required: true }
        role: { type: string, required: true }
```
- **Server** (`src/routes/users.ts`):
```typescript
app.get("/api/users/:id", async (req, res) => {
  const user = await db.findUser(req.params.id);
  res.json({
    id: user.id,
    name: user.name,
    email: user.email,
    // MISSING: role — manifest says it's required
  });
});
```
- **Expected violation:** Manifest requires field "role" in response, but server handler does not include it.

**TP-MC-03: Type mismatch — manifest says integer, code uses string**

- **Manifest:**
```yaml
endpoints:
  POST /api/orders:
    request:
      fields:
        quantity: { type: integer }
```
- **Client** (`src/services/order-client.ts`):
```typescript
async function createOrder(data: { quantity: string; productId: string }) {
  return fetch("/api/orders", { method: "POST", body: JSON.stringify(data) });
}
```
- **Expected violation:** Field "quantity" type mismatch. Manifest declares integer; client code uses string.

**TP-MC-04: Field name mismatch — camelCase vs snake_case**

- **Manifest:**
```yaml
endpoints:
  GET /api/users/:id:
    response:
      fields:
        created_at: { type: string, format: iso8601 }
```
- **Server** (`src/routes/users.ts`):
```typescript
interface UserResponse {
  id: number;
  name: string;
  createdAt: string; // camelCase, but manifest says snake_case "created_at"
}
```
- **Expected violation:** Field name mismatch. Manifest declares "created_at"; code uses "createdAt". Serialized JSON name does not match manifest.

**TP-MC-05: Manifest has endpoint that does not exist in code**

- **Manifest:**
```yaml
endpoints:
  GET /api/users/:id:
    response: { fields: { id: { type: integer } } }
  DELETE /api/users/:id:
    response: { fields: { success: { type: boolean } } }
```
- **Server** (`src/routes/users.ts`):
```typescript
app.get("/api/users/:id", (req, res) => { /* ... */ });
// DELETE /api/users/:id is NOT registered
```
- **Expected violation:** Manifest declares endpoint `DELETE /api/users/:id` but no matching handler exists in code.

**TP-MC-06: Code has endpoint not in manifest**

- **Manifest:**
```yaml
endpoints:
  GET /api/users/:id:
    response: { fields: { id: { type: integer }, name: { type: string } } }
```
- **Server** (`handlers.go`):
```go
mux.HandleFunc("GET /api/users/{id}", getUser)
mux.HandleFunc("PUT /api/users/{id}", updateUser) // NOT in manifest
```
- **Expected violation:** Handler `PUT /api/users/{id}` exists in code but is not declared in the manifest.

**TP-MC-07: Go struct tags do not match manifest field names**

- **Manifest:**
```yaml
endpoints:
  GET /api/products/:id:
    response:
      fields:
        product_name: { type: string }
        unit_price: { type: number }
```
- **Server** (`models.go`):
```go
type ProductResponse struct {
    ProductName string  `json:"productName"` // Manifest says "product_name"
    UnitPrice   float64 `json:"unitPrice"`   // Manifest says "unit_price"
}
```
- **Expected violation:** JSON tag "productName" does not match manifest field name "product_name". JSON tag "unitPrice" does not match manifest field name "unit_price".

**TP-MC-08: TypeScript interface has optional field, manifest says required**

- **Manifest:**
```yaml
endpoints:
  POST /api/users:
    request:
      fields:
        name: { type: string, required: true }
        email: { type: string, required: true }
        role: { type: string, required: true }
```
- **Client** (`src/services/api-client.ts`):
```typescript
interface CreateUserRequest {
  name: string;
  email: string;
  role?: string; // Optional in code, required in manifest
}
```
- **Expected violation:** Field "role" is optional in code but required by manifest.

**TP-MC-09: Response type built dynamically — spread operator adds unmanifested fields**

- **Manifest:**
```yaml
endpoints:
  GET /api/users/:id:
    response:
      fields:
        id: { type: integer }
        name: { type: string }
```
- **Server** (`src/routes/users.ts`):
```typescript
app.get("/api/users/:id", async (req, res) => {
  const user = await db.findUser(req.params.id);
  const metadata = { lastLogin: new Date(), ipAddress: req.ip };
  res.json({ ...user, ...metadata }); // Spreads unknown fields — may leak data
});
```
- **Expected violation:** Response is built dynamically via spread. Cannot verify conformance to manifest. Fields "lastLogin" and "ipAddress" are not in manifest. Potential data leak.

**TP-MC-10: Handler returns different types in different code paths**

- **Manifest:**
```yaml
endpoints:
  GET /api/users/:id:
    response:
      fields:
        id: { type: integer }
        name: { type: string }
        email: { type: string }
```
- **Server** (`handlers.go`):
```go
func getUser(w http.ResponseWriter, r *http.Request) {
    user, err := repo.FindUser(r.PathValue("id"))
    if err != nil {
        // Path A: error response — different shape
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }
    // Path B: success response — matches manifest
    json.NewEncoder(w).Encode(UserResponse{ID: user.ID, Name: user.Name, Email: user.Email})
}
```
- **Expected violation:** Handler has multiple response paths. Success path matches manifest, but error path returns `{"error": string}` which does not match the manifest response shape.

**TP-MC-11: Client request sends field with wrong nested type**

- **Manifest:**
```yaml
endpoints:
  POST /api/orders:
    request:
      fields:
        items:
          type: array
          items:
            fields:
              productId: { type: integer }
              quantity: { type: integer }
```
- **Client** (`src/services/order-client.ts`):
```typescript
interface OrderItem {
  productId: string; // Manifest says integer
  quantity: number;
}
async function createOrder(items: OrderItem[]) {
  return fetch("/api/orders", { method: "POST", body: JSON.stringify({ items }) });
}
```
- **Expected violation:** Nested field "items[].productId" type mismatch. Manifest: integer. Code: string.

**TP-MC-12: Go handler uses Object.assign equivalent (struct embedding adds extra fields)**

- **Manifest:**
```yaml
endpoints:
  GET /api/users/:id:
    response:
      fields:
        id: { type: integer }
        name: { type: string }
```
- **Server** (`handlers.go`):
```go
type BaseResponse struct {
    RequestID string `json:"request_id"`
    Timestamp int64  `json:"timestamp"`
}
type UserResponse struct {
    BaseResponse          // Embeds request_id and timestamp — not in manifest
    ID           int      `json:"id"`
    Name         string   `json:"name"`
}
```
- **Expected violation:** Embedded struct `BaseResponse` adds fields "request_id" and "timestamp" to response. These fields are not in the manifest.

### 30.2 True Negative Cases

**TN-MC-01:** Server response type has exactly the fields declared in the manifest, with matching types and names. No violation.
**TN-MC-02:** Client request type matches manifest request fields exactly. No violation.
**TN-MC-03:** Go struct JSON tags match manifest field names. No violation.
**TN-MC-04:** Server sends a subset of manifest fields and `allowResponseSubset: true` is configured. No violation.
**TN-MC-05:** Endpoint exists in code but is not covered by any manifest contract (unmatched endpoints are informational, not violations).
**TN-MC-06:** Error response shape differs from success response, but the manifest has a separate `errors` section declaring the error shape. Both match. No violation.

### 30.3 False Positive Risks

**FP-MC-01:** Server uses middleware to strip extra fields before serialization (e.g., a response sanitizer). The code type has extra fields, but they never reach the wire.
**FP-MC-02:** Server uses `json:"-"` tags to exclude fields from serialization. The struct has extra fields, but they are not serialized. Rule engine must check JSON tags before flagging.
**FP-MC-03:** Manifest is outdated but code is correct. A new field was added to the code and the manifest has not been updated yet. The violation is valid per the rule but may frustrate developers.
**FP-MC-04:** Server conditionally includes fields based on API version headers. Manifest covers v2, code serves both v1 and v2. v1 handler is flagged for missing v2 fields.
**FP-MC-05:** Go struct has computed/transient fields with `json:"-"` or no JSON tag. These are internal and not serialized. Should not be flagged.
**FP-MC-06:** TypeScript response is built via a mapper function that explicitly picks only manifest-declared fields (`pick(user, ["id", "name", "email"])`). The source type has extra fields, but the mapper ensures conformance.

### 30.4 False Negative Risks

**FN-MC-01:** Response is built with `res.json(Object.fromEntries(...))` or similar dynamic construction. Static analysis cannot determine the field set.
**FN-MC-02:** Go handler serializes with `json.Marshal(map[string]interface{}{...})` where keys are built at runtime. Field set is unknowable statically.
**FN-MC-03:** Handler delegates response construction to another function in a different file. Cross-function, cross-file tracing is needed but not performed.
**FN-MC-04:** Response type is a generic (`Response<T>`) and `T` is resolved at the call site. Rule engine may not perform generic instantiation.
**FN-MC-05:** Manifest endpoint uses path parameter patterns that differ from code (manifest: `/api/users/:id`, code: `/api/users/{id}`). Pattern normalization is needed for matching.

### 30.5 Edge Cases

**EC-MC-01:** Manifest declares a field as `type: any` or `type: object` without sub-fields. Code has a specific type. Should this pass or fail? (Pass -- manifest is less specific, code is more specific.)
**EC-MC-02:** Manifest has duplicate field names at different nesting levels (e.g., top-level `id` and nested `items[].id`). Rule must compare at the correct level.
**EC-MC-03:** Go struct uses custom JSON marshaler (`MarshalJSON`). The struct fields do not reflect the serialized output. Rule should warn that custom marshaler prevents static analysis.
**EC-MC-04:** Manifest endpoint path has optional segments (`/api/users/:id?`). Code registers separate handlers for `/api/users` and `/api/users/:id`.
**EC-MC-05:** Server returns a paginated response wrapping manifest-declared fields: `{ data: [...], total: 10, page: 1 }`. Manifest only declares fields inside `data`. Rule must understand pagination wrapper.
**EC-MC-06:** TypeScript code uses `Omit<>` or `Pick<>` utility types to derive the response type from a larger interface. Rule must resolve utility types to compare actual field sets.
**EC-MC-07:** Manifest declares a field with `oneOf` (union type): `status: { oneOf: [{ type: string }, { type: integer }] }`. Code uses `string | number`. Should match.
**EC-MC-08:** Go handler writes response headers and body separately. Some "fields" are in headers (e.g., `X-Request-ID`), not in the JSON body. Manifest should distinguish header vs body fields.
**EC-MC-09:** Manifest endpoint uses HTTP method override (`X-HTTP-Method-Override: DELETE`). Code handler checks the header. Path matching must account for method overrides.
**EC-MC-10:** Server returns different response shapes based on query parameters (e.g., `?fields=id,name` for sparse fieldsets). Manifest declares the full shape. Rule should not flag when server returns a subset if sparse fieldsets are documented.
**EC-MC-11:** Manifest declares a response field as `nullable: true`. Code type uses a pointer (`*string` in Go, `string | null` in TS). These should be recognized as conformant.
**EC-MC-12:** Client and server are in different repositories. Manifest is in a third shared repository. Rule must resolve the manifest path across repo boundaries or accept a configured path.

### 30.6 Configuration Interaction

**CI-MC-01:** `strictExtraFields: true` -- fields in code but not in manifest are errors (default). When false, extra fields are warnings.
**CI-MC-02:** `allowResponseSubset: false` -- server must return ALL manifest fields (default). When true, server may return a subset.
**CI-MC-03:** `allowResponseSubset: true` -- server may omit optional manifest fields from the response without triggering a violation.
**CI-MC-04:** `manifestPath: "custom/path/manifest.yml"` -- override the default manifest file location. Rule is a no-op if the manifest file does not exist.
**CI-MC-05:** Rule is a no-op when no `manifest` section exists in `.stricture.yml`. Enabling the rule without a manifest produces no violations.
**CI-MC-06:** Combined with `CTR-strictness-parity`: constraint mismatches (right field, right type, wrong validation bounds) are deferred to CTR-strictness-parity and not double-reported by CTR-manifest-conformance.

### 30.7 Inline Suppression Testing

**IS-MC-01: Next-line suppression on struct definition (Go)**
```go
// stricture-disable-next-line CTR-manifest-conformance
type UserResponse struct {
    ID            int    `json:"id"`
    Name          string `json:"name"`
    InternalNotes string `json:"internal_notes"` // Extra field — suppressed
}
```
- **Expected:** No violation reported for this struct.

**IS-MC-02: Block suppression around handler (TypeScript)**
```typescript
// stricture-disable CTR-manifest-conformance
app.get("/api/users/:id", async (req, res) => {
  const user = await db.findUser(req.params.id);
  res.json({ ...user, debugInfo: process.env.NODE_ENV }); // Extra field — suppressed
});
// stricture-enable CTR-manifest-conformance
```
- **Expected:** No violation reported within the block.

**IS-MC-03: File-level suppression**
- File begins with `// stricture-disable-file CTR-manifest-conformance`.
- **Expected:** No violations for the entire file.

**IS-MC-04: Suppression on one endpoint, violation on another**
- Handler file has two endpoints. One is suppressed, the other is not.
- **Expected:** Only the unsuppressed endpoint is checked and may report violations.

**IS-MC-05: Suppression with reason**
```go
// stricture-disable-next-line CTR-manifest-conformance -- manifest update pending in PR #342
type LegacyResponse struct {
    OldField string `json:"old_field"` // Will be removed when manifest merges
}
```
- **Expected:** Suppression accepted. Reason is recorded in lint output.

---

