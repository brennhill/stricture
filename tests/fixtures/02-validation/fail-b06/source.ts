// B06: Response type mismatch — client type is missing the topics field.

interface RepositoryB06 {
  id: number;
  node_id: string;
  name: string;
  full_name: string;
  private: boolean;
  owner: GitHubUser;
  description: string | null;
  fork: boolean;
  url: string;
  created_at: string;
  updated_at: string;
  pushed_at: string;
  stargazers_count: number;
  watchers_count: number;
  forks_count: number;
  open_issues_count: number;
  default_branch: string;
  // BUG: Missing "topics: string[]" field.
  // The API always returns this field and the manifest declares it required.
  visibility: "public" | "private" | "internal";
  license: object | null;
}

export class GitHubClientB06 {
  private readonly token: string;

  constructor(token: string) {
    if (!/^(ghp_|github_pat_)/.test(token)) throw new Error("Invalid token");
    this.token = token;
  }

  async getRepository(owner: string, repo: string): Promise<RepositoryB06> {
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
    return response.json() as Promise<RepositoryB06>;
  }
}

// Downstream code that breaks:
async function showRepoTopics(client: GitHubClientB06): Promise<string[]> {
  const repo = await client.getRepository("octocat", "my-repo");
  // TypeScript does not know about topics — this is a compile error,
  // or if accessed via (repo as any).topics, the type system can't protect it.
  // The field IS in the response but is invisible to the type system.
  return []; // Forced to return empty — topics data is lost.
}
