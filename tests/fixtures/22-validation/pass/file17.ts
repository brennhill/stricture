// BUG: created_at is Date type but server sends string
interface Product {
    id: string;
    product_name: string;
    created_at: Date;  // WRONG — server sends ISO string, not Date object
}

async function getProduct(productId: string): Promise<Product> {
    const response = await fetch(`${API_BASE}/api/products/${productId}`);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    const data = await response.json();
    return data;  // created_at is "2024-01-15T12:34:56.789Z" (string), not Date
}

// Usage:
const product = await getProduct("...");
console.log(product.created_at.getFullYear());  // CRASH — created_at is string
