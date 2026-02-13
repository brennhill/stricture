// billing-service.ts -- Billing orchestration.
// V07: Imports directly from billing module's internal directory.

import { logger } from "../shared/logger";
import { ValidationError } from "../shared/errors";

// VIOLATION: Importing from the internal directory of the billing module.
// The billing module's public API is at src/modules/billing/index.ts, which
// re-exports createPaymentIntent, capturePayment, refundPayment.
// This import bypasses the public API and reaches into the internal directory.
import { createPaymentIntent } from "../modules/billing/internal/stripe-adapter";  // <-- VIOLATION

export async function processPayment(
  orderId: string,
  amountCents: number,
  currency: string,
  customerId: string
): Promise<{ paymentId: string; status: string }> {
  if (amountCents <= 0) {
    throw new ValidationError("Amount must be positive");
  }

  logger.info("Processing payment", { orderId, amountCents, currency });

  // BUG: By importing from internal/stripe-adapter directly, this service
  // is coupled to the Stripe implementation. If billing switches to a
  // different payment provider, this import path breaks even though the
  // public API (index.ts) would remain stable.
  const result = await createPaymentIntent(amountCents, currency, customerId);

  return { paymentId: result.paymentId, status: result.status };
}

export async function refundOrder(
  orderId: string,
  paymentId: string,
  amountCents?: number
): Promise<void> {
  logger.info("Refunding order", { orderId, paymentId, amountCents });

  // This import is also wrong but we only need to demonstrate one.
  // In the PERFECT version, this would be:
  //   import { refundPayment } from "../modules/billing";
  const { refundPayment } = await import("../modules/billing/internal/stripe-adapter");
  await refundPayment(paymentId, amountCents);
}
