// src/parsers/response-parser.ts
export async function parseResponse(response: Response): Promise<ParsedData> {
  try {
    const json = await response.json();
    return {
      id: json.id,
      value: json.value,
      timestamp: new Date(json.timestamp),
    };
  } catch (error) {
    throw new Error('Failed to parse response: malformed JSON');
  }
}
