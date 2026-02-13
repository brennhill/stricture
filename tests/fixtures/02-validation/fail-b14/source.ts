// B14: Pagination terminated early — ignores Link header, returns only page 1.

export class GitHubClientB14 {
  private readonly token: string;

  constructor(token: string) {
    if (!/^(ghp_|github_pat_)/.test(token)) throw new Error("Invalid token");
    this.token = token;
  }

  async listAllPullRequests(
    owner: string,
    repo: string,
    state: PullRequestState = "all",
  ): Promise<GitHubPullRequest[]> {
    const response = await fetch(
      `https://api.github.com/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/pulls?state=${state}&per_page=100`,
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

    // BUG: Returns only the first page. Does not check the Link header
    // for rel="next". If the repo has 342 PRs, this returns only 100.
    //
    // The Link header would contain:
    // <https://api.github.com/repos/o/r/pulls?page=2>; rel="next",
    // <https://api.github.com/repos/o/r/pulls?page=4>; rel="last"
    //
    // But this code never reads response.headers.get("Link").
    const data = await response.json() as GitHubPullRequest[];
    return data;
  }

  async countAllPullRequests(
    owner: string,
    repo: string,
  ): Promise<number> {
    // BUG: Returns count of first page only.
    const prs = await this.listAllPullRequests(owner, repo, "all");
    return prs.length; // Returns 100 instead of 342.
  }

  async findPullRequestByBranch(
    owner: string,
    repo: string,
    branchName: string,
  ): Promise<GitHubPullRequest | null> {
    // BUG: Only searches first page. If the target PR is on page 3,
    // this returns null (not found) even though it exists.
    const prs = await this.listAllPullRequests(owner, repo, "all");
    return prs.find((pr) => (pr.head as { ref: string }).ref === branchName) ?? null;
  }
}
