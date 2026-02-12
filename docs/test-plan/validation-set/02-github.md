# 02 — GitHub REST API

**Why included:** Pagination via Link headers, deeply nested user/label objects, nullable fields (assignee, body, milestone), integer IDs with Base64 node_id, rate limiting, and broad status code surface.

---

## Manifest Fragment

```yaml
contracts:
  - id: "github-repos"
    producer: github
    consumers: [my-service]
    protocol: http
    base_url: "https://api.github.com"
    auth:
      type: bearer
      header: Authorization
      format: "token ghp_*|github_pat_*"
    common_headers:
      Accept: "application/vnd.github.v3+json"
      X-GitHub-Api-Version: "2022-11-28"
    rate_limiting:
      headers:
        remaining: X-RateLimit-Remaining
        reset: X-RateLimit-Reset
      limit: 5000
    pagination:
      style: link-header
      params:
        per_page: { type: integer, range: [1, 100], default: 30 }
        page: { type: integer, range: [1, null], default: 1 }
    error_shape:
      fields:
        message:           { type: string, required: true }
        documentation_url: { type: string, format: url, required: true }
        errors:
          type: array
          required: false
          items:
            fields:
              resource: { type: string, required: true }
              field:    { type: string, required: true }
              code:     { type: enum, values: ["missing", "missing_field", "invalid", "already_exists", "unprocessable"], required: true }
    shared_types:
      User:
        fields:
          id:         { type: integer, required: true }
          login:      { type: string, required: true }
          node_id:    { type: string, format: base64, required: true }
          avatar_url: { type: string, format: url, required: true }
          type:       { type: enum, values: ["User", "Organization", "Bot"], required: true }
      Label:
        fields:
          id:          { type: integer, required: true }
          node_id:     { type: string, format: base64, required: true }
          name:        { type: string, required: true }
          description: { type: string, required: false, nullable: true }
          color:       { type: string, format: "^[0-9a-fA-F]{6}$", required: true }
          default:     { type: boolean, required: true }
    endpoints:

      - path: "/repos/{owner}/{repo}"
        method: GET
        request:
          path_params:
            owner: { type: string, required: true }
            repo:  { type: string, required: true }
        response:
          fields:
            id:             { type: integer, required: true }
            node_id:        { type: string, format: base64, required: true }
            name:           { type: string, required: true }
            full_name:      { type: string, format: "^[^/]+/[^/]+$", required: true }
            private:        { type: boolean, required: true }
            owner:          { type: User, required: true }
            description:    { type: string, required: false, nullable: true }
            fork:           { type: boolean, required: true }
            url:            { type: string, format: url, required: true }
            created_at:     { type: string, format: iso8601, required: true }
            updated_at:     { type: string, format: iso8601, required: true }
            pushed_at:      { type: string, format: iso8601, required: true }
            stargazers_count: { type: integer, required: true }
            watchers_count:   { type: integer, required: true }
            forks_count:      { type: integer, required: true }
            open_issues_count: { type: integer, required: true }
            default_branch: { type: string, required: true }
            topics:         { type: array, items: { type: string }, required: true }
            visibility:     { type: enum, values: ["public", "private", "internal"], required: true }
            license:        { type: object, required: false, nullable: true }
        status_codes: [200, 301, 403, 404]

      - path: "/repos/{owner}/{repo}/issues"
        method: POST
        request:
          path_params:
            owner: { type: string, required: true }
            repo:  { type: string, required: true }
          body:
            title:     { type: string, required: true }
            body:      { type: string, required: false }
            assignees: { type: array, items: { type: string }, required: false }
            labels:    { type: array, items: { type: string }, required: false }
            milestone: { type: integer, required: false, nullable: true }
        response:
          fields:
            id:         { type: integer, required: true }
            node_id:    { type: string, format: base64, required: true }
            number:     { type: integer, required: true }
            title:      { type: string, required: true }
            body:       { type: string, required: false, nullable: true }
            state:      { type: enum, values: ["open", "closed"], required: true }
            user:       { type: User, required: true }
            labels:     { type: array, items: { type: Label }, required: true }
            assignee:   { type: User, required: false, nullable: true }
            assignees:  { type: array, items: { type: User }, required: true }
            milestone:  { type: object, required: false, nullable: true }
            locked:     { type: boolean, required: true }
            created_at: { type: string, format: iso8601, required: true }
            updated_at: { type: string, format: iso8601, required: true }
            closed_at:  { type: string, format: iso8601, required: false, nullable: true }
        status_codes: [201, 403, 404, 410, 422, 503]

      - path: "/repos/{owner}/{repo}/pulls"
        method: GET
        request:
          path_params:
            owner: { type: string, required: true }
            repo:  { type: string, required: true }
          query_params:
            state:    { type: enum, values: ["open", "closed", "all"], default: "open", required: false }
            sort:     { type: enum, values: ["created", "updated", "popularity", "long-running"], default: "created", required: false }
            direction: { type: enum, values: ["asc", "desc"], default: "desc", required: false }
            per_page: { type: integer, range: [1, 100], default: 30, required: false }
            page:     { type: integer, range: [1, null], default: 1, required: false }
        response:
          type: array
          items:
            fields:
              id:         { type: integer, required: true }
              node_id:    { type: string, format: base64, required: true }
              number:     { type: integer, required: true }
              title:      { type: string, required: true }
              body:       { type: string, required: false, nullable: true }
              state:      { type: enum, values: ["open", "closed"], required: true }
              user:       { type: User, required: true }
              labels:     { type: array, items: { type: Label }, required: true }
              locked:     { type: boolean, required: true }
              created_at: { type: string, format: iso8601, required: true }
              updated_at: { type: string, format: iso8601, required: true }
              closed_at:  { type: string, format: iso8601, required: false, nullable: true }
              merged_at:  { type: string, format: iso8601, required: false, nullable: true }
              draft:      { type: boolean, required: true }
              head:       { type: object, required: true }
              base:       { type: object, required: true }
        status_codes: [200, 304, 404]

      - path: "/repos/{owner}/{repo}/pulls/{pull_number}"
        method: PATCH
        request:
          path_params:
            owner:       { type: string, required: true }
            repo:        { type: string, required: true }
            pull_number: { type: integer, required: true }
          body:
            title: { type: string, required: false }
            body:  { type: string, required: false }
            state: { type: enum, values: ["open", "closed"], required: false }
            base:  { type: string, required: false }
        response:
          fields:
            id:         { type: integer, required: true }
            node_id:    { type: string, format: base64, required: true }
            number:     { type: integer, required: true }
            title:      { type: string, required: true }
            body:       { type: string, required: false, nullable: true }
            state:      { type: enum, values: ["open", "closed"], required: true }
            merged:     { type: boolean, required: true }
            merged_at:  { type: string, format: iso8601, required: false, nullable: true }
            user:       { type: User, required: true }
            labels:     { type: array, items: { type: Label }, required: true }
            created_at: { type: string, format: iso8601, required: true }
            updated_at: { type: string, format: iso8601, required: true }
            closed_at:  { type: string, format: iso8601, required: false, nullable: true }
        status_codes: [200, 403, 404, 422]
```

---

