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
