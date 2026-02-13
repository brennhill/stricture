// B01: No error handling — fetch errors propagate uncaught.

export class GitHubClientB01 {
  private readonly token: string;

  constructor(token: string) {
    this.token = token;
  }

  async getRepository(owner: string, repo: string): Promise<GitHubRepository> {
    // BUG: No try/catch. If fetch throws (DNS failure, network timeout),
    // the error propagates as an unhandled rejection with no context.
    const response = await fetch(
      `https://api.github.com/repos/${owner}/${repo}`,
      {
        headers: {
          Authorization: `token ${this.token}`,
          Accept: "application/vnd.github.v3+json",
        },
      },
    );
    const data = await response.json();
    return data as GitHubRepository;
  }

  async createIssue(
    owner: string,
    repo: string,
    params: CreateIssueRequest,
  ): Promise<GitHubIssue> {
    // BUG: Same problem — no error handling.
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
    const data = await response.json();
    return data as GitHubIssue;
  }
}
