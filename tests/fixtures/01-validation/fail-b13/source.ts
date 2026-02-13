// B13: Accepts Stripe webhook payloads without verifying the signature.

import type { Request, Response as ExpressResponse } from "express";

// BUG: No signature verification at all.
// An attacker can POST a fake webhook event to trigger actions:
//   - Fake "charge.succeeded" to mark unpaid orders as paid
//   - Fake "customer.updated" to change account details
//   - Fake "payment_intent.succeeded" to fulfill unprocessed orders

async function handleStripeWebhook(req: Request, res: ExpressResponse): Promise<void> {
  // BUG: The Stripe-Signature header is completely ignored.
  // The manifest specifies verification:
  //   method: hmac-sha256
  //   header: Stripe-Signature
  //   format: "t={timestamp},v1={signature}"
  //   tolerance_seconds: 300
  // None of this is checked.

  const event = req.body;  // BUG: Raw body parsed as trusted event.

  switch (event.type) {
    case "charge.succeeded": {
      const charge = event.data.object as StripeCharge;
      await markOrderAsPaid(charge.id, charge.amount);
      break;
    }
    case "charge.failed": {
      const charge = event.data.object as StripeCharge;
      await notifyPaymentFailed(charge.id, charge.failure_message);
      break;
    }
    case "payment_intent.succeeded": {
      const intent = event.data.object as StripePaymentIntent;
      await fulfillOrder(intent.id, intent.amount);
      break;
    }
    default:
      // Ignore unhandled event types
      break;
  }

  res.status(200).json({ received: true });
}

// Placeholder functions -- in production these modify real data.
async function markOrderAsPaid(chargeId: string, amount: number): Promise<void> { /* ... */ }
async function notifyPaymentFailed(chargeId: string, message: string | null): Promise<void> { /* ... */ }
async function fulfillOrder(intentId: string, amount: number): Promise<void> { /* ... */ }

export { handleStripeWebhook };
