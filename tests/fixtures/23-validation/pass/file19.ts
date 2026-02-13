private mapResponseToTask(data: unknown): Task {
    if (!this.isTaskResponse(data)) {
        throw new Error("Invalid task response structure");
    }

    return {
        id: data.id,
        title: data.title,
        description: data.description,
        assignee: data.assignee,
        status: data.status,
        priority: data.priority,
        dueDate: data.dueDate, // BUG: Java sends due_date, not dueDate
        createdAt: data.createdAt, // BUG: Java sends created_at, not createdAt
        updatedAt: data.updatedAt, // BUG: Java sends updated_at, not updatedAt
        version: data.version,
    };
}

private isTaskResponse(data: unknown): data is {
    id: number;
    title: string;
    description: string;
    assignee: string;
    status: TaskStatus;
    priority: Priority;
    dueDate: string; // BUG: Should be due_date
    createdAt: string; // BUG: Should be created_at
    updatedAt: string; // BUG: Should be updated_at
    version: number;
} {
    // ... type guard expecting camelCase fields
}
