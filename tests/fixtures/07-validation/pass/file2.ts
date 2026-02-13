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
