# Stricture Dataflow Analysis: What Complex Multi-Protocol Systems Reveal

> Analysis of the Gasoline MCP codebase (`/Users/brenn/dev/gasoline`) against the Stricture product spec. This document identifies what Stricture's current 32 rules would catch, what they would miss, and proposes new capabilities for cross-language, cross-protocol contract enforcement.

---

## 1. Gasoline Data Flow Map

Gasoline's data flows through 6 protocol boundaries, 3 languages, and 4 process contexts. Below is the actual chain with concrete types, file paths, and serialization points at each boundary.

### 1.1 The Full Chain

```
LLM (Claude/GPT)
  |  stdio (JSON-RPC 2.0 / newline-delimited JSON)
  v
Go MCP Server (cmd/dev-console/)
  |  HTTP POST /mcp (JSON-RPC 2.0 over HTTP)
  |  OR stdio (JSON-RPC 2.0 directly)
  v
Go HTTP Server (internal/capture/)
  |  HTTP POST /sync, /pending-queries, /query-result, etc.
  v
Chrome Extension Background (src/background/)
  |  chrome.runtime.sendMessage / chrome.tabs.sendMessage
  v
Chrome Extension Content Script (src/content/)
  |  window.postMessage (page <-> content script)
  v
Chrome Extension Inject Script (src/inject/)
  |  Intercepts browser APIs (console, fetch, WebSocket, etc.)
  v
[Back up the chain: data flows from inject -> content -> background -> HTTP -> Go -> stdio -> LLM]
```

### 1.2 Boundary 1: LLM <-> Go MCP Server (stdio / JSON-RPC 2.0)

**Protocol:** JSON-RPC 2.0 over stdio (newline-delimited) or HTTP POST `/mcp`

**Go types (request side):**
- `/Users/brenn/dev/gasoline/cmd/dev-console/types.go:8` - `JSONRPCRequest`
  ```go
  type JSONRPCRequest struct {
      JSONRPC  string          `json:"jsonrpc"`
      ID       any             `json:"id"`
      Method   string          `json:"method"`
      Params   json.RawMessage `json:"params,omitempty"`
      ClientID string          `json:"-"`
  }
  ```

**Go types (response side):**
- `/Users/brenn/dev/gasoline/cmd/dev-console/types.go:18` - `JSONRPCResponse`
  ```go
  type JSONRPCResponse struct {
      JSONRPC string          `json:"jsonrpc"`
      ID      any             `json:"id"`
      Result  json.RawMessage `json:"result,omitempty"`
      Error   *JSONRPCError   `json:"error,omitempty"`
  }
  ```

**Tool call payload:**
- `/Users/brenn/dev/gasoline/cmd/dev-console/tools_core.go:38` - `MCPToolResult`
  ```go
  type MCPToolResult struct {
      Content []MCPContentBlock `json:"content"`
      IsError bool              `json:"isError"` // SPEC:MCP - camelCase
  }
  ```

**Tool schemas (the contract between LLM and server):**
- `/Users/brenn/dev/gasoline/cmd/dev-console/tools_schema.go:8` - `ToolsList()`
- Schemas are defined as `map[string]any` -- no typed struct, no shared schema file.
- Tool argument schemas use snake_case field names (e.g., `what`, `after_cursor`, `min_level`).

**Serialization:** `json.Marshal` / `json.Unmarshal` with `encoding/json`.

**Key insight:** The tool schemas in `tools_schema.go` are the only contract the LLM sees. They are hand-coded `map[string]any` literals, not derived from any typed Go struct. The actual argument parsing happens in each tool handler (e.g., `toolObserve`, `toolAnalyze`) with separate local struct definitions. There is no compile-time guarantee that the schema matches the handler's expected fields.

### 1.3 Boundary 2: Go MCP Handler <-> Go HTTP Server (internal dispatch)

**Protocol:** In-process Go function calls.

**Types:**
- Handler receives `JSONRPCRequest` -> dispatches to `ToolHandler.HandleToolCall()`
  (`/Users/brenn/dev/gasoline/cmd/dev-console/tools_core.go:272`)
- Tool arguments are `json.RawMessage` (opaque bytes) -> each tool handler unmarshals into a local struct.
- Response is `JSONRPCResponse` with `Result` as `json.RawMessage` containing `MCPToolResult`.

**Example: observe tool argument parsing pattern:**
Each tool function (e.g., `toolObserve`) defines its own local argument struct:
```go
var args struct {
    What    string `json:"what"`
    Limit   int    `json:"limit"`
    // ... more fields
}
json.Unmarshal(rawArgs, &args)
```

**Key insight:** The argument struct is defined inline in the handler function, not shared with the schema definition in `tools_schema.go`. If someone adds a field to the schema but forgets to add it to the handler's struct, there is no compile-time error -- the field is silently ignored.

### 1.4 Boundary 3: Go Server <-> Chrome Extension (HTTP)

**Protocol:** HTTP POST/GET with JSON bodies.