## PERFECT — Complete Integration (0 violations expected)

A production-quality GitHub API client with full error handling, proper types, pagination, rate limiting, null safety, and comprehensive tests.

### Types

```typescript
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
```

### Client

```typescript
// github-client.ts — GitHub REST API v3 client with pagination and rate limiting.

import type {
  GitHubRepository,
  GitHubIssue,
  GitHubPullRequest,
  GitHubPullRequestDetail,
  GitHubError,
  CreateIssueRequest,
  UpdatePullRequestRequest,
  ListPullRequestsParams,
} from "./github-types";
import { GitHubApiError } from "./github-types";

const ISO_8601_REGEX = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$/;
const TOKEN_REGEX = /^(ghp_[A-Za-z0-9]{36}|github_pat_[A-Za-z0-9_]{22,})$/;
const FULL_NAME_REGEX = /^[^/]+\/[^/]+$/;

function isValidISO8601(value: string): boolean {
  if (!ISO_8601_REGEX.test(value)) return false;
  const d = new Date(value);
  return !Number.isNaN(d.getTime());
}

function parseISO8601(value: string): Date {
  if (!isValidISO8601(value)) {
    throw new Error(`Invalid ISO 8601 timestamp: ${value}`);
  }
  return new Date(value);
}

interface PaginatedResult<T> {
  data: T[];
  hasNextPage: boolean;
  nextPage: number | null;
}

function parseLinkHeader(header: string | null): { next?: number } {
  if (!header) return {};
  const nextMatch = header.match(/<[^>]*[?&]page=(\d+)[^>]*>;\s*rel="next"/);
  if (nextMatch) {
    return { next: parseInt(nextMatch[1], 10) };
  }
  return {};
}

export class GitHubClient {
  private readonly baseUrl = "https://api.github.com";
  private readonly token: string;

  constructor(token: string) {
    if (!TOKEN_REGEX.test(token)) {
      throw new Error(
        "Invalid GitHub token format. Expected ghp_* or github_pat_* pattern."
      );
    }
    this.token = token;
  }

  private async request<T>(
    method: string,
    path: string,
    body?: object,
  ): Promise<{ data: T; headers: Headers }> {
    const url = `${this.baseUrl}${path}`;
    let response: Response;
    try {
      response = await fetch(url, {
        method,
        headers: {
          Authorization: `token ${this.token}`,
          Accept: "application/vnd.github.v3+json",
          "X-GitHub-Api-Version": "2022-11-28",
          ...(body ? { "Content-Type": "application/json" } : {}),
        },
        body: body ? JSON.stringify(body) : undefined,
      });
    } catch (err) {
      throw new Error(
        `Network error calling GitHub API ${method} ${path}: ${err instanceof Error ? err.message : String(err)}`
      );
    }

    if (!response.ok) {
      let errorBody: GitHubError;
      try {
        errorBody = await response.json() as GitHubError;
      } catch {
        errorBody = {
          message: `HTTP ${response.status} ${response.statusText}`,
          documentation_url: "https://docs.github.com/rest",
        };
      }
      throw new GitHubApiError(response.status, errorBody, response.headers);
    }

    const data = await response.json() as T;
    return { data, headers: response.headers };
  }

  async getRepository(owner: string, repo: string): Promise<GitHubRepository> {
    const { data } = await this.request<GitHubRepository>(
      "GET",
      `/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}`
    );

    // Validate critical response shape invariants
    if (typeof data.id !== "number") {
      throw new Error(`Repository id must be a number, got ${typeof data.id}`);
    }
    if (!FULL_NAME_REGEX.test(data.full_name)) {
      throw new Error(`Repository full_name must match owner/repo format, got "${data.full_name}"`);
    }
    if (!Array.isArray(data.topics)) {
      throw new Error(`Repository topics must be an array`);
    }
    if (!isValidISO8601(data.created_at)) {
      throw new Error(`Repository created_at is not valid ISO 8601: ${data.created_at}`);
    }

    return data;
  }

  async createIssue(
    owner: string,
    repo: string,
    params: CreateIssueRequest,
  ): Promise<GitHubIssue> {
    if (!params.title || params.title.trim().length === 0) {
      throw new Error("Issue title is required and cannot be empty.");
    }

    const { data } = await this.request<GitHubIssue>(
      "POST",
      `/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/issues`,
      params,
    );

    if (typeof data.id !== "number") {
      throw new Error(`Issue id must be a number, got ${typeof data.id}`);
    }
    if (typeof data.number !== "number") {
      throw new Error(`Issue number must be a number, got ${typeof data.number}`);
    }

    return data;
  }

  async listPullRequests(
    owner: string,
    repo: string,
    params: ListPullRequestsParams = {},
  ): Promise<PaginatedResult<GitHubPullRequest>> {
    const perPage = params.per_page ?? 30;
    if (perPage < 1 || perPage > 100) {
      throw new Error(`per_page must be between 1 and 100, got ${perPage}`);
    }
    const page = params.page ?? 1;
    if (page < 1) {
      throw new Error(`page must be >= 1, got ${page}`);
    }

    const queryParts: string[] = [];
    if (params.state) queryParts.push(`state=${params.state}`);
    if (params.sort) queryParts.push(`sort=${params.sort}`);
    if (params.direction) queryParts.push(`direction=${params.direction}`);
    queryParts.push(`per_page=${perPage}`);
    queryParts.push(`page=${page}`);
    const query = queryParts.join("&");

    const path = `/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/pulls?${query}`;
    const { data, headers } = await this.request<GitHubPullRequest[]>("GET", path);

    const linkInfo = parseLinkHeader(headers.get("Link"));

    return {
      data,
      hasNextPage: linkInfo.next !== undefined,
      nextPage: linkInfo.next ?? null,
    };
  }

  async listAllPullRequests(
    owner: string,
    repo: string,
    params: Omit<ListPullRequestsParams, "page"> = {},
  ): Promise<GitHubPullRequest[]> {
    const allPRs: GitHubPullRequest[] = [];
    let page = 1;

    while (true) {
      const result = await this.listPullRequests(owner, repo, {
        ...params,
        per_page: params.per_page ?? 100,
        page,
      });
      allPRs.push(...result.data);

      if (!result.hasNextPage || result.nextPage === null) {
        break;
      }
      page = result.nextPage;
    }

    return allPRs;
  }

  async updatePullRequest(
    owner: string,
    repo: string,
    pullNumber: number,
    params: UpdatePullRequestRequest,
  ): Promise<GitHubPullRequestDetail> {
    if (!Number.isInteger(pullNumber) || pullNumber < 1) {
      throw new Error(`pull_number must be a positive integer, got ${pullNumber}`);
    }

    const { data } = await this.request<GitHubPullRequestDetail>(
      "PATCH",
      `/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/pulls/${pullNumber}`,
      params,
    );

    if (typeof data.id !== "number") {
      throw new Error(`PR id must be a number, got ${typeof data.id}`);
    }

    return data;
  }

  isPullRequestMerged(pr: GitHubPullRequestDetail): boolean {
    return pr.merged === true && pr.merged_at !== null;
  }

  getPullRequestEffectiveState(
    pr: GitHubPullRequestDetail,
  ): "open" | "closed" | "merged" {
    if (pr.state === "open") return "open";
    if (pr.merged === true && pr.merged_at !== null) return "merged";
    return "closed";
  }

  parseTimestamp(isoString: string): Date {
    return parseISO8601(isoString);
  }

  compareTimestamps(a: string, b: string): number {
    const dateA = parseISO8601(a);
    const dateB = parseISO8601(b);
    return dateA.getTime() - dateB.getTime();
  }

  async addLabelsToIssue(
    owner: string,
    repo: string,
    issueNumber: number,
    labels: string[],
  ): Promise<GitHubLabel[]> {
    // Uses POST to ADD labels (not PUT which replaces all).
    // This avoids the read-modify-write race condition.
    const { data } = await this.request<GitHubLabel[]>(
      "POST",
      `/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/issues/${issueNumber}/labels`,
      { labels },
    );
    return data;
  }
}
```

