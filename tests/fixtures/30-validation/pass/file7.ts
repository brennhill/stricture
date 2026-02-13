// notification-service.ts -- Notification dispatch.
// V05: Part of a 3-node circular dependency.

import { logger } from "../shared/logger";
import { getUserById } from "./user-service";  // <-- COMPLETES THE CYCLE

export async function sendWelcomeNotification(userId: string, email: string): Promise<void> {
  logger.info("Sending welcome notification", { userId, email });

  // BUG: This service imports from user-service to "enrich" the notification
  // with user details. But user-service imports from notification-service,
  // creating: user-service -> notification-service -> user-service.
  const user = await getUserById(userId);
  logger.info("Notification sent", { userName: user.name });
}

export async function sendOrderNotification(userId: string, orderId: string): Promise<void> {
  const user = await getUserById(userId);
  logger.info("Order notification sent", { userName: user.name, orderId });
}
