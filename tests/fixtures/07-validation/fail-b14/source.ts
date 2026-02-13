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
