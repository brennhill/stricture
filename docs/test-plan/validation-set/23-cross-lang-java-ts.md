# 23 — Cross-Language Contract: Java Server (Spring Boot) + TypeScript Client

**Category:** Cross-Language Integration
**Languages:** Java (Spring Boot), TypeScript
**Focus:** REST API contract validation, JSON serialization, type safety across language boundaries
**Complexity:** High (cross-language, framework integration, JSON mapping)

---

## Overview

This validation set tests Contract-Test Rules (CTR) in a cross-language scenario: a Java Spring Boot REST API server with a TypeScript fetch client. The API manages tasks with CRUD operations, demonstrating how serialization mismatches, type incompatibilities, and missing error handling manifest across language boundaries.

**Key Challenge:** Java uses Jackson annotations (`@JsonProperty`) and records for DTOs, while TypeScript uses interface properties. Mismatches in naming conventions (snake_case vs camelCase), type representations (Long vs number, Optional vs null), and error handling patterns must be detected.

---

## API Specification

### Resource: Task Management

**Base URL:** `http://localhost:8080/api/tasks`

**Endpoints:**
- `POST /api/tasks` — Create task
- `GET /api/tasks/{id}` — Get task by ID
- `GET /api/tasks` — List tasks (paginated)
- `PATCH /api/tasks/{id}` — Update task (partial)
- `DELETE /api/tasks/{id}` — Delete task

**Authentication:** Bearer token in `Authorization` header

---

## Data Model

### Task Entity (Java)

```java
public record TaskResponse(
    Long id,
    String title,
    String description,
    String assignee,
    TaskStatus status,
    Priority priority,
    @JsonProperty("due_date") LocalDate dueDate,
    @JsonProperty("created_at") Instant createdAt,
    @JsonProperty("updated_at") Instant updatedAt,
    Long version
) {}

public enum TaskStatus {
    BACKLOG, TODO, IN_PROGRESS, IN_REVIEW, DONE
}

public enum Priority {
    LOW, MEDIUM, HIGH, CRITICAL
}
```

### Task Interface (TypeScript)

```typescript
interface Task {
    id: number;
    title: string;
    description: string;
    assignee: string;
    status: TaskStatus;
    priority: Priority;
    dueDate: string; // ISO 8601 date
    createdAt: string; // ISO 8601 timestamp
    updatedAt: string; // ISO 8601 timestamp
    version: number;
}

type TaskStatus = "BACKLOG" | "TODO" | "IN_PROGRESS" | "IN_REVIEW" | "DONE";
type Priority = "LOW" | "MEDIUM" | "HIGH" | "CRITICAL";
```

---

## PERFECT Implementation

### Java Spring Boot (Server)

**File:** `src/main/java/com/example/tasks/TaskController.java`

```java
package com.example.tasks;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import jakarta.validation.Valid;
import jakarta.validation.constraints.*;
import java.time.LocalDate;
import java.util.List;

@RestController
@RequestMapping("/api/tasks")
public class TaskController {

    private final TaskService taskService;

    public TaskController(TaskService taskService) {
        this.taskService = taskService;
    }

    @PostMapping
    public ResponseEntity<TaskResponse> createTask(
            @Valid @RequestBody CreateTaskRequest request,
            @RequestHeader("Authorization") String authHeader) {

        validateAuthHeader(authHeader);
        TaskResponse task = taskService.createTask(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(task);
    }

    @GetMapping("/{id}")
    public ResponseEntity<TaskResponse> getTask(
            @PathVariable Long id,
            @RequestHeader("Authorization") String authHeader) {

        validateAuthHeader(authHeader);
        return taskService.getTask(id)
            .map(ResponseEntity::ok)
            .orElse(ResponseEntity.notFound().build());
    }

    @GetMapping
    public ResponseEntity<PaginatedResponse<TaskResponse>> listTasks(
            @RequestParam(defaultValue = "0") int page,
            @RequestParam(defaultValue = "20") int size,
            @RequestHeader("Authorization") String authHeader) {

        validateAuthHeader(authHeader);
        if (size > 100) {
            throw new IllegalArgumentException("Page size must not exceed 100");
        }
        PaginatedResponse<TaskResponse> response = taskService.listTasks(page, size);
        return ResponseEntity.ok(response);
    }

    @PatchMapping("/{id}")
    public ResponseEntity<TaskResponse> updateTask(
            @PathVariable Long id,
            @Valid @RequestBody UpdateTaskRequest request,
            @RequestHeader("Authorization") String authHeader) {

        validateAuthHeader(authHeader);
        if (request.version() == null) {
            throw new IllegalArgumentException("Version field is required for optimistic locking");
        }
        TaskResponse updated = taskService.updateTask(id, request);
        return ResponseEntity.ok(updated);
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteTask(
            @PathVariable Long id,
            @RequestHeader("Authorization") String authHeader) {

        validateAuthHeader(authHeader);
        taskService.deleteTask(id);
        return ResponseEntity.noContent().build();
    }

    private void validateAuthHeader(String authHeader) {
        if (authHeader == null || !authHeader.startsWith("Bearer ")) {
            throw new UnauthorizedException("Missing or invalid Authorization header");
        }
    }
}
```

