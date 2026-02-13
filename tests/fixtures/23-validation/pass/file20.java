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
