// src/webhooks/types.ts
export interface WebhookEvent {
  eventId: string;
  eventType: string;
  payload: Record<string, unknown>;
  timestamp: number;
  signature: string;
}
