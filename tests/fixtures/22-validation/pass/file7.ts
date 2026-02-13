// BUG: No response.ok check — 404 error treated as valid Product
async function getProduct(productId: string): Promise<Product> {
    try {
        const response = await fetch(`${API_BASE}/api/products/${productId}`);
        // Missing: if (!response.ok) { throw new Error(...) }
        const data = await response.json();
        return data;  // Could be {detail: "not found"} instead of Product
    } catch (error) {
        throw new Error(`Failed to fetch product: ${error.message}`);
    }
}
