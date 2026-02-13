// B11: Precision loss -- token counts accumulated in float accumulator.
export class OpenAIClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;
  // BUG: tokenUsage is typed as number (float64 in JS).
  // For small counts this works fine. But after thousands of requests,
  // the accumulated total can exceed Number.MAX_SAFE_INTEGER territory
  // or accumulate floating-point drift if any intermediate computation
  // introduces a non-integer.
  private tokenUsage: { prompt: number; completion: number; total: number };

  constructor(apiKey: string) {
    validateApiKey(apiKey);
    this.apiKey = apiKey;
    this.baseUrl = "https://api.openai.com";
    this.tokenUsage = { prompt: 0.0, completion: 0.0, total: 0.0 };
  }

  async createChatCompletion(request: ChatCompletionRequest): Promise<ChatCompletionResponse> {
    // ... fetch and validation ...
    const completion: ChatCompletionResponse = await response.json();

    // BUG: Using += on floats. While token counts from the API are integers,
    // if any middleware or transformation introduces a fractional value
    // (e.g., dividing by a rate), the accumulator drifts.
    // More critically: at scale, after millions of requests, even pure
    // integer addition in float64 loses precision above 2^53.
    //
    // Example scenario:
    //   this.tokenUsage.total = 9007199254740992; // 2^53
    //   this.tokenUsage.total += 1; // Still 9007199254740992 (precision lost)
    this.tokenUsage.prompt += completion.usage.prompt_tokens;
    this.tokenUsage.completion += completion.usage.completion_tokens;
    this.tokenUsage.total += completion.usage.total_tokens;

    // BUG: Cost calculation introduces fractional cents.
    // $0.01 per 1000 tokens: 18 tokens = 0.00018
    // After 10000 requests: 0.00018 * 10000 = 1.7999999999999998 (not 1.8)
    const costUsd = this.tokenUsage.total * 0.00001;
    console.log(`Total cost: $${costUsd}`);

    return completion;
  }
}
