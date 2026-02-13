interface Task {
    id: number;
    title: string;
    description: string;
    assignee: string;
    status: TaskStatus;
    priority: Priority;
    dueDate: string;
    createdAt: string;
    updatedAt: string;
    version: number;
}

type TaskStatus = "BACKLOG" | "TODO" | "IN_PROGRESS" | "IN_REVIEW" | "DONE";
type Priority = "LOW" | "MEDIUM" | "HIGH" | "CRITICAL";

interface CreateTaskRequest {
    title: string;
    description: string;
    assignee: string;
    status: TaskStatus;
    priority: Priority;
    due_date: string; // CRITICAL: snake_case for JSON
}

interface UpdateTaskRequest {
    title?: string;
    description?: string;
    assignee?: string;
    status?: TaskStatus;
    priority?: Priority;
    due_date?: string; // CRITICAL: snake_case for JSON
    version: number;
}

interface PaginatedResponse<T> {
    data: T[];
    page: number;
    size: number;
    total: number;
}

class TaskClient {
    constructor(
        private readonly baseUrl: string,
        private readonly authToken: string
    ) {}

    async createTask(request: CreateTaskRequest): Promise<Task> {
        try {
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
                throw new Error(
                    `HTTP ${response.status}: ${errorText || response.statusText}`
                );
            }

            const data = await response.json();
            return this.mapResponseToTask(data);
        } catch (error) {
            if (error instanceof Error) {
                throw new Error(`Failed to create task: ${error.message}`);
            }
            throw error;
        }
    }

    async getTask(id: number): Promise<Task> {
        try {
            const response = await fetch(`${this.baseUrl}/api/tasks/${id}`, {
                method: "GET",
                headers: {
                    "Authorization": `Bearer ${this.authToken}`,
                },
            });

            if (response.status === 404) {
                throw new Error(`Task not found: ${id}`);
            }

            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(
                    `HTTP ${response.status}: ${errorText || response.statusText}`
                );
            }

            const data = await response.json();
            return this.mapResponseToTask(data);
        } catch (error) {
            if (error instanceof Error) {
                throw new Error(`Failed to get task: ${error.message}`);
            }
            throw error;
        }
    }

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
                    "Authorization": `Bearer ${this.authToken}`,
                },
            });

            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(
                    `HTTP ${response.status}: ${errorText || response.statusText}`
                );
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

    async updateTask(id: number, request: UpdateTaskRequest): Promise<Task> {
        try {
            const response = await fetch(`${this.baseUrl}/api/tasks/${id}`, {
                method: "PATCH",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": `Bearer ${this.authToken}`,
                },
                body: JSON.stringify(request),
            });

            if (response.status === 404) {
                throw new Error(`Task not found: ${id}`);
            }

            if (response.status === 409) {
                const errorText = await response.text();
                throw new Error(`Version conflict: ${errorText}`);
            }

            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(
                    `HTTP ${response.status}: ${errorText || response.statusText}`
                );
            }

            const data = await response.json();
            return this.mapResponseToTask(data);
        } catch (error) {
            if (error instanceof Error) {
                throw new Error(`Failed to update task: ${error.message}`);
            }
            throw error;
        }
    }

    async deleteTask(id: number): Promise<void> {
        try {
            const response = await fetch(`${this.baseUrl}/api/tasks/${id}`, {
                method: "DELETE",
                headers: {
                    "Authorization": `Bearer ${this.authToken}`,
                },
            });

            if (response.status === 404) {
                throw new Error(`Task not found: ${id}`);
            }

            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(
                    `HTTP ${response.status}: ${errorText || response.statusText}`
                );
            }
        } catch (error) {
            if (error instanceof Error) {
                throw new Error(`Failed to delete task: ${error.message}`);
            }
            throw error;
        }
    }

    private mapResponseToTask(data: unknown): Task {
        // Type guard: ensure response has expected structure
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
            dueDate: data.due_date, // Map snake_case to camelCase
            createdAt: data.created_at,
            updatedAt: data.updated_at,
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
        due_date: string;
        created_at: string;
        updated_at: string;
        version: number;
    } {
        if (typeof data !== "object" || data === null) return false;
        const obj = data as Record<string, unknown>;

        return (
            typeof obj.id === "number" &&
            typeof obj.title === "string" &&
            typeof obj.description === "string" &&
            typeof obj.assignee === "string" &&
            typeof obj.status === "string" &&
            typeof obj.priority === "string" &&
            typeof obj.due_date === "string" &&
            typeof obj.created_at === "string" &&
            typeof obj.updated_at === "string" &&
            typeof obj.version === "number"
        );
    }
}

export { TaskClient, Task, TaskStatus, Priority, CreateTaskRequest, UpdateTaskRequest };
