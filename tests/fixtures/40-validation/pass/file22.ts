// tests/services/order-service.test.ts
it('should get order items', () => {
  const items = getOrderItems('order-456');

  // TQ-assertion-depth: Only checks array length
  expect(items).toHaveLength(3);
  // Missing: items[0].productId, items[0].quantity, items[0].price, etc.
});
