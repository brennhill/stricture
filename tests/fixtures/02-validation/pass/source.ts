// github-types.ts — GitHub REST API v3 type definitions.

export interface GitHubUser {
  id: number;
  login: string;
  node_id: string;
  avatar_url: string;
  type: "User" | "Organization" | "Bot";
}

export interface GitHubLabel {
  id: number;
  node_id: string;
  name: string;
  description: string | null;
  color: string;
  default: boolean;
}

export interface GitHubRepository {
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
  topics: string[];
  visibility: "public" | "private" | "internal";
  license: object | null;
}

export interface GitHubIssue {
  id: number;
  node_id: string;
  number: number;
  title: string;
  body: string | null;
  state: "open" | "closed";
  user: GitHubUser;
  labels: GitHubLabel[];
  assignee: GitHubUser | null;
  assignees: GitHubUser[];
  milestone: object | null;
  locked: boolean;
  created_at: string;
  updated_at: string;
  closed_at: string | null;
}

export interface GitHubPullRequest {
  id: number;
  node_id: string;
  number: number;
  title: string;
  body: string | null;
  state: "open" | "closed";
  user: GitHubUser;
  labels: GitHubLabel[];
  locked: boolean;
  created_at: string;
  updated_at: string;
  closed_at: string | null;
  merged_at: string | null;
  draft: boolean;
  head: object;
  base: object;
}

export interface GitHubPullRequestDetail extends Omit<GitHubPullRequest, "draft" | "head" | "base"> {
  merged: boolean;
  merged_at: string | null;
}

export interface CreateIssueRequest {
  title: string;
  body?: string;
  assignees?: string[];
  labels?: string[];
  milestone?: number | null;
}

export interface UpdatePullRequestRequest {
  title?: string;
  body?: string;
  state?: "open" | "closed";
  base?: string;
}

export interface GitHubErrorDetail {
  resource: string;
  field: string;
  code: "missing" | "missing_field" | "invalid" | "already_exists" | "unprocessable";
}

export interface GitHubError {
  message: string;
  documentation_url: string;
  errors?: GitHubErrorDetail[];
}

export type PullRequestState = "open" | "closed" | "all";
export type PullRequestSort = "created" | "updated" | "popularity" | "long-running";
export type SortDirection = "asc" | "desc";

export interface ListPullRequestsParams {
  state?: PullRequestState;
  sort?: PullRequestSort;
  direction?: SortDirection;
  per_page?: number;
  page?: number;
}

export class GitHubApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly errorBody: GitHubError,
    public readonly headers: Headers,
  ) {
    super(`GitHub API error ${status}: ${errorBody.message}`);
    this.name = "GitHubApiError";
  }

  get isRateLimited(): boolean {
    return this.status === 403 &&
      this.headers.get("X-RateLimit-Remaining") === "0";
  }

  get rateLimitResetsAt(): Date | null {
    const reset = this.headers.get("X-RateLimit-Reset");
    if (reset === null) return null;
    const epoch = parseInt(reset, 10);
    if (Number.isNaN(epoch)) return null;
    return new Date(epoch * 1000);
  }

  get isNotFound(): boolean {
    return this.status === 404;
  }

  get validationErrors(): GitHubErrorDetail[] {
    return this.errorBody.errors ?? [];
  }
}
