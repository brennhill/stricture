// src/client/api-client.ts — Ignores hasMore flag
async getAllUsers(): Promise<User[]> {
  try {
    const firstPage = await this.listUsers(undefined, 10);

    // BUG: Only returns first page, ignores hasMore flag
    // If hasMore=true, client should loop and fetch next pages using cursor
    return firstPage.data;

    // Correct implementation:
    // const allUsers = [...firstPage.data];
    // let cursor = firstPage.cursor;
    // while (firstPage.hasMore && cursor) {
    //   const nextPage = await this.listUsers(cursor, 10);
    //   allUsers.push(...nextPage.data);
    //   cursor = nextPage.cursor;
    // }
    // return allUsers;
  } catch (err) {
    throw new Error(`Failed to get all users: ${(err as Error).message}`);
  }
}
