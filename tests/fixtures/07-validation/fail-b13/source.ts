// B13: Missing API key validation -- accepts any string as API key.
export class OpenAIClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;

  constructor(apiKey: string) {
    // BUG: No validation of apiKey format.
    // The manifest declares auth format: "sk-*" but the constructor
    // accepts any string. This means:
    // 1. A session token ("sess_abc123") gets sent as Bearer token
    // 2. A password ("hunter2") gets sent to OpenAI's servers
    // 3. An empty string ("") produces "Authorization: Bearer "
    // 4. A different provider's key ("anthropic-sk-...") gets leaked
    this.apiKey = apiKey;
    this.baseUrl = "https://api.openai.com";
  }

  async createChatCompletion(request: ChatCompletionRequest): Promise<ChatCompletionResponse> {
    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/chat/completions`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          // BUG: Sends whatever string was provided, even if it is
          // a Stripe key, AWS secret, or database password.
          // OpenAI's servers receive the credential in plaintext.
          "Authorization": `Bearer ${this.apiKey}`,
        },
        body: JSON.stringify(request),
      });
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) {
      // The 401 error here is the ONLY indication that the key was wrong.
      // By this point, the wrong credential has already been transmitted.
      const errorBody: OpenAIError = await response.json();
      throw new Error(errorBody.error.message);
    }

    return await response.json();
  }
}

// Example of how this bug manifests:
// const client = new OpenAIClient(process.env.STRIPE_SECRET_KEY!);
// Sends Stripe key to OpenAI -- credential leaked to third party.
