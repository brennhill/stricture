interface Task {
    id: number; // BUG: JavaScript number has 53-bit precision, Java Long is 64-bit
    title: string;
    // ... other fields
}

private mapResponseToTask(data: unknown): Task {
    if (!this.isTaskResponse(data)) {
        throw new Error("Invalid task response structure");
    }

    return {
        id: data.id, // BUG: If Java sends Long > Number.MAX_SAFE_INTEGER, precision lost
        title: data.title,
        // ...
    };
}
