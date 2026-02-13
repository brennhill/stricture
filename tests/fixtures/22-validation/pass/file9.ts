// BUG: Shallow assertion — doesn't verify contract fields
test("getProduct returns product", async () => {
    const product = await client.getProduct("123e4567-e89b-12d3-a456-426614174000");
    expect(product).toBeDefined();  // SHALLOW — what about id, created_at, version?
});
