// tests/server/users.test.ts (correct, deep assertions)
it("should create user with valid fields", async () => {
  const response = await request(app)
    .post("/api/users")
    .send({ name: "Alice", email: "alice@example.com", role: "editor" });

  expect(response.status).toBe(201);
  expect(response.body).toMatchObject({
    name: "Alice",
    email: "alice@example.com",
    role: "editor",
  });
  expect(response.body.id).toBeDefined();
  expect(response.body.avatar).toBeNull();
});
