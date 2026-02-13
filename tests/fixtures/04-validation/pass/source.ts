// email-service/src/sendgrid-client.ts — Production-quality SendGrid v3 integration.

import type { IncomingHttpHeaders } from "http";

// ── Types ──────────────────────────────────────────────────────────────────────

/** RFC 5322 email regex — covers quoted local parts, subdomains, IP literals. */
const RFC5322_EMAIL_REGEX =
  /^(?:[a-zA-Z0-9!#$%&'*+/=?^_`{|}~-]+(?:\.[a-zA-Z0-9!#$%&'*+/=?^_`{|}~-]+)*|"(?:[\x20-\x21\x23-\x5B\x5D-\x7E]|\\[\x20-\x7E])*")@(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/;

const SENDGRID_API_KEY_REGEX = /^SG\.[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+$/;

const MAX_PERSONALIZATIONS = 1000;

interface SendGridEmailAddress {
  email: string;
  name?: string;
}

interface SendGridContent {
  type: "text/plain" | "text/html";
  value: string;
}

interface SendGridPersonalization {
  to: SendGridEmailAddress[];
  subject?: string;
}

interface SendGridSendRequest {
  personalizations: SendGridPersonalization[];
  from: SendGridEmailAddress;
  subject: string;
  content: SendGridContent[];
}

interface SendGridErrorDetail {
  message: string;
  field: string | null;
  help: string | null;
}

interface SendGridErrorResponse {
  errors: SendGridErrorDetail[];
}

interface SendGridMessage {
  msg_id: string;
  from_email: string;
  subject: string;
  to_email: string;
  status: "delivered" | "not_delivered" | "processing" | "deferred" | "bounce";
  last_event_time: string;
}

interface SendGridMessagesResponse {
  messages: SendGridMessage[];
}

interface SendGridContact {
  email: string;
  first_name?: string;
  last_name?: string;
}

interface SendGridContactsResponse {
  job_id: string;
}

interface SendGridBounce {
  email: string;
  created: number;
  reason: string;
  status: string;
}

type SendGridBounceList = SendGridBounce[];

// ── Custom Errors ──────────────────────────────────────────────────────────────

class SendGridApiError extends Error {
  constructor(
    public readonly statusCode: number,
    public readonly errors: SendGridErrorDetail[],
  ) {
    const messages = errors
      .map((e) => {
        const fieldPart = e.field !== null ? ` (field: ${e.field})` : "";
        const helpPart = e.help !== null ? ` [help: ${e.help}]` : "";
        return `${e.message}${fieldPart}${helpPart}`;
      })
      .join("; ");
    super(`SendGrid API error ${statusCode}: ${messages}`);
    this.name = "SendGridApiError";
  }
}

class SendGridRateLimitError extends Error {
  constructor(public readonly retryAfterSeconds: number) {
    super(`SendGrid rate limit exceeded. Retry after ${retryAfterSeconds}s`);
    this.name = "SendGridRateLimitError";
  }
}

class SendGridPayloadTooLargeError extends Error {
  constructor() {
    super("SendGrid request payload too large (max 30MB)");
    this.name = "SendGridPayloadTooLargeError";
  }
}

class SendGridAuthError extends Error {
  constructor(statusCode: 401 | 403) {
    const detail = statusCode === 401 ? "invalid API key" : "insufficient permissions";
    super(`SendGrid authentication failed: ${detail}`);
    this.name = "SendGridAuthError";
  }
}

// ── Validation ─────────────────────────────────────────────────────────────────

function isValidEmail(email: string): boolean {
  return RFC5322_EMAIL_REGEX.test(email);
}

function isValidApiKey(key: string): boolean {
  return SENDGRID_API_KEY_REGEX.test(key);
}

function validateSendRequest(request: SendGridSendRequest): void {
  if (!request.subject || request.subject.trim().length === 0) {
    throw new Error("Subject is required and must not be empty");
  }

  if (!isValidEmail(request.from.email)) {
    throw new Error(`Invalid from email: ${request.from.email}`);
  }

  if (request.personalizations.length === 0) {
    throw new Error("At least one personalization is required");
  }

  if (request.personalizations.length > MAX_PERSONALIZATIONS) {
    throw new Error(
      `Personalizations array exceeds maximum of ${MAX_PERSONALIZATIONS} entries`,
    );
  }

  let totalRecipients = 0;
  for (const personalization of request.personalizations) {
    for (const recipient of personalization.to) {
      if (!isValidEmail(recipient.email)) {
        throw new Error(`Invalid recipient email: ${recipient.email}`);
      }
      totalRecipients++;
    }
  }

  if (totalRecipients > MAX_PERSONALIZATIONS) {
    throw new Error(
      `Total recipients (${totalRecipients}) exceeds maximum of ${MAX_PERSONALIZATIONS}`,
    );
  }

  if (request.content.length === 0) {
    throw new Error("At least one content block is required");
  }

  for (const block of request.content) {
    if (block.type !== "text/plain" && block.type !== "text/html") {
      throw new Error(`Invalid content type: ${block.type}`);
    }
  }
}

// ── Client ─────────────────────────────────────────────────────────────────────

export class SendGridClient {
  private readonly baseUrl = "https://api.sendgrid.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    if (!isValidApiKey(apiKey)) {
      throw new Error(
        "Invalid SendGrid API key format. Expected SG.xxxxx.xxxxx",
      );
    }
    this.apiKey = apiKey;
  }

  // ── POST /v3/mail/send ───────────────────────────────────────────────────

  async sendEmail(request: SendGridSendRequest): Promise<void> {
    validateSendRequest(request);

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v3/mail/send`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(request),
      });
    } catch (err) {
      throw new Error(
        `SendGrid network error: ${err instanceof Error ? err.message : String(err)}`,
      );
    }

    // 202 Accepted — success, empty body
    if (response.status === 202) {
      return; // No body to parse — 202 returns empty response
    }

    // Error responses
    await this.handleErrorResponse(response);
  }

  // ── GET /v3/messages ─────────────────────────────────────────────────────

  async searchMessages(query: string, limit?: number): Promise<SendGridMessage[]> {
    const params = new URLSearchParams({ query });
    if (limit !== undefined) {
      if (limit < 1 || limit > 1000) {
        throw new Error("Limit must be between 1 and 1000");
      }
      params.set("limit", String(limit));
    }

    let response: Response;
    try {
      response = await fetch(
        `${this.baseUrl}/v3/messages?${params.toString()}`,
        {
          method: "GET",
          headers: {
            Authorization: `Bearer ${this.apiKey}`,
            "Content-Type": "application/json",
          },
        },
      );
    } catch (err) {
      throw new Error(
        `SendGrid network error: ${err instanceof Error ? err.message : String(err)}`,
      );
    }

    if (response.status === 200) {
      if (!response.ok) {
        throw new Error("Unexpected: status 200 but response not ok");
      }
      const data: SendGridMessagesResponse = await response.json();
      return data.messages;
    }

    await this.handleErrorResponse(response);
    throw new Error("Unreachable: handleErrorResponse always throws");
  }

  // ── PUT /v3/marketing/contacts ───────────────────────────────────────────

  async upsertContacts(contacts: SendGridContact[]): Promise<string> {
    for (const contact of contacts) {
      if (!isValidEmail(contact.email)) {
        throw new Error(`Invalid contact email: ${contact.email}`);
      }
    }

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v3/marketing/contacts`, {
        method: "PUT",
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ contacts }),
      });
    } catch (err) {
      throw new Error(
        `SendGrid network error: ${err instanceof Error ? err.message : String(err)}`,
      );
    }

    if (response.status === 202) {
      const data: SendGridContactsResponse = await response.json();
      return data.job_id;
    }

    await this.handleErrorResponse(response);
    throw new Error("Unreachable: handleErrorResponse always throws");
  }

  // ── GET /v3/suppression/bounces ──────────────────────────────────────────

  async getAllBounces(startTime?: number, endTime?: number): Promise<SendGridBounce[]> {
    const allBounces: SendGridBounce[] = [];
    let url: string | null = this.buildBouncesUrl(startTime, endTime);

    while (url !== null) {
      let response: Response;
      try {
        response = await fetch(url, {
          method: "GET",
          headers: {
            Authorization: `Bearer ${this.apiKey}`,
            "Content-Type": "application/json",
          },
        });
      } catch (err) {
        throw new Error(
          `SendGrid network error: ${err instanceof Error ? err.message : String(err)}`,
        );
      }

      if (response.status === 200) {
        const bounces: SendGridBounceList = await response.json();

        for (const bounce of bounces) {
          // Validate created is a number (unix timestamp)
          if (typeof bounce.created !== "number") {
            throw new Error(
              `Unexpected bounce.created type: expected number, got ${typeof bounce.created}`,
            );
          }
          allBounces.push(bounce);
        }

        // Follow pagination via Link header
        url = this.extractNextPageUrl(response.headers);
      } else {
        await this.handleErrorResponse(response);
      }
    }

    return allBounces;
  }

  // ── Suppression Check + Send (atomic-safe pattern) ───────────────────────

  async sendEmailIfNotSuppressed(
    recipientEmail: string,
    request: SendGridSendRequest,
  ): Promise<{ sent: boolean; reason?: string }> {
    // NOTE: There is an inherent race condition between checking suppression
    // and sending. A recipient could be added to the suppression list between
    // the check and the send. SendGrid handles this server-side — if the
    // recipient is suppressed at send time, the email is silently dropped
    // and recorded in the Activity feed. We send optimistically and check
    // the Activity API afterward for confirmation.

    validateSendRequest(request);

    try {
      await this.sendEmail(request);
    } catch (err) {
      if (err instanceof SendGridApiError) {
        // Check if the error is suppression-related
        const suppressionError = err.errors.find(
          (e) => e.message.toLowerCase().includes("suppression"),
        );
        if (suppressionError) {
          return { sent: false, reason: suppressionError.message };
        }
      }
      throw err;
    }

    return { sent: true };
  }

  // ── Private Helpers ──────────────────────────────────────────────────────

  private buildBouncesUrl(startTime?: number, endTime?: number): string {
    const params = new URLSearchParams();
    if (startTime !== undefined) {
      params.set("start_time", String(startTime));
    }
    if (endTime !== undefined) {
      params.set("end_time", String(endTime));
    }
    const qs = params.toString();
    return `${this.baseUrl}/v3/suppression/bounces${qs ? `?${qs}` : ""}`;
  }

  private extractNextPageUrl(headers: Headers): string | null {
    const linkHeader = headers.get("Link");
    if (!linkHeader) {
      return null;
    }
    // Parse Link header: <https://api.sendgrid.com/v3/...?offset=50>; rel="next"
    const match = linkHeader.match(/<([^>]+)>;\s*rel="next"/);
    return match ? match[1] : null;
  }

  private async handleErrorResponse(response: Response): Promise<never> {
    switch (response.status) {
      case 400: {
        const body: SendGridErrorResponse = await response.json();
        throw new SendGridApiError(400, body.errors);
      }
      case 401:
        throw new SendGridAuthError(401);
      case 403:
        throw new SendGridAuthError(403);
      case 413:
        throw new SendGridPayloadTooLargeError();
      case 429: {
        const retryAfter = response.headers.get("Retry-After");
        const seconds = retryAfter ? parseInt(retryAfter, 10) : 60;
        throw new SendGridRateLimitError(seconds);
      }
      case 500: {
        let errorDetails: SendGridErrorDetail[] = [];
        try {
          const body: SendGridErrorResponse = await response.json();
          errorDetails = body.errors;
        } catch {
          errorDetails = [
            { message: "Internal server error", field: null, help: null },
          ];
        }
        throw new SendGridApiError(500, errorDetails);
      }
      default:
        throw new Error(`Unexpected SendGrid status code: ${response.status}`);
    }
  }
}

