package com.example.tasks;

import java.util.List;

public record PaginatedResponse<T>(
    List<T> data,
    int page,
    int size,
    long total
) {}
