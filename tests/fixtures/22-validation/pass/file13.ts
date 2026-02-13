// BUG: Missing required field "category"
interface CreateProductRequest {
    product_name: string;
    price: number;
    stock_quantity: number;
    // MISSING: category (required by Pydantic)
}

async function createProduct(req: CreateProductRequest): Promise<Product> {
    const response = await fetch(`${API_BASE}/api/products`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(req),  // Server will return 422 — category missing
    });
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    return response.json();
}