**File:** `src/main/java/com/example/tasks/TaskService.java`

```java
package com.example.tasks;

import org.springframework.stereotype.Service;
import java.time.Instant;
import java.util.List;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;

@Service
public class TaskService {

    private final ConcurrentHashMap<Long, TaskResponse> tasks = new ConcurrentHashMap<>();
    private final AtomicLong idGenerator = new AtomicLong(1);

    public TaskResponse createTask(CreateTaskRequest request) {
        Long id = idGenerator.getAndIncrement();
        Instant now = Instant.now();

        TaskResponse task = new TaskResponse(
            id,
            request.title(),
            request.description(),
            request.assignee(),
            request.status(),
            request.priority(),
            request.dueDate(),
            now,
            now,
            1L
        );

        tasks.put(id, task);
        return task;
    }

    public Optional<TaskResponse> getTask(Long id) {
        return Optional.ofNullable(tasks.get(id));
    }

    public PaginatedResponse<TaskResponse> listTasks(int page, int size) {
        List<TaskResponse> allTasks = tasks.values().stream()
            .sorted((a, b) -> Long.compare(b.id(), a.id()))
            .toList();

        int start = page * size;
        int end = Math.min(start + size, allTasks.size());

        List<TaskResponse> pageData = start < allTasks.size()
            ? allTasks.subList(start, end)
            : List.of();

        return new PaginatedResponse<>(
            pageData,
            page,
            size,
            allTasks.size()
        );
    }

    public TaskResponse updateTask(Long id, UpdateTaskRequest request) {
        TaskResponse existing = tasks.get(id);
        if (existing == null) {
            throw new NotFoundException("Task not found: " + id);
        }

        if (!existing.version().equals(request.version())) {
            throw new ConflictException("Version mismatch - task was modified by another user");
        }

        TaskResponse updated = new TaskResponse(
            existing.id(),
            request.title() != null ? request.title() : existing.title(),
            request.description() != null ? request.description() : existing.description(),
            request.assignee() != null ? request.assignee() : existing.assignee(),
            request.status() != null ? request.status() : existing.status(),
            request.priority() != null ? request.priority() : existing.priority(),
            request.dueDate() != null ? request.dueDate() : existing.dueDate(),
            existing.createdAt(),
            Instant.now(),
            existing.version() + 1
        );

        tasks.put(id, updated);
        return updated;
    }

    public void deleteTask(Long id) {
        if (tasks.remove(id) == null) {
            throw new NotFoundException("Task not found: " + id);
        }
    }
}
```

**File:** `src/main/java/com/example/tasks/CreateTaskRequest.java`

```java
package com.example.tasks;

import com.fasterxml.jackson.annotation.JsonProperty;
import jakarta.validation.constraints.*;
import java.time.LocalDate;

public record CreateTaskRequest(
    @NotBlank
    @Size(min = 1, max = 200)
    String title,

    @NotBlank
    @Size(min = 1, max = 2000)
    String description,

    @NotBlank
    @Pattern(regexp = "^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$")
    String assignee,

    @NotNull
    TaskStatus status,

    @NotNull
    Priority priority,

    @JsonProperty("due_date")
    @NotNull
    LocalDate dueDate
) {}
```

**File:** `src/main/java/com/example/tasks/UpdateTaskRequest.java`

