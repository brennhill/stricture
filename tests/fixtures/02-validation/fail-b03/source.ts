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