### Tests

```typescript
// github-client.test.ts — Tests for GitHub REST API client.

import { describe, it, expect, vi, beforeEach } from "vitest";
import { GitHubClient, GitHubApiError } from "./github-types";

// --- Test fixtures ---

function makeRepoFixture(overrides: Partial<GitHubRepository> = {}): GitHubRepository {
  return {
    id: 123456,
    node_id: "R_kgDOAbCdEf",
    name: "my-repo",
    full_name: "octocat/my-repo",
    private: false,
    owner: {
      id: 1,
      login: "octocat",
      node_id: "MDQ6VXNlcjE=",
      avatar_url: "https://avatars.githubusercontent.com/u/1?v=4",
      type: "User" as const,
    },
    description: "A test repository",
    fork: false,
    url: "https://api.github.com/repos/octocat/my-repo",
    created_at: "2020-01-15T08:30:00Z",
    updated_at: "2026-02-10T12:00:00Z",
    pushed_at: "2026-02-10T12:00:00Z",
    stargazers_count: 42,
    watchers_count: 42,
    forks_count: 7,
    open_issues_count: 3,
    default_branch: "main",
    topics: ["typescript", "api"],
    visibility: "public" as const,
    license: { spdx_id: "MIT", name: "MIT License" },
    ...overrides,
  };
}

function makeIssueFixture(overrides: Partial<GitHubIssue> = {}): GitHubIssue {
  return {
    id: 789012,
    node_id: "I_kwDOAbCdEf",
    number: 42,
    title: "Fix the widget",
    body: "The widget is broken.",
    state: "open" as const,
    user: {
      id: 1,
      login: "octocat",
      node_id: "MDQ6VXNlcjE=",
      avatar_url: "https://avatars.githubusercontent.com/u/1?v=4",
      type: "User" as const,
    },
    labels: [],
    assignee: null,
    assignees: [],
    milestone: null,
    locked: false,
    created_at: "2026-02-01T10:00:00Z",
    updated_at: "2026-02-01T10:00:00Z",
    closed_at: null,
    ...overrides,
  };
}

function makePRFixture(overrides: Partial<GitHubPullRequest> = {}): GitHubPullRequest {
  return {
    id: 345678,
    node_id: "PR_kwDOAbCdEf",
    number: 99,
    title: "Add feature X",
    body: "This PR adds feature X.",
    state: "open" as const,
    user: {
      id: 1,
      login: "octocat",
      node_id: "MDQ6VXNlcjE=",
      avatar_url: "https://avatars.githubusercontent.com/u/1?v=4",
      type: "User" as const,
    },
    labels: [],
    locked: false,
    created_at: "2026-02-05T09:00:00Z",
    updated_at: "2026-02-05T09:00:00Z",
    closed_at: null,
    merged_at: null,
    draft: false,
    head: { ref: "feature-x", sha: "abc123" },
    base: { ref: "main", sha: "def456" },
    ...overrides,
  };
}

function makePRDetailFixture(
  overrides: Partial<GitHubPullRequestDetail> = {},
): GitHubPullRequestDetail {
  return {
    id: 345678,
    node_id: "PR_kwDOAbCdEf",
    number: 99,
    title: "Add feature X",
    body: "This PR adds feature X.",
    state: "open" as const,
    merged: false,
    merged_at: null,
    user: {
      id: 1,
      login: "octocat",
      node_id: "MDQ6VXNlcjE=",
      avatar_url: "https://avatars.githubusercontent.com/u/1?v=4",
      type: "User" as const,
    },
    labels: [],
    created_at: "2026-02-05T09:00:00Z",
    updated_at: "2026-02-05T09:00:00Z",
    closed_at: null,
    ...overrides,
  };
}

function mockFetchSuccess(body: unknown, headers: Record<string, string> = {}): void {
  vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
    new Response(JSON.stringify(body), {
      status: 200,
      statusText: "OK",
      headers: new Headers(headers),
    }),
  );
}

function mockFetchCreated(body: unknown): void {
  vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
    new Response(JSON.stringify(body), {
      status: 201,
      statusText: "Created",
      headers: new Headers(),
    }),
  );
}

function mockFetchError(status: number, body: GitHubError, headers: Record<string, string> = {}): void {
  vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
    new Response(JSON.stringify(body), {
      status,
      statusText: "Error",
      headers: new Headers(headers),
    }),
  );
}

function mockFetchNetworkError(message: string): void {
  vi.spyOn(globalThis, "fetch").mockRejectedValueOnce(new Error(message));
}

const VALID_TOKEN = "ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdef01";

// --- Tests ---

describe("GitHubClient", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  // -- Constructor --

  describe("constructor", () => {
    it("accepts a valid ghp_ token", () => {
      expect(() => new GitHubClient(VALID_TOKEN)).not.toThrow();
    });

    it("accepts a valid github_pat_ token", () => {
      const pat = "github_pat_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdef01234567890123";
      expect(() => new GitHubClient(pat)).not.toThrow();
    });

    it("rejects an empty token", () => {
      expect(() => new GitHubClient("")).toThrow("Invalid GitHub token format");
    });

    it("rejects a malformed token", () => {
      expect(() => new GitHubClient("not-a-token")).toThrow("Invalid GitHub token format");
    });
  });

  // -- getRepository --

  describe("getRepository", () => {
    it("returns a fully typed repository on success", async () => {
      const repo = makeRepoFixture();
      mockFetchSuccess(repo);
      const client = new GitHubClient(VALID_TOKEN);

      const result = await client.getRepository("octocat", "my-repo");

      expect(result.id).toBe(123456);
      expect(typeof result.id).toBe("number");
      expect(result.full_name).toBe("octocat/my-repo");
      expect(result.owner.login).toBe("octocat");
      expect(result.owner.type).toBe("User");
      expect(Array.isArray(result.topics)).toBe(true);
      expect(result.topics).toEqual(["typescript", "api"]);
      expect(result.visibility).toBe("public");
      expect(result.description).toBe("A test repository");
      expect(result.license).toEqual({ spdx_id: "MIT", name: "MIT License" });
    });

    it("handles a repository with null description and license", async () => {
      const repo = makeRepoFixture({ description: null, license: null });
      mockFetchSuccess(repo);
      const client = new GitHubClient(VALID_TOKEN);

      const result = await client.getRepository("octocat", "my-repo");

      expect(result.description).toBeNull();
      expect(result.license).toBeNull();
    });

    it("throws GitHubApiError on 404", async () => {
      mockFetchError(404, {
        message: "Not Found",
        documentation_url: "https://docs.github.com/rest/repos/repos#get-a-repository",
      });
      const client = new GitHubClient(VALID_TOKEN);

      await expect(client.getRepository("octocat", "nonexistent"))
        .rejects.toThrow(GitHubApiError);

      try {
        await client.getRepository("octocat", "nonexistent");
      } catch (err) {
        // Second call to inspect error properties - remock
        mockFetchError(404, {
          message: "Not Found",
          documentation_url: "https://docs.github.com/rest/repos/repos#get-a-repository",
        });
        try {
          await client.getRepository("octocat", "nonexistent");
        } catch (e) {
          const apiErr = e as GitHubApiError;
          expect(apiErr.status).toBe(404);
          expect(apiErr.isNotFound).toBe(true);
          expect(apiErr.errorBody.message).toBe("Not Found");
          expect(apiErr.errorBody.documentation_url).toContain("docs.github.com");
        }
      }
    });

    it("throws GitHubApiError on 403 rate limit", async () => {
      mockFetchError(
        403,
        {
          message: "API rate limit exceeded",
          documentation_url: "https://docs.github.com/rest/rate-limit",
        },
        {
          "X-RateLimit-Remaining": "0",
          "X-RateLimit-Reset": "1707580800",
        },
      );
      const client = new GitHubClient(VALID_TOKEN);

      try {
        await client.getRepository("octocat", "my-repo");
        expect.unreachable("Should have thrown");
      } catch (err) {
        const apiErr = err as GitHubApiError;
        expect(apiErr.status).toBe(403);
        expect(apiErr.isRateLimited).toBe(true);
        expect(apiErr.rateLimitResetsAt).toBeInstanceOf(Date);
      }
    });

    it("throws on network failure", async () => {
      mockFetchNetworkError("DNS resolution failed");
      const client = new GitHubClient(VALID_TOKEN);

      await expect(client.getRepository("octocat", "my-repo"))
        .rejects.toThrow("Network error calling GitHub API");
    });

    it("validates that id is a number in the response", async () => {
      const badRepo = makeRepoFixture({ id: "not-a-number" as unknown as number });
      mockFetchSuccess(badRepo);
      const client = new GitHubClient(VALID_TOKEN);

      await expect(client.getRepository("octocat", "my-repo"))
        .rejects.toThrow("Repository id must be a number");
    });

    it("validates full_name format", async () => {
      const badRepo = makeRepoFixture({ full_name: "no-slash-here" });
      mockFetchSuccess(badRepo);
      const client = new GitHubClient(VALID_TOKEN);

      await expect(client.getRepository("octocat", "my-repo"))
        .rejects.toThrow("must match owner/repo format");
    });

    it("validates created_at is ISO 8601", async () => {
      const badRepo = makeRepoFixture({ created_at: "not-a-date" });
      mockFetchSuccess(badRepo);
      const client = new GitHubClient(VALID_TOKEN);

      await expect(client.getRepository("octocat", "my-repo"))
        .rejects.toThrow("not valid ISO 8601");
    });
  });

  // -- createIssue --

  describe("createIssue", () => {
    it("creates an issue and returns typed response", async () => {
      const issue = makeIssueFixture();
      mockFetchCreated(issue);
      const client = new GitHubClient(VALID_TOKEN);

      const result = await client.createIssue("octocat", "my-repo", {
        title: "Fix the widget",
        body: "The widget is broken.",
        labels: ["bug"],
      });

      expect(result.id).toBe(789012);
      expect(typeof result.id).toBe("number");
      expect(result.number).toBe(42);
      expect(typeof result.number).toBe("number");
      expect(result.title).toBe("Fix the widget");
      expect(result.state).toBe("open");
      expect(result.assignee).toBeNull();
      expect(result.milestone).toBeNull();
      expect(result.body).toBe("The widget is broken.");
    });

    it("rejects empty title before calling API", async () => {
      const client = new GitHubClient(VALID_TOKEN);

      await expect(
        client.createIssue("octocat", "my-repo", { title: "" }),
      ).rejects.toThrow("title is required and cannot be empty");
    });

    it("rejects whitespace-only title", async () => {
      const client = new GitHubClient(VALID_TOKEN);

      await expect(
        client.createIssue("octocat", "my-repo", { title: "   " }),
      ).rejects.toThrow("title is required and cannot be empty");
    });

    it("handles 422 validation error", async () => {
      mockFetchError(422, {
        message: "Validation Failed",
        documentation_url: "https://docs.github.com/rest/issues/issues#create-an-issue",
        errors: [
          { resource: "Issue", field: "title", code: "missing_field" },
        ],
      });
      const client = new GitHubClient(VALID_TOKEN);

      try {
        await client.createIssue("octocat", "my-repo", { title: "test" });
        expect.unreachable("Should have thrown");
      } catch (err) {
        const apiErr = err as GitHubApiError;
        expect(apiErr.status).toBe(422);
        expect(apiErr.validationErrors).toHaveLength(1);
        expect(apiErr.validationErrors[0].field).toBe("title");
        expect(apiErr.validationErrors[0].code).toBe("missing_field");
      }
    });

    it("handles 410 Gone for locked repository", async () => {
      mockFetchError(410, {
        message: "Issues are disabled for this repo",
        documentation_url: "https://docs.github.com/rest/issues/issues#create-an-issue",
      });
      const client = new GitHubClient(VALID_TOKEN);

      try {
        await client.createIssue("octocat", "archived-repo", { title: "test" });
        expect.unreachable("Should have thrown");
      } catch (err) {
        const apiErr = err as GitHubApiError;
        expect(apiErr.status).toBe(410);
      }
    });

    it("handles 503 service unavailable", async () => {
      mockFetchError(503, {
        message: "Service temporarily unavailable",
        documentation_url: "https://docs.github.com/rest",
      });
      const client = new GitHubClient(VALID_TOKEN);

      await expect(
        client.createIssue("octocat", "my-repo", { title: "test" }),
      ).rejects.toThrow(GitHubApiError);
    });
  });

  // -- listPullRequests --

  describe("listPullRequests", () => {
    it("returns paginated pull request list", async () => {
      const prs = [makePRFixture()];
      mockFetchSuccess(prs, {
        Link: '<https://api.github.com/repos/octocat/my-repo/pulls?page=2>; rel="next"',
      });
      const client = new GitHubClient(VALID_TOKEN);

      const result = await client.listPullRequests("octocat", "my-repo");

      expect(result.data).toHaveLength(1);
      expect(result.data[0].number).toBe(99);
      expect(result.data[0].state).toBe("open");
      expect(result.data[0].draft).toBe(false);
      expect(result.hasNextPage).toBe(true);
      expect(result.nextPage).toBe(2);
    });

    it("returns hasNextPage=false when no Link header", async () => {
      mockFetchSuccess([makePRFixture()]);
      const client = new GitHubClient(VALID_TOKEN);

      const result = await client.listPullRequests("octocat", "my-repo");

      expect(result.hasNextPage).toBe(false);
      expect(result.nextPage).toBeNull();
    });

    it("validates per_page range (too high)", async () => {
      const client = new GitHubClient(VALID_TOKEN);

      await expect(
        client.listPullRequests("octocat", "my-repo", { per_page: 500 }),
      ).rejects.toThrow("per_page must be between 1 and 100");
    });

    it("validates per_page range (too low)", async () => {
      const client = new GitHubClient(VALID_TOKEN);

      await expect(
        client.listPullRequests("octocat", "my-repo", { per_page: 0 }),
      ).rejects.toThrow("per_page must be between 1 and 100");
    });

    it("validates page range", async () => {
      const client = new GitHubClient(VALID_TOKEN);

      await expect(
        client.listPullRequests("octocat", "my-repo", { page: 0 }),
      ).rejects.toThrow("page must be >= 1");
    });

    it("handles 304 Not Modified", async () => {
      mockFetchError(304, {
        message: "Not Modified",
        documentation_url: "https://docs.github.com/rest",
      });
      const client = new GitHubClient(VALID_TOKEN);

      await expect(
        client.listPullRequests("octocat", "my-repo"),
      ).rejects.toThrow(GitHubApiError);
    });

    it("handles PR with null body and null closed_at", async () => {
      const pr = makePRFixture({ body: null, closed_at: null });
      mockFetchSuccess([pr]);
      const client = new GitHubClient(VALID_TOKEN);

      const result = await client.listPullRequests("octocat", "my-repo");

      expect(result.data[0].body).toBeNull();
      expect(result.data[0].closed_at).toBeNull();
    });
  });

  // -- listAllPullRequests (full pagination) --

  describe("listAllPullRequests", () => {
    it("fetches all pages until no next link", async () => {
      const page1 = [makePRFixture({ number: 1 })];
      const page2 = [makePRFixture({ number: 2 })];
      const page3 = [makePRFixture({ number: 3 })];

      mockFetchSuccess(page1, {
        Link: '<https://api.github.com/repos/o/r/pulls?page=2>; rel="next"',
      });
      mockFetchSuccess(page2, {
        Link: '<https://api.github.com/repos/o/r/pulls?page=3>; rel="next"',
      });
      mockFetchSuccess(page3); // No Link header = last page

      const client = new GitHubClient(VALID_TOKEN);
      const result = await client.listAllPullRequests("octocat", "my-repo");

      expect(result).toHaveLength(3);
      expect(result[0].number).toBe(1);
      expect(result[1].number).toBe(2);
      expect(result[2].number).toBe(3);
    });
  });

  // -- updatePullRequest --

  describe("updatePullRequest", () => {
    it("updates title and returns full PR detail", async () => {
      const detail = makePRDetailFixture({ title: "Updated title" });
      mockFetchSuccess(detail);
      const client = new GitHubClient(VALID_TOKEN);

      const result = await client.updatePullRequest("octocat", "my-repo", 99, {
        title: "Updated title",
      });

      expect(result.title).toBe("Updated title");
      expect(result.merged).toBe(false);
      expect(result.merged_at).toBeNull();
    });

    it("rejects non-integer pull number", async () => {
      const client = new GitHubClient(VALID_TOKEN);

      await expect(
        client.updatePullRequest("octocat", "my-repo", 1.5, { title: "x" }),
      ).rejects.toThrow("must be a positive integer");
    });

    it("rejects zero pull number", async () => {
      const client = new GitHubClient(VALID_TOKEN);

      await expect(
        client.updatePullRequest("octocat", "my-repo", 0, { title: "x" }),
      ).rejects.toThrow("must be a positive integer");
    });

    it("handles 422 for invalid base branch", async () => {
      mockFetchError(422, {
        message: "Validation Failed",
        documentation_url: "https://docs.github.com/rest",
        errors: [{ resource: "PullRequest", field: "base", code: "invalid" }],
      });
      const client = new GitHubClient(VALID_TOKEN);

      try {
        await client.updatePullRequest("octocat", "my-repo", 99, {
          base: "nonexistent-branch",
        });
        expect.unreachable("Should have thrown");
      } catch (err) {
        const apiErr = err as GitHubApiError;
        expect(apiErr.status).toBe(422);
        expect(apiErr.validationErrors[0].code).toBe("invalid");
      }
    });
  });

  // -- PR state helpers --

  describe("getPullRequestEffectiveState", () => {
    it("returns 'open' for open PR", () => {
      const client = new GitHubClient(VALID_TOKEN);
      const pr = makePRDetailFixture({ state: "open", merged: false, merged_at: null });
      expect(client.getPullRequestEffectiveState(pr)).toBe("open");
    });

    it("returns 'merged' for merged PR", () => {
      const client = new GitHubClient(VALID_TOKEN);
      const pr = makePRDetailFixture({
        state: "closed",
        merged: true,
        merged_at: "2026-02-10T15:00:00Z",
      });
      expect(client.getPullRequestEffectiveState(pr)).toBe("merged");
    });

    it("returns 'closed' for closed-not-merged PR", () => {
      const client = new GitHubClient(VALID_TOKEN);
      const pr = makePRDetailFixture({
        state: "closed",
        merged: false,
        merged_at: null,
        closed_at: "2026-02-10T15:00:00Z",
      });
      expect(client.getPullRequestEffectiveState(pr)).toBe("closed");
    });
  });

  // -- Timestamp handling --

  describe("timestamp handling", () => {
    it("parses valid ISO 8601 timestamps", () => {
      const client = new GitHubClient(VALID_TOKEN);
      const date = client.parseTimestamp("2026-02-10T12:30:00Z");
      expect(date).toBeInstanceOf(Date);
      expect(date.getUTCFullYear()).toBe(2026);
      expect(date.getUTCMonth()).toBe(1); // 0-indexed
      expect(date.getUTCDate()).toBe(10);
    });

    it("rejects invalid timestamp format", () => {
      const client = new GitHubClient(VALID_TOKEN);
      expect(() => client.parseTimestamp("2026-1-9")).toThrow("Invalid ISO 8601");
    });

    it("correctly compares two timestamps", () => {
      const client = new GitHubClient(VALID_TOKEN);
      const cmp = client.compareTimestamps(
        "2026-01-09T00:00:00Z",
        "2026-01-10T00:00:00Z",
      );
      expect(cmp).toBeLessThan(0);
    });
  });

  // -- addLabelsToIssue (race-condition-safe) --

  describe("addLabelsToIssue", () => {
    it("uses POST to add labels without overwriting existing ones", async () => {
      const labels = [
        { id: 1, node_id: "LA_1", name: "bug", description: null, color: "d73a4a", default: true },
        { id: 2, node_id: "LA_2", name: "enhancement", description: "New feature", color: "a2eeef", default: false },
      ];
      mockFetchSuccess(labels);
      const client = new GitHubClient(VALID_TOKEN);

      const result = await client.addLabelsToIssue("octocat", "my-repo", 42, ["bug", "enhancement"]);

      expect(result).toHaveLength(2);
      expect(result[0].name).toBe("bug");

      // Verify POST was used (not PUT)
      const fetchCall = vi.mocked(fetch).mock.calls[0];
      expect(fetchCall[1]?.method).toBe("POST");
    });
  });
});
```

