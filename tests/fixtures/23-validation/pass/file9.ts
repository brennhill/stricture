async createTask(request: CreateTaskRequest): Promise<Task> {
    // BUG: No try/catch wrapper
    const response = await fetch(`${this.baseUrl}/api/tasks`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${this.authToken}`,
        },
        body: JSON.stringify(request),
    });

    if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`HTTP ${response.status}: ${errorText || response.statusText}`);
    }

    const data = await response.json();
    return this.mapResponseToTask(data);
}
