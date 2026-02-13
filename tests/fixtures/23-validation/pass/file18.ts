async createTask(request: CreateTaskRequest): Promise<Task> {
    try {
        // BUG: No UUID validation - Java expects v4 UUID regex
        // Allows assignee="not-a-uuid" or assignee="12345"

        const response = await fetch(`${this.baseUrl}/api/tasks`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${this.authToken}`,
            },
            body: JSON.stringify(request),
        });
        // ...
    }
}
