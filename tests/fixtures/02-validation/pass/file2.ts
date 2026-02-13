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
