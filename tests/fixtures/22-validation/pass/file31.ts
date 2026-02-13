// BUG: Field name mismatch (total vs totalCount)
interface PaginatedProducts {
    items: Product[];
    totalCount: number;  // Server sends "total", not "totalCount"
    offset: number;
    limit: number;
    hasMore: boolean;    // Server sends "has_more", not "hasMore"
}

async function listProducts(offset: number, limit: number): Promise<PaginatedProducts> {
    const response = await fetch(`${API_BASE}/api/products?offset=${offset}&limit=${limit}`);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    return response.json();  // Runtime: {total: 42, has_more: true}
}

// Usage:
const result = await listProducts(0, 20);
console.log(`Total: ${result.totalCount}`);  // undefined — actual key is "total"
