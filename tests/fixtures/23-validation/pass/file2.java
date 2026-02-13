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
