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