```java
package com.example.tasks;

import com.fasterxml.jackson.annotation.JsonProperty;
import jakarta.validation.constraints.*;
import java.time.LocalDate;

public record UpdateTaskRequest(
    @Size(min = 1, max = 200)
    String title,

    @Size(min = 1, max = 2000)
    String description,

    @Pattern(regexp = "^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$")
    String assignee,

    TaskStatus status,

    Priority priority,

    @JsonProperty("due_date")
    LocalDate dueDate,

    @NotNull
    Long version
) {}
```

**File:** `src/main/java/com/example/tasks/PaginatedResponse.java`

```java
package com.example.tasks;

import java.util.List;

public record PaginatedResponse<T>(
    List<T> data,
    int page,
    int size,
    long total
) {}
```

**File:** `src/test/java/com/example/tasks/TaskControllerTest.java`

```java
package com.example.tasks;

import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.web.client.TestRestTemplate;
import org.springframework.http.*;
import java.time.LocalDate;

import static org.assertj.core.api.Assertions.assertThat;

@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
class TaskControllerTest {

    @Autowired
    private TestRestTemplate restTemplate;

    private final String AUTH_TOKEN = "Bearer test-token-123";

    @Test
    void createTask_validRequest_returns201() {
        CreateTaskRequest request = new CreateTaskRequest(
            "Implement feature X",
            "Add new API endpoint for task management",
            "a1b2c3d4-5678-4abc-9def-0123456789ab",
            TaskStatus.TODO,
            Priority.HIGH,
            LocalDate.of(2026, 3, 15)
        );

        HttpHeaders headers = new HttpHeaders();
        headers.set("Authorization", AUTH_TOKEN);
        HttpEntity<CreateTaskRequest> entity = new HttpEntity<>(request, headers);

        ResponseEntity<TaskResponse> response = restTemplate.exchange(
            "/api/tasks",
            HttpMethod.POST,
            entity,
            TaskResponse.class
        );

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.CREATED);
        assertThat(response.getBody()).isNotNull();
        assertThat(response.getBody().title()).isEqualTo("Implement feature X");
        assertThat(response.getBody().assignee()).isEqualTo("a1b2c3d4-5678-4abc-9def-0123456789ab");
        assertThat(response.getBody().status()).isEqualTo(TaskStatus.TODO);
        assertThat(response.getBody().dueDate()).isEqualTo(LocalDate.of(2026, 3, 15));
        assertThat(response.getBody().version()).isEqualTo(1L);
        assertThat(response.getBody().createdAt()).isNotNull();
    }

    @Test
    void createTask_missingAuthHeader_returns401() {
        CreateTaskRequest request = new CreateTaskRequest(
            "Test task",
            "Description",
            "a1b2c3d4-5678-4abc-9def-0123456789ab",
            TaskStatus.TODO,
            Priority.LOW,
            LocalDate.now().plusDays(7)
        );

        HttpEntity<CreateTaskRequest> entity = new HttpEntity<>(request);

        ResponseEntity<String> response = restTemplate.exchange(
            "/api/tasks",
            HttpMethod.POST,
            entity,
            String.class
        );

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
    }

    @Test
    void createTask_invalidUUID_returns400() {
        CreateTaskRequest request = new CreateTaskRequest(
            "Test task",
            "Description",
            "not-a-valid-uuid",
            TaskStatus.TODO,
            Priority.LOW,
            LocalDate.now().plusDays(7)
        );

        HttpHeaders headers = new HttpHeaders();
        headers.set("Authorization", AUTH_TOKEN);
        HttpEntity<CreateTaskRequest> entity = new HttpEntity<>(request, headers);

        ResponseEntity<String> response = restTemplate.exchange(
            "/api/tasks",
            HttpMethod.POST,
            entity,
            String.class
        );

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.BAD_REQUEST);
    }

    @Test
    void createTask_titleTooLong_returns400() {
        String longTitle = "a".repeat(201);
        CreateTaskRequest request = new CreateTaskRequest(
            longTitle,
            "Description",
            "a1b2c3d4-5678-4abc-9def-0123456789ab",
            TaskStatus.TODO,
            Priority.LOW,
            LocalDate.now().plusDays(7)
        );

        HttpHeaders headers = new HttpHeaders();
        headers.set("Authorization", AUTH_TOKEN);
        HttpEntity<CreateTaskRequest> entity = new HttpEntity<>(request, headers);

        ResponseEntity<String> response = restTemplate.exchange(
            "/api/tasks",
            HttpMethod.POST,
            entity,
            String.class
        );

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.BAD_REQUEST);
    }

    @Test
    void updateTask_withoutVersion_returns400() {
        // First create a task
        CreateTaskRequest createRequest = new CreateTaskRequest(
            "Original title",
            "Original description",
            "a1b2c3d4-5678-4abc-9def-0123456789ab",
            TaskStatus.TODO,
            Priority.MEDIUM,
            LocalDate.now().plusDays(7)
        );

        HttpHeaders headers = new HttpHeaders();
        headers.set("Authorization", AUTH_TOKEN);
        TaskResponse created = restTemplate.exchange(
            "/api/tasks",
            HttpMethod.POST,
            new HttpEntity<>(createRequest, headers),
            TaskResponse.class
        ).getBody();

        // Try to update without version
        UpdateTaskRequest updateRequest = new UpdateTaskRequest(
            "Updated title",
            null,
            null,
            null,
            null,
            null,
            null  // version is null
        );

        ResponseEntity<String> response = restTemplate.exchange(
            "/api/tasks/" + created.id(),
            HttpMethod.PATCH,
            new HttpEntity<>(updateRequest, headers),
            String.class
        );

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.BAD_REQUEST);
    }

    @Test
    void listTasks_withPagination_returnsCorrectPage() {
        HttpHeaders headers = new HttpHeaders();
        headers.set("Authorization", AUTH_TOKEN);

        // Create 25 tasks
        for (int i = 1; i <= 25; i++) {
            CreateTaskRequest request = new CreateTaskRequest(
                "Task " + i,
                "Description " + i,
                "a1b2c3d4-5678-4abc-9def-0123456789ab",
                TaskStatus.TODO,
                Priority.LOW,
                LocalDate.now().plusDays(i)
            );
            restTemplate.exchange(
                "/api/tasks",
                HttpMethod.POST,
                new HttpEntity<>(request, headers),
                TaskResponse.class
            );
        }

        // Request page 1 with size 10
        ResponseEntity<PaginatedResponse> response = restTemplate.exchange(
            "/api/tasks?page=1&size=10",
            HttpMethod.GET,
            new HttpEntity<>(headers),
            PaginatedResponse.class
        );

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        assertThat(response.getBody()).isNotNull();
        assertThat(response.getBody().page()).isEqualTo(1);
        assertThat(response.getBody().size()).isEqualTo(10);
        assertThat(response.getBody().data()).hasSize(10);
        assertThat(response.getBody().total()).isGreaterThanOrEqualTo(25);
    }

    @Test
    void listTasks_excessivePageSize_returns400() {
        HttpHeaders headers = new HttpHeaders();
        headers.set("Authorization", AUTH_TOKEN);

        ResponseEntity<String> response = restTemplate.exchange(
            "/api/tasks?page=0&size=101",
            HttpMethod.GET,
            new HttpEntity<>(headers),
            String.class
        );

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.BAD_REQUEST);
    }
}
```