**Expected result: 0 violations.**

---

## B01 — No Error Handling

No try/catch around fetch calls. Network errors and non-2xx responses crash the caller with unhandled exceptions.

```typescript
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
```

**Expected violation:**

- **Rule:** `TQ-error-path-coverage`
- **Message:** `fetch() call in getRepository has no try/catch or .catch() handler. Network errors will propagate as unhandled rejections.`
- **Production impact:** DNS failures, TLS errors, or timeouts cause unhandled promise rejections that crash the Node.js process (with `--unhandled-rejections=throw`) or silently disappear. No retry, no logging, no graceful degradation.

---

## B02 — No Status Code Check

Calls the API and parses the response body without checking `response.ok` or `response.status`. A 404 or 422 error body is silently treated as a valid repository or issue.

```typescript
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
```

**Expected violation:**

- **Rule:** `CTR-status-code-handling`
- **Message:** `Response from GET /repos/{owner}/{repo} is used without checking status code. Server returns status codes [200, 301, 403, 404] but client handles none of them distinctly.`
- **Production impact:** A 404 error body `{ message: "Not Found" }` is returned as a GitHubRepository object. Downstream code accesses `repo.id` (undefined), `repo.full_name` (undefined), and `repo.owner.login` (TypeError: Cannot read properties of undefined). The error surfaces far from its origin, making debugging extremely difficult.

