// BUG: "sku" field doesn't exist in Python response
interface Product {
    id: string;
    product_name: string;
    price: number;
    stock_quantity: number;
    category: string;
    created_at: string;
    updated_at: string;
    version: number;
    sku: string;  // NOT in Pydantic model — will be undefined at runtime
}

async function getProduct(productId: string): Promise<Product> {
    const response = await fetch(`${API_BASE}/api/products/${productId}`);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    return response.json();  // Runtime data missing "sku"
}
