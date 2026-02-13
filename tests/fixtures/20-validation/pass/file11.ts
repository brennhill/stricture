// tests/server/users.test.ts (has negative test)
it("should return 404 for non-existent user", async () => {
  const response = await request(app).get("/api/users/non-existent-id");
  expect(response.status).toBe(404);
  expect(response.body.error).toContain("user not found");
});
