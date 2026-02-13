// B07: Wrong field types -- created timestamp stored as string.
interface ChatCompletionResponse {
  id: string;
  object: "chat.completion";
  // BUG: created is typed as string, but the API returns a Unix timestamp
  // as a number (e.g., 1700000000). When deserialized from JSON, the value
  // IS a number, but the TypeScript type says string. Code that does
  // `new Date(completion.created * 1000)` gets NaN because
  // TypeScript allows string * number (NaN).
  created: string;
  model: string;
  choices: ChatCompletionChoice[];
  usage: {
    prompt_tokens: number;
    completion_tokens: number;
    total_tokens: number;
  };
}

function formatCompletionTimestamp(completion: ChatCompletionResponse): string {
  // BUG: If created is treated as string by downstream code:
  // new Date("1700000000" * 1000) works in JS (implicit coercion)
  // but new Date(completion.created + "000") produces "1700000000000"
  // which is a valid date by accident -- masking the type error.
  const timestamp = Number(completion.created) * 1000;
  return new Date(timestamp).toISOString();
}
