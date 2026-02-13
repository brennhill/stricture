// BUG: No runtime validation for discount_percentage range
interface CreateProductRequest {
    product_name: string;
    price: number;
    stock_quantity: number;
    category: string;
    discount_percentage?: number;  // No ge=0, le=100 constraint
}

async function createProduct(req: CreateProductRequest): Promise<Product> {
    // BUG: Sends invalid discount_percentage without client-side validation
    const response = await fetch(`${API_BASE}/api/products`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(req),  // Could send discount_percentage: 150
    });
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    return response.json();
}

// Usage:
await createProduct({
    product_name: "Widget",
    price: 99.99,
    stock_quantity: 50,
    category: "Tools",
    discount_percentage: 150,  // Server will reject with 422 — should fail client-side
});
