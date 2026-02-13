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
