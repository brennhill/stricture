// B15 test -- demonstrates the false safety:
describe("createChatCompletionSafe", () => {
  it("verifies model before sending request", async () => {
    // Mock: model exists in list
    global.fetch = jest.fn()
      .mockResolvedValueOnce({
        ok: true, status: 200,
        json: async () => ({
          object: "list",
          data: [{ id: "gpt-4o", object: "model", created: 1700000000, owned_by: "openai" }],
        }),
        headers: new Headers({}),
      })
      // Mock: completion succeeds
      .mockResolvedValueOnce({
        ok: true, status: 200,
        json: async () => VALID_COMPLETION,
        headers: new Headers({}),
      });

    const client = new OpenAIClient("sk-test123");
    const result = await client.createChatCompletionSafe({
      model: "gpt-4o",
      messages: [{ role: "user", content: "Hi" }],
    });

    expect(result.id).toMatch(/^chatcmpl-/);
    // BUG: This test does not cover the race condition.
    // It mocks both calls as successful, which is the non-race path.
    // No test for: model check succeeds, then completion returns 404.
    // No test for: concurrent calls exhausting rate limit on listModels.
  });
});
