# 07 -- OpenAI Chat Completions API

Stricture validation set for the OpenAI Chat Completions API. Tests streaming SSE parsing, token counting, nullable fields, model enum validation, and rate limit handling.

**Why included:** Streaming (Server-Sent Events), nullable response fields, token arithmetic, model lifecycle (deprecation races), and multi-format content arrays make OpenAI a uniquely challenging contract surface.

---

## Manifest Fragment

```yaml
contracts:
  - id: "openai-chat-completions"
    producer: openai
    consumers: [my-service]
    protocol: http
    auth:
      type: bearer
      header: Authorization
      format: "sk-*"
    endpoints:
      - path: "/v1/chat/completions"
        method: POST
        request:
          fields:
            model:        { type: enum, values: ["gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "gpt-3.5-turbo", "o1", "o1-mini", "o3-mini"], required: true }
            messages:     { type: array, items: { type: object, fields: { role: { type: enum, values: ["system", "user", "assistant", "tool"], required: true }, content: { type: "string|null|array", required: true }, tool_call_id: { type: string, required: false } } }, required: true, minItems: 1 }
            temperature:  { type: number, range: [0, 2], required: false, default: 1 }
            max_tokens:   { type: integer, range: [1, null], required: false, nullable: true }
            stream:       { type: boolean, required: false, default: false }
            tools:        { type: array, required: false }
            tool_choice:  { type: "string|object", required: false }
            n:            { type: integer, range: [1, 128], required: false, default: 1 }
            stop:         { type: "string|array", required: false, nullable: true }
        response:
          fields:
            id:           { type: string, format: "chatcmpl-*", required: true }
            object:       { type: enum, values: ["chat.completion"], required: true }
            created:      { type: integer, format: unix_timestamp, required: true }
            model:        { type: string, required: true }
            choices:      { type: array, items: { type: object, fields: { index: { type: integer, required: true }, message: { type: object, fields: { role: { type: enum, values: ["assistant"], required: true }, content: { type: "string|null", required: true }, tool_calls: { type: array, required: false } }, required: true }, finish_reason: { type: enum, values: ["stop", "length", "content_filter", "tool_calls"], required: true, nullable: true } } }, required: true }
            usage:        { type: object, fields: { prompt_tokens: { type: integer, range: [0, null], required: true }, completion_tokens: { type: integer, range: [0, null], required: true }, total_tokens: { type: integer, range: [0, null], required: true } }, required: true }
          streaming:
            protocol: sse
            prefix: "data: "
            terminator: "data: [DONE]"
            chunk_fields:
              id:         { type: string, format: "chatcmpl-*", required: true }
              object:     { type: enum, values: ["chat.completion.chunk"], required: true }
              choices:    { type: array, items: { type: object, fields: { index: { type: integer }, delta: { type: object, fields: { role: { type: string, required: false }, content: { type: "string|null", required: false }, tool_calls: { type: array, required: false } } }, finish_reason: { type: enum, values: ["stop", "length", "content_filter", "tool_calls"], nullable: true } } } }
        status_codes: [200, 400, 401, 403, 429, 500, 503]
        error_shape:
          fields:
            error:        { type: object, fields: { message: { type: string, required: true }, type: { type: enum, values: ["invalid_request_error", "authentication_error", "permission_error", "rate_limit_error", "server_error"], required: true }, param: { type: "string|null", required: true }, code: { type: "string|null", required: true } }, required: true }
        rate_limit_headers:
          - "x-ratelimit-limit-requests"
          - "x-ratelimit-limit-tokens"
          - "x-ratelimit-remaining-requests"
          - "x-ratelimit-remaining-tokens"
          - "x-ratelimit-reset-requests"
          - "x-ratelimit-reset-tokens"

  - id: "openai-embeddings"
    producer: openai
    consumers: [my-service]
    protocol: http
    endpoints:
      - path: "/v1/embeddings"
        method: POST
        request:
          fields:
            model:        { type: enum, values: ["text-embedding-3-small", "text-embedding-3-large", "text-embedding-ada-002"], required: true }
            input:        { type: "string|array", required: true }
            dimensions:   { type: integer, required: false }
            encoding_format: { type: enum, values: ["float", "base64"], required: false, default: "float" }
        response:
          fields:
            object:       { type: enum, values: ["list"], required: true }
            data:         { type: array, items: { type: object, fields: { object: { type: enum, values: ["embedding"] }, index: { type: integer }, embedding: { type: array } } }, required: true }
            model:        { type: string, required: true }
            usage:        { type: object, fields: { prompt_tokens: { type: integer }, total_tokens: { type: integer } }, required: true }
        status_codes: [200, 400, 401, 429, 500]

  - id: "openai-models"
    producer: openai
    consumers: [my-service]
    protocol: http
    endpoints:
      - path: "/v1/models"
        method: GET
        response:
          fields:
            object:       { type: enum, values: ["list"], required: true }
            data:         { type: array, items: { type: object, fields: { id: { type: string }, object: { type: enum, values: ["model"] }, created: { type: integer }, owned_by: { type: string } } }, required: true }
        status_codes: [200, 401, 500]
```

---

## PERFECT -- Zero Violations

This integration correctly handles all response shapes, streaming, nullable fields, all finish reasons, token counting as integers, error handling for every status code, and rate limit headers.

### Source File (`src/services/openai-client.ts`)

