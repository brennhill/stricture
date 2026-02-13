// BUG: description is string, not string | null
interface Product {
    id: string;
    product_name: string;
    description: string;  // WRONG — should be "string | null"
}

async function getProduct(productId: string): Promise<Product> {
    const response = await fetch(`${API_BASE}/api/products/${productId}`);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    return response.json();
}

// Usage:
const product = await getProduct("123");
console.log(product.description.toUpperCase());  // CRASH if description is null
