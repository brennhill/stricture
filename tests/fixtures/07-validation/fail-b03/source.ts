// B03: Shallow assertions -- tests prove nothing about correctness.
describe("OpenAIClient", () => {
  it("creates a completion", async () => {
    mockFetchSuccess(VALID_COMPLETION);
    const client = new OpenAIClient("sk-test123");

    const result = await client.createChatCompletion({
      model: "gpt-4o",
      messages: [{ role: "user", content: "Hi" }],
    });

    // BUG: These assertions verify existence, not correctness.
    // The response could have wrong id format, wrong object type,
    // missing usage, zero-length choices -- all would pass.
    expect(result).toBeDefined();
    expect(result.choices).toBeTruthy();
    expect(result.usage).not.toBeNull();
  });

  it("handles errors", async () => {
    mockFetchError(400, VALID_ERROR_BODY);
    const client = new OpenAIClient("sk-test123");

    // BUG: Only checks that it throws, not WHAT it throws.
    // Could throw "undefined is not a function" and this passes.
    await expect(
      client.createChatCompletion({ model: "gpt-4o", messages: [{ role: "user", content: "Hi" }] })
    ).rejects.toBeDefined();
  });
});
