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
