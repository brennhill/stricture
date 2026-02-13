// DEFECT B15: Missing If-Match header for optimistic locking
async function updateOrderStatus(
  orderId: string,
  status: OrderStatus,
  version: number
): Promise<Order> {
  const response = await fetch(`/api/orders/${orderId}/status`, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
      // BUG: Missing If-Match header
      // Should have: "If-Match": version.toString()
    },
    body: JSON.stringify({ status, version }),
  });

  return response.json();  // Server returns 428 Precondition Required
}

// Race condition scenario:
// User A fetches order (version=0)
// User B fetches order (version=0)
// User A updates status → version=1 (succeeds)
// User B updates status → should fail with 409 Conflict, but gets 428 instead