---

## B03 — Shallow Test Assertions

Tests exist but only assert that the result is defined, never inspecting field values, types, or shape.

```typescript
// B03: Shallow test assertions — only checks existence, never shape.

describe("GitHubClient (B03 — shallow)", () => {
  it("gets a repository", async () => {
    mockFetchSuccess(makeRepoFixture());
    const client = new GitHubClient(VALID_TOKEN);
    const repo = await client.getRepository("octocat", "my-repo");

    // BUG: Only checks that repo exists. Does not verify id is a number,
    // full_name matches format, topics is an array, owner is a User, etc.
    expect(repo).toBeDefined();
    expect(repo).not.toBeNull();
  });

  it("creates an issue", async () => {
    mockFetchCreated(makeIssueFixture());
    const client = new GitHubClient(VALID_TOKEN);
    const issue = await client.createIssue("octocat", "my-repo", {
      title: "Test",
    });

    // BUG: Truthy check passes for any non-null object — even an error body.
    expect(issue).toBeTruthy();
  });

  it("lists pull requests", async () => {
    mockFetchSuccess([makePRFixture()]);
    const client = new GitHubClient(VALID_TOKEN);
    const result = await client.listPullRequests("octocat", "my-repo");

    // BUG: Checks array has items but not what those items contain.
    expect(result.data.length).toBeGreaterThan(0);
  });

  it("updates a pull request", async () => {
    mockFetchSuccess(makePRDetailFixture());
    const client = new GitHubClient(VALID_TOKEN);
    const pr = await client.updatePullRequest("octocat", "my-repo", 99, {
      title: "New title",
    });

    // BUG: Only checks existence. merged could be any type.
    expect(pr).toBeDefined();
    expect(pr.merged).toBeDefined();
  });
});
```

