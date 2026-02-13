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