**Server endpoints** (`/Users/brenn/dev/gasoline/cmd/dev-console/server_routes.go:542`):
| Endpoint | Direction | Go Handler | TS Caller |
|----------|-----------|------------|-----------|
| `POST /sync` | Extension -> Server | `capture.HandleSync` | `SyncClient.doSync()` |
| `POST /logs` | Extension -> Server | `server.handleLogs` | `sendLogsToServer()` |
| `POST /network-bodies` | Extension -> Server | `capture.HandleNetworkBodies` | `sendNetworkBodiesToServer()` |
| `POST /network-waterfall` | Extension -> Server | `capture.HandleNetworkWaterfall` | `sendNetworkWaterfallToServer()` |
| `POST /enhanced-actions` | Extension -> Server | `capture.HandleEnhancedActions` | `sendEnhancedActionsToServer()` |
| `POST /websocket-events` | Extension -> Server | `capture.HandleWebSocketEvents` | `sendWSEventsToServer()` |
| `POST /performance-snapshots` | Extension -> Server | `capture.HandlePerformanceSnapshots` | `sendPerformanceSnapshotsToServer()` |
| `POST /query-result` | Extension -> Server | `capture.HandleQueryResult` | `postQueryResult()` / `postAsyncCommandResult()` |
| `POST /screenshots` | Extension -> Server | `server.handleScreenshot` | `captureScreenshot()` |

**The `/sync` contract pair (most critical):**

Go request type (`/Users/brenn/dev/gasoline/internal/capture/sync.go:24`):
```go
type SyncRequest struct {
    SessionID        string              `json:"session_id"`
    ExtensionVersion string              `json:"extension_version,omitempty"`
    Settings         *SyncSettings       `json:"settings,omitempty"`
    ExtensionLogs    []ExtensionLog      `json:"extension_logs,omitempty"`
    LastCommandAck   string              `json:"last_command_ack,omitempty"`
    CommandResults   []SyncCommandResult `json:"command_results,omitempty"`
}
```

TypeScript request type (`/Users/brenn/dev/gasoline/src/background/sync-client.ts:45`):
```typescript
interface SyncRequest {
    session_id: string
    extension_version?: string
    settings?: SyncSettings
    extension_logs?: SyncExtensionLog[]
    last_command_ack?: string
    command_results?: SyncCommandResult[]
}
```

Go response type (`/Users/brenn/dev/gasoline/internal/capture/sync.go:67`):
```go
type SyncResponse struct {
    Ack              bool              `json:"ack"`
    Commands         []SyncCommand     `json:"commands"`
    NextPollMs       int               `json:"next_poll_ms"`
    ServerTime       string            `json:"server_time"`
    ServerVersion    string            `json:"server_version,omitempty"`
    CaptureOverrides map[string]string `json:"capture_overrides"`
}
```

TypeScript response type (`/Users/brenn/dev/gasoline/src/background/sync-client.ts:63`):
```typescript
interface SyncResponse {
    ack: boolean
    commands: SyncCommand[]
    next_poll_ms: number
    server_time: string
    server_version?: string
    capture_overrides?: Record<string, string>
}
```

**Serialization:** Go uses `json:"field_name"` struct tags. TypeScript uses property names matching the JSON wire format.

### 1.5 Boundary 4: Background <-> Content Script (chrome.runtime messaging)

**Protocol:** `chrome.runtime.sendMessage()` / `chrome.tabs.sendMessage()`

**Types** (`/Users/brenn/dev/gasoline/src/types/runtime-messages.ts`):
- `BackgroundMessage` (line 203): discriminated union of 20+ message types, discriminated on `type` field
- `ContentMessage` (line 387): discriminated union of 12+ message types, discriminated on `type` field

**Example content-bound message:**
```typescript
interface HighlightMessage {
    readonly type: 'GASOLINE_HIGHLIGHT'
    readonly params: {
        readonly selector: string
        readonly duration_ms?: number
    }
}
```

