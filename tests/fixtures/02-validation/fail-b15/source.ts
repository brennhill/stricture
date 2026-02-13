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
