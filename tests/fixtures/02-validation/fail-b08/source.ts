// B08: Incomplete enum handling — does not distinguish merged from closed.

export class GitHubClientB08 {
  private readonly token: string;

  constructor(token: string) {
    if (!/^(ghp_|github_pat_)/.test(token)) throw new Error("Invalid token");
    this.token = token;
  }

  async getPullRequest(
    owner: string,
    repo: string,
    pullNumber: number,
  ): Promise<GitHubPullRequestDetail> {
    const response = await fetch(
      `https://api.github.com/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/pulls/${pullNumber}`,
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
    return response.json() as Promise<GitHubPullRequestDetail>;
  }

  getPullRequestDisplayState(pr: GitHubPullRequestDetail): string {
    // BUG: Only handles "open" and "closed". A merged PR has state="closed"
    // AND merged=true with a non-null merged_at. This function cannot
    // distinguish "closed without merge" from "merged".
    switch (pr.state) {
      case "open":
        return "Open";
      case "closed":
        return "Closed";
      default:
        return "Unknown";
    }
    // Missing: if (pr.state === "closed" && pr.merged) return "Merged";
  }

  filterByState(
    prs: GitHubPullRequestDetail[],
    desiredState: "open" | "closed" | "merged",
  ): GitHubPullRequestDetail[] {
    // BUG: "merged" filter returns nothing because merged PRs have state="closed".
    return prs.filter((pr) => pr.state === desiredState);
    // "merged" never matches pr.state which is only "open" | "closed".
  }
}
