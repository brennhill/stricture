// DEFECT B08: TS only handles 3 of 5 statuses
test("order status transitions", async () => {
  const order = await client.createOrder(validRequest);

  // BUG: Only tests 3 statuses, missing "processing" and "delivered"
  switch (order.status) {
    case "pending":
      // test logic
      break;
    case "confirmed":
      // test logic
      break;
    case "shipped":
      // test logic
      break;
    // MISSING: "processing" and "delivered" cases
    default:
      throw new Error("Unknown status");  // Will crash on valid statuses
  }
});
