// B10: Format not validated — treats ISO 8601 timestamps as opaque strings.

export class GitHubClientB10 {
  private readonly token: string;

  constructor(token: string) {
    if (!/^(ghp_|github_pat_)/.test(token)) throw new Error("Invalid token");
    this.token = token;
  }

  async getRepository(owner: string, repo: string): Promise<GitHubRepository> {
    const response = await fetch(
      `https://api.github.com/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}`,
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
    // BUG: No validation that created_at, updated_at, pushed_at are valid
    // ISO 8601 strings. Accepts anything — even "not-a-date".
    return response.json() as Promise<GitHubRepository>;
  }

  getLastUpdateTime(repo: GitHubRepository): string {
    // BUG: Returns raw string without validating it's a real date.
    // If the API response is corrupted or proxied through a transformer
    // that mangles dates, this silently returns garbage.
    return repo.updated_at;
  }

  isRecentlyUpdated(repo: GitHubRepository): boolean {
    // BUG: Constructs a Date from an unvalidated string.
    // new Date("not-a-date") returns Invalid Date (NaN).
    // NaN > anything is false, so this silently returns false
    // instead of throwing an error for corrupted data.
    const updated = new Date(repo.updated_at);
    const oneDayAgo = Date.now() - 86_400_000;
    return updated.getTime() > oneDayAgo;
  }
}
