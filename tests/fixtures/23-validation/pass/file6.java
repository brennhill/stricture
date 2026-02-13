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