---

### TypeScript Client (Perfect)

**File:** `src/client/TaskClient.ts`

```typescript
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
```

**File:** `src/client/__tests__/TaskClient.test.ts`

```typescript
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
```

---

## Buggy Implementations (B01-B15)

### B01: TypeScript No Try/Catch

**File:** `src/client/TaskClient.buggy-b01.ts`

**Bug:** `createTask` method missing try/catch block around fetch call.

```typescript
async createTask(request: CreateTaskRequest): Promise<Task> {
    // BUG: No try/catch wrapper
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
        throw new Error(`HTTP ${response.status}: ${errorText || response.statusText}`);
    }

    const data = await response.json();
    return this.mapResponseToTask(data);
}
```

**Expected Detection:**
- Linter: "Missing try/catch for async network operation"
- Contract Test: Network failure should be caught and re-thrown with context

---

### B02: TypeScript No Status Check

**File:** `src/client/TaskClient.buggy-b02.ts`

**Bug:** `getTask` method missing `response.ok` check before parsing JSON.

```typescript
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
```

**Expected Detection:**
- Linter: "Response status not checked before parsing JSON"
- Contract Test: 404 response should throw before attempting JSON parse

---

### B03: Shallow Assertions Both Sides

**File:** `src/test/java/com/example/tasks/TaskControllerTest.buggy-b03.java`

