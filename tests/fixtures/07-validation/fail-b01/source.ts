// B01: No error handling -- fetch failures crash the process.
async createChatCompletion(request: ChatCompletionRequest): Promise<ChatCompletionResponse> {
  validateMessages(request.messages);
  validateTemperature(request.temperature);

  // BUG: no try/catch -- ECONNREFUSED, DNS failures, and timeouts
  // produce unhandled rejections that crash the Node.js process.
  const response = await fetch(`${this.baseUrl}/v1/chat/completions`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "Authorization": `Bearer ${this.apiKey}`,
    },
    body: JSON.stringify({ ...request, stream: false }),
  });

  const completion: ChatCompletionResponse = await response.json();
  return completion;
}
