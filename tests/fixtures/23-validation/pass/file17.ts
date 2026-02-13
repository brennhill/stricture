async createTask(request: CreateTaskRequest): Promise<Task> {
    try {
        // BUG: No validation - Java expects @Size(min=1, max=200)
        // Allows title="" or title="a".repeat(500)

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
