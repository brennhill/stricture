// B08: Incomplete enum handling -- only handles "stop" finish_reason.
private handleFinishReason(reason: string): void {
  // BUG: Only handles "stop". The API returns 4 possible values:
  // "stop", "length", "content_filter", "tool_calls", plus null for streaming chunks.
  // Missing "length" means truncated responses are silently accepted.
  // Missing "content_filter" means filtered content is treated as valid.
  // Missing "tool_calls" means function calling responses are not processed.
  if (reason === "stop") {
    return; // Normal completion
  }
  // All other finish_reasons fall through with no handling.
  // No warning, no error, no logging.
}

async createChatCompletion(request: ChatCompletionRequest): Promise<ChatCompletionResponse> {
  // ...
  for (const choice of completion.choices) {
    this.handleFinishReason(choice.finish_reason);

    // BUG: Assumes content is always a string for all finish_reasons.
    // When finish_reason is "tool_calls", content is null.
    const text = choice.message.content.toUpperCase(); // null.toUpperCase() -> TypeError
  }
  return completion;
}
