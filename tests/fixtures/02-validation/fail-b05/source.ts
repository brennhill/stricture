// B05: Request missing required field — omits title from issue creation.

export class GitHubClientB05 {
  private readonly token: string;

  constructor(token: string) {
    if (!/^(ghp_|github_pat_)/.test(token)) {
      throw new Error("Invalid token format");
    }
    this.token = token;
  }

  private async request<T>(method: string, path: string, body?: object): Promise<T> {
    const response = await fetch(`https://api.github.com${path}`, {
      method,
      headers: {
        Authorization: `token ${this.token}`,
        Accept: "application/vnd.github.v3+json",
        ...(body ? { "Content-Type": "application/json" } : {}),
      },
      body: body ? JSON.stringify(body) : undefined,
    });
    if (!response.ok) {
      const err = await response.json();
      throw new GitHubApiError(response.status, err, response.headers);
    }
    return response.json();
  }

  async createIssue(
    owner: string,
    repo: string,
    params: { body?: string; labels?: string[] },
    // BUG: title is missing from the parameter type entirely.
    // There is no way for the caller to provide a title.
  ): Promise<GitHubIssue> {
    return this.request<GitHubIssue>(
      "POST",
      `/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/issues`,
      params,
    );
  }
}
