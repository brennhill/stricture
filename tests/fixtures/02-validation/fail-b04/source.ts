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
