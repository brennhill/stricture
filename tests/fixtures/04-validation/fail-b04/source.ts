// email-service/tests/sendgrid-client-b04.test.ts — No negative test cases.

import { describe, it, expect, vi } from "vitest";

describe("SendGridClient", () => {
  const client = new SendGridClient("SG.test_key.test_secret");

  // BUG: Only tests the success path. No tests for:
  // - 400 Bad Request (invalid email, missing fields)
  // - 401 Unauthorized (bad API key)
  // - 403 Forbidden (insufficient permissions)
  // - 413 Payload Too Large (oversized batch)
  // - 429 Rate Limited (too many requests)
  // - 500 Internal Server Error
  // - Network errors (fetch throws TypeError)
  // - Invalid input validation (bad email format, empty subject, >1000 personalizations)

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

    expect(result).toBeUndefined();
  });

  it("fetches bounces", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(JSON.stringify([]), { status: 200 }),
    );

    const bounces = await client.getAllBounces();
    expect(bounces).toEqual([]);
  });
});