**Expected violation:**

- **Rule:** `TQ-no-shallow-assertions`
- **Message:** `Test "gets a repository" uses only toBeDefined()/toBeTruthy() assertions. No field values, types, or shapes are verified. This test passes even if the response is { message: "Not Found" }.`
- **Production impact:** Tests provide a false sense of security. A type mismatch in the response (e.g., `id` is suddenly a string, `topics` is missing) would not be caught. Regressions ship to production undetected because tests never actually verify the contract.

---

## B04 — Missing Negative Tests

Only happy-path tests exist. No tests for 404, 403, 422, 410, rate limiting, network errors, or invalid inputs.

```typescript
// B04: Missing negative tests — only tests the happy path.

describe("GitHubClient (B04 — no negative tests)", () => {
  it("gets a repository", async () => {
    mockFetchSuccess(makeRepoFixture());
    const client = new GitHubClient(VALID_TOKEN);
    const repo = await client.getRepository("octocat", "my-repo");
    expect(repo.id).toBe(123456);
    expect(repo.full_name).toBe("octocat/my-repo");
  });

  it("creates an issue", async () => {
    mockFetchCreated(makeIssueFixture());
    const client = new GitHubClient(VALID_TOKEN);
    const issue = await client.createIssue("octocat", "my-repo", {
      title: "Test issue",
      body: "Body text",
    });
    expect(issue.number).toBe(42);
    expect(issue.state).toBe("open");
  });

  it("lists pull requests", async () => {
    mockFetchSuccess([makePRFixture()]);
    const client = new GitHubClient(VALID_TOKEN);
    const result = await client.listPullRequests("octocat", "my-repo");
    expect(result.data).toHaveLength(1);
  });

  // BUG: No tests for any of these scenarios:
  //   - 404 Not Found
  //   - 403 Forbidden / rate limited
  //   - 422 Validation Failed
  //   - 410 Gone
  //   - 503 Service Unavailable
  //   - Network timeout / DNS failure
  //   - Empty string title
  //   - Invalid per_page values
  //   - Invalid token format
  //   - Null assignee/body/milestone handling
});
```

**Expected violation:**

- **Rule:** `TQ-negative-cases`
- **Message:** `createIssue has 1 test but 0 negative tests. Manifest declares status codes [201, 403, 404, 410, 422, 503] — none of the error codes are tested.`
- **Production impact:** Error handling code (if it exists) is never exercised. A regression in 422 handling could cause the client to crash on validation errors instead of surfacing them. Rate limit handling is untested, meaning a production outage from mishandled 403s would not be caught pre-deployment.

---

## B05 — Request Missing Required Fields

Creates an issue without the `title` field, which is marked as required in the manifest. The client sends a malformed request that the server will reject with 422.

```typescript
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
```

**Expected violation:**

- **Rule:** `CTR-request-shape`
- **Message:** `POST /repos/{owner}/{repo}/issues request body is missing required field "title" (type: string). Manifest declares title as required but client type omits it.`
- **Production impact:** Every call to `createIssue` results in a 422 Validation Failed error from GitHub. The feature is completely non-functional. Since the parameter type lacks `title`, callers cannot fix this without modifying the client library.

---

## B06 — Response Type Mismatch

The client's `Repository` type is missing the `topics` array field that the API always returns and the manifest declares as required.

```typescript
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
```

**Expected violation:**

- **Rule:** `CTR-response-shape`
- **Message:** `Client type RepositoryB06 for GET /repos/{owner}/{repo} is missing required field "topics" (type: array<string>). The API response includes this field but the client type does not declare it, causing data loss.`
- **Production impact:** The `topics` array is present in every API response but inaccessible through the client type. Features that depend on repository topics (search, filtering, display) cannot function. Developers resort to `as any` casts or raw JSON access, defeating TypeScript's type safety.

---

## B07 — Wrong Field Types

Repository `id` is stored as a string instead of a number. The manifest and API both return integers, but the client type declares string.

```typescript
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
```

**Expected violation:**

