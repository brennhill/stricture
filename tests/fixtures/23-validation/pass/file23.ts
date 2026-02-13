interface UpdateTaskRequest {
    title?: string;
    description?: string;
    assignee?: string;
    status?: TaskStatus;
    priority?: Priority;
    due_date?: string;
    version?: number; // BUG: Should be required, not optional
}

async updateTask(id: number, request: UpdateTaskRequest): Promise<Task> {
    try {
        // BUG: No check that request.version is provided
        const response = await fetch(`${this.baseUrl}/api/tasks/${id}`, {
            method: "PATCH",
            headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${this.authToken}`,
            },
            body: JSON.stringify(request), // Sends request with version=undefined
        });
        // ...
    }
}
