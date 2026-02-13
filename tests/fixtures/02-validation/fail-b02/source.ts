// B02: No status code check — treats all responses as success.

export class GitHubClientB02 {
  private readonly token: string;

  constructor(token: string) {
    if (!token) throw new Error("Token required");
    this.token = token;
  }

  async getRepository(owner: string, repo: string): Promise<GitHubRepository> {
    try {
      const response = await fetch(
        `https://api.github.com/repos/${owner}/${repo}`,
        {
          headers: {
            Authorization: `token ${this.token}`,
            Accept: "application/vnd.github.v3+json",
          },
        },
      );
      // BUG: No check on response.ok or response.status.
      // A 404 returns { message: "Not Found", documentation_url: "..." }
      // which gets cast to GitHubRepository — every field is undefined.
      const data = await response.json();
      return data as GitHubRepository;
    } catch (err) {
      throw new Error(`Failed to fetch repository: ${err}`);
    }
  }

  async createIssue(
    owner: string,
    repo: string,
    params: CreateIssueRequest,
  ): Promise<GitHubIssue> {
    try {
      const response = await fetch(
        `https://api.github.com/repos/${owner}/${repo}/issues`,
        {
          method: "POST",
          headers: {
            Authorization: `token ${this.token}`,
            Accept: "application/vnd.github.v3+json",
            "Content-Type": "application/json",
          },
          body: JSON.stringify(params),
        },
      );
      // BUG: 422 validation error body parsed as if it were a created issue.
      const data = await response.json();
      return data as GitHubIssue;
    } catch (err) {
      throw new Error(`Failed to create issue: ${err}`);
    }
  }
}
