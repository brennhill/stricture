import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { TaskClient, CreateTaskRequest, UpdateTaskRequest } from "../TaskClient";
import { setupServer } from "msw/node";
import { http, HttpResponse } from "msw";

const server = setupServer();

beforeEach(() => server.listen());
afterEach(() => server.close());

describe("TaskClient", () => {
    const baseUrl = "http://localhost:8080";
    const authToken = "test-token-123";
    let client: TaskClient;

    beforeEach(() => {
        client = new TaskClient(baseUrl, authToken);
    });

    it("creates task with valid request", async () => {
        const request: CreateTaskRequest = {
            title: "Implement feature X",
            description: "Add new API endpoint",
            assignee: "a1b2c3d4-5678-4abc-9def-0123456789ab",
            status: "TODO",
            priority: "HIGH",
            due_date: "2026-03-15",
        };

        server.use(
            http.post(`${baseUrl}/api/tasks`, async ({ request: req }) => {
                const body = await req.json();
                return HttpResponse.json({
                    id: 1,
                    title: body.title,
                    description: body.description,
                    assignee: body.assignee,
                    status: body.status,
                    priority: body.priority,
                    due_date: body.due_date,
                    created_at: "2026-02-13T10:00:00Z",
                    updated_at: "2026-02-13T10:00:00Z",
                    version: 1,
                }, { status: 201 });
            })
        );

        const task = await client.createTask(request);

        expect(task.id).toBe(1);
        expect(task.title).toBe("Implement feature X");
        expect(task.assignee).toBe("a1b2c3d4-5678-4abc-9def-0123456789ab");
        expect(task.dueDate).toBe("2026-03-15");
        expect(task.version).toBe(1);
    });

    it("throws error when creating task without auth header", async () => {
        server.use(
            http.post(`${baseUrl}/api/tasks`, () => {
                return HttpResponse.json(
                    { error: "Unauthorized" },
                    { status: 401 }
                );
            })
        );

        const request: CreateTaskRequest = {
            title: "Test",
            description: "Test",
            assignee: "a1b2c3d4-5678-4abc-9def-0123456789ab",
            status: "TODO",
            priority: "LOW",
            due_date: "2026-03-15",
        };

        await expect(client.createTask(request)).rejects.toThrow("HTTP 401");
    });

    it("throws error when UUID is invalid", async () => {
        server.use(
            http.post(`${baseUrl}/api/tasks`, () => {
                return HttpResponse.json(
                    { error: "Invalid UUID format" },
                    { status: 400 }
                );
            })
        );

        const request: CreateTaskRequest = {
            title: "Test",
            description: "Test",
            assignee: "not-a-uuid",
            status: "TODO",
            priority: "LOW",
            due_date: "2026-03-15",
        };

        await expect(client.createTask(request)).rejects.toThrow("HTTP 400");
    });

    it("updates task with version", async () => {
        const updateRequest: UpdateTaskRequest = {
            title: "Updated title",
            status: "IN_PROGRESS",
            version: 1,
        };

        server.use(
            http.patch(`${baseUrl}/api/tasks/1`, async ({ request: req }) => {
                const body = await req.json();
                return HttpResponse.json({
                    id: 1,
                    title: body.title,
                    description: "Original description",
                    assignee: "a1b2c3d4-5678-4abc-9def-0123456789ab",
                    status: body.status,
                    priority: "MEDIUM",
                    due_date: "2026-03-15",
                    created_at: "2026-02-13T10:00:00Z",
                    updated_at: "2026-02-13T10:30:00Z",
                    version: 2,
                });
            })
        );

        const task = await client.updateTask(1, updateRequest);

        expect(task.title).toBe("Updated title");
        expect(task.status).toBe("IN_PROGRESS");
        expect(task.version).toBe(2);
    });

    it("throws error when version conflicts", async () => {
        server.use(
            http.patch(`${baseUrl}/api/tasks/1`, () => {
                return HttpResponse.json(
                    { error: "Version mismatch" },
                    { status: 409 }
                );
            })
        );

        const updateRequest: UpdateTaskRequest = {
            title: "Updated",
            version: 1,
        };

        await expect(client.updateTask(1, updateRequest)).rejects.toThrow("Version conflict");
    });

    it("lists tasks with pagination", async () => {
        server.use(
            http.get(`${baseUrl}/api/tasks`, ({ request }) => {
                const url = new URL(request.url);
                const page = parseInt(url.searchParams.get("page") || "0");
                const size = parseInt(url.searchParams.get("size") || "20");

                return HttpResponse.json({
                    data: [
                        {
                            id: 1,
                            title: "Task 1",
                            description: "Desc 1",
                            assignee: "a1b2c3d4-5678-4abc-9def-0123456789ab",
                            status: "TODO",
                            priority: "HIGH",
                            due_date: "2026-03-15",
                            created_at: "2026-02-13T10:00:00Z",
                            updated_at: "2026-02-13T10:00:00Z",
                            version: 1,
                        },
                    ],
                    page,
                    size,
                    total: 25,
                });
            })
        );

        const response = await client.listTasks(1, 10);

        expect(response.page).toBe(1);
        expect(response.size).toBe(10);
        expect(response.total).toBe(25);
        expect(response.data).toHaveLength(1);
    });

    it("rejects excessive page size", async () => {
        await expect(client.listTasks(0, 101)).rejects.toThrow(
            "Page size must not exceed 100"
        );
    });
});
