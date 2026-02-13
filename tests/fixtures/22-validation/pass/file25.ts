// BUG: camelCase field names don't match server JSON (snake_case)
interface Product {
    id: string;
    productName: string;  // Server sends "product_name"
    createdAt: string;    // Server sends "created_at"
    updatedAt: string;    // Server sends "updated_at"
}

async function getProduct(productId: string): Promise<Product> {
    const response = await fetch(`${API_BASE}/api/products/${productId}`);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    const data = await response.json();
    return data;  // Runtime: {created_at: "...", product_name: "..."} — keys don't match
}

// Usage:
const product = await getProduct("123");
console.log(product.createdAt);  // undefined — actual key is "created_at"
