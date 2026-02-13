// BUG: Only tests happy path — no 422 validation error tests
describe("createProduct", () => {
    test("creates product successfully", async () => {
        const product = await client.createProduct({
            product_name: "Widget",
            price: 19.99,
            stock_quantity: 100,
            category: "Tools",
        });
        expect(product.id).toBeDefined();
    });

    // MISSING: test("returns 422 for invalid price", ...)
    // MISSING: test("returns 422 for empty product_name", ...)
    // MISSING: test("returns 422 for negative stock_quantity", ...)
});