```typescript
// openai-client.ts -- OpenAI Chat Completions API client with streaming support.

interface ChatMessage {
  role: "system" | "user" | "assistant" | "tool";
  content: string | null | ContentPart[];
  tool_call_id?: string;
}

interface ContentPart {
  type: "text" | "image_url";
  text?: string;
  image_url?: { url: string; detail?: "auto" | "low" | "high" };
}

interface ToolCall {
  id: string;
  type: "function";
  function: { name: string; arguments: string };
}

interface ChatCompletionRequest {
  model: "gpt-4o" | "gpt-4o-mini" | "gpt-4-turbo" | "gpt-3.5-turbo" | "o1" | "o1-mini" | "o3-mini";
  messages: ChatMessage[];
  temperature?: number;
  max_tokens?: number | null;
  stream?: boolean;
  tools?: Tool[];
  tool_choice?: "none" | "auto" | "required" | { type: "function"; function: { name: string } };
  n?: number;
  stop?: string | string[] | null;
}

interface Tool {
  type: "function";
  function: {
    name: string;
    description?: string;
    parameters?: Record<string, unknown>;
  };
}

interface ChatCompletionChoice {
  index: number;
  message: {
    role: "assistant";
    content: string | null;
    tool_calls?: ToolCall[];
  };
  finish_reason: "stop" | "length" | "content_filter" | "tool_calls" | null;
}

interface ChatCompletionResponse {
  id: string;
  object: "chat.completion";
  created: number;
  model: string;
  choices: ChatCompletionChoice[];
  usage: {
    prompt_tokens: number;
    completion_tokens: number;
    total_tokens: number;
  };
}

interface ChatCompletionChunkDelta {
  role?: string;
  content?: string | null;
  tool_calls?: ToolCall[];
}

interface ChatCompletionChunk {
  id: string;
  object: "chat.completion.chunk";
  created: number;
  model: string;
  choices: Array<{
    index: number;
    delta: ChatCompletionChunkDelta;
    finish_reason: "stop" | "length" | "content_filter" | "tool_calls" | null;
  }>;
}

interface OpenAIError {
  error: {
    message: string;
    type: "invalid_request_error" | "authentication_error" | "permission_error" | "rate_limit_error" | "server_error";
    param: string | null;
    code: string | null;
  };
}

interface RateLimitInfo {
  limitRequests: number;
  limitTokens: number;
  remainingRequests: number;
  remainingTokens: number;
  resetRequests: string;
  resetTokens: string;
}

interface TokenAccumulator {
  promptTokens: number;
  completionTokens: number;
  totalTokens: number;
}

function validateApiKey(apiKey: string): void {
  if (!apiKey.startsWith("sk-")) {
    throw new Error("Invalid API key format: must start with 'sk-'");
  }
}

function validateTemperature(temperature: number | undefined): void {
  if (temperature !== undefined && (temperature < 0 || temperature > 2)) {
    throw new Error(`Temperature must be between 0 and 2, got ${temperature}`);
  }
}

function validateMessages(messages: ChatMessage[]): void {
  if (!Array.isArray(messages) || messages.length === 0) {
    throw new Error("Messages array must contain at least one message");
  }
  for (const msg of messages) {
    if (!["system", "user", "assistant", "tool"].includes(msg.role)) {
      throw new Error(`Invalid message role: ${msg.role}`);
    }
  }
}

function validateCompletionId(id: string): void {
  if (!id.startsWith("chatcmpl-")) {
    throw new Error(`Invalid completion ID format: expected chatcmpl-* prefix, got "${id}"`);
  }
}

function parseRateLimitHeaders(headers: Headers): RateLimitInfo {
  return {
    limitRequests: parseInt(headers.get("x-ratelimit-limit-requests") ?? "0", 10),
    limitTokens: parseInt(headers.get("x-ratelimit-limit-tokens") ?? "0", 10),
    remainingRequests: parseInt(headers.get("x-ratelimit-remaining-requests") ?? "0", 10),
    remainingTokens: parseInt(headers.get("x-ratelimit-remaining-tokens") ?? "0", 10),
    resetRequests: headers.get("x-ratelimit-reset-requests") ?? "",
    resetTokens: headers.get("x-ratelimit-reset-tokens") ?? "",
  };
}

export class OpenAIClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;
  private tokenAccumulator: TokenAccumulator;

  constructor(apiKey: string, baseUrl = "https://api.openai.com") {
    validateApiKey(apiKey);
    this.apiKey = apiKey;
    this.baseUrl = baseUrl;
    this.tokenAccumulator = { promptTokens: 0, completionTokens: 0, totalTokens: 0 };
  }

  async createChatCompletion(request: ChatCompletionRequest): Promise<ChatCompletionResponse> {
    validateMessages(request.messages);
    validateTemperature(request.temperature);

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/chat/completions`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${this.apiKey}`,
        },
        body: JSON.stringify({ ...request, stream: false }),
      });
    } catch (err) {
      throw new Error(`Network error calling OpenAI: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody: OpenAIError = await response.json();
      const errType = errorBody.error.type;
      const errMsg = errorBody.error.message;

      switch (response.status) {
        case 400:
          throw new Error(`Invalid request: ${errMsg} (param: ${errorBody.error.param})`);
        case 401:
          throw new Error(`Authentication failed: ${errMsg}`);
        case 403:
          throw new Error(`Permission denied: ${errMsg}`);
        case 429: {
          const rateLimits = parseRateLimitHeaders(response.headers);
          throw new Error(`Rate limited (${errType}): ${errMsg}. Resets at: requests=${rateLimits.resetRequests}, tokens=${rateLimits.resetTokens}`);
        }
        case 500:
          throw new Error(`OpenAI server error: ${errMsg}`);
        case 503:
          throw new Error(`OpenAI service unavailable: ${errMsg}. Retry after backoff.`);
        default:
          throw new Error(`Unexpected status ${response.status}: ${errMsg}`);
      }
    }

    const completion: ChatCompletionResponse = await response.json();

    validateCompletionId(completion.id);

    if (completion.object !== "chat.completion") {
      throw new Error(`Unexpected object type: ${completion.object}`);
    }

    if (typeof completion.created !== "number") {
      throw new Error(`Invalid created timestamp: expected number, got ${typeof completion.created}`);
    }

    // Integer accumulation -- no floating point
    this.tokenAccumulator.promptTokens += completion.usage.prompt_tokens;
    this.tokenAccumulator.completionTokens += completion.usage.completion_tokens;
    this.tokenAccumulator.totalTokens += completion.usage.total_tokens;

    for (const choice of completion.choices) {
      this.handleFinishReason(choice.finish_reason);

      // Safely handle nullable content (tool_calls produce null content)
      const contentLength = choice.message.content !== null ? choice.message.content.length : 0;
      if (choice.finish_reason === "length" && contentLength > 0) {
        console.warn(`Response truncated at ${contentLength} characters (finish_reason: length)`);
      }
    }

    return completion;
  }

  async *streamChatCompletion(request: ChatCompletionRequest): AsyncGenerator<ChatCompletionChunk, void, undefined> {
    validateMessages(request.messages);
    validateTemperature(request.temperature);

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/chat/completions`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${this.apiKey}`,
        },
        body: JSON.stringify({ ...request, stream: true }),
      });
    } catch (err) {
      throw new Error(`Network error calling OpenAI streaming: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody: OpenAIError = await response.json();
      switch (response.status) {
        case 429: {
          const rateLimits = parseRateLimitHeaders(response.headers);
          throw new Error(`Rate limited: ${errorBody.error.message}. Resets: ${rateLimits.resetRequests}`);
        }
        case 503:
          throw new Error(`Service unavailable: ${errorBody.error.message}`);
        default:
          throw new Error(`OpenAI error ${response.status}: ${errorBody.error.message}`);
      }
    }

    if (!response.body) {
      throw new Error("Response body is null -- streaming not supported by this environment");
    }

    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = "";

    try {
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split("\n");
        buffer = lines.pop() ?? "";

        for (const line of lines) {
          const trimmed = line.trim();
          if (trimmed === "") continue;
          if (trimmed === "data: [DONE]") return;
          if (!trimmed.startsWith("data: ")) continue;

          const json = trimmed.slice(6);
          const chunk: ChatCompletionChunk = JSON.parse(json);
          validateCompletionId(chunk.id);
          yield chunk;
        }
      }

      // Process any remaining buffer
      if (buffer.trim() !== "" && buffer.trim() !== "data: [DONE]") {
        if (buffer.trim().startsWith("data: ")) {
          const json = buffer.trim().slice(6);
          const chunk: ChatCompletionChunk = JSON.parse(json);
          yield chunk;
        }
      }
    } finally {
      reader.releaseLock();
    }
  }

  async listModels(): Promise<Array<{ id: string; object: "model"; created: number; owned_by: string }>> {
    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/models`, {
        method: "GET",
        headers: { "Authorization": `Bearer ${this.apiKey}` },
      });
    } catch (err) {
      throw new Error(`Network error listing models: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody: OpenAIError = await response.json();
      throw new Error(`Failed to list models (${response.status}): ${errorBody.error.message}`);
    }

    const body = await response.json();
    return body.data;
  }

  getTokenUsage(): TokenAccumulator {
    return { ...this.tokenAccumulator };
  }

  private handleFinishReason(reason: "stop" | "length" | "content_filter" | "tool_calls" | null): void {
    switch (reason) {
      case "stop":
        break;
      case "length":
        console.warn("Completion truncated due to max_tokens limit");
        break;
      case "content_filter":
        console.warn("Content was filtered by OpenAI safety system");
        break;
      case "tool_calls":
        break;
      case null:
        // Streaming chunks before the final chunk have null finish_reason
        break;
    }
  }
}
```

### Test File (`tests/services/openai-client.test.ts`)

```typescript
// openai-client.test.ts -- Tests for OpenAI Chat Completions client.

import { OpenAIClient } from "../../src/services/openai-client";

function mockFetchSuccess(body: unknown, headers?: Record<string, string>): void {
  global.fetch = jest.fn().mockResolvedValue({
    ok: true,
    status: 200,
    json: async () => body,
    headers: new Headers(headers ?? {}),
  });
}

function mockFetchError(status: number, errorBody: unknown, headers?: Record<string, string>): void {
  global.fetch = jest.fn().mockResolvedValue({
    ok: false,
    status,
    json: async () => errorBody,
    headers: new Headers(headers ?? {}),
  });
}

function mockFetchStream(chunks: string[]): void {
  const encoder = new TextEncoder();
  let chunkIndex = 0;
  const readable = new ReadableStream({
    pull(controller) {
      if (chunkIndex < chunks.length) {
        controller.enqueue(encoder.encode(chunks[chunkIndex]));
        chunkIndex++;
      } else {
        controller.close();
      }
    },
  });
  global.fetch = jest.fn().mockResolvedValue({
    ok: true,
    status: 200,
    body: readable,
    headers: new Headers({}),
  });
}

function mockFetchNetworkError(): void {
  global.fetch = jest.fn().mockRejectedValue(new Error("ECONNREFUSED"));
}

const VALID_COMPLETION: Record<string, unknown> = {
  id: "chatcmpl-abc123def456",
  object: "chat.completion",
  created: 1700000000,
  model: "gpt-4o",
  choices: [
    {
      index: 0,
      message: { role: "assistant", content: "Hello! How can I help?" },
      finish_reason: "stop",
    },
  ],
  usage: { prompt_tokens: 10, completion_tokens: 8, total_tokens: 18 },
};

const VALID_TOOL_CALL_COMPLETION: Record<string, unknown> = {
  id: "chatcmpl-tool789",
  object: "chat.completion",
  created: 1700000001,
  model: "gpt-4o",
  choices: [
    {
      index: 0,
      message: {
        role: "assistant",
        content: null,
        tool_calls: [
          {
            id: "call_abc123",
            type: "function",
            function: { name: "get_weather", arguments: '{"location":"NYC"}' },
          },
        ],
      },
      finish_reason: "tool_calls",
    },
  ],
  usage: { prompt_tokens: 15, completion_tokens: 12, total_tokens: 27 },
};

const VALID_ERROR_BODY = {
  error: {
    message: "Invalid request",
    type: "invalid_request_error" as const,
    param: "temperature",
    code: "invalid_value",
  },
};

describe("OpenAIClient", () => {
  const API_KEY = "sk-test-key-abc123";

  afterEach(() => {
    jest.restoreAllMocks();
  });

  describe("constructor", () => {
    it("rejects API keys not matching sk-* format", () => {
      expect(() => new OpenAIClient("bad-key")).toThrow("Invalid API key format: must start with 'sk-'");
    });

    it("accepts valid sk-* API keys", () => {
      const client = new OpenAIClient(API_KEY);
      expect(client).toBeInstanceOf(OpenAIClient);
    });
  });

  describe("createChatCompletion", () => {
    it("sends correct request shape with all required fields", async () => {
      mockFetchSuccess(VALID_COMPLETION);
      const client = new OpenAIClient(API_KEY);

      await client.createChatCompletion({
        model: "gpt-4o",
        messages: [{ role: "user", content: "Hi" }],
      });

      const fetchCall = (global.fetch as jest.Mock).mock.calls[0];
      expect(fetchCall[0]).toBe("https://api.openai.com/v1/chat/completions");
      expect(fetchCall[1].method).toBe("POST");
      expect(fetchCall[1].headers["Authorization"]).toBe(`Bearer ${API_KEY}`);
      expect(fetchCall[1].headers["Content-Type"]).toBe("application/json");

      const body = JSON.parse(fetchCall[1].body);
      expect(body.model).toBe("gpt-4o");
      expect(body.messages).toEqual([{ role: "user", content: "Hi" }]);
      expect(body.stream).toBe(false);
    });

    it("returns response with validated completion ID format (chatcmpl-*)", async () => {
      mockFetchSuccess(VALID_COMPLETION);
      const client = new OpenAIClient(API_KEY);

      const result = await client.createChatCompletion({
        model: "gpt-4o",
        messages: [{ role: "user", content: "Hi" }],
      });

      expect(result.id).toMatch(/^chatcmpl-/);
      expect(result.object).toBe("chat.completion");
      expect(typeof result.created).toBe("number");
      expect(result.model).toBe("gpt-4o");
    });

    it("validates response usage contains integer token counts", async () => {
      mockFetchSuccess(VALID_COMPLETION);
      const client = new OpenAIClient(API_KEY);

      const result = await client.createChatCompletion({
        model: "gpt-4o",
        messages: [{ role: "user", content: "Hi" }],
      });

      expect(Number.isInteger(result.usage.prompt_tokens)).toBe(true);
      expect(Number.isInteger(result.usage.completion_tokens)).toBe(true);
      expect(Number.isInteger(result.usage.total_tokens)).toBe(true);
      expect(result.usage.prompt_tokens).toBe(10);
      expect(result.usage.completion_tokens).toBe(8);
      expect(result.usage.total_tokens).toBe(18);
    });

    it("accumulates token counts as integers across multiple requests", async () => {
      mockFetchSuccess(VALID_COMPLETION);
      const client = new OpenAIClient(API_KEY);

      await client.createChatCompletion({ model: "gpt-4o", messages: [{ role: "user", content: "Hi" }] });
      await client.createChatCompletion({ model: "gpt-4o", messages: [{ role: "user", content: "Hi" }] });

      const usage = client.getTokenUsage();
      expect(usage.promptTokens).toBe(20);
      expect(usage.completionTokens).toBe(16);
      expect(usage.totalTokens).toBe(36);
      expect(Number.isInteger(usage.totalTokens)).toBe(true);
    });

    it("handles null content in tool_calls response without crashing", async () => {
      mockFetchSuccess(VALID_TOOL_CALL_COMPLETION);
      const client = new OpenAIClient(API_KEY);

      const result = await client.createChatCompletion({
        model: "gpt-4o",
        messages: [{ role: "user", content: "What is the weather in NYC?" }],
        tools: [{ type: "function", function: { name: "get_weather", parameters: {} } }],
      });

      expect(result.choices[0].message.content).toBeNull();
      expect(result.choices[0].message.tool_calls).toHaveLength(1);
      expect(result.choices[0].message.tool_calls![0].function.name).toBe("get_weather");
      expect(result.choices[0].finish_reason).toBe("tool_calls");
    });

    it("handles all finish_reason values", async () => {
      for (const reason of ["stop", "length", "content_filter", "tool_calls", null] as const) {
        const body = {
          ...VALID_COMPLETION,
          choices: [{
            index: 0,
            message: { role: "assistant", content: reason === "tool_calls" ? null : "text" },
            finish_reason: reason,
          }],
        };
        mockFetchSuccess(body);
        const client = new OpenAIClient(API_KEY);
        const result = await client.createChatCompletion({
          model: "gpt-4o",
          messages: [{ role: "user", content: "Hi" }],
        });
        expect(result.choices[0].finish_reason).toBe(reason);
      }
    });

    it("rejects temperature outside [0, 2] range", async () => {
      const client = new OpenAIClient(API_KEY);
      await expect(
        client.createChatCompletion({
          model: "gpt-4o",
          messages: [{ role: "user", content: "Hi" }],
          temperature: 5.0,
        })
      ).rejects.toThrow("Temperature must be between 0 and 2, got 5");
    });

    it("rejects empty messages array", async () => {
      const client = new OpenAIClient(API_KEY);
      await expect(
        client.createChatCompletion({
          model: "gpt-4o",
          messages: [],
        })
      ).rejects.toThrow("Messages array must contain at least one message");
    });

    it("throws on network error with descriptive message", async () => {
      mockFetchNetworkError();
      const client = new OpenAIClient(API_KEY);
      await expect(
        client.createChatCompletion({
          model: "gpt-4o",
          messages: [{ role: "user", content: "Hi" }],
        })
      ).rejects.toThrow("Network error calling OpenAI: ECONNREFUSED");
    });

    it("handles 400 with error shape including param and code", async () => {
      mockFetchError(400, VALID_ERROR_BODY);
      const client = new OpenAIClient(API_KEY);
      await expect(
        client.createChatCompletion({
          model: "gpt-4o",
          messages: [{ role: "user", content: "Hi" }],
        })
      ).rejects.toThrow("Invalid request: Invalid request (param: temperature)");
    });

    it("handles 401 authentication error", async () => {
      mockFetchError(401, {
        error: { message: "Incorrect API key", type: "authentication_error", param: null, code: "invalid_api_key" },
      });
      const client = new OpenAIClient(API_KEY);
      await expect(
        client.createChatCompletion({ model: "gpt-4o", messages: [{ role: "user", content: "Hi" }] })
      ).rejects.toThrow("Authentication failed: Incorrect API key");
    });

    it("handles 429 rate limit with rate limit headers", async () => {
      mockFetchError(429, {
        error: { message: "Rate limit exceeded", type: "rate_limit_error", param: null, code: "rate_limit_exceeded" },
      }, {
        "x-ratelimit-limit-requests": "60",
        "x-ratelimit-remaining-requests": "0",
        "x-ratelimit-reset-requests": "1s",
        "x-ratelimit-limit-tokens": "150000",
        "x-ratelimit-remaining-tokens": "0",
        "x-ratelimit-reset-tokens": "6s",
      });
      const client = new OpenAIClient(API_KEY);
      await expect(
        client.createChatCompletion({ model: "gpt-4o", messages: [{ role: "user", content: "Hi" }] })
      ).rejects.toThrow(/Rate limited.*Resets at/);
    });

    it("handles 500 server error", async () => {
      mockFetchError(500, {
        error: { message: "Internal error", type: "server_error", param: null, code: null },
      });
      const client = new OpenAIClient(API_KEY);
      await expect(
        client.createChatCompletion({ model: "gpt-4o", messages: [{ role: "user", content: "Hi" }] })
      ).rejects.toThrow("OpenAI server error: Internal error");
    });

    it("handles 503 service unavailable", async () => {
      mockFetchError(503, {
        error: { message: "Overloaded", type: "server_error", param: null, code: null },
      });
      const client = new OpenAIClient(API_KEY);
      await expect(
        client.createChatCompletion({ model: "gpt-4o", messages: [{ role: "user", content: "Hi" }] })
      ).rejects.toThrow("OpenAI service unavailable: Overloaded. Retry after backoff.");
    });

    it("rejects completion ID not matching chatcmpl-* format", async () => {
      mockFetchSuccess({ ...VALID_COMPLETION, id: "wrong-prefix-123" });
      const client = new OpenAIClient(API_KEY);
      await expect(
        client.createChatCompletion({ model: "gpt-4o", messages: [{ role: "user", content: "Hi" }] })
      ).rejects.toThrow('Invalid completion ID format: expected chatcmpl-* prefix');
    });
  });

  describe("streamChatCompletion", () => {
    it("yields all chunks and terminates on [DONE]", async () => {
      mockFetchStream([
        'data: {"id":"chatcmpl-s1","object":"chat.completion.chunk","created":1700000000,"model":"gpt-4o","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}\n\n',
        'data: {"id":"chatcmpl-s1","object":"chat.completion.chunk","created":1700000000,"model":"gpt-4o","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}\n\n',
        'data: {"id":"chatcmpl-s1","object":"chat.completion.chunk","created":1700000000,"model":"gpt-4o","choices":[{"index":0,"delta":{"content":"!"},"finish_reason":null}]}\n\n',
        'data: {"id":"chatcmpl-s1","object":"chat.completion.chunk","created":1700000000,"model":"gpt-4o","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}\n\n',
        "data: [DONE]\n\n",
      ]);

      const client = new OpenAIClient(API_KEY);
      const chunks: unknown[] = [];
      for await (const chunk of client.streamChatCompletion({
        model: "gpt-4o",
        messages: [{ role: "user", content: "Hi" }],
      })) {
        chunks.push(chunk);
      }

      expect(chunks).toHaveLength(4);
      expect((chunks[0] as ChatCompletionChunk).choices[0].delta.role).toBe("assistant");
      expect((chunks[1] as ChatCompletionChunk).choices[0].delta.content).toBe("Hello");
      expect((chunks[2] as ChatCompletionChunk).choices[0].delta.content).toBe("!");
      expect((chunks[3] as ChatCompletionChunk).choices[0].finish_reason).toBe("stop");
    });

    it("handles streaming error responses", async () => {
      global.fetch = jest.fn().mockResolvedValue({
        ok: false,
        status: 429,
        json: async () => ({
          error: { message: "Too many requests", type: "rate_limit_error", param: null, code: null },
        }),
        headers: new Headers({ "x-ratelimit-reset-requests": "2s" }),
      });

      const client = new OpenAIClient(API_KEY);
      const gen = client.streamChatCompletion({
        model: "gpt-4o",
        messages: [{ role: "user", content: "Hi" }],
      });

      await expect(gen.next()).rejects.toThrow(/Rate limited/);
    });
  });

  describe("listModels", () => {
    it("returns model list with correct shape", async () => {
      mockFetchSuccess({
        object: "list",
        data: [
          { id: "gpt-4o", object: "model", created: 1700000000, owned_by: "openai" },
          { id: "gpt-3.5-turbo", object: "model", created: 1690000000, owned_by: "openai" },
        ],
      });
      const client = new OpenAIClient(API_KEY);
      const models = await client.listModels();

      expect(models).toHaveLength(2);
      expect(models[0].id).toBe("gpt-4o");
      expect(models[0].object).toBe("model");
      expect(typeof models[0].created).toBe("number");
      expect(models[0].owned_by).toBe("openai");
    });
  });
});
```

**Expected Stricture result:** 0 violations. All rules pass.

---

## B01 -- No Error Handling

**Bug:** No try/catch on the fetch call. Network errors crash the caller with an unhandled promise rejection.

**Stricture rule:** `TQ-error-path-coverage`

```typescript
// B01: No error handling -- fetch failures crash the process.
async createChatCompletion(request: ChatCompletionRequest): Promise<ChatCompletionResponse> {
  validateMessages(request.messages);
  validateTemperature(request.temperature);

  // BUG: no try/catch -- ECONNREFUSED, DNS failures, and timeouts
  // produce unhandled rejections that crash the Node.js process.
  const response = await fetch(`${this.baseUrl}/v1/chat/completions`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "Authorization": `Bearer ${this.apiKey}`,
    },
    body: JSON.stringify({ ...request, stream: false }),
  });

  const completion: ChatCompletionResponse = await response.json();
  return completion;
}
```

**Expected violation:** `TQ-error-path-coverage` -- `fetch()` call at line 8 has no try/catch or `.catch()` handler. Network-level errors (ECONNREFUSED, DNS failure, timeout) will produce unhandled rejections.

**Production impact:** The process crashes on any network interruption. No retry, no fallback, no error logging. In a serverless environment, this means a 500 to the end user with no diagnostic information. In a long-running server, repeated crashes trigger container restarts.

---

## B02 -- No Status Code Check

**Bug:** Treats every HTTP response as success. A 429 rate limit or 503 overloaded response gets parsed as a valid completion.

**Stricture rule:** `CTR-status-code-handling`

```typescript
// B02: No status code check -- 429, 500, 503 all treated as success.
async createChatCompletion(request: ChatCompletionRequest): Promise<ChatCompletionResponse> {
  validateMessages(request.messages);

  let response: Response;
  try {
    response = await fetch(`${this.baseUrl}/v1/chat/completions`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${this.apiKey}`,
      },
      body: JSON.stringify({ ...request, stream: false }),
    });
  } catch (err) {
    throw new Error(`Network error: ${(err as Error).message}`);
  }

  // BUG: Never checks response.ok or response.status.
  // A 429 response body is { error: { message, type, ... } },
  // which gets destructured as a ChatCompletionResponse, producing
  // undefined values for id, choices, usage, etc.
  const completion: ChatCompletionResponse = await response.json();
  return completion;
}
```

**Expected violation:** `CTR-status-code-handling` -- Manifest declares status_codes [200, 400, 401, 403, 429, 500, 503] but code only handles the 200 path. No branching on `response.status` or `response.ok`.

**Production impact:** Rate-limited responses (429) are silently treated as completions. The `usage` field is undefined, causing NaN token counts. The `choices` array is undefined, causing a TypeError when the caller iterates results. Worse: the error body leaks to the UI as a "completion."

---

## B03 -- Shallow Assertions

**Bug:** Tests verify only that a result exists, not its shape, values, or contract conformance.

**Stricture rule:** `TQ-no-shallow-assertions`

```typescript
// B03: Shallow assertions -- tests prove nothing about correctness.
describe("OpenAIClient", () => {
  it("creates a completion", async () => {
    mockFetchSuccess(VALID_COMPLETION);
    const client = new OpenAIClient("sk-test123");

    const result = await client.createChatCompletion({
      model: "gpt-4o",
      messages: [{ role: "user", content: "Hi" }],
    });

    // BUG: These assertions verify existence, not correctness.
    // The response could have wrong id format, wrong object type,
    // missing usage, zero-length choices -- all would pass.
    expect(result).toBeDefined();
    expect(result.choices).toBeTruthy();
    expect(result.usage).not.toBeNull();
  });

  it("handles errors", async () => {
    mockFetchError(400, VALID_ERROR_BODY);
    const client = new OpenAIClient("sk-test123");

    // BUG: Only checks that it throws, not WHAT it throws.
    // Could throw "undefined is not a function" and this passes.
    await expect(
      client.createChatCompletion({ model: "gpt-4o", messages: [{ role: "user", content: "Hi" }] })
    ).rejects.toBeDefined();
  });
});
```

**Expected violation:** `TQ-no-shallow-assertions` at three locations:
1. `expect(result).toBeDefined()` -- shallow assertion on typed return value `ChatCompletionResponse` with 6 fields.
2. `expect(result.choices).toBeTruthy()` -- shallow assertion; does not verify array length, element shape, or finish_reason.
3. `expect(result.usage).not.toBeNull()` -- shallow assertion; does not verify prompt_tokens, completion_tokens, or total_tokens values.
4. `rejects.toBeDefined()` -- shallow assertion on error; does not verify error message or type.

**Production impact:** A regression that changes the response shape (e.g., `usage` becomes `{tokens: 18}` instead of `{prompt_tokens: 10, completion_tokens: 8, total_tokens: 18}`) would not be caught by these tests. The test suite gives false confidence.

---

## B04 -- Missing Negative Tests

**Bug:** Only tests the happy path. No tests for content filtering, rate limiting, authentication failure, streaming errors, or edge cases.

**Stricture rule:** `TQ-negative-cases`

```typescript
// B04: Missing negative tests -- only happy path tested.
describe("OpenAIClient", () => {
  it("creates a completion successfully", async () => {
    mockFetchSuccess(VALID_COMPLETION);
    const client = new OpenAIClient("sk-test123");

    const result = await client.createChatCompletion({
      model: "gpt-4o",
      messages: [{ role: "user", content: "Hello" }],
    });

    expect(result.id).toMatch(/^chatcmpl-/);
    expect(result.choices[0].message.content).toBe("Hello! How can I help?");
    expect(result.usage.total_tokens).toBe(18);
  });

  it("streams a completion successfully", async () => {
    mockFetchStream([
      'data: {"id":"chatcmpl-s1","object":"chat.completion.chunk","created":1700000000,"model":"gpt-4o","choices":[{"index":0,"delta":{"content":"Hi"},"finish_reason":null}]}\n\n',
      "data: [DONE]\n\n",
    ]);
    const client = new OpenAIClient("sk-test123");

    const chunks: unknown[] = [];
    for await (const chunk of client.streamChatCompletion({
      model: "gpt-4o",
      messages: [{ role: "user", content: "Hi" }],
    })) {
      chunks.push(chunk);
    }

    expect(chunks).toHaveLength(1);
  });

  // BUG: No tests for:
  // - 401 authentication error
  // - 429 rate limiting
  // - 503 service unavailable
  // - content_filter finish_reason
  // - tool_calls with null content
  // - Invalid temperature (out of range)
  // - Empty messages array
  // - Network failure (ECONNREFUSED)
  // - Invalid completion ID format
  // - Streaming error mid-stream
});
```

**Expected violation:** `TQ-negative-cases` -- The test file covers 2 positive scenarios but 0 negative scenarios. The source handles 7 error status codes (400, 401, 403, 429, 500, 503, and unexpected), 4 finish_reason variants (stop, length, content_filter, tool_calls), input validation (temperature range, empty messages, API key format), and network errors. None are tested.

**Production impact:** Regressions in error handling go undetected. If the 429 handler is accidentally removed, no test fails. If content_filter handling breaks, users see raw errors instead of graceful messages.

---

## B05 -- Request Missing Required Fields

**Bug:** The client omits the required `messages` array from the request body, sending only `model`.

**Stricture rule:** `CTR-request-shape`

```typescript
// B05: Request missing required fields -- omits messages array.
interface MinimalCompletionRequest {
  model: string;
  // BUG: messages field is missing entirely from the type.
  // The OpenAI API requires messages as a non-empty array.
  temperature?: number;
  max_tokens?: number;
}

