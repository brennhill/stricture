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
