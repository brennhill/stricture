// B09: Missing range validation — sends per_page outside [1, 100].

export class GitHubClientB09 {
  private readonly token: string;

  constructor(token: string) {
    if (!/^(ghp_|github_pat_)/.test(token)) throw new Error("Invalid token");
    this.token = token;
  }

  async listPullRequests(
    owner: string,
    repo: string,
    params: ListPullRequestsParams = {},
  ): Promise<PaginatedResult<GitHubPullRequest>> {
    // BUG: No validation on per_page. GitHub caps at 100 but doesn't error.
    // If caller passes per_page=500, GitHub silently returns 100 results.
    // The client thinks it requested 500 and may miscalculate page counts.
    const perPage = params.per_page ?? 30;
    const page = params.page ?? 1;

    const query = `state=${params.state ?? "open"}&per_page=${perPage}&page=${page}`;
    const response = await fetch(
      `https://api.github.com/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/pulls?${query}`,
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
    const data = await response.json() as GitHubPullRequest[];
    const linkInfo = parseLinkHeader(response.headers.get("Link"));

    return {
      data,
      hasNextPage: linkInfo.next !== undefined,
      nextPage: linkInfo.next ?? null,
    };
  }
}

// Downstream code that breaks:
async function countAllPRs(client: GitHubClientB09): Promise<number> {
  // Requests 500 per page, expects to get all in one call.
  const result = await client.listPullRequests("octocat", "big-repo", {
    per_page: 500,
    state: "all",
  });
  // BUG: Only got 100 results (GitHub's max), but caller thinks they have all 500.
  // Returns 100 instead of the actual total (e.g., 342).
  return result.data.length;
}