**Serialization:** Chrome serializes/deserializes using structured clone algorithm (supports more than JSON but constrained by Chrome's API).

### 1.6 Boundary 5: Content Script <-> Inject Script (window.postMessage)

**Protocol:** `window.postMessage()` with origin checking.

**Types** (`/Users/brenn/dev/gasoline/src/types/runtime-messages.ts:411`):
```typescript
type PageMessageType =
    | 'GASOLINE_LOG'
    | 'GASOLINE_WS'
    | 'GASOLINE_NETWORK_BODY'
    | 'GASOLINE_ENHANCED_ACTION'
    | 'GASOLINE_PERFORMANCE_SNAPSHOT'
    | 'GASOLINE_HIGHLIGHT_RESPONSE'
    | 'GASOLINE_EXECUTE_JS_RESULT'
    // ... etc
```

**Serialization:** `window.postMessage` uses structured clone. Data is serialized to JSON-compatible objects at the inject script level.

### 1.7 Cross-Language Type Mismatches Found

**Mismatch 1: EnhancedAction timestamp field**

TypeScript (`/Users/brenn/dev/gasoline/src/types/actions.ts:38`):
```typescript
readonly ts: string  // field name: "ts", type: string
```

Go (`/Users/brenn/dev/gasoline/internal/capture/enhanced-actions-types.go:11`):
```go
Timestamp int64 `json:"timestamp"`  // field name: "timestamp", type: int64
```

The TypeScript type uses `ts` (string) while the Go type uses `timestamp` (int64). These are completely different JSON field names AND types. This works because the Go handler uses `json.RawMessage` and `map[string]any` for flexible deserialization rather than the typed struct, but it means the Go struct type does NOT accurately represent what the extension actually sends.

**Mismatch 2: WaterfallEntry structure**

TypeScript (`/Users/brenn/dev/gasoline/src/types/network.ts:20`):
```typescript
interface WaterfallEntry {
    readonly url: string
    readonly initiatorType: string  // camelCase
    readonly startTime: number      // camelCase
    readonly duration: number
    readonly phases: WaterfallPhases  // nested object
    readonly transferSize: number    // camelCase
    readonly encodedBodySize: number // camelCase
    readonly decodedBodySize: number // camelCase
    readonly cached?: boolean
}
```

Go (`/Users/brenn/dev/gasoline/internal/capture/network-types.go:14`):
```go
type NetworkWaterfallEntry struct {
    Name            string    `json:"name"`
    URL             string    `json:"url"`
    InitiatorType   string    `json:"initiator_type"`    // snake_case
    Duration        float64   `json:"duration"`
    StartTime       float64   `json:"start_time"`        // snake_case
    TransferSize    int       `json:"transfer_size"`     // snake_case
    DecodedBodySize int       `json:"decoded_body_size"` // snake_case
    EncodedBodySize int       `json:"encoded_body_size"` // snake_case
    // No "phases" field, no "cached" field
    // Extra fields: Name, FetchStart, ResponseEnd, PageURL, Timestamp
}
```

The TypeScript type uses camelCase property names (`initiatorType`, `startTime`), but the Go type uses snake_case JSON tags (`initiator_type`, `start_time`). Additionally, the TypeScript type has a `phases` field that the Go type lacks entirely, and the Go type has fields (`Name`, `FetchStart`, `ResponseEnd`, `PageURL`, `Timestamp`) that the TypeScript type does not include. These are different types describing what is supposed to be the same data.

**Mismatch 3: NetworkBodyPayload field naming**

TypeScript (`/Users/brenn/dev/gasoline/src/types/network.ts:45`):
```typescript
interface NetworkBodyPayload {
    readonly url: string
    readonly method: string
    readonly status: number
    readonly contentType: string      // camelCase
    readonly requestBody?: string     // camelCase
    readonly responseBody?: string    // camelCase
    readonly duration: number
    readonly tabId?: number           // camelCase
}
```

Go (`/Users/brenn/dev/gasoline/internal/types/network.go:137`):
```go
type NetworkBody struct {
    Timestamp         string            `json:"ts,omitempty"`
    Method            string            `json:"method"`
    URL               string            `json:"url"`
    Status            int               `json:"status"`
    RequestBody       string            `json:"request_body,omitempty"`   // snake_case
    ResponseBody      string            `json:"response_body,omitempty"` // snake_case
    ContentType       string            `json:"content_type,omitempty"`  // snake_case
    Duration          int               `json:"duration,omitempty"`
    // Go has extra fields: RequestTruncated, ResponseTruncated, ResponseHeaders, etc.
}
```

The TypeScript type uses camelCase (`contentType`, `requestBody`, `responseBody`, `tabId`) but the Go JSON tags use snake_case (`content_type`, `request_body`, `response_body`, `tab_id`). If the extension sends `{"contentType": "application/json"}`, Go's JSON decoder will NOT populate the `ContentType` field because the tag is `content_type`. **This is a potential data loss bug** -- unless there is runtime conversion happening elsewhere (e.g., the TypeScript batchers convert camelCase to snake_case before sending).

**Mismatch 4: WebSocketEvent field mapping**

TypeScript (`/Users/brenn/dev/gasoline/src/types/websocket.ts:19`):
```typescript
interface WebSocketEvent {
    readonly type: WebSocketEventType  // 'open' | 'close' | 'error' | 'message'
    readonly url: string
    readonly ts: string
    readonly connectionId?: string     // camelCase
    readonly data?: string
    readonly size?: number
    readonly direction?: 'sent' | 'received'
    readonly code?: number
    readonly reason?: string
}
```

Go (`/Users/brenn/dev/gasoline/internal/capture/websocket-types.go:13`):
```go
type WebSocketEvent struct {
    Timestamp        string        `json:"ts,omitempty"`
    Type             string        `json:"type,omitempty"`
    Event            string        `json:"event"`
    ID               string        `json:"id"`         // "id" vs "connectionId"
    URL              string        `json:"url,omitempty"`
    Direction        string        `json:"direction,omitempty"`
    Data             string        `json:"data,omitempty"`
    Size             int           `json:"size,omitempty"`
    CloseCode        int           `json:"code,omitempty"`
    CloseReason      string        `json:"reason,omitempty"`
    // Go has extra fields: Sampled, BinaryFormat, FormatConfidence, TabId, TestIDs
}
```

The TypeScript type uses `connectionId` (camelCase) while Go uses `id` (different name entirely). Go also has an `Event` field that TypeScript lacks. The TypeScript type defines `direction` values as `'sent' | 'received'` while Go has no such constraint on the type level.

**Mismatch 5: Screenshot upload field names**

The extension's `captureScreenshot()` function (`/Users/brenn/dev/gasoline/src/background/communication.ts:143`) sends:
```typescript
body: JSON.stringify({
    dataUrl,     // camelCase
    url: tab.url,
    errorId: relatedErrorId || '',     // camelCase
    errorType: errorType || ''         // camelCase
})
```

But the Go handler (`/Users/brenn/dev/gasoline/cmd/dev-console/server_routes.go:149`) expects:
```go
var body struct {
    DataURL       string `json:"data_url"`       // snake_case
    URL           string `json:"url"`
    CorrelationID string `json:"correlation_id"`  // different field name entirely
    QueryID       string `json:"query_id"`
}
```

The TS sends `dataUrl` but Go expects `data_url`. The TS sends `errorId`/`errorType` but Go expects `correlation_id`/`query_id`. These are entirely different contracts. However, the `pending-queries.ts:242` does use `data_url` snake_case, suggesting there are two different screenshot upload paths with different field naming conventions.

---

## 2. What Stricture Would Catch Today

### 2.1 ARCH rules (would find significant issues)

**ARCH-max-file-lines:** Many files in Gasoline exceed 800 LOC. The Go capture package and tool handler files are large. Stricture would flag these.

**ARCH-import-boundary:** The `src/content/` files import from `src/types/` (shared types) but never from `src/background/` -- this is correct MV3 isolation. Stricture could enforce this as a rule.

**ARCH-no-circular-deps:** The TypeScript types are cleanly separated into modules (`src/types/telemetry.ts`, `src/types/network.ts`, etc.) with a facade re-export in `src/types/messages.ts`. No circular dependencies exist. Stricture would confirm this.

### 2.2 CONV rules (would find issues)

**CONV-file-header:** The Go files consistently use `// filename.go -- Purpose.` headers. The TypeScript files use `@fileoverview` JSDoc headers. Stricture could enforce both patterns per language.

**CONV-error-format:** Error messages in Go follow a pattern but not always the structured `{OPERATION}: {ROOT_CAUSE}. {RECOVERY_ACTION}` format. Many use `fmt.Errorf("verb: %w", err)` which is close but not exactly the prescribed format.

**CONV-file-naming:** Go files use snake_case (`server_routes.go`), TypeScript files use kebab-case (`sync-client.ts`). Stricture could enforce per-language.

### 2.3 TQ rules (would find issues in tests)

**TQ-no-shallow-assertions:** Many Go tests use the pattern `if err != nil { t.Fatal(err) }` without checking the error type or message. Stricture would flag these.

**TQ-return-type-verified:** The `SyncResponse` type has 6 fields, but many sync tests only check `Ack` and `Commands`. Stricture would flag incomplete field coverage.

**TQ-error-path-coverage:** The `HandleSync` handler has 3 error exits (method not allowed, max body exceeded, invalid JSON) but not all are tested in every test file. Stricture would flag uncovered error paths.

### 2.4 CTR rules (partially effective)

**CTR-json-tag-match:** Within Go-only contract pairs, Stricture would catch mismatched JSON tags between types in different packages. For example, the `capture.NetworkWaterfallEntry` and `types.NetworkWaterfallEntry` are duplicated -- Stricture's `CTR-shared-type-sync` rule would detect the duplicated type definitions.

**CTR-request-shape / CTR-response-shape:** Within Go-only code, Stricture could detect if a Go HTTP client sends a struct with different JSON tags than what the server handler expects. But this would miss the cross-language boundary.

---

## 3. What Stricture Would Miss

### 3.1 Cross-Language Contract Mismatches (CRITICAL GAP)

**Stricture's current design processes each language separately through its LanguageAdapter.** The Go adapter produces `UnifiedFileModel` for Go files. The TypeScript adapter produces `UnifiedFileModel` for TypeScript files. But **no rule crosses the language boundary to compare them.**

**Concrete example -- the WaterfallEntry mismatch:**

The TypeScript `WaterfallEntry` (`src/types/network.ts:20`) uses camelCase property names:
```typescript
readonly initiatorType: string
readonly startTime: number
readonly transferSize: number
```

The Go `NetworkWaterfallEntry` (`internal/capture/network-types.go:14`) uses snake_case JSON tags:
```go
InitiatorType string `json:"initiator_type"`
StartTime     float64 `json:"start_time"`
TransferSize  int     `json:"transfer_size"`
```

When the extension sends `{"initiatorType": "fetch"}` over HTTP, Go's `json.Unmarshal` looks for `"initiator_type"` in the JSON. The `InitiatorType` field stays empty. **This is a silent data loss bug that no existing linter catches.** Stricture's `CTR-json-tag-match` rule only compares Go struct tags against other Go struct tags.

**Scale of the gap:** I found at least 5 cross-language field naming mismatches in the Gasoline codebase (documented in Section 1.7). Each of these would require a human code reviewer to manually compare Go JSON tags against TypeScript property names. This is exactly the kind of tedious, error-prone task a linter should automate.

### 3.2 Schema-to-Handler Conformance (NO RULE EXISTS)

The MCP tool schemas in `tools_schema.go` are hand-coded `map[string]any` literals:
```go
"what": map[string]any{
    "type": "string",
    "enum": []string{"errors", "logs", "network_waterfall", ...},
},
```

The actual handler parses arguments into a local struct:
```go
var args struct {
    What string `json:"what"`
    // ...
}
```

**What could go wrong:**
1. Schema declares a field but handler's struct doesn't include it -- field is silently ignored
2. Handler's struct expects a field not in the schema -- the LLM will never send it
3. Schema says `"type": "string"` but handler's struct uses `int` -- parse error at runtime
4. Schema `enum` values don't match handler's `switch` cases -- unhandled enum variants

**No Stricture rule addresses this.** This is a contract between a JSON Schema (declared as a Go `map[string]any`) and a Go struct used to unmarshal that same JSON. It is not an HTTP client/server pair -- it is a schema/implementation pair.

### 3.3 Chrome Extension Message Contracts (NO RULE EXISTS)

The Chrome extension uses `chrome.runtime.sendMessage()` and `window.postMessage()` for inter-context communication. These are typed in TypeScript:

```typescript
// Background -> Content
type ContentMessage =
    | ContentPingMessage
    | HighlightMessage
    | ExecuteJsMessage
    // ...

// Content -> Background
type BackgroundMessage =
    | GetTabIdMessage
    | WsEventMessage
    | EnhancedActionMessage
    // ...
```

But the actual message dispatch uses `type` string discrimination:
```typescript
if (message.type === 'GASOLINE_HIGHLIGHT') {
    // Handle highlight
}
```

**What could go wrong:**
1. A message type string is misspelled in the sender vs receiver
2. The payload shape doesn't match what the receiver destructures
3. A new message type is added to the union but no handler exists
4. The handler accesses a field that the sender doesn't include

Stricture's `CTR-*` rules only detect HTTP contract pairs. Chrome extension messaging is a completely different protocol with no URL/method pattern to match on.

### 3.4 Multi-Hop Data Transformation Fidelity (NO RULE EXISTS)

Data transforms at each hop. A console log entry starts as:
1. **Inject script** (page context): `console.log("hello")` intercepted, wrapped as `{ type: 'GASOLINE_LOG', level: 'log', args: ['hello'], ts: '...' }`
2. **Content script**: forwarded via `window.postMessage` -> received, forwarded via `chrome.runtime.sendMessage`
3. **Background**: batched into `LogEntry[]`, sent via `fetch('/logs', { body: JSON.stringify({ entries }) })`
4. **Go server**: parsed as `[]LogEntry` (type alias for `map[string]any`)
5. **MCP response**: serialized back through `MCPToolResult.Content[0].Text` as a formatted string

At step 4, the Go type is `map[string]any` -- all type information is lost. The Go server stores logs as opaque JSON. When the MCP `observe` tool returns them, it re-serializes whatever was stored. If the extension changes a field name (e.g., `ts` to `timestamp`), the Go server doesn't notice -- it stores whatever comes in. But if the LLM expects `ts` and gets `timestamp`, behavior changes silently.

**No Stricture rule tracks data through multiple serialization hops.** The tool sees each file in isolation.

### 3.5 Protocol Envelope Consistency (NO RULE EXISTS)

The data has nested envelopes:
```
JSON-RPC 2.0 envelope:
  { "jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {
      "name": "observe",
      "arguments": {                      // MCP tool call envelope
          "what": "logs"                   // Gasoline-specific payload
      }
  }}
```

The response has similar nesting:
```
JSON-RPC 2.0 envelope:
  { "jsonrpc": "2.0", "id": 1, "result": {
      "content": [{                        // MCP tool result envelope
          "type": "text",
          "text": "..."                    // Gasoline-specific response
      }],
      "isError": false
  }}
```

Validation at each layer:
- JSON-RPC 2.0: `jsonrpc` must be `"2.0"`, `id` must match request, exactly one of `result`/`error`
- MCP: `content` must be array of `{type, text}` blocks, `isError` is boolean
- Gasoline: response text contains the actual data (formatted as text or JSON string within text)

**No Stricture rule validates envelope nesting.** The tool call parameters are `json.RawMessage` at each layer -- opaque bytes that bypass type checking.

### 3.6 Serialization Fidelity Across Type Systems (NO RULE EXISTS)

Several type mismatches between Go and TypeScript are not just naming but fundamental type differences:

| Data | TypeScript Type | Go Type | JSON Wire | Potential Issue |
|------|----------------|---------|-----------|-----------------|
| Timestamp | `string` (ISO 8601) | `time.Time` | `"2024-01-01T00:00:00Z"` | Go's `time.Time` marshals with nanosecond precision; TS `Date.toISOString()` has millisecond precision |
| Enhanced Action timestamp | `string` (`ts`) | `int64` (`json:"timestamp"`) | Different field name AND type | Complete mismatch |
| Duration | `number` (float) | `int` or `float64` | Varies | Go `int` truncates fractional milliseconds |
| Tab ID | `number` | `int` | `42` | JavaScript `number` is float64; Go `int` is platform-dependent width |
| Selectors | `SelectorStrategies` (typed interface) | `map[string]any` | `{...}` | Go loses all type information |

---

## 4. Proposed New Capabilities

### 4.1 New Rule Category: Cross-Language Contract (XLC)

A new rule category for enforcing contracts across language boundaries.

#### XLC-json-field-match

**Purpose:** Compare JSON field names between a Go struct with `json:"..."` tags and a TypeScript interface/type whose properties represent the same JSON wire format.

**What it catches:**
```go
// Go: internal/capture/network-types.go
type NetworkWaterfallEntry struct {
    InitiatorType string `json:"initiator_type"`  // snake_case
}
```
```typescript
// TypeScript: src/types/network.ts
interface WaterfallEntry {
    readonly initiatorType: string  // camelCase -- VIOLATION
}
```

**Detection:**
1. User declares contract pairs in config (Go file <-> TS file)
2. For each pair, extract Go struct JSON tag names and TS interface property names
3. Compare: mismatched names where fuzzy match suggests they represent the same field -> ERROR

**Config:**
```yaml
XLC-json-field-match:
  - error
  - pairs:
      - go: "internal/capture/network-types.go::NetworkWaterfallEntry"
        ts: "src/types/network.ts::WaterfallEntry"
      - go: "internal/capture/sync.go::SyncRequest"
        ts: "src/background/sync-client.ts::SyncRequest"
      - go: "internal/capture/sync.go::SyncResponse"
        ts: "src/background/sync-client.ts::SyncResponse"
    convention: snake_case  # Enforce that both sides use snake_case on the wire
```

#### XLC-type-fidelity

**Purpose:** Detect type mismatches across language boundaries where the same JSON field has different types in Go and TypeScript.

**What it catches:**
```go
// Go
Timestamp int64 `json:"timestamp"`
```
```typescript
// TypeScript
readonly ts: string  // Different field name AND different type
```

**Detection:**
1. For each declared contract pair, compare field types
2. Flag when: Go `int64` paired with TS `string`, Go `time.Time` paired with TS `number`, etc.
3. Provide a compatibility matrix: which Go types are compatible with which TS types

#### XLC-field-coverage

**Purpose:** Detect fields present on one side of a cross-language contract but missing on the other.

**What it catches:**
- Go `NetworkWaterfallEntry` has `Name`, `FetchStart`, `ResponseEnd`, `PageURL`, `Timestamp` fields that TypeScript `WaterfallEntry` lacks
- TypeScript `WaterfallEntry` has `phases` and `cached` fields that Go `NetworkWaterfallEntry` lacks

### 4.2 New Rule: SCHEMA-handler-conformance

**Purpose:** Verify that a JSON Schema definition (expressed as code or a spec file) matches the handler that processes data conforming to that schema.

**What it catches in Gasoline:**
```go
// Schema: tools_schema.go line 18
"what": map[string]any{
    "type": "string",
    "enum": []string{"errors", "logs", "network_waterfall", ...},
}

// Handler: toolObserve function
var args struct {
    What string `json:"what"`
}
switch args.What {
case "errors":
case "logs":
// Missing cases that schema declares: "extension_logs", "recordings", etc.
```

**Detection:**
1. Parse schema definition (map literal, JSON Schema file, or OpenAPI spec)
2. Parse handler's argument struct and switch/if-else dispatch
3. Compare: schema fields vs struct fields, schema enum values vs switch cases
4. Flag: missing struct fields, unhandled enum values, type mismatches

**Config:**
```yaml
SCHEMA-handler-conformance:
  - error
  - schemas:
      - definition: "cmd/dev-console/tools_schema.go"
        handler: "cmd/dev-console/tools_observe.go"
        type: "mcp-tool-schema"  # Parser knows how to extract from map literals
```

### 4.3 New Rule: MSG-contract-match

**Purpose:** Enforce that Chrome extension message senders and receivers agree on message shape.

**What it catches:**
```typescript
// Sender (background -> content)
chrome.tabs.sendMessage(tabId, {
    type: 'GASOLINE_HIGHLIGHT',
    params: { selector: '.foo', duration_ms: 5000 }
})

// Receiver (content script)
if (message.type === 'GASOLINE_HIGHLIGHT') {
    const { selector, durationMs } = message.params  // VIOLATION: "durationMs" vs "duration_ms"
}
```

**Detection:**
1. Parse message union types from TypeScript
2. For each message type, find all send sites and receive sites
3. Compare the fields accessed in send vs receive
4. Flag mismatches

### 4.4 Enhancement to CTR-request-shape: Cross-Language HTTP Contracts

Extend the existing `CTR-request-shape` rule to work across languages by adding cross-language contract pair declaration:

```yaml
contracts:
  cross_language_pairs:
    - server:
        language: go
        file: "internal/capture/sync.go"
        handler: "HandleSync"
        request_type: "SyncRequest"
        response_type: "SyncResponse"
      client:
        language: typescript
        file: "src/background/sync-client.ts"
        caller: "doSync"
        request_type: "SyncRequest"
        response_type: "SyncResponse"
      endpoint: "POST /sync"
```

---

## 5. Dataflow Config Proposal

### 5.1 The `.stricture.yml` `dataflow` Section

```yaml
# .stricture.yml for Gasoline

dataflow:
  # Declare the data flow chain
  chain:
    name: "MCP Tool Call"
    hops:
      - id: stdio-mcp
        protocol: json-rpc-2.0
        transport: stdio
        from: "LLM (external)"
        to: "cmd/dev-console/handler.go::HandleHTTP"
        request_type: "cmd/dev-console/types.go::JSONRPCRequest"
        response_type: "cmd/dev-console/types.go::JSONRPCResponse"

      - id: mcp-dispatch
        protocol: internal
        transport: function-call
        from: "cmd/dev-console/handler.go::handleToolsCall"
        to: "cmd/dev-console/tools_core.go::HandleToolCall"
        schema: "cmd/dev-console/tools_schema.go::ToolsList"
        note: "Schema defined as map[string]any; handler uses local struct"

      - id: server-extension-sync
        protocol: http
        transport: "POST /sync"
        from: "src/background/sync-client.ts::SyncClient.doSync"
        to: "internal/capture/sync.go::HandleSync"
        request:
          go: "internal/capture/sync.go::SyncRequest"
          ts: "src/background/sync-client.ts::SyncRequest"
        response:
          go: "internal/capture/sync.go::SyncResponse"
          ts: "src/background/sync-client.ts::SyncResponse"

      - id: extension-data-ingest
        protocol: http
        transport: "POST /network-bodies"
        from: "src/background/server.ts::sendNetworkBodiesToServer"
        to: "internal/capture/handlers.go::HandleNetworkBodies"
        request:
          go: "internal/types/network.go::NetworkBody"
          ts: "src/types/network.ts::NetworkBodyPayload"

      - id: extension-data-waterfall
        protocol: http
        transport: "POST /network-waterfall"
        from: "src/background/server.ts::sendNetworkWaterfallToServer"
        to: "internal/capture/handlers.go::HandleNetworkWaterfall"
        request:
          go: "internal/capture/network-types.go::NetworkWaterfallEntry"
          ts: "src/types/network.ts::WaterfallEntry"

      - id: bg-content-messages
        protocol: chrome-runtime-message
        transport: "chrome.tabs.sendMessage"
        from: "src/background/*.ts"
        to: "src/content/runtime-message-listener.ts"
        message_union: "src/types/runtime-messages.ts::ContentMessage"

      - id: content-inject-messages
        protocol: window-postmessage
        transport: "window.postMessage"
        from: "src/content/*.ts"
        to: "src/inject/*.ts"
        message_types: "src/types/runtime-messages.ts::PageMessageType"

  # Cross-language type pairs (compared field-by-field)
  type_pairs:
    - go: "internal/capture/sync.go::SyncRequest"
      ts: "src/background/sync-client.ts::SyncRequest"
    - go: "internal/capture/sync.go::SyncResponse"
      ts: "src/background/sync-client.ts::SyncResponse"
    - go: "internal/capture/sync.go::SyncCommand"
      ts: "src/background/sync-client.ts::SyncCommand"
    - go: "internal/capture/sync.go::SyncCommandResult"
      ts: "src/background/sync-client.ts::SyncCommandResult"
    - go: "internal/capture/sync.go::SyncSettings"
      ts: "src/background/sync-client.ts::SyncSettings"
    - go: "internal/types/network.go::NetworkBody"
      ts: "src/types/network.ts::NetworkBodyPayload"
    - go: "internal/capture/network-types.go::NetworkWaterfallEntry"
      ts: "src/types/network.ts::WaterfallEntry"
    - go: "internal/capture/enhanced-actions-types.go::EnhancedAction"
      ts: "src/types/actions.ts::EnhancedAction"
    - go: "internal/capture/websocket-types.go::WebSocketEvent"
      ts: "src/types/websocket.ts::WebSocketEvent"

  # JSON naming convention enforcement across all pairs
  convention:
    wire_format: snake_case
    go_fields: PascalCase  # Enforced by Go itself
    ts_properties: snake_case  # Must match wire format
    exceptions:
      - pattern: "SPEC:MCP"  # Allow camelCase for MCP spec fields
        convention: camelCase
```

### 5.2 What This Enables

With this config, Stricture can:

1. **Automatically compare Go JSON tags against TypeScript property names** for every declared type pair
2. **Flag camelCase/snake_case mismatches** like `initiatorType` vs `initiator_type`
3. **Detect missing fields** where one side has a field the other lacks
4. **Detect type incompatibilities** where Go uses `int64` and TypeScript uses `string` for the same wire field
5. **Validate message unions** by checking that every variant in a discriminated union has a corresponding handler
6. **Track data through hops** to ensure a field name that changes at one hop is updated at all subsequent hops
7. **Report violations with full context** including which hop in the chain is affected

---

## 6. Bigger Vision: Beyond Traditional Static Analysis

### 6.1 Dataflow Diagrams as Config Input

Stricture could accept Mermaid sequence diagrams (already common in tech specs) as an additional input:

```yaml
dataflow:
  diagram: "docs/features/mcp-persistent-server/architecture.md"
  diagram_format: mermaid
```

The linter parses the Mermaid diagram, extracts participants and message flows, and validates that:
- Every participant in the diagram maps to an actual file/module in the codebase
- Every message in the diagram maps to an actual function call, HTTP request, or message send
- Messages not in the diagram (undocumented data flows) are flagged as violations

This turns architecture documentation from a passive artifact into an active contract.

### 6.2 Runtime Trace Validation

Stricture could optionally ingest runtime traces (HTTP access logs, captured network traffic, test execution traces) and validate them against declared types:

```yaml
traces:
  - source: "test-output/sync-trace.har"
    validate_against:
      - go: "internal/capture/sync.go::SyncRequest"
      - go: "internal/capture/sync.go::SyncResponse"
```

This catches the most insidious bugs: where the declared types are wrong because the code works around them. If the HAR trace shows `{"initiatorType": "fetch"}` but the Go struct expects `"initiator_type"`, Stricture can flag that the runtime data doesn't match the declared type.

### 6.3 Schema Registry Integration

For systems with formal API specs (OpenAPI, protobuf, JSON Schema, AsyncAPI), Stricture could treat the spec file as the source of truth and validate both sides:

```yaml
specs:
  - path: "api/openapi.yaml"
    servers: ["internal/capture/handlers.go"]
    clients: ["src/background/server.ts"]
    validate: [request-shape, response-shape, status-codes, field-names]
```

For Gasoline specifically, the MCP tool schemas in `tools_schema.go` are essentially a programmatic JSON Schema. Stricture could parse them and validate conformance.

### 6.4 State Machine Verification

Gasoline has implicit state machines (connection state, circuit breaker state, recording state). Stricture could verify that declared state machines match actual state transitions in code:

```yaml
state_machines:
  - name: "SyncClient connection"
    file: "src/background/sync-client.ts::SyncClient"
    states: [disconnected, connected]
    transitions:
      - from: disconnected
        to: connected
        trigger: "onSuccess()"
      - from: connected
        to: disconnected
        trigger: "onFailure()"
    verify:
      - all_states_reachable: true
      - no_dead_states: true
      - every_transition_has_test: true
```

### 6.5 Cross-Repo Analysis

For microservice architectures, Stricture could analyze multiple repos together:

```yaml
# In a monorepo or multi-repo config
repos:
  - path: "../gasoline-server"
    language: go
  - path: "../gasoline-extension"
    language: typescript

contracts:
  cross_repo:
    - server_repo: "../gasoline-server"
      client_repo: "../gasoline-extension"
      endpoint: "POST /sync"
```

This is the logical extension of cross-language analysis to cross-repository analysis.

### 6.6 Protocol-Aware Linting

Different protocols have different constraints:
- **JSON-RPC 2.0**: `id` must match, `jsonrpc` must be `"2.0"`, notifications have no `id`
- **MCP**: tool results must have `content` array, `isError` boolean
- **Chrome messaging**: `sendResponse` must be called synchronously or return `true` for async
- **window.postMessage**: must include origin checking on receive side

Stricture could have protocol-specific rules:

```yaml
protocols:
  json-rpc-2.0:
    files: ["cmd/dev-console/handler.go"]
    rules:
      - notification-no-response: error    # Notifications must not receive responses
      - id-echo: error                     # Response ID must match request ID
      - error-codes: error                 # Error codes must be in valid range

  chrome-messaging:
    files: ["src/background/**", "src/content/**"]
    rules:
      - async-response-return-true: error  # If handler is async, must return true
      - message-type-exhaustive: error     # All union variants must have handlers
      - sender-validation: error           # Must validate sender before processing
```

### 6.7 The Meta-Insight

The fundamental limitation of every existing linter (ESLint, golangci-lint, Codacy, SonarQube) is that they operate on a single file or a single language. They verify syntax and local semantics. **The bugs that matter most in complex systems are at the boundaries** -- where data crosses process boundaries, language boundaries, serialization boundaries, and protocol boundaries.

Traditional linters verify that "this code compiles." Stricture (with TQ rules) verifies that "this code is tested." The proposed dataflow extensions would verify that **"this code agrees with the other code it talks to."**

This is the gap between "each service works in isolation" and "the system works." It is the gap that causes production incidents. And it is entirely automatable with static analysis -- if the linter knows about the dataflow.

---

## Appendix: Files Referenced

| File | Purpose |
|------|---------|
| `/Users/brenn/dev/gasoline/cmd/dev-console/types.go` | JSON-RPC request/response types |
| `/Users/brenn/dev/gasoline/cmd/dev-console/tools_core.go` | MCP tool result types, tool dispatch |
| `/Users/brenn/dev/gasoline/cmd/dev-console/tools_schema.go` | MCP tool schema definitions (LLM contract) |
| `/Users/brenn/dev/gasoline/cmd/dev-console/handler.go` | MCP protocol handler, HTTP transport |
| `/Users/brenn/dev/gasoline/cmd/dev-console/server_routes.go` | HTTP route setup, endpoint handlers |
| `/Users/brenn/dev/gasoline/cmd/dev-console/server_middleware.go` | CORS, origin validation, security |
| `/Users/brenn/dev/gasoline/internal/capture/sync.go` | Unified sync endpoint (Go server side) |
| `/Users/brenn/dev/gasoline/internal/capture/handlers.go` | Extension data ingestion handlers |
| `/Users/brenn/dev/gasoline/internal/capture/network-types.go` | Go network waterfall/body types |
| `/Users/brenn/dev/gasoline/internal/capture/enhanced-actions-types.go` | Go enhanced action types |
| `/Users/brenn/dev/gasoline/internal/capture/websocket-types.go` | Go WebSocket event types |
| `/Users/brenn/dev/gasoline/internal/types/network.go` | Canonical Go network/WS/action types |
| `/Users/brenn/dev/gasoline/src/types/messages.ts` | TypeScript message type facade |
| `/Users/brenn/dev/gasoline/src/types/runtime-messages.ts` | Chrome runtime message types |
| `/Users/brenn/dev/gasoline/src/types/network.ts` | TypeScript network types |
| `/Users/brenn/dev/gasoline/src/types/actions.ts` | TypeScript action types |
| `/Users/brenn/dev/gasoline/src/types/websocket.ts` | TypeScript WebSocket types |
| `/Users/brenn/dev/gasoline/src/types/telemetry.ts` | TypeScript log entry types |
| `/Users/brenn/dev/gasoline/src/types/performance.ts` | TypeScript performance types |
| `/Users/brenn/dev/gasoline/src/types/queries.ts` | TypeScript pending query types |
| `/Users/brenn/dev/gasoline/src/background/server.ts` | Extension -> server HTTP functions |
| `/Users/brenn/dev/gasoline/src/background/sync-client.ts` | Unified sync client (TS side) |
| `/Users/brenn/dev/gasoline/src/background/communication.ts` | Communication facade |
| `/Users/brenn/dev/gasoline/src/background/message-handlers.ts` | Background message routing |
| `/Users/brenn/dev/gasoline/src/content/runtime-message-listener.ts` | Content script message handler |
