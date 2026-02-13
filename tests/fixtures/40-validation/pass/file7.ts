// src/parsers/json-parser.ts
export interface ParsedData {
  id: string;
  metadata: {
    version: string;
    timestamp: number;
  };
  items: string[];
}

export function parseData(input: string): ParsedData {
  const raw = JSON.parse(input);
  return {
    id: raw.id,
    metadata: { version: raw.version, timestamp: raw.ts },
    items: raw.items || [],
  };
}
