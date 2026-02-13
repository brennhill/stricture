// BUG: version sent in body, but should use If-Match header per REST best practices
interface UpdateProductRequest {
    product_name?: string;
    price?: number;
    version: number;  // Should be If-Match header, not body field
}

async function updateProduct(
    productId: string,
    updates: UpdateProductRequest
): Promise<Product> {
    const response = await fetch(`${API_BASE}/api/products/${productId}`, {
        method: "PUT",
        headers: {
            "Content-Type": "application/json",
            // MISSING: "If-Match": String(updates.version)
        },
        body: JSON.stringify(updates),  // version in body works, but not idiomatic
    });
    if (!response.ok) {
        if (response.status === 409) {
            throw new Error("Concurrent modification — retry with latest version");
        }
        throw new Error(`HTTP ${response.status}`);
    }
    return response.json();
}
