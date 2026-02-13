async getTask(id: number): Promise<Task> {
    try {
        const response = await fetch(`${this.baseUrl}/api/tasks/${id}`, {
            method: "GET",
            headers: {
                "Authorization": `Bearer ${this.authToken}`,
            },
        });

        // BUG: Missing response.ok check - will call .json() on 404/500
        const data = await response.json();
        return this.mapResponseToTask(data);
    } catch (error) {
        if (error instanceof Error) {
            throw new Error(`Failed to get task: ${error.message}`);
        }
        throw error;
    }
}
