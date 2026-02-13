// tests/webhooks/handler.test.ts
it('should process webhook event', async () => {
  // TQ-schema-conformance: Mock missing 'signature' field
  const mockEvent = {
    eventId: 'evt-123',
    eventType: 'user.created',
    payload: { userId: 'user-456' },
    timestamp: 1234567890,
    // Missing: signature (required by WebhookEvent)
  };

  await handleWebhook(mockEvent as WebhookEvent);
  expect(eventProcessor.received).toHaveBeenCalled();
});