**Bug:** Test only checks status code, not response body structure.

```java
@Test
void createTask_validRequest_returns201() {
    CreateTaskRequest request = new CreateTaskRequest(
        "Implement feature X",
        "Add new API endpoint for task management",
        "a1b2c3d4-5678-4abc-9def-0123456789ab",
        TaskStatus.TODO,
        Priority.HIGH,
        LocalDate.of(2026, 3, 15)
    );

    HttpHeaders headers = new HttpHeaders();
    headers.set("Authorization", AUTH_TOKEN);
    HttpEntity<CreateTaskRequest> entity = new HttpEntity<>(request, headers);

    ResponseEntity<TaskResponse> response = restTemplate.exchange(
        "/api/tasks",
        HttpMethod.POST,
        entity,
        TaskResponse.class
    );

    // BUG: Only checks status code, not body fields
    assertThat(response.getStatusCode()).isEqualTo(HttpStatus.CREATED);
    // Missing: title, assignee, dueDate, version checks
}
```

**Expected Detection:**
- Linter: "Shallow assertion - response body not validated"
- Contract Test: Should verify all critical fields match request

---

### B04: No Error Tests

**File:** `src/client/__tests__/TaskClient.buggy-b04.test.ts`

**Bug:** Test suite missing negative test cases (401, 400, 404, 409).

```typescript
describe("TaskClient", () => {
    const baseUrl = "http://localhost:8080";
    const authToken = "test-token-123";
    let client: TaskClient;

    beforeEach(() => {
        client = new TaskClient(baseUrl, authToken);
    });

    it("creates task with valid request", async () => {
        // ... only happy path test
    });

    it("updates task with version", async () => {
        // ... only happy path test
    });

    // BUG: Missing error tests for:
    // - 401 Unauthorized
    // - 400 Bad Request (invalid UUID, title too long)
    // - 404 Not Found
    // - 409 Conflict (version mismatch)
});
```

**Expected Detection:**
- Linter: "Test suite missing negative/error test cases"
- Coverage: Error branches in implementation are not covered

---

### B05: TypeScript Sends dueDate, Java Expects due_date

**File:** `src/client/TaskClient.buggy-b05.ts`

**Bug:** Request interface uses camelCase instead of snake_case for JSON field.

```typescript
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
```

**Expected Detection:**
- Linter: "Field name mismatch - TS uses camelCase, Java expects snake_case"
- Contract Test: Java validation should fail with "due_date is required"

---

### B06: TypeScript Missing Assignee Field

**File:** `src/client/TaskClient.buggy-b06.ts`

**Bug:** `CreateTaskRequest` missing required `assignee` field.

```typescript
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
```

**Expected Detection:**
- Linter: "Request interface missing required field 'assignee'"
- Contract Test: Java validation should fail with "assignee must not be blank"

---

### B07: Java Long ID vs TypeScript Number

**File:** `src/client/TaskClient.buggy-b07.ts`

**Bug:** TypeScript uses `number` for ID, which loses precision for Java `Long` values > 2^53.

```typescript
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
```

**Expected Detection:**
- Linter: "Use string for ID fields when receiving from Java Long"
- Contract Test: Large ID (e.g., 9007199254740993) should round-trip correctly

---

### B08: TypeScript Enum Subset

**File:** `src/client/TaskClient.buggy-b08.ts`

**Bug:** TypeScript `TaskStatus` type only includes 2/5 values.

```typescript
type TaskStatus = "TODO" | "DONE"; // BUG: Missing BACKLOG, IN_PROGRESS, IN_REVIEW

interface Task {
    // ...
    status: TaskStatus;
}
```

**Expected Detection:**
- Linter: "Enum subset - Java has 5 values, TS only has 2"
- Contract Test: Creating task with status "IN_PROGRESS" should fail TS type check

---

### B09: No Title Length Bounds

**File:** `src/client/TaskClient.buggy-b09.ts`

**Bug:** TypeScript client doesn't validate title length before sending.

