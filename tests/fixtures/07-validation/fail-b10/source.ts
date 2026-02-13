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