- **Rule:** `CTR-manifest-conformance`
- **Message:** `Field "id" in RepositoryB07 is declared as string but manifest specifies type: integer. The API returns a numeric id (e.g., 123456) which will mismatch the client's string type.`
- **Production impact:** Equality comparisons fail silently. `repos.find(r => r.id === String(targetId))` never matches because the runtime value is a number, not a string. Map lookups, Set membership checks, and database foreign key comparisons all produce incorrect results. The bug is invisible in logs because `console.log(repo.id)` outputs `123456` regardless of type.

---

## B08 — Incomplete Enum Handling

Handles "open" and "closed" states for pull requests but does not check `merged`/`merged_at` to detect the merged state. A merged PR shows as "closed".

```typescript
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
```

**Expected violation:**

- **Rule:** `CTR-strictness-parity`
- **Message:** `getPullRequestDisplayState handles PR states ["open", "closed"] but does not check the "merged" boolean field. GitHub represents merged PRs as state="closed" + merged=true. The effective state space is {open, closed, merged} but only {open, closed} is handled.`
- **Production impact:** Merged PRs display as "Closed" in the UI. Users cannot distinguish merged contributions from abandoned PRs. The `filterByState("merged")` function returns an empty array, making the "show merged PRs" feature non-functional. Metrics dashboards undercount merge rates and overcount closures.

---

## B09 — Missing Range Validation

Sends `per_page=500` to the list pull requests endpoint. GitHub silently clamps this to 100, but the client does not validate the range before sending, causing inconsistency between requested and actual page sizes.

```typescript
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
```

**Expected violation:**

- **Rule:** `CTR-strictness-parity`
- **Message:** `Parameter "per_page" sent with value 500 but manifest declares range [1, 100]. Client does not validate this range. GitHub silently clamps to 100, causing a mismatch between requested and actual page size.`
- **Production impact:** Pagination logic breaks. If the client requests 500 items per page and gets 100, it may conclude that only 100 items exist total and skip remaining pages. This causes data loss: reports show partial PR lists, export features miss entries, and aggregation queries produce incorrect counts.

---

## B10 — Format Not Validated

Accepts any string as an ISO 8601 timestamp without parsing or validating the format. Stores and compares timestamps as raw strings.

```typescript
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
```

**Expected violation:**

- **Rule:** `CTR-strictness-parity`
- **Message:** `Field "updated_at" in GET /repos/{owner}/{repo} response has format: iso8601 in manifest but client performs no format validation. The string is used directly in Date constructor without verifying ISO 8601 compliance.`
- **Production impact:** Corrupted or transformed timestamps silently produce `Invalid Date` objects. `isRecentlyUpdated` always returns `false` for invalid dates instead of signaling an error. A proxy server that reformats dates (e.g., to local timezone format "Feb 10, 2026") breaks the client silently. Sorting by date produces arbitrary orderings.

---

## B11 — Precision Loss on Date Comparison

Compares dates as raw strings instead of parsing them into `Date` objects. String comparison works for most ISO 8601 dates but fails for non-zero-padded formats or timezone-offset variants.

```typescript
// B11: Precision loss — compares ISO 8601 dates as raw strings.

export class GitHubClientB11 {
  private readonly token: string;

  constructor(token: string) {
    if (!/^(ghp_|github_pat_)/.test(token)) throw new Error("Invalid token");
    this.token = token;
  }

  sortIssuesByDate(issues: GitHubIssue[]): GitHubIssue[] {
    // BUG: Sorting by string comparison of ISO 8601 timestamps.
    // This works for standard GitHub responses ("2026-01-09T00:00:00Z")
    // because zero-padded ISO 8601 strings sort lexicographically.
    //
    // But it breaks when:
    // - Dates come from a transform that drops zero-padding: "2026-1-9" < "2026-1-10"
    //   String comparison: "2026-1-9" > "2026-1-10" because "9" > "1" (wrong!)
    // - Timezone offsets are present: "2026-01-10T00:00:00+05:00" vs "2026-01-09T23:00:00Z"
    //   These are the same instant, but string comparison says first > second.
    return [...issues].sort((a, b) => {
      if (a.created_at < b.created_at) return -1;
      if (a.created_at > b.created_at) return 1;
      return 0;
    });
  }

  findIssuesCreatedBetween(
    issues: GitHubIssue[],
    start: string,
    end: string,
  ): GitHubIssue[] {
    // BUG: String range comparison fails for the same reasons as above.
    return issues.filter(
      (issue) => issue.created_at >= start && issue.created_at <= end,
    );
  }

  getMostRecentIssue(issues: GitHubIssue[]): GitHubIssue | null {
    if (issues.length === 0) return null;
    // BUG: String max comparison — same problem.
    return issues.reduce((latest, issue) =>
      issue.created_at > latest.created_at ? issue : latest,
    );
  }
}
```

**Expected violation:**

- **Rule:** `CTR-strictness-parity`
- **Message:** `Timestamp fields "created_at" compared as raw strings instead of parsed Date objects. String comparison of ISO 8601 timestamps produces incorrect ordering for non-zero-padded values or timezone-offset variants.`
- **Production impact:** Issue sorting appears correct in test fixtures (which use zero-padded UTC timestamps) but fails in edge cases. If any middleware, cache, or transform layer produces non-standard date formatting, the sort order becomes incorrect. A timeline view shows issues out of order, "most recent" queries return wrong results, and date-range filters include or exclude wrong entries. The bug is intermittent and data-dependent, making it extremely hard to reproduce.

---

## B12 — Nullable Field Crash

Accesses `issue.assignee.login` without checking that `assignee` can be null. The manifest declares `assignee` as `nullable: true`.

```typescript
// B12: Nullable field crash — accesses assignee.login without null check.

export class GitHubClientB12 {
  private readonly token: string;

  constructor(token: string) {
    if (!/^(ghp_|github_pat_)/.test(token)) throw new Error("Invalid token");
    this.token = token;
  }

  async getIssue(
    owner: string,
    repo: string,
    issueNumber: number,
  ): Promise<GitHubIssue> {
    const response = await fetch(
      `https://api.github.com/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/issues/${issueNumber}`,
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
    return response.json() as Promise<GitHubIssue>;
  }

  getAssigneeLogin(issue: GitHubIssue): string {
    // BUG: issue.assignee is nullable (GitHubUser | null).
    // When no one is assigned, assignee is null.
    // null.login throws TypeError: Cannot read properties of null.
    return issue.assignee.login;
  }

  formatIssueAssignment(issue: GitHubIssue): string {
    // BUG: Same null dereference in string interpolation.
    return `Issue #${issue.number} assigned to ${issue.assignee.login}`;
  }

  getAssigneeAvatarUrl(issue: GitHubIssue): string {
    // BUG: Triple null-unsafe access chain. Even if assignee exists,
    // this pattern is fragile and would crash on null.
    return issue.assignee.avatar_url;
  }
}
```

**Expected violation:**

- **Rule:** `CTR-response-shape`
- **Message:** `Field "assignee" in GitHubIssue is declared nullable (type: User, nullable: true) but client accesses "issue.assignee.login" without null check. This will throw TypeError when assignee is null.`
- **Production impact:** Unassigned issues (which are extremely common -- most new issues have no assignee) crash the application with `TypeError: Cannot read properties of null (reading 'login')`. This is a hard crash, not a graceful degradation. In a web server, it returns a 500 to the user. In a CLI tool, it terminates the process. The bug affects the most common case (unassigned issues) rather than an edge case.

---

## B13 — Missing Auth Token Validation

Sends API requests without validating that the token is present or has the correct format. An empty string or malformed token is silently sent as the Authorization header.

```typescript
// B13: Missing auth token validation — no check on token format or presence.

