interface CreateTaskRequest {
    title: string;
    description: string;
    assignee: string;
    status: TaskStatus;
    priority: Priority;
    dueDate: string; // BUG: Should be due_date for Java @JsonProperty
}

async createTask(request: CreateTaskRequest): Promise<Task> {
    try {
        const response = await fetch(`${this.baseUrl}/api/tasks`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${this.authToken}`,
            },
            body: JSON.stringify(request), // Sends {"dueDate": "..."} instead of {"due_date": "..."}
        });
        // ...
    }
}
