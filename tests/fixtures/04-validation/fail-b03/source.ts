// email-service/tests/sendgrid-client-b03.test.ts — Shallow assertions.

import { describe, it, expect, vi } from "vitest";

describe("SendGridClient", () => {
  const client = new SendGridClient("SG.test_key.test_secret");

  it("sends email successfully", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(null, { status: 202 }),
    );

    const result = await client.sendEmail({
      personalizations: [{ to: [{ email: "a@example.com" }] }],
      from: { email: "b@example.com" },
      subject: "Test",
      content: [{ type: "text/plain", value: "Hello" }],
    });

    // BUG: Shallow assertion — sendEmail returns void on success.
    // This assertion passes even if the function throws and is caught,
    // returns a random object, or returns undefined for wrong reasons.
    expect(result).toBeDefined();
  });

  it("fetches bounces", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(
        JSON.stringify([
          { email: "x@example.com", created: 1700000000, reason: "bounce", status: "5.1.1" },
        ]),
        { status: 200 },
      ),
    );

    const bounces = await client.getAllBounces();

    // BUG: Shallow — only checks that something was returned.
    // Does not verify array length, field values, field types,
    // or that `created` is a number (not a string).
    expect(bounces).toBeTruthy();
  });

  it("upserts contacts", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify({ job_id: "job_abc123" }), { status: 202 }),
    );

    const jobId = await client.upsertContacts([{ email: "c@example.com" }]);

    // BUG: Shallow — doesn't verify the job_id value or format.
    expect(jobId).toBeDefined();
  });
});
