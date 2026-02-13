// tests/parsers/json-parser.test.ts
it('should parse JSON data', () => {
  const input = '{"id":"123","version":"1.0","ts":1234567890,"items":["a","b"]}';
  const result = parseData(input);

  // TQ-return-type-verified: Only checks one field, not full return type
  expect(result.id).toBe('123');
  // Missing: metadata.version, metadata.timestamp, items array content
});