export class GitHubClientB13 {
  private readonly token: string;

  constructor(token: string) {
    // BUG: No validation. Accepts empty string, undefined coerced to string,
    // arbitrary strings, expired tokens, tokens with whitespace, etc.
    this.token = token;
  }

  async getRepository(owner: string, repo: string): Promise<GitHubRepository> {
    try {
      const response = await fetch(
        `https://api.github.com/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}`,
        {
          headers: {
            // BUG: Sends "token " (empty) or "token not-a-real-token".
            // GitHub returns 401 Unauthorized, which is not in the
            // manifest's status_codes [200, 301, 403, 404].
            Authorization: `token ${this.token}`,
            Accept: "application/vnd.github.v3+json",
          },
        },
      );
      if (!response.ok) {
        const err = await response.json();
        throw new GitHubApiError(response.status, err, response.headers);
      }
      return response.json() as Promise<GitHubRepository>;
    } catch (err) {
      if (err instanceof GitHubApiError) throw err;
      throw new Error(`Network error: ${err}`);
    }
  }
}

// Usage that silently fails:
const client = new GitHubClientB13("");  // No error thrown
// First API call fails with 401, wasting a round-trip and leaking
// the fact that we're making unauthenticated requests.

const client2 = new GitHubClientB13("some random string");
// Sends "Authorization: token some random string" — immediately
// flagged as suspicious by GitHub's abuse detection.
```

**Expected violation:**

- **Rule:** `CTR-request-shape`
- **Message:** `Authorization header sends token without validating format. Manifest declares auth format "token ghp_*|github_pat_*" but constructor accepts any string, including empty. Invalid tokens cause 401 responses not declared in the endpoint's status_codes.`
- **Production impact:** Misconfigured environments (missing `GITHUB_TOKEN` env var, expired token, copy-paste error with whitespace) are not caught at construction time. The error surfaces as an opaque 401 on the first API call, after the network round-trip. Worse, repeatedly sending malformed tokens triggers GitHub's abuse detection, which can IP-ban the entire CI/CD fleet. Fail-fast validation at construction time prevents all of these issues.

---

## B14 — Pagination Terminated Early

Fetches the first page of pull requests but ignores the `Link` header's `rel="next"` indicator. Returns only the first page as the complete result.

```typescript
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
```

**Expected violation:**

- **Rule:** `CTR-response-shape`
- **Message:** `listAllPullRequests reads response body but ignores the Link response header. Manifest declares pagination style "link-header" with rel="next" but client does not parse or follow pagination links. Only the first page of results is returned.`
- **Production impact:** Any repository with more than 100 pull requests returns incomplete data. `countAllPullRequests` reports 100 instead of the actual count. `findPullRequestByBranch` fails to find PRs that exist but are on subsequent pages, causing CI/CD integrations to create duplicate PRs or skip status checks. The bug is invisible on small repositories and only manifests at scale, making it a ticking time bomb.

---

## B15 — Race Condition (Read-Modify-Write)

Reads current issue labels, appends a new label to the array, then writes the full label set back with PUT. This overwrites any labels added by concurrent operations between the read and write.

```typescript
// B15: Race condition — read-modify-write on issue labels overwrites concurrent changes.

export class GitHubClientB15 {
  private readonly token: string;

  constructor(token: string) {
    if (!/^(ghp_|github_pat_)/.test(token)) throw new Error("Invalid token");
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

  async addLabelToIssue(
    owner: string,
    repo: string,
    issueNumber: number,
    newLabel: string,
  ): Promise<GitHubLabel[]> {
    // Step 1: READ current labels
    const currentLabels = await this.request<GitHubLabel[]>(
      "GET",
      `/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/issues/${issueNumber}/labels`,
    );

    // BUG: Race condition window. Between the GET above and the PUT below,
    // another process/user/webhook could add or remove labels.
    // Those changes will be silently overwritten.

    // Step 2: MODIFY — append new label to the list
    const labelNames = currentLabels.map((l) => l.name);
    if (!labelNames.includes(newLabel)) {
      labelNames.push(newLabel);
    }

    // Step 3: WRITE — replace ALL labels with the modified list (PUT)
    // This overwrites whatever is currently on the issue, including any
    // labels added between Step 1 and Step 3.
    const updatedLabels = await this.request<GitHubLabel[]>(
      "PUT",
      `/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/issues/${issueNumber}/labels`,
      { labels: labelNames },
    );

    return updatedLabels;

    // CORRECT approach: Use POST /repos/{owner}/{repo}/issues/{number}/labels
    // which atomically ADDS labels without replacing existing ones:
    //   POST /repos/o/r/issues/42/labels { labels: ["new-label"] }
    // This is what the PERFECT implementation does.
  }

  async removeLabelFromIssue(
    owner: string,
    repo: string,
    issueNumber: number,
    labelToRemove: string,
  ): Promise<GitHubLabel[]> {
    // Same read-modify-write pattern with the same race condition.
    const currentLabels = await this.request<GitHubLabel[]>(
      "GET",
      `/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/issues/${issueNumber}/labels`,
    );

    const labelNames = currentLabels
      .map((l) => l.name)
      .filter((name) => name !== labelToRemove);

    // BUG: Overwrites concurrent label additions.
    const updatedLabels = await this.request<GitHubLabel[]>(
      "PUT",
      `/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/issues/${issueNumber}/labels`,
      { labels: labelNames },
    );

    return updatedLabels;

    // CORRECT approach: Use DELETE /repos/{owner}/{repo}/issues/{number}/labels/{name}
    // which atomically removes a single label:
    //   DELETE /repos/o/r/issues/42/labels/bug
  }
}
```

**Expected violation:**

- **Rule:** `CTR-request-shape`
- **Message:** `addLabelToIssue uses read-modify-write pattern (GET then PUT) on /repos/{owner}/{repo}/issues/{number}/labels. This creates a race condition where concurrent label modifications are silently overwritten. The API provides atomic POST (add) and DELETE (remove) endpoints that avoid this.`
- **Production impact:** In a CI/CD pipeline with multiple workflows (test, build, deploy) running concurrently, each adds its own label (e.g., "tests-passed", "build-ready", "deployed"). Workflow A reads labels `["bug"]`, adds `"tests-passed"`, writes `["bug", "tests-passed"]`. Between A's read and write, Workflow B already wrote `["bug", "build-ready"]`. A's PUT overwrites B's label, producing `["bug", "tests-passed"]` -- the `"build-ready"` label is silently lost. The deploy workflow, which triggers on `"build-ready"`, never fires. The bug is non-deterministic and depends on timing, making it nearly impossible to reproduce in testing but frequent in production under load.
