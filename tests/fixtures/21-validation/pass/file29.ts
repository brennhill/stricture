// DEFECT B14: Uses offset pagination instead of cursor
async function listOrders(page: number = 1, limit: number = 20): Promise<Order[]> {
  const offset = (page - 1) * limit;

  // BUG: Sends offset/limit params, but Go expects cursor param
  const response = await fetch(
    `/api/orders?offset=${offset}&limit=${limit}`,  // WRONG params
    { method: "GET" }
  );

  const data = await response.json();
  return data.orders;  // Ignores next_cursor field

  // Should use:
  // `/api/orders?cursor=${lastCursor}&limit=${limit}`
}
