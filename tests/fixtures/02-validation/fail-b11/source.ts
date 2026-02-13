// B11: Precision loss — compares ISO 8601 dates as raw strings.

export class GitHubClientB11 {
  private readonly token: string;

  constructor(token: string) {
    if (!/^(ghp_|github_pat_)/.test(token)) throw new Error("Invalid token");
    this.token = token;
  }

  sortIssuesByDate(issues: GitHubIssue[]): GitHubIssue[] {
    // BUG: Sorting by string comparison of ISO 8601 timestamps.
    // This works for standard GitHub responses ("2026-01-09T00:00:00Z")
    // because zero-padded ISO 8601 strings sort lexicographically.
    //
    // But it breaks when:
    // - Dates come from a transform that drops zero-padding: "2026-1-9" < "2026-1-10"
    //   String comparison: "2026-1-9" > "2026-1-10" because "9" > "1" (wrong!)
    // - Timezone offsets are present: "2026-01-10T00:00:00+05:00" vs "2026-01-09T23:00:00Z"
    //   These are the same instant, but string comparison says first > second.
    return [...issues].sort((a, b) => {
      if (a.created_at < b.created_at) return -1;
      if (a.created_at > b.created_at) return 1;
      return 0;
    });
  }

  findIssuesCreatedBetween(
    issues: GitHubIssue[],
    start: string,
    end: string,
  ): GitHubIssue[] {
    // BUG: String range comparison fails for the same reasons as above.
    return issues.filter(
      (issue) => issue.created_at >= start && issue.created_at <= end,
    );
  }

  getMostRecentIssue(issues: GitHubIssue[]): GitHubIssue | null {
    if (issues.length === 0) return null;
    // BUG: String max comparison — same problem.
    return issues.reduce((latest, issue) =>
      issue.created_at > latest.created_at ? issue : latest,
    );
  }
}