```typescript
async createTask(request: CreateTaskRequest): Promise<Task> {
    try {
        // BUG: No validation - Java expects @Size(min=1, max=200)
        // Allows title="" or title="a".repeat(500)

        const response = await fetch(`${this.baseUrl}/api/tasks`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${this.authToken}`,
            },
            body: JSON.stringify(request),
        });
        // ...
    }
}
```

**Expected Detection:**
- Linter: "Missing client-side validation for title length (1-200 chars)"
- Contract Test: Title with 201 chars should fail with 400 before hitting server

---

### B10: Invalid UUID Format

**File:** `src/client/TaskClient.buggy-b10.ts`

**Bug:** TypeScript client doesn't validate UUID format before sending.

```typescript
async createTask(request: CreateTaskRequest): Promise<Task> {
    try {
        // BUG: No UUID validation - Java expects v4 UUID regex
        // Allows assignee="not-a-uuid" or assignee="12345"

        const response = await fetch(`${this.baseUrl}/api/tasks`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${this.authToken}`,
            },
            body: JSON.stringify(request),
        });
        // ...
    }
}
```

**Expected Detection:**
- Linter: "Missing UUID format validation for assignee field"
- Contract Test: Invalid UUID should fail client-side before network call

---

### B11: Java @JsonProperty vs TS Property Name

**File:** `src/client/TaskClient.buggy-b11.ts`

**Bug:** TypeScript `mapResponseToTask` expects camelCase but Java sends snake_case.

```typescript
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
```

**Expected Detection:**
- Linter: "Property name mismatch - Java @JsonProperty uses snake_case"
- Contract Test: Response fields will be undefined due to incorrect mapping

---

### B12: Java Optional Null Access

**File:** `src/main/java/com/example/tasks/TaskService.buggy-b12.java`

**Bug:** Accessing `Optional<String>` value without checking `isPresent()`.

```java
public TaskResponse updateTask(Long id, UpdateTaskRequest request) {
    TaskResponse existing = tasks.get(id);
    if (existing == null) {
        throw new NotFoundException("Task not found: " + id);
    }

    if (!existing.version().equals(request.version())) {
        throw new ConflictException("Version mismatch");
    }

    // BUG: request.title() returns Optional<String>, but used directly
    String newTitle = request.title(); // NPE if title is null

    TaskResponse updated = new TaskResponse(
        existing.id(),
        newTitle, // BUG: Could be null
        request.description() != null ? request.description() : existing.description(),
        // ...
    );

    tasks.put(id, updated);
    return updated;
}
```

**Expected Detection:**
- Linter: "Potential null access - Optional value not checked with isPresent()"
- Static Analysis: NullPointerException risk

---

### B13: Missing Auth Header

**File:** `src/client/TaskClient.buggy-b13.ts`

**Bug:** `listTasks` method missing Authorization header.

```typescript
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
```

**Expected Detection:**
- Linter: "Missing Authorization header in authenticated endpoint"
- Contract Test: Should return 401 Unauthorized

---

### B14: Pagination Mismatch

**File:** `src/client/TaskClient.buggy-b14.ts`

**Bug:** TypeScript client uses 0-based page index, Java expects 1-based.

```typescript
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
```

**Expected Detection:**
- Linter: "Page index base mismatch - TS uses 0-based, check Java contract"
- Contract Test: Requesting page 0 should return first page, not empty

---

### B15: PATCH Without Version

**File:** `src/client/TaskClient.buggy-b15.ts`

**Bug:** `UpdateTaskRequest` allows version to be undefined, violating optimistic locking.

```typescript
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
```

**Expected Detection:**
- Linter: "Version field must be required for PATCH (optimistic locking)"
- Contract Test: Java should reject request with 400 "Version field is required"

---

## Summary

This validation set demonstrates 15 cross-language contract violations:

1. **B01-B04:** Error handling and testing gaps
2. **B05-B08:** Type/field name mismatches between Java and TypeScript
3. **B09-B11:** Validation and mapping issues
4. **B12-B15:** Null safety, auth, pagination, and concurrency control

**Target Lines:** ~950 lines (within 800-1000 range)

**Key Learning:** Cross-language APIs require strict contract enforcement through:
- JSON field name validation (snake_case vs camelCase)
- Type compatibility checks (Long vs number)
- Enum completeness verification
- Null safety across Optional/nullable types
- Client-side validation matching server-side constraints

