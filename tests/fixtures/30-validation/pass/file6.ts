import { sendWelcomeNotification } from "./notification-service";

export async function createUser(input: CreateUserInput): Promise<User> {
  // ... (create user logic) ...
  const user = await userRepo.createUser(hashedInput);
  await sendWelcomeNotification(user.id, user.email);  // <-- starts the chain
  return user;
}
