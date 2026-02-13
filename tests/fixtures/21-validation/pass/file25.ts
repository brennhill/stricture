// DEFECT B12: Doesn't handle null before string operations
async function displayOrderNote(orderId: string): Promise<void> {
  const order = await client.getOrder(orderId);

  // BUG: order.note can be null, but code assumes it's always a string
  const trimmedNote = order.note.trim();  // Runtime error: Cannot read property 'trim' of null
  console.log(`Note: ${trimmedNote}`);

  // Should be:
  // const trimmedNote = order.note?.trim() ?? "No note";
}
