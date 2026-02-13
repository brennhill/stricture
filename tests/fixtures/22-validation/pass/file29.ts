// BUG: No Authorization header — server requires Bearer token
async function createProduct(req: CreateProductRequest): Promise<Product> {
    const response = await fetch(`${API_BASE}/api/products`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            // MISSING: "Authorization": "Bearer valid-token-123"
        },
        body: JSON.stringify(req),
    });
    if (!response.ok) throw new Error(`HTTP ${response.status}`);  // Will get 401
    return response.json();
}
