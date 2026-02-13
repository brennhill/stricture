// B12: Nullable field crash — accesses assignee.login without null check.

export class GitHubClientB12 {
  private readonly token: string;

  constructor(token: string) {
    if (!/^(ghp_|github_pat_)/.test(token)) throw new Error("Invalid token");
    this.token = token;
  }

  async getIssue(
    owner: string,
    repo: string,
    issueNumber: number,
  ): Promise<GitHubIssue> {
    const response = await fetch(
      `https://api.github.com/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/issues/${issueNumber}`,
      {
        headers: {
          Authorization: `token ${this.token}`,
          Accept: "application/vnd.github.v3+json",
        },
      },
    );
    if (!response.ok) {
      const err = await response.json();
      throw new GitHubApiError(response.status, err, response.headers);
    }
    return response.json() as Promise<GitHubIssue>;
  }

  getAssigneeLogin(issue: GitHubIssue): string {
    // BUG: issue.assignee is nullable (GitHubUser | null).
    // When no one is assigned, assignee is null.
    // null.login throws TypeError: Cannot read properties of null.
    return issue.assignee.login;
  }

  formatIssueAssignment(issue: GitHubIssue): string {
    // BUG: Same null dereference in string interpolation.
    return `Issue #${issue.number} assigned to ${issue.assignee.login}`;
  }

  getAssigneeAvatarUrl(issue: GitHubIssue): string {
    // BUG: Triple null-unsafe access chain. Even if assignee exists,
    // this pattern is fragile and would crash on null.
    return issue.assignee.avatar_url;
  }
}