async createChatCompletion(request: MinimalCompletionRequest): Promise<ChatCompletionResponse> {
  let response: Response;
  try {
    response = await fetch(`${this.baseUrl}/v1/chat/completions`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${this.apiKey}`,
      },
      // BUG: Serializes { model: "gpt-4o", temperature: 1 }
      // without messages. OpenAI returns 400:
      // { error: { message: "'messages' is a required property", type: "invalid_request_error" } }
      body: JSON.stringify(request),
    });
  } catch (err) {
    throw new Error(`Network error: ${(err as Error).message}`);
  }

  if (!response.ok) {
    const errorBody: OpenAIError = await response.json();
    throw new Error(errorBody.error.message);
  }

  return await response.json();
}
```

**Expected violation:** `CTR-request-shape` -- Manifest declares `messages` as `required: true` on `POST /v1/chat/completions`. Client type `MinimalCompletionRequest` has no `messages` field. Every request will receive a 400 response.

**Production impact:** Every single API call fails with a 400 error. This is caught immediately in development but demonstrates that Stricture catches structural omissions before runtime.

---

## B06 -- Response Type Mismatch

**Bug:** The client's response type omits the `usage` field, so token tracking is impossible and accessing `result.usage.total_tokens` throws at runtime.

**Stricture rule:** `CTR-response-shape`

```typescript
// B06: Response type mismatch -- usage field missing from client type.
interface ChatCompletionResponse {
  id: string;
  object: "chat.completion";
  created: number;
  model: string;
  choices: ChatCompletionChoice[];
  // BUG: usage field is missing from the type definition.
  // The API always returns usage: { prompt_tokens, completion_tokens, total_tokens }
  // but the client type does not include it. TypeScript won't complain
  // because JSON.parse returns `any` from response.json(), but any code
  // that tries to access result.usage will see `undefined`.
}

async createChatCompletion(request: ChatCompletionRequest): Promise<ChatCompletionResponse> {
  // ... fetch and error handling ...
  const completion: ChatCompletionResponse = await response.json();

  // This line compiles but crashes at runtime:
  // this.tokenAccumulator.totalTokens += completion.usage.total_tokens;
  // TypeError: Cannot read properties of undefined (reading 'total_tokens')

  return completion;
}
```

**Expected violation:** `CTR-response-shape` -- Manifest declares `usage` as `required: true` in the response. Client type `ChatCompletionResponse` does not include `usage`. Field count mismatch: manifest has 6 top-level fields, client type has 5.

**Production impact:** Token usage tracking is broken. Any code downstream that relies on `usage.total_tokens` for billing, quota management, or cost monitoring crashes with a TypeError. Since TypeScript trusts `response.json()` as `any`, this is not caught at compile time.

---

## B07 -- Wrong Field Types

**Bug:** The `created` timestamp is stored as a string instead of a number. Arithmetic on it produces NaN or string concatenation.

**Stricture rule:** `CTR-manifest-conformance`

```typescript
// B07: Wrong field types -- created timestamp stored as string.
interface ChatCompletionResponse {
  id: string;
  object: "chat.completion";
  // BUG: created is typed as string, but the API returns a Unix timestamp
  // as a number (e.g., 1700000000). When deserialized from JSON, the value
  // IS a number, but the TypeScript type says string. Code that does
  // `new Date(completion.created * 1000)` gets NaN because
  // TypeScript allows string * number (NaN).
  created: string;
  model: string;
  choices: ChatCompletionChoice[];
  usage: {
    prompt_tokens: number;
    completion_tokens: number;
    total_tokens: number;
  };
}

function formatCompletionTimestamp(completion: ChatCompletionResponse): string {
  // BUG: If created is treated as string by downstream code:
  // new Date("1700000000" * 1000) works in JS (implicit coercion)
  // but new Date(completion.created + "000") produces "1700000000000"
  // which is a valid date by accident -- masking the type error.
  const timestamp = Number(completion.created) * 1000;
  return new Date(timestamp).toISOString();
}
```

**Expected violation:** `CTR-manifest-conformance` -- Manifest declares `created` as `type: integer, format: unix_timestamp`. Client type declares `created: string`. Type mismatch: integer vs string.

**Production impact:** Date formatting works inconsistently. Some code paths coerce correctly (implicit `Number()`) while others concatenate ("1700000000" + 60 = "170000000060"). Sorting by creation time fails because string comparison produces lexicographic order, not chronological order.

---

## B08 -- Incomplete Enum Handling

**Bug:** The `finish_reason` handler only covers `"stop"`. Responses with `"length"`, `"content_filter"`, or `"tool_calls"` fall through silently with no action.

**Stricture rule:** `CTR-strictness-parity`

```typescript
// B08: Incomplete enum handling -- only handles "stop" finish_reason.
private handleFinishReason(reason: string): void {
  // BUG: Only handles "stop". The API returns 4 possible values:
  // "stop", "length", "content_filter", "tool_calls", plus null for streaming chunks.
  // Missing "length" means truncated responses are silently accepted.
  // Missing "content_filter" means filtered content is treated as valid.
  // Missing "tool_calls" means function calling responses are not processed.
  if (reason === "stop") {
    return; // Normal completion
  }
  // All other finish_reasons fall through with no handling.
  // No warning, no error, no logging.
}

async createChatCompletion(request: ChatCompletionRequest): Promise<ChatCompletionResponse> {
  // ...
  for (const choice of completion.choices) {
    this.handleFinishReason(choice.finish_reason);

    // BUG: Assumes content is always a string for all finish_reasons.
    // When finish_reason is "tool_calls", content is null.
    const text = choice.message.content.toUpperCase(); // null.toUpperCase() -> TypeError
  }
  return completion;
}
```

**Expected violation:** `CTR-strictness-parity` -- Manifest declares `finish_reason` as enum with values `["stop", "length", "content_filter", "tool_calls"]` plus nullable. Client code only handles `"stop"` (1 of 4 enum values). Missing branches: `"length"`, `"content_filter"`, `"tool_calls"`.

**Production impact:** Truncated responses (`finish_reason: "length"`) are silently accepted as complete answers. Users receive half-finished responses. Content-filtered responses are treated as valid, potentially showing "[Content blocked]" placeholder text to users. Tool call responses crash with TypeError on null content access.

---

## B09 -- Missing Range Validation

**Bug:** Temperature is sent to the API without validating the [0, 2] range. Values like 5.0 or -1 are sent, causing a 400 error at runtime.

**Stricture rule:** `CTR-strictness-parity`

```typescript
// B09: Missing range validation -- temperature accepted without bounds check.
async createChatCompletion(request: ChatCompletionRequest): Promise<ChatCompletionResponse> {
  validateMessages(request.messages);
  // BUG: No temperature validation. The manifest declares range [0, 2]
  // but no code enforces it. A caller can pass temperature: 5.0 or
  // temperature: -1.0, which OpenAI rejects with:
  // { error: { message: "5 is greater than the maximum of 2", param: "temperature" } }
  //
  // This validation gap means the error is discovered at the API boundary
  // (400 response) instead of at the call site, making debugging harder
  // and wasting a network round trip.

  let response: Response;
  try {
    response = await fetch(`${this.baseUrl}/v1/chat/completions`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${this.apiKey}`,
      },
      body: JSON.stringify(request), // temperature: 5.0 sent as-is
    });
  } catch (err) {
    throw new Error(`Network error: ${(err as Error).message}`);
  }

  if (!response.ok) {
    const errorBody: OpenAIError = await response.json();
    throw new Error(errorBody.error.message);
  }

  return await response.json();
}
```

**Expected violation:** `CTR-strictness-parity` -- Manifest declares `temperature` with `range: [0, 2]`. Client code does not validate this range before sending the request. The constraint exists in the contract but is not enforced by the client.

**Production impact:** Invalid temperature values produce confusing 400 errors from the API. The error message mentions "temperature" but the call stack does not show where the bad value originated. In a pipeline where temperature is user-configurable, this allows injection of invalid values that fail silently until runtime.

---

## B10 -- Format Not Validated

**Bug:** The completion ID is accepted as any string. IDs not matching the `chatcmpl-*` pattern are silently accepted, which could indicate a proxy, mock, or man-in-the-middle response.

**Stricture rule:** `CTR-strictness-parity`

```typescript
// B10: Format not validated -- completion ID accepted as any string.
async createChatCompletion(request: ChatCompletionRequest): Promise<ChatCompletionResponse> {
  validateMessages(request.messages);
  validateTemperature(request.temperature);

  let response: Response;
  try {
    response = await fetch(`${this.baseUrl}/v1/chat/completions`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${this.apiKey}`,
      },
      body: JSON.stringify({ ...request, stream: false }),
    });
  } catch (err) {
    throw new Error(`Network error: ${(err as Error).message}`);
  }

  if (!response.ok) {
    const errorBody: OpenAIError = await response.json();
    throw new Error(errorBody.error.message);
  }

  const completion: ChatCompletionResponse = await response.json();

  // BUG: No validation of completion.id format.
  // Manifest requires format: "chatcmpl-*" but this code accepts
  // any string: "fake-id", "", "null", "undefined", etc.
  // A misconfigured proxy could return { id: "proxy-cached-123" }
  // and the client would treat it as a valid OpenAI response.

  return completion;
}
```

**Expected violation:** `CTR-strictness-parity` -- Manifest declares `id` with `format: "chatcmpl-*"`. Client code does not validate the ID format. The format constraint exists in the contract but is unenforced.

**Production impact:** If the client is routed through an API gateway or caching proxy that rewrites response IDs, the client silently accepts non-OpenAI responses. This breaks audit trails that depend on chatcmpl-* IDs for tracing. It also means a compromised proxy can inject arbitrary responses without detection.

---

## B11 -- Precision Loss

**Bug:** Token counts are accumulated using a floating-point variable. After many requests, integer precision is lost due to IEEE 754 limitations.

**Stricture rule:** `CTR-strictness-parity`

```typescript
// B11: Precision loss -- token counts accumulated in float accumulator.
export class OpenAIClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;
  // BUG: tokenUsage is typed as number (float64 in JS).
  // For small counts this works fine. But after thousands of requests,
  // the accumulated total can exceed Number.MAX_SAFE_INTEGER territory
  // or accumulate floating-point drift if any intermediate computation
  // introduces a non-integer.
  private tokenUsage: { prompt: number; completion: number; total: number };

  constructor(apiKey: string) {
    validateApiKey(apiKey);
    this.apiKey = apiKey;
    this.baseUrl = "https://api.openai.com";
    this.tokenUsage = { prompt: 0.0, completion: 0.0, total: 0.0 };
  }

  async createChatCompletion(request: ChatCompletionRequest): Promise<ChatCompletionResponse> {
    // ... fetch and validation ...
    const completion: ChatCompletionResponse = await response.json();

    // BUG: Using += on floats. While token counts from the API are integers,
    // if any middleware or transformation introduces a fractional value
    // (e.g., dividing by a rate), the accumulator drifts.
    // More critically: at scale, after millions of requests, even pure
    // integer addition in float64 loses precision above 2^53.
    //
    // Example scenario:
    //   this.tokenUsage.total = 9007199254740992; // 2^53
    //   this.tokenUsage.total += 1; // Still 9007199254740992 (precision lost)
    this.tokenUsage.prompt += completion.usage.prompt_tokens;
    this.tokenUsage.completion += completion.usage.completion_tokens;
    this.tokenUsage.total += completion.usage.total_tokens;

    // BUG: Cost calculation introduces fractional cents.
    // $0.01 per 1000 tokens: 18 tokens = 0.00018
    // After 10000 requests: 0.00018 * 10000 = 1.7999999999999998 (not 1.8)
    const costUsd = this.tokenUsage.total * 0.00001;
    console.log(`Total cost: $${costUsd}`);

    return completion;
  }
}
```

**Expected violation:** `CTR-strictness-parity` -- Manifest declares `prompt_tokens`, `completion_tokens`, and `total_tokens` as `type: integer`. Client accumulates them in a float-initialized variable (`0.0`). Arithmetic on the accumulator introduces a precision risk. Integer accumulation should use `0` (not `0.0`) and avoid float multiplication in the accumulation path.

**Production impact:** Token billing becomes inaccurate at scale. A SaaS application tracking usage across thousands of users accumulates rounding errors that compound over time. Monthly billing reports show fractional token counts (e.g., 1847293.0000000002 instead of 1847293), breaking invoice systems that expect integers. At extreme scale (>2^53 total tokens), additions are silently dropped.

---

## B12 -- Nullable Field Crash

**Bug:** Accesses `choice.message.content.length` without checking for null. When the finish_reason is `"tool_calls"`, content is null, and `.length` throws a TypeError.

**Stricture rule:** `CTR-response-shape`

```typescript
// B12: Nullable field crash -- content.length on null content.
async createChatCompletion(request: ChatCompletionRequest): Promise<ChatCompletionResponse> {
  // ... fetch and error handling ...
  const completion: ChatCompletionResponse = await response.json();

  for (const choice of completion.choices) {
    // BUG: Accesses .length on content without null check.
    // When finish_reason is "tool_calls", the message contains:
    //   { role: "assistant", content: null, tool_calls: [...] }
    // Calling null.length throws:
    //   TypeError: Cannot read properties of null (reading 'length')
    //
    // This crashes the entire completion handler, even though the
    // tool_calls data is perfectly valid and usable.
    const contentLength = choice.message.content.length;
    console.log(`Response length: ${contentLength} characters`);

    if (choice.finish_reason === "length") {
      console.warn(`Truncated at ${contentLength} chars`);
    }
  }

  // BUG: Also crashes when building response summaries:
  const summary = completion.choices
    .map((c) => c.message.content.substring(0, 100)) // null.substring() -> TypeError
    .join("\n");

  return completion;
}
```

**Expected violation:** `CTR-response-shape` -- Manifest declares `content` as `type: "string|null"`. Client code accesses `choice.message.content.length` and `choice.message.content.substring()` without null guard. Property access on nullable field without null check.

**Production impact:** Any request that triggers tool calling crashes the application. Since tool calling is a core feature of GPT-4o, this means function-calling workflows are completely broken. The crash happens in the response handler, so the valid tool_calls data is lost even though it was successfully received from the API.

---

## B13 -- Missing API Key Validation

**Bug:** The client sends requests without validating that the API key matches the `sk-*` format. Non-API strings (session tokens, passwords, random strings) are sent as Bearer tokens, leaking credentials to OpenAI's servers.

**Stricture rule:** `CTR-request-shape`

```typescript
// B13: Missing API key validation -- accepts any string as API key.
export class OpenAIClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;

  constructor(apiKey: string) {
    // BUG: No validation of apiKey format.
    // The manifest declares auth format: "sk-*" but the constructor
    // accepts any string. This means:
    // 1. A session token ("sess_abc123") gets sent as Bearer token
    // 2. A password ("hunter2") gets sent to OpenAI's servers
    // 3. An empty string ("") produces "Authorization: Bearer "
    // 4. A different provider's key ("anthropic-sk-...") gets leaked
    this.apiKey = apiKey;
    this.baseUrl = "https://api.openai.com";
  }

  async createChatCompletion(request: ChatCompletionRequest): Promise<ChatCompletionResponse> {
    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/chat/completions`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          // BUG: Sends whatever string was provided, even if it is
          // a Stripe key, AWS secret, or database password.
          // OpenAI's servers receive the credential in plaintext.
          "Authorization": `Bearer ${this.apiKey}`,
        },
        body: JSON.stringify(request),
      });
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) {
      // The 401 error here is the ONLY indication that the key was wrong.
      // By this point, the wrong credential has already been transmitted.
      const errorBody: OpenAIError = await response.json();
      throw new Error(errorBody.error.message);
    }

    return await response.json();
  }
}

