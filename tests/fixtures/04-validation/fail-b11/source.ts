// email-service/src/sendgrid-client-b11.ts — Timestamp stored as 32-bit int.

// BUG: Using Int32Array to store timestamps. This overflows on 2038-01-19.
// Manifest declares created as { type: integer, format: unix_timestamp }.
// While the manifest does not specify bit width, standard practice for
// unix timestamps requires at least 64-bit storage to avoid Y2038.

export class SendGridClientB11 {
  private readonly baseUrl = "https://api.sendgrid.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async getBounces(): Promise<Array<{ email: string; created: number; reason: string; status: string }>> {
    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v3/suppression/bounces`, {
        method: "GET",
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "application/json",
        },
      });
    } catch (err) {
      throw new Error(`SendGrid network error: ${(err as Error).message}`);
    }

    if (response.status === 200) {
      const bounces: SendGridBounce[] = await response.json();

      // BUG: Storing timestamps in a Int32Array for "efficient" batch processing.
      // Unix timestamp 2145916800 (2038-01-19) overflows signed 32-bit int.
      // Int32Array clamps/wraps: 2147483648 becomes -2147483648.
      const timestamps = new Int32Array(bounces.length);
      const results: Array<{ email: string; created: number; reason: string; status: string }> = [];

      for (let i = 0; i < bounces.length; i++) {
        timestamps[i] = bounces[i].created; // Overflow for post-2038 timestamps
        results.push({
          email: bounces[i].email,
          created: timestamps[i], // Negative number after 2038
          reason: bounces[i].reason,
          status: bounces[i].status,
        });
      }

      return results;
    }

    throw new Error(`SendGrid error: ${response.status}`);
  }

  async getRecentBounces(afterDate: Date): Promise<Array<{ email: string; created: number; reason: string; status: string }>> {
    const bounces = await this.getBounces();
    const cutoff = Math.floor(afterDate.getTime() / 1000);

    // BUG: After 2038, all stored timestamps are negative.
    // Every bounce appears to be "before" any positive cutoff timestamp,
    // so this filter returns an empty array even when bounces exist.
    return bounces.filter((b) => b.created > cutoff);
  }
}