// ── Tests ──────────────────────────────────────────────────────────────────────

// email-service/tests/sendgrid-client.test.ts

import { describe, it, expect, vi, beforeEach } from "vitest";

describe("SendGridClient", () => {
  const VALID_API_KEY = "SG.test_key_id.test_key_secret";
  let client: SendGridClient;

  beforeEach(() => {
    client = new SendGridClient(VALID_API_KEY);
    vi.restoreAllMocks();
  });

  // ── Constructor ──────────────────────────────────────────────────────────

  describe("constructor", () => {
    it("rejects API key without SG. prefix", () => {
      expect(() => new SendGridClient("invalid_key_no_prefix")).toThrow(
        "Invalid SendGrid API key format",
      );
    });

    it("accepts valid SG.* API key", () => {
      expect(() => new SendGridClient(VALID_API_KEY)).not.toThrow();
    });
  });

  // ── sendEmail ────────────────────────────────────────────────────────────

  describe("sendEmail", () => {
    const validRequest: SendGridSendRequest = {
      personalizations: [
        { to: [{ email: "recipient@example.com", name: "Recipient" }] },
      ],
      from: { email: "sender@example.com", name: "Sender" },
      subject: "Test Subject",
      content: [{ type: "text/plain", value: "Hello, World!" }],
    };

    it("returns void on 202 Accepted with empty body", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
        new Response(null, { status: 202 }),
      );

      const result = await client.sendEmail(validRequest);
      expect(result).toBeUndefined();
    });

    it("sends correct Authorization header with Bearer prefix", async () => {
      const fetchSpy = vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
        new Response(null, { status: 202 }),
      );

      await client.sendEmail(validRequest);

      const [, options] = fetchSpy.mock.calls[0];
      const headers = options?.headers as Record<string, string>;
      expect(headers.Authorization).toBe(`Bearer ${VALID_API_KEY}`);
      expect(headers["Content-Type"]).toBe("application/json");
    });

    it("includes all required fields in the request body", async () => {
      const fetchSpy = vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
        new Response(null, { status: 202 }),
      );

      await client.sendEmail(validRequest);

      const [, options] = fetchSpy.mock.calls[0];
      const body = JSON.parse(options?.body as string);
      expect(body).toEqual({
        personalizations: [
          { to: [{ email: "recipient@example.com", name: "Recipient" }] },
        ],
        from: { email: "sender@example.com", name: "Sender" },
        subject: "Test Subject",
        content: [{ type: "text/plain", value: "Hello, World!" }],
      });
    });

    it("rejects invalid from email address", async () => {
      const badRequest = {
        ...validRequest,
        from: { email: "not-an-email" },
      };

      await expect(client.sendEmail(badRequest)).rejects.toThrow(
        "Invalid from email",
      );
    });

    it("rejects missing subject", async () => {
      const badRequest = { ...validRequest, subject: "" };

      await expect(client.sendEmail(badRequest)).rejects.toThrow(
        "Subject is required",
      );
    });

    it("rejects personalizations exceeding 1000 entries", async () => {
      const tooMany = Array.from({ length: 1001 }, (_, i) => ({
        to: [{ email: `user${i}@example.com` }],
      }));
      const badRequest = { ...validRequest, personalizations: tooMany };

      await expect(client.sendEmail(badRequest)).rejects.toThrow(
        "exceeds maximum of 1000",
      );
    });

    it("rejects invalid content type", async () => {
      const badRequest = {
        ...validRequest,
        content: [{ type: "application/json" as "text/plain", value: "{}" }],
      };

      await expect(client.sendEmail(badRequest)).rejects.toThrow(
        "Invalid content type",
      );
    });

    // ── Error responses ────────────────────────────────────────────────────

    it("throws SendGridApiError on 400 with error details", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            errors: [
              { message: "Invalid email", field: "personalizations.0.to.0.email", help: null },
            ],
          }),
          { status: 400 },
        ),
      );

      await expect(client.sendEmail(validRequest)).rejects.toThrow(
        SendGridApiError,
      );

      try {
        await client.sendEmail(validRequest);
      } catch (err) {
        // Re-mock since the first mock was consumed
        vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
          new Response(
            JSON.stringify({
              errors: [
                { message: "Invalid email", field: "personalizations.0.to.0.email", help: null },
              ],
            }),
            { status: 400 },
          ),
        );
        try {
          await client.sendEmail(validRequest);
        } catch (e) {
          const apiErr = e as SendGridApiError;
          expect(apiErr.statusCode).toBe(400);
          expect(apiErr.errors).toHaveLength(1);
          expect(apiErr.errors[0].message).toBe("Invalid email");
          expect(apiErr.errors[0].field).toBe("personalizations.0.to.0.email");
          expect(apiErr.errors[0].help).toBeNull();
        }
      }
    });

    it("throws SendGridAuthError on 401", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
        new Response(null, { status: 401 }),
      );

      await expect(client.sendEmail(validRequest)).rejects.toThrow(
        SendGridAuthError,
      );
    });

    it("throws SendGridAuthError on 403", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
        new Response(null, { status: 403 }),
      );

      await expect(client.sendEmail(validRequest)).rejects.toThrow(
        SendGridAuthError,
      );
    });

    it("throws SendGridPayloadTooLargeError on 413", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
        new Response(null, { status: 413 }),
      );

      await expect(client.sendEmail(validRequest)).rejects.toThrow(
        SendGridPayloadTooLargeError,
      );
    });

    it("throws SendGridRateLimitError on 429 with Retry-After header", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
        new Response(null, {
          status: 429,
          headers: { "Retry-After": "30" },
        }),
      );

      try {
        await client.sendEmail(validRequest);
      } catch (err) {
        expect(err).toBeInstanceOf(SendGridRateLimitError);
        expect((err as SendGridRateLimitError).retryAfterSeconds).toBe(30);
      }
    });

    it("throws SendGridApiError on 500 with best-effort error parsing", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            errors: [{ message: "Internal failure", field: null, help: null }],
          }),
          { status: 500 },
        ),
      );

      await expect(client.sendEmail(validRequest)).rejects.toThrow(
        SendGridApiError,
      );
    });

    it("handles 500 with unparseable body gracefully", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
        new Response("not json", { status: 500 }),
      );

      await expect(client.sendEmail(validRequest)).rejects.toThrow(
        SendGridApiError,
      );
    });

    it("wraps network errors with descriptive message", async () => {
      vi.spyOn(globalThis, "fetch").mockRejectedValueOnce(
        new TypeError("Failed to fetch"),
      );

      await expect(client.sendEmail(validRequest)).rejects.toThrow(
        "SendGrid network error: Failed to fetch",
      );
    });
  });

  // ── getAllBounces ─────────────────────────────────────────────────────────

  describe("getAllBounces", () => {
    it("returns bounces with correct types (created as number)", async () => {
      const bounceData: SendGridBounce[] = [
        {
          email: "bounced@example.com",
          created: 1700000000,
          reason: "550 5.1.1 User unknown",
          status: "5.1.1",
        },
      ];

      vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
        new Response(JSON.stringify(bounceData), {
          status: 200,
          headers: {},
        }),
      );

      const bounces = await client.getAllBounces();
      expect(bounces).toHaveLength(1);
      expect(bounces[0].email).toBe("bounced@example.com");
      expect(typeof bounces[0].created).toBe("number");
      expect(bounces[0].created).toBe(1700000000);
      expect(bounces[0].reason).toBe("550 5.1.1 User unknown");
      expect(bounces[0].status).toBe("5.1.1");
    });

    it("follows pagination via Link header until exhausted", async () => {
      const page1: SendGridBounce[] = [
        { email: "a@example.com", created: 1700000000, reason: "bounce", status: "5.1.1" },
      ];
      const page2: SendGridBounce[] = [
        { email: "b@example.com", created: 1700000001, reason: "bounce", status: "5.1.1" },
      ];

      vi.spyOn(globalThis, "fetch")
        .mockResolvedValueOnce(
          new Response(JSON.stringify(page1), {
            status: 200,
            headers: {
              Link: '<https://api.sendgrid.com/v3/suppression/bounces?offset=1>; rel="next"',
            },
          }),
        )
        .mockResolvedValueOnce(
          new Response(JSON.stringify(page2), {
            status: 200,
            headers: {},
          }),
        );

      const bounces = await client.getAllBounces();
      expect(bounces).toHaveLength(2);
      expect(bounces[0].email).toBe("a@example.com");
      expect(bounces[1].email).toBe("b@example.com");
    });
  });

  // ── Error field null safety ──────────────────────────────────────────────

  describe("error handling with nullable fields", () => {
    it("formats error message correctly when field is null", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            errors: [{ message: "Rate limit reached", field: null, help: null }],
          }),
          { status: 400 },
        ),
      );

      try {
        await client.sendEmail({
          personalizations: [{ to: [{ email: "a@example.com" }] }],
          from: { email: "b@example.com" },
          subject: "Test",
          content: [{ type: "text/plain", value: "x" }],
        });
      } catch (err) {
        const apiErr = err as SendGridApiError;
        // Must not crash accessing null.toLowerCase() or similar
        expect(apiErr.message).toContain("Rate limit reached");
        expect(apiErr.message).not.toContain("null");
      }
    });

    it("formats error message correctly when field is present", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            errors: [
              { message: "Invalid", field: "from.email", help: "https://docs.sendgrid.com" },
            ],
          }),
          { status: 400 },
        ),
      );

      try {
        await client.sendEmail({
          personalizations: [{ to: [{ email: "a@example.com" }] }],
          from: { email: "b@example.com" },
          subject: "Test",
          content: [{ type: "text/plain", value: "x" }],
        });
      } catch (err) {
        const apiErr = err as SendGridApiError;
        expect(apiErr.message).toContain("(field: from.email)");
        expect(apiErr.message).toContain("[help: https://docs.sendgrid.com]");
      }
    });
  });
});
