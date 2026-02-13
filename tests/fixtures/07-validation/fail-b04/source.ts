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
