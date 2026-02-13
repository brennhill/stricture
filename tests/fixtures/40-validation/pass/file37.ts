// src/auth/resource-guard.ts
export async function accessResource(userId: string, resourceId: string): Promise<Resource> {
  const user = await userService.getUser(userId);
  if (!user) {
    throw new Error('User not found');
  }

  const resource = await resourceService.getResource(resourceId);
  if (resource.ownerId !== userId) {
    throw new Error('Unauthorized: You do not own this resource');
  }

  return resource;
}
