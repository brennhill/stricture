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
