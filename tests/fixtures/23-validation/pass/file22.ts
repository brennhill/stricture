async listTasks(page = 0, size = 20): Promise<PaginatedResponse<Task>> {
    try {
        const url = new URL(`${this.baseUrl}/api/tasks`);
        url.searchParams.set("page", page.toString()); // BUG: Sends page=0, Java interprets as page 0
        url.searchParams.set("size", size.toString());

        // If Java uses 1-based indexing (page 1 = first page),
        // this will request the wrong page
        const response = await fetch(url.toString(), {
            method: "GET",
            headers: {
                "Authorization": `Bearer ${this.authToken}`,
            },
        });
        // ...
    }
}
