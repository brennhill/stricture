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
