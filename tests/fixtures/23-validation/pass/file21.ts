async listTasks(page = 0, size = 20): Promise<PaginatedResponse<Task>> {
    try {
        if (size > 100) {
            throw new Error("Page size must not exceed 100");
        }

        const url = new URL(`${this.baseUrl}/api/tasks`);
        url.searchParams.set("page", page.toString());
        url.searchParams.set("size", size.toString());

        const response = await fetch(url.toString(), {
            method: "GET",
            headers: {
                // BUG: Missing Authorization header
                "Content-Type": "application/json",
            },
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`HTTP ${response.status}: ${errorText || response.statusText}`);
        }

        const data = await response.json();
        return {
            data: data.data.map((item: unknown) => this.mapResponseToTask(item)),
            page: data.page,
            size: data.size,
            total: data.total,
        };
    } catch (error) {
        if (error instanceof Error) {
            throw new Error(`Failed to list tasks: ${error.message}`);
        }
        throw error;
    }
}
