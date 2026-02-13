interface CreateTaskRequest {
    title: string;
    description: string;
    // BUG: Missing assignee field
    status: TaskStatus;
    priority: Priority;
    due_date: string;
}

async createTask(request: CreateTaskRequest): Promise<Task> {
    try {
        const response = await fetch(`${this.baseUrl}/api/tasks`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${this.authToken}`,
            },
            body: JSON.stringify(request), // Sends request without assignee
        });
        // ...
    }
}
