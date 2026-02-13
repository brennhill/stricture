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
