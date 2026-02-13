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
