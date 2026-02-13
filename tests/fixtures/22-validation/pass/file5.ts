// BUG: No try/catch — network errors crash the app
async function getProduct(productId: string): Promise<Product> {
    const response = await fetch(`${API_BASE}/api/products/${productId}`);
    const data = await response.json();
    return data;
}
