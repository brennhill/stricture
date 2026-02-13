# Incremental Analysis Testing

Scenarios for validating `stricture --changed` and `stricture --staged` modes, where only a subset of files is analyzed.

---

## Why This Matters

In CI, `stricture --changed` is the primary use case. It must detect cross-file violations even when only one file in a contract pair has changed. If it misses violations because it only analyzes the changed file without context from unchanged files, it defeats the purpose.

---

## Scenario 1: Client Type Changed, Server Unchanged

**Setup:**
```
src/server/routes/users.ts    (unchanged)
src/client/api-client.ts      (changed — added new field "avatar")
```

**Expected behavior:** Stricture loads the unchanged server file as context. `CTR-request-shape` should detect that the client now sends "avatar" but the server doesn't expect it (if `strictExtraFields: true`).

**Test steps:**
1. Create both files with matching types
2. Modify client to add extra field
3. Run `stricture --changed` (only client file in changed set)
4. Assert: `CTR-request-shape` violation on client file

---

## Scenario 2: Server Response Updated, Client Not Updated

**Setup:**
```
src/server/routes/users.ts    (changed — added "role" to response)
src/client/api-client.ts      (unchanged — still uses old type without "role")
```

**Expected behavior:** `CTR-shared-type-sync` should detect that the server now returns a field the client doesn't handle.

**Test steps:**
1. Both files with matching types
2. Server adds new required field to response
3. Run `stricture --changed` (only server file changed)
4. Assert: `CTR-shared-type-sync` warning on the contract pair

---

## Scenario 3: Shared Type Modified But Only One Consumer Changed

**Setup:**
```
src/shared/types.ts            (changed — renamed field "name" to "fullName")
src/server/routes/users.ts     (changed — updated to use "fullName")
src/client/api-client.ts       (unchanged — still uses "name")
```

**Expected behavior:** Even though the client is unchanged, Stricture should detect the stale reference via the changed shared type.

**Test steps:**
1. All three files in sync
2. Rename field in shared type and server
3. Run `stricture --changed` (shared type + server changed)
4. Assert: `CTR-shared-type-sync` violation referencing the unchanged client

---

## Scenario 4: Test File Changed But Source Unchanged

**Setup:**
```
src/user-service.ts            (unchanged)
tests/user-service.test.ts     (changed — developer weakened an assertion)
```

**Expected behavior:** TQ rules should still run. `TQ-no-shallow-assertions` should detect that the test now uses `toBeDefined()` instead of field-level checks, even though the source file is unchanged.

**Test steps:**
1. Source + test file both correct
2. Change test to use shallow assertion
3. Run `stricture --changed` (only test file changed)
4. Assert: `TQ-no-shallow-assertions` violation

---

## Scenario 5: Manifest Changed But No Source Files Changed

**Setup:**
```
.stricture-manifest.yml        (changed — tightened amount range from [1, 999999] to [50, 99999999])
src/services/payment.ts        (unchanged — validates amount >= 1)
```

**Expected behavior:** `CTR-strictness-parity` should detect that the source code's validation (amount >= 1) is now weaker than the manifest's constraint (amount >= 50).

**Test steps:**
1. Manifest and source in sync with old range
2. Update manifest range to be stricter
3. Run `stricture --changed` (only manifest changed)
4. Assert: `CTR-strictness-parity` violation on payment.ts

---

## Scenario 6: Architecture Violation in Changed File

**Setup:**
```
src/services/user-service.ts   (changed — added import from routes/)
src/routes/users.ts            (unchanged)
```

**Expected behavior:** `ARCH-dependency-direction` should detect the reverse import even though only one file changed.

**Test steps:**
1. Service file with correct imports
2. Add `import { handler } from "../routes/users"`
3. Run `stricture --changed`
4. Assert: `ARCH-dependency-direction` violation

---

## Scenario 7: Staged Files Only (Pre-commit Hook)

**Setup:**
```
git add src/client/api-client.ts     (staged — has shallow assertion)
src/client/other-client.ts           (modified but NOT staged — also has issues)
```

**Expected behavior:** `stricture --staged` should only analyze the staged file. The unstaged file's issues should NOT appear in output.

**Test steps:**
1. Stage one file with violation
2. Modify another file with different violation (don't stage)
3. Run `stricture --staged`
4. Assert: only the staged file's violation appears

---

## Scenario 8: Go JSON Tag Changed, TypeScript Client Unchanged

**Setup:**
```
server/models/order.go         (changed — renamed json:"order_id" to json:"id")
client/src/types.ts            (unchanged — still uses order_id)
```

**Expected behavior:** `CTR-json-tag-match` should detect the cross-language mismatch.

**Test steps:**
1. Go server and TS client with matching json tags
2. Rename Go json tag
3. Run `stricture --changed` (only Go file changed)
4. Assert: `CTR-json-tag-match` violation

---

## Implementation Notes

- Stricture must load the **project context** (all files) even in `--changed` mode
- Only the changed files are analyzed for violations, but unchanged files provide context
- The dependency graph must be built from all files to detect cross-file issues
- Performance: loading context is O(all files), but rule execution is O(changed files)
- Cache: the project context can be cached between runs (`.stricture-cache/`)
