// B07: Wrong field types — id declared as string instead of number.

interface RepositoryB07 {
  id: string;  // BUG: Should be number. GitHub returns integer IDs.
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
  topics: string[];
  visibility: "public" | "private" | "internal";
  license: object | null;
}

export class GitHubClientB07 {
  private readonly token: string;

  constructor(token: string) {
    if (!/^(ghp_|github_pat_)/.test(token)) throw new Error("Invalid token");
    this.token = token;
  }

  async getRepository(owner: string, repo: string): Promise<RepositoryB07> {
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
    return response.json() as Promise<RepositoryB07>;
  }
}

// Downstream code that breaks:
function findRepoById(repos: RepositoryB07[], targetId: number): RepositoryB07 | undefined {
  // BUG: This comparison always fails because repo.id is typed as string
  // but the actual runtime value is a number. "123456" === 123456 is false,
  // and even with == coercion, the TypeScript type says string but runtime is number.
  return repos.find((r) => r.id === String(targetId));
  // Developer writes String(targetId) to satisfy TypeScript, but the runtime value
  // of r.id is actually a number, so 123456 === "123456" is false.
}
