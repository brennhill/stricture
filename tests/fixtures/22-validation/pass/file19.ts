// BUG: Missing "pre_order" enum value
type ProductStatus = "available" | "out_of_stock" | "discontinued";
// MISSING: "pre_order"

interface Product {
    status: ProductStatus;
}

async function filterByStatus(status: ProductStatus): Promise<Product[]> {
    const response = await fetch(`${API_BASE}/api/products?status=${status}`);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    const data = await response.json();
    return data.items;
}

// Usage:
const products = await filterByStatus("pre_order");
// TypeScript error: "pre_order" not assignable to ProductStatus
// But server DOES accept "pre_order" — client/server enum mismatch
