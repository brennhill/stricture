describe("TaskClient", () => {
    const baseUrl = "http://localhost:8080";
    const authToken = "test-token-123";
    let client: TaskClient;

    beforeEach(() => {
        client = new TaskClient(baseUrl, authToken);
    });

    it("creates task with valid request", async () => {
        // ... only happy path test
    });

    it("updates task with version", async () => {
        // ... only happy path test
    });

    // BUG: Missing error tests for:
    // - 401 Unauthorized
    // - 400 Bad Request (invalid UUID, title too long)
    // - 404 Not Found
    // - 409 Conflict (version mismatch)
});