// Example of how this bug manifests:
// const client = new OpenAIClient(process.env.STRIPE_SECRET_KEY!);
// Sends Stripe key to OpenAI -- credential leaked to third party.
```

**Expected violation:** `CTR-request-shape` -- Manifest declares `auth.format: "sk-*"`. Client constructor does not validate the API key format before storing it. Non-conforming strings are accepted and transmitted to the API endpoint.

**Production impact:** Credential leakage to a third-party service. If environment variables are misconfigured (e.g., `OPENAI_API_KEY` is set to a Stripe secret key), the Stripe key is sent to OpenAI's servers in the Authorization header. This is a security incident requiring key rotation. The error is only discovered when OpenAI returns a 401, by which time the credential is already in OpenAI's access logs.

---

## B14 -- Streaming Terminated Early

**Bug:** The streaming reader stops after the first SSE chunk instead of reading until the `data: [DONE]` sentinel. Multi-token responses are truncated to a single token.

**Stricture rule:** `CTR-response-shape`

```typescript
// B14: Streaming terminated early -- stops on first chunk instead of [DONE].
async *streamChatCompletion(request: ChatCompletionRequest): AsyncGenerator<ChatCompletionChunk, void, undefined> {
  validateMessages(request.messages);

  let response: Response;
  try {
    response = await fetch(`${this.baseUrl}/v1/chat/completions`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${this.apiKey}`,
      },
      body: JSON.stringify({ ...request, stream: true }),
    });
  } catch (err) {
    throw new Error(`Network error: ${(err as Error).message}`);
  }

  if (!response.ok) {
    const errorBody: OpenAIError = await response.json();
    throw new Error(`OpenAI error: ${errorBody.error.message}`);
  }

  if (!response.body) {
    throw new Error("Response body is null");
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();

  // BUG: Reads exactly ONE chunk, then returns.
  // The SSE protocol sends multiple "data: {...}" lines followed by
  // "data: [DONE]". This code reads the first chunk and stops,
  // discarding all subsequent tokens.
  //
  // For a response "Hello, how are you?", only "Hello" (or even just "H")
  // is yielded. The rest of the response is abandoned in the stream.
  const { done, value } = await reader.read();
  if (!done && value) {
    const text = decoder.decode(value);
    const lines = text.split("\n");
    for (const line of lines) {
      const trimmed = line.trim();
      if (trimmed === "" || trimmed === "data: [DONE]") continue;
      if (trimmed.startsWith("data: ")) {
        const chunk: ChatCompletionChunk = JSON.parse(trimmed.slice(6));
        yield chunk;
      }
    }
  }

  // BUG: reader.releaseLock() never called -- the ReadableStream
  // remains locked, preventing any other code from reading it.
  // Also: no finally block, so errors during parsing leave the lock held.
}
```

**Expected violation:** `CTR-response-shape` -- Manifest declares streaming terminator as `data: [DONE]`. Client reads exactly one chunk from the stream without looping until the `[DONE]` sentinel. The streaming contract requires reading all chunks until termination.

**Production impact:** Every streaming response is truncated. A chatbot shows only the first word or fragment of each response. Users see "I" instead of "I would be happy to help you with that." The truncation is silent -- no error, no warning. The abandoned stream also leaks the ReadableStream lock, causing memory leaks in long-running processes.

---

## B15 -- Race Condition

**Bug:** Checks model availability via `GET /v1/models`, then sends a completion request. The model can be deprecated between the check and the use, causing a 404 that the code does not handle.

**Stricture rule:** `CTR-request-shape`

```typescript
// B15: Race condition -- TOCTOU between model check and completion request.
export class OpenAIClient {
  // ... constructor and other methods ...

  async createChatCompletionSafe(
    request: ChatCompletionRequest
  ): Promise<ChatCompletionResponse> {
    // Step 1: Check if the model is available
    // BUG: TOCTOU (time-of-check-to-time-of-use) race condition.
    // Between the listModels() call and the createChatCompletion() call,
    // OpenAI can deprecate or remove the model. This is not hypothetical:
    // OpenAI has deprecated models with as little as 2 weeks notice
    // (gpt-3.5-turbo-0301, gpt-4-0314, etc.)
    let models: Array<{ id: string }>;
    try {
      models = await this.listModels();
    } catch (err) {
      throw new Error(`Failed to verify model availability: ${(err as Error).message}`);
    }

    const modelExists = models.some((m) => m.id === request.model);
    if (!modelExists) {
      throw new Error(`Model "${request.model}" is not available`);
    }

    // Step 2: Send the completion request
    // BUG: Between Step 1 and Step 2, the model could be:
    // - Deprecated (returns 404)
    // - Rate-limited for this specific model (returns 429)
    // - Under maintenance (returns 503)
    //
    // The code assumes the model check guarantees success, so it does
    // NOT handle 404 in the completion response handler.
    //
    // Additional race: if this method is called concurrently by multiple
    // request handlers, each one makes a redundant listModels() call,
    // consuming rate limit quota unnecessarily. Under load, the model
    // check itself triggers rate limiting, causing the completion
    // request to fail.
    return this.createChatCompletion(request);
  }

  // BUG: The "safe" method is actually LESS safe than calling
  // createChatCompletion directly, because:
  // 1. It doubles the number of API calls (rate limit risk)
  // 2. It introduces a TOCTOU window
  // 3. It gives callers false confidence that the model check prevents errors
  // 4. It does not handle the 404 that occurs when the model is deprecated
  //    between check and use
}
```

```typescript
// B15 test -- demonstrates the false safety:
describe("createChatCompletionSafe", () => {
  it("verifies model before sending request", async () => {
    // Mock: model exists in list
    global.fetch = jest.fn()
      .mockResolvedValueOnce({
        ok: true, status: 200,
        json: async () => ({
          object: "list",
          data: [{ id: "gpt-4o", object: "model", created: 1700000000, owned_by: "openai" }],
        }),
        headers: new Headers({}),
      })
      // Mock: completion succeeds
      .mockResolvedValueOnce({
        ok: true, status: 200,
        json: async () => VALID_COMPLETION,
        headers: new Headers({}),
      });

    const client = new OpenAIClient("sk-test123");
    const result = await client.createChatCompletionSafe({
      model: "gpt-4o",
      messages: [{ role: "user", content: "Hi" }],
    });

    expect(result.id).toMatch(/^chatcmpl-/);
    // BUG: This test does not cover the race condition.
    // It mocks both calls as successful, which is the non-race path.
    // No test for: model check succeeds, then completion returns 404.
    // No test for: concurrent calls exhausting rate limit on listModels.
  });
});
```

**Expected violation:** `CTR-request-shape` -- TOCTOU pattern detected: `listModels()` result is used as a precondition for `createChatCompletion()` without atomicity guarantee. The model availability check and the model usage are two separate API calls with no transactional binding. Additionally, `TQ-negative-cases` -- the test does not cover the race condition path (model check passes, completion returns 404 due to deprecation between calls).

**Production impact:** Under normal conditions, this works. During an OpenAI model deprecation event (which has happened multiple times), the "safe" method gives false confidence. The listModels call succeeds, the completion call returns 404, and the error handler does not expect 404 because the precondition "guaranteed" the model exists. The doubled API calls also increase rate limit consumption by 2x, making the application more likely to hit 429 errors during high-traffic periods.
