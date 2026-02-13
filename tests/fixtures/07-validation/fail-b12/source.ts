// B12: Nullable field crash -- content.length on null content.
async createChatCompletion(request: ChatCompletionRequest): Promise<ChatCompletionResponse> {
  // ... fetch and error handling ...
  const completion: ChatCompletionResponse = await response.json();

  for (const choice of completion.choices) {
    // BUG: Accesses .length on content without null check.
    // When finish_reason is "tool_calls", the message contains:
    //   { role: "assistant", content: null, tool_calls: [...] }
    // Calling null.length throws:
    //   TypeError: Cannot read properties of null (reading 'length')
    //
    // This crashes the entire completion handler, even though the
    // tool_calls data is perfectly valid and usable.
    const contentLength = choice.message.content.length;
    console.log(`Response length: ${contentLength} characters`);

    if (choice.finish_reason === "length") {
      console.warn(`Truncated at ${contentLength} chars`);
    }
  }

  // BUG: Also crashes when building response summaries:
  const summary = completion.choices
    .map((c) => c.message.content.substring(0, 100)) // null.substring() -> TypeError
    .join("\n");

  return completion;
}
