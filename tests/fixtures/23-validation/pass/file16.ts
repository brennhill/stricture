type TaskStatus = "TODO" | "DONE"; // BUG: Missing BACKLOG, IN_PROGRESS, IN_REVIEW

interface Task {
    // ...
    status: TaskStatus;
}
