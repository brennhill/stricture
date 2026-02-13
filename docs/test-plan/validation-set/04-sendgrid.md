# 04 — SendGrid Email API (v3)

> **API:** SendGrid v3 Mail Send, Activity, Contacts, Suppressions
> **Why included:** Batch operations, 202-no-body semantics, bounce handling, RFC 5322 email validation, rate limiting, nullable error fields

---

## Manifest Fragment

```yaml
# .stricture-manifest.yml — SendGrid Email API v3 contract
contracts:
  - id: "sendgrid-mail-send"
    producer: sendgrid
    consumers: [email-service]
    protocol: http
    endpoints:
      - path: "/v3/mail/send"
        method: POST
        request:
          headers:
            Authorization: { type: string, format: "Bearer SG\\..*", required: true }
            Content-Type: { type: string, values: ["application/json"], required: true }
          fields:
            personalizations:
              type: array
              required: true
              maxItems: 1000
              items:
                fields:
                  to:
                    type: array
                    required: true
                    items:
                      fields:
                        email: { type: string, format: email_rfc5322, required: true }
                        name:  { type: string, required: false }
                  subject: { type: string, required: false }
            from:
              type: object
              required: true
              fields:
                email: { type: string, format: email_rfc5322, required: true }
                name:  { type: string, required: false }
            subject: { type: string, required: true }
            content:
              type: array
              required: true
              items:
                fields:
                  type:  { type: enum, values: ["text/plain", "text/html"], required: true }
                  value: { type: string, required: true }
        response:
          status_codes: [202, 400, 401, 403, 413, 429, 500]
          success:
            status: 202
            body: null  # 202 Accepted returns empty body
          error:
            fields:
              errors:
                type: array
                required: true
                items:
                  fields:
                    message: { type: string, required: true }
                    field:   { type: string, nullable: true, required: false }
                    help:    { type: string, nullable: true, required: false }

  - id: "sendgrid-messages-search"
    producer: sendgrid
    consumers: [email-service]
    protocol: http
    endpoints:
      - path: "/v3/messages"
        method: GET
        request:
          headers:
            Authorization: { type: string, format: "Bearer SG\\..*", required: true }
          query:
            query:  { type: string, required: true }
            limit:  { type: integer, range: [1, 1000], required: false }
        response:
          status_codes: [200, 400, 401, 403, 429, 500]
          success:
            status: 200
            fields:
              messages:
                type: array
                items:
                  fields:
                    msg_id:        { type: string, required: true }
                    from_email:    { type: string, format: email_rfc5322, required: true }
                    subject:       { type: string, required: true }
                    to_email:      { type: string, format: email_rfc5322, required: true }
                    status:        { type: enum, values: ["delivered", "not_delivered", "processing", "deferred", "bounce"], required: true }
                    last_event_time: { type: string, format: iso8601, required: true }

  - id: "sendgrid-marketing-contacts"
    producer: sendgrid
    consumers: [email-service]
    protocol: http
    endpoints:
      - path: "/v3/marketing/contacts"
        method: PUT
        request:
          headers:
            Authorization: { type: string, format: "Bearer SG\\..*", required: true }
          fields:
            contacts:
              type: array
              required: true
              items:
                fields:
                  email:      { type: string, format: email_rfc5322, required: true }
                  first_name: { type: string, required: false }
                  last_name:  { type: string, required: false }
        response:
          status_codes: [202, 400, 401, 403, 429, 500]
          success:
            status: 202
            fields:
              job_id: { type: string, required: true }

  - id: "sendgrid-suppression-bounces"
    producer: sendgrid
    consumers: [email-service]
    protocol: http
    endpoints:
      - path: "/v3/suppression/bounces"
        method: GET
        request:
          headers:
            Authorization: { type: string, format: "Bearer SG\\..*", required: true }
          query:
            start_time: { type: integer, required: false }
            end_time:   { type: integer, required: false }
        response:
          status_codes: [200, 401, 403, 429, 500]
          success:
            status: 200
            pagination:
              cursor: header  # Link header with rel="next"
            fields:
              type: array
              items:
                fields:
                  email:   { type: string, format: email_rfc5322, required: true }
                  created: { type: integer, format: unix_timestamp, required: true }
                  reason:  { type: string, required: true }
                  status:  { type: string, format: "^[245]\\.[0-9]+\\.[0-9]+$", required: true }
```

---

## PERFECT Integration (Zero Violations Expected)

This integration correctly handles all SendGrid contract nuances: 202 empty-body responses, RFC 5322 email validation, nullable error fields, pagination, rate limiting, and proper typing.

```typescript
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
```

**Stricture expected result:** ZERO violations. All rules pass:
- `TQ-error-path-coverage` -- try/catch on every fetch, all status codes handled
- `TQ-no-shallow-assertions` -- all assertions check specific values, types, and shapes
- `TQ-negative-cases` -- tests cover 400, 401, 403, 413, 429, 500, network errors, invalid inputs
- `CTR-request-shape` -- request body includes all required fields per manifest
- `CTR-response-shape` -- 202 returns void (not a parsed body), 200 parses correct shape
- `CTR-status-code-handling` -- explicit handling for 202, 200, 400, 401, 403, 413, 429, 500
- `CTR-strictness-parity` -- email format validated, personalizations range checked, enums covered
- `CTR-manifest-conformance` -- all types match manifest declarations exactly

---

## B01 -- No Error Handling

**Bug:** No try/catch around the `fetch()` call. Network errors, DNS failures, and timeouts crash the caller with an unhandled promise rejection.

**Stricture rule:** `TQ-error-path-coverage`

```typescript
// email-service/src/sendgrid-client-b01.ts — Missing error handling on fetch.

export class SendGridClientB01 {
  private readonly baseUrl = "https://api.sendgrid.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async sendEmail(request: SendGridSendRequest): Promise<void> {
    // BUG: No try/catch — fetch() can throw on network errors, DNS failures,
    // timeouts, or AbortController signals. These propagate as unhandled
    // rejections and crash the calling service.
    const response = await fetch(`${this.baseUrl}/v3/mail/send`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${this.apiKey}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });

    if (response.status === 202) {
      return;
    }

    throw new Error(`SendGrid error: ${response.status}`);
  }
}
```

**Expected violation:**
```
TQ-error-path-coverage: fetch() call at line 14 has no try/catch or .catch() handler.
Network-layer failures (TypeError, AbortError) will propagate as unhandled rejections.
```

**Production impact:** In Node.js, unhandled promise rejections terminate the process (since Node 15+). A transient DNS failure or network timeout kills the entire email service instead of being caught and retried.

---

## B02 -- No Status Code Check (Expects 200 Instead of 202)

**Bug:** Checks for `response.status === 200` instead of `202`. SendGrid's `/v3/mail/send` returns `202 Accepted` on success, not `200 OK`. Every successful send is treated as an error.

**Stricture rule:** `CTR-status-code-handling`

```typescript
// email-service/src/sendgrid-client-b02.ts — Wrong success status code.

export class SendGridClientB02 {
  private readonly baseUrl = "https://api.sendgrid.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async sendEmail(request: SendGridSendRequest): Promise<void> {
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
      throw new Error(`SendGrid network error: ${(err as Error).message}`);
    }

    // BUG: SendGrid returns 202 Accepted on success, not 200 OK.
    // This condition is never true for successful sends.
    if (response.status === 200) {
      return;
    }

    throw new Error(`SendGrid error: ${response.status}`);
  }
}
```

**Expected violation:**
```
CTR-status-code-handling: Manifest declares success status 202 for POST /v3/mail/send.
Client checks for status 200 but not 202. Unhandled status code: 202.
```

**Production impact:** Every successfully queued email is reported as a failure. The calling code may retry indefinitely, sending duplicate emails or erroneously alerting operators that email delivery is broken.

---

## B03 -- Shallow Test Assertions

**Bug:** Tests use `expect(result).toBeDefined()` and `expect(result).toBeTruthy()` instead of verifying actual response shapes, status codes, and field values.

**Stricture rule:** `TQ-no-shallow-assertions`

```typescript
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
```

**Expected violation:**
```
TQ-no-shallow-assertions at line 20: expect(result).toBeDefined() — shallow assertion on void return.
TQ-no-shallow-assertions at line 36: expect(bounces).toBeTruthy() — shallow assertion on SendGridBounce[].
TQ-no-shallow-assertions at line 45: expect(jobId).toBeDefined() — shallow assertion on string return.
```

**Production impact:** Tests pass with 100% line coverage but verify nothing. A refactor that changes the return type, drops fields from the bounce response, or returns the wrong job_id format passes all tests silently.

---

## B04 -- Missing Negative Tests

**Bug:** Only the happy path (202 success) is tested. No tests for 400, 401, 403, 413, 429, 500 error responses, network failures, or invalid inputs.

**Stricture rule:** `TQ-negative-cases`

```typescript
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
```

**Expected violation:**
```
TQ-negative-cases: sendEmail handles status codes [202, 400, 401, 403, 413, 429, 500] per manifest,
but tests only cover [202]. Missing negative tests for: 400, 401, 403, 413, 429, 500.
TQ-negative-cases: getAllBounces handles status codes [200, 401, 403, 429, 500] per manifest,
but tests only cover [200]. Missing negative tests for: 401, 403, 429, 500.
```

**Production impact:** Error handling code paths are completely untested. A regression that breaks 429 rate-limit handling (e.g., misreading the Retry-After header) goes undetected. In production, the service hammers SendGrid without backoff, gets permanently blocked, and no email can be sent.

---

## B05 -- Request Missing Required Fields

**Bug:** The `subject` field is omitted from the send request. SendGrid requires `subject` at the top level (or per-personalization). The request will be rejected with a 400 error.

**Stricture rule:** `CTR-request-shape`

```typescript
// email-service/src/sendgrid-client-b05.ts — Missing required 'subject' field.

interface SendGridSendRequestB05 {
  personalizations: Array<{
    to: Array<{ email: string; name?: string }>;
  }>;
  from: { email: string; name?: string };
  // BUG: 'subject' field is missing from the interface entirely.
  // The manifest declares subject as required: true.
  content: Array<{ type: string; value: string }>;
}

export class SendGridClientB05 {
  private readonly baseUrl = "https://api.sendgrid.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async sendEmail(request: SendGridSendRequestB05): Promise<void> {
    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v3/mail/send`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "application/json",
        },
        // BUG: Serialized body never contains 'subject' because the
        // interface doesn't have it. SendGrid returns 400.
        body: JSON.stringify(request),
      });
    } catch (err) {
      throw new Error(`SendGrid network error: ${(err as Error).message}`);
    }

    if (response.status === 202) {
      return;
    }

    throw new Error(`SendGrid error: ${response.status}`);
  }
}
```

**Expected violation:**
```
CTR-request-shape: POST /v3/mail/send request type is missing required field "subject"
declared in manifest. Field is required: true, type: string.
```

**Production impact:** Every email send attempt fails with a 400 Bad Request. The error message from SendGrid (`"The subject is required"`) is not surfaced to the caller because the error handler is generic. Operators see a flood of `SendGrid error: 400` without knowing why.

---

## B06 -- Response Type Mismatch (Expects Body from 202)

**Bug:** Attempts to parse a JSON body from the 202 Accepted response. SendGrid's `/v3/mail/send` returns an empty body on 202. Calling `response.json()` on an empty body throws a SyntaxError.

**Stricture rule:** `CTR-response-shape`

```typescript
// email-service/src/sendgrid-client-b06.ts — Tries to parse empty 202 body.

interface SendGridSendResponseB06 {
  // BUG: This type does not exist — 202 Accepted has no body.
  // The manifest declares: success.body = null
  message_id: string;
  status: string;
}

export class SendGridClientB06 {
  private readonly baseUrl = "https://api.sendgrid.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async sendEmail(request: SendGridSendRequest): Promise<SendGridSendResponseB06> {
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
      throw new Error(`SendGrid network error: ${(err as Error).message}`);
    }

    if (response.status === 202) {
      // BUG: 202 Accepted returns an empty body. response.json() will throw:
      // SyntaxError: Unexpected end of JSON input
      const data: SendGridSendResponseB06 = await response.json();
      return data;
    }

    throw new Error(`SendGrid error: ${response.status}`);
  }
}
```

**Expected violation:**
```
CTR-response-shape: POST /v3/mail/send manifest declares success response body as null
(status 202). Client attempts to parse response body as SendGridSendResponseB06 with
fields [message_id, status]. Type mismatch: expected no body, got object parse.
```

**Production impact:** Every successful email send throws a `SyntaxError: Unexpected end of JSON input`. The function never returns successfully. Callers see failures even though emails are actually being delivered, leading to duplicate retry sends and confused monitoring.

---

## B07 -- Wrong Field Types (Bounce Timestamp as String)

**Bug:** The `created` field on bounce records is stored as a `string` instead of a `number`. SendGrid returns it as a Unix timestamp (integer). Storing it as a string breaks numeric comparisons and date arithmetic.

**Stricture rule:** `CTR-manifest-conformance`

```typescript
// email-service/src/sendgrid-client-b07.ts — Wrong type for bounce.created field.

interface SendGridBounceB07 {
  email: string;
  // BUG: Manifest declares created as { type: integer, format: unix_timestamp }.
  // Storing as string means numeric comparisons fail:
  //   "1700000000" > "200000000" is FALSE (string comparison)
  //   1700000000 > 200000000 is TRUE (numeric comparison)
  created: string;
  reason: string;
  status: string;
}

export class SendGridClientB07 {
  private readonly baseUrl = "https://api.sendgrid.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async getBounces(): Promise<SendGridBounceB07[]> {
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
      const bounces: SendGridBounceB07[] = await response.json();
      return bounces;
    }

    throw new Error(`SendGrid error: ${response.status}`);
  }

  async getBouncesAfter(timestamp: number): Promise<SendGridBounceB07[]> {
    const bounces = await this.getBounces();
    // BUG: String comparison instead of numeric comparison.
    // "9" > "1700000000" is TRUE in string comparison because "9" > "1".
    // This filter returns wrong results.
    return bounces.filter((b) => b.created > String(timestamp));
  }
}
```

**Expected violation:**
```
CTR-manifest-conformance: Field "created" in SendGridBounceB07 has type string,
but manifest declares type integer with format unix_timestamp.
Type mismatch: expected number, got string.
```

**Production impact:** Bounce filtering by date is silently incorrect. String comparison of timestamps produces wrong results: bounces from 2023 may be filtered out while bounces from 1970 are included. A "delete bounces older than 30 days" operation may delete recent bounces and keep ancient ones.

---

## B08 -- Incomplete Enum Handling (Missing 413 and 429)

**Bug:** Error handling covers 400 and 401 but not 413 (Payload Too Large) or 429 (Rate Limit). These are status codes that require specific remediation: 413 means the batch must be split; 429 means the caller must back off.

**Stricture rule:** `CTR-strictness-parity`

```typescript
// email-service/src/sendgrid-client-b08.ts — Missing 413 and 429 handling.

export class SendGridClientB08 {
  private readonly baseUrl = "https://api.sendgrid.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async sendEmail(request: SendGridSendRequest): Promise<void> {
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
      throw new Error(`SendGrid network error: ${(err as Error).message}`);
    }

    if (response.status === 202) {
      return;
    }

    // BUG: Only handles 400 and 401. Manifest declares status codes:
    // [202, 400, 401, 403, 413, 429, 500].
    // Missing: 403, 413, 429, 500 — each requires different remediation.
    switch (response.status) {
      case 400: {
        const body = await response.json();
        throw new Error(`Bad request: ${body.errors[0].message}`);
      }
      case 401:
        throw new Error("Unauthorized: check API key");
      default:
        // 413 and 429 fall through to this generic handler.
        // 413: caller does not know to split the batch.
        // 429: caller does not know to wait and retry with backoff.
        throw new Error(`SendGrid error: ${response.status}`);
    }
  }
}
```

**Expected violation:**
```
CTR-strictness-parity: POST /v3/mail/send manifest declares status codes
[202, 400, 401, 403, 413, 429, 500]. Client explicitly handles [202, 400, 401].
Missing explicit handling for: 403, 413, 429, 500. Status codes 413 (payload too large)
and 429 (rate limit) require specific remediation strategies.
```

**Production impact:** When SendGrid returns 429 (rate limited), the generic error handler throws a nondescript error. The calling code retries immediately without backoff, making the rate limiting worse. When 413 is returned for a large batch, the caller retries the same oversized payload indefinitely instead of splitting it.

---

## B09 -- Missing Range Validation (Personalizations Not Bounded)

**Bug:** No validation that the `personalizations` array does not exceed 1000 entries. SendGrid enforces a maximum of 1000 recipients per send call. Exceeding this results in a 400 error that could be prevented client-side.

**Stricture rule:** `CTR-strictness-parity`

```typescript
// email-service/src/sendgrid-client-b09.ts — No personalizations limit check.

export class SendGridClientB09 {
  private readonly baseUrl = "https://api.sendgrid.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async sendEmail(request: SendGridSendRequest): Promise<void> {
    // BUG: No validation on personalizations array size.
    // Manifest declares: personalizations.maxItems = 1000.
    // A caller can pass 50,000 personalizations. This:
    // 1. Creates a massive JSON payload (may trigger 413)
    // 2. Gets rejected by SendGrid with a 400
    // 3. Wastes network bandwidth and API rate limit quota
    // 4. Could cause OOM if the serialized body is very large

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
      throw new Error(`SendGrid network error: ${(err as Error).message}`);
    }

    if (response.status === 202) {
      return;
    }

    throw new Error(`SendGrid error: ${response.status}`);
  }

  async sendBulkEmails(
    recipients: Array<{ email: string; name?: string }>,
    from: { email: string },
    subject: string,
    htmlContent: string,
  ): Promise<void> {
    // BUG: Builds a single request with all recipients.
    // Should batch into groups of 1000.
    const request: SendGridSendRequest = {
      personalizations: recipients.map((r) => ({ to: [r] })),
      from,
      subject,
      content: [{ type: "text/html", value: htmlContent }],
    };

    await this.sendEmail(request);
  }
}
```

**Expected violation:**
```
CTR-strictness-parity: Manifest declares personalizations.maxItems = 1000.
Producer (SendGrid) enforces this limit server-side. Consumer (email-service)
does not validate the array size before sending. Missing range validation
on field "personalizations".
```

**Production impact:** A marketing campaign targeting 10,000 users calls `sendBulkEmails` with all recipients at once. The request is rejected by SendGrid. No emails are sent. If the caller retries, it retries the same oversized payload. The correct behavior is to batch into groups of 1000.

---

## B10 -- Format Not Validated (Email Addresses Accepted as Any String)

**Bug:** Email addresses are accepted as any string with no RFC 5322 format validation. Strings like `"not-an-email"`, `""`, or `"foo bar @baz"` are sent to SendGrid, which rejects them.

**Stricture rule:** `CTR-strictness-parity`

```typescript
// email-service/src/sendgrid-client-b10.ts — No email format validation.

export class SendGridClientB10 {
  private readonly baseUrl = "https://api.sendgrid.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async sendEmail(request: SendGridSendRequest): Promise<void> {
    // BUG: No email format validation on from.email or personalizations[].to[].email.
    // Manifest declares email fields as: { type: string, format: email_rfc5322 }.
    // SendGrid validates server-side and returns 400, but:
    // 1. The error message from SendGrid may be cryptic
    // 2. We waste an API call and rate limit quota
    // 3. We cannot provide a user-friendly error message
    // 4. Batch sends fail entirely if one email is invalid

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
      throw new Error(`SendGrid network error: ${(err as Error).message}`);
    }

    if (response.status === 202) {
      return;
    }

    throw new Error(`SendGrid error: ${response.status}`);
  }

  async sendWelcomeEmail(userEmail: string, userName: string): Promise<void> {
    // BUG: userEmail is used directly without validation.
    // If userEmail is "John Doe" (a name, not an email), this sends
    // a malformed request to SendGrid.
    await this.sendEmail({
      personalizations: [{ to: [{ email: userEmail, name: userName }] }],
      from: { email: "welcome@myapp.com" },
      subject: `Welcome, ${userName}!`,
      content: [{ type: "text/html", value: `<h1>Welcome, ${userName}!</h1>` }],
    });
  }
}
```

**Expected violation:**
```
CTR-strictness-parity: Manifest declares field "from.email" with format email_rfc5322.
Producer (SendGrid) validates this format server-side. Consumer (email-service) does
not validate email format before sending. Missing format validation on fields:
"from.email", "personalizations[].to[].email".
```

**Production impact:** A user registration form allows any string as an email. The welcome email is sent with `"John Doe"` as the recipient address. SendGrid returns 400. The user never receives their welcome email. No client-side error message explains what went wrong. At scale, invalid emails consume API quota and fill error logs with unhelpful SendGrid error responses.

---

## B11 -- Precision Loss (Unix Timestamp as 32-bit Int)

**Bug:** The bounce `created` timestamp is stored as a 32-bit integer. Unix timestamps will overflow a signed 32-bit integer on January 19, 2038 (Y2038 problem). JavaScript's `number` type is a 64-bit float and can represent timestamps beyond 2038, but if the value passes through a system that uses 32-bit integers (a database column, a Protocol Buffer int32 field, or a C FFI binding), the timestamp wraps around to a negative number.

**Stricture rule:** `CTR-strictness-parity`

```typescript
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
```

**Expected violation:**
```
CTR-strictness-parity: Field "created" (format: unix_timestamp) is stored via Int32Array
which has a maximum value of 2,147,483,647 (2038-01-19T03:14:07Z). Unix timestamps
beyond this date will overflow. Use standard number type or BigInt for timestamp storage.
```

**Production impact:** After January 19, 2038, all bounce timestamps wrap to negative values. Date-based bounce queries return incorrect results. A "purge bounces older than 90 days" job deletes all bounces (they all appear to be from 1901). Suppression list management breaks silently.

---

## B12 -- Nullable Field Crash (Accessing error.field When Null)

**Bug:** The error handler calls `error.field.toLowerCase()` without checking if `field` is `null`. The SendGrid error response schema declares `field` as `string | null`. When field is null (e.g., for rate limit errors or authentication errors), this crashes with `TypeError: Cannot read properties of null`.

**Stricture rule:** `CTR-response-shape`

```typescript
// email-service/src/sendgrid-client-b12.ts — Nullable field access crash.

export class SendGridClientB12 {
  private readonly baseUrl = "https://api.sendgrid.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async sendEmail(request: SendGridSendRequest): Promise<void> {
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
      throw new Error(`SendGrid network error: ${(err as Error).message}`);
    }

    if (response.status === 202) {
      return;
    }

    if (response.status === 400 || response.status === 500) {
      const body: SendGridErrorResponse = await response.json();
      this.formatAndThrowErrors(body.errors);
    }

    throw new Error(`SendGrid error: ${response.status}`);
  }

  private formatAndThrowErrors(errors: SendGridErrorDetail[]): never {
    const formatted = errors.map((err) => {
      // BUG: err.field can be null according to the manifest:
      //   field: { type: string, nullable: true, required: false }
      // When SendGrid returns an error like "Rate limit exceeded" or
      // "Authorization required", the field property is null.
      // Calling .toLowerCase() on null throws TypeError.
      const fieldName = err.field.toLowerCase(); // CRASH when field is null

      // BUG: Same issue with err.help — also nullable.
      const helpLink = err.help.trim(); // CRASH when help is null

      return `[${fieldName}] ${err.message} (see: ${helpLink})`;
    });

    throw new Error(`SendGrid validation errors:\n${formatted.join("\n")}`);
  }
}
```

**Expected violation:**
```
CTR-response-shape: Field "errors[].field" is declared as nullable: true in manifest.
Code accesses err.field.toLowerCase() at line 47 without null check.
TypeError will occur when field is null.
CTR-response-shape: Field "errors[].help" is declared as nullable: true in manifest.
Code accesses err.help.trim() at line 50 without null check.
TypeError will occur when help is null.
```

**Production impact:** Every non-field-specific error (rate limits, auth failures, server errors) crashes the error handler. Instead of a meaningful error message, the caller gets `TypeError: Cannot read properties of null (reading 'toLowerCase')`. Error logging and monitoring see the TypeError instead of the actual SendGrid error, making debugging impossible.

---

## B13 -- Missing API Key Format Validation

**Bug:** The constructor accepts any string as an API key. SendGrid API keys follow the format `SG.xxxxx.xxxxx` (two base64url segments separated by dots, prefixed with `SG.`). Accepting arbitrary strings means typos, partial keys, or accidentally passing a different service's key are not caught until the first API call fails with a 401.

**Stricture rule:** `CTR-request-shape`

```typescript
// email-service/src/sendgrid-client-b13.ts — No API key format validation.

export class SendGridClientB13 {
  private readonly baseUrl = "https://api.sendgrid.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    // BUG: No format validation on API key.
    // Manifest declares: Authorization header format: "Bearer SG\\..*"
    // The constructor should verify the key matches the SG.* pattern.
    // Without this check:
    // 1. A Stripe key ("sk_live_...") is silently accepted
    // 2. An empty string is accepted
    // 3. A key with trailing whitespace is accepted
    // 4. The error only surfaces on the first API call as a 401
    this.apiKey = apiKey;
  }

  async sendEmail(request: SendGridSendRequest): Promise<void> {
    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v3/mail/send`, {
        method: "POST",
        headers: {
          // BUG: If apiKey is "sk_live_stripe_key", this sends:
          // Authorization: Bearer sk_live_stripe_key
          // SendGrid returns 401 with a cryptic error.
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(request),
      });
    } catch (err) {
      throw new Error(`SendGrid network error: ${(err as Error).message}`);
    }

    if (response.status === 202) {
      return;
    }

    if (response.status === 401) {
      // The error message is correct but the root cause is unclear.
      // Was the key revoked? Was the wrong key used? Was it malformed?
      throw new Error("SendGrid authentication failed");
    }

    throw new Error(`SendGrid error: ${response.status}`);
  }
}
```

**Expected violation:**
```
CTR-request-shape: Manifest declares Authorization header format "Bearer SG\\..*".
Client does not validate API key format before use. Any string is accepted as the
API key, including keys from other services, empty strings, and malformed values.
```

**Production impact:** During deployment, an environment variable mixup sets `SENDGRID_API_KEY` to the Stripe secret key. The service starts successfully (no validation at construction time). The first email send attempt fails with a 401. Because the error message says "authentication failed" without mentioning the key format, the operator spends hours checking SendGrid dashboard permissions, revoking and regenerating keys, before realizing the wrong environment variable was used.

---

## B14 -- Pagination Issue (Ignores Link Header)

**Bug:** The bounces endpoint returns paginated results. The first page is fetched, but the `Link` header containing the URL for the next page is ignored. Only the first page of results is returned.

**Stricture rule:** `CTR-response-shape`

```typescript
// email-service/src/sendgrid-client-b14.ts — Pagination not followed.

export class SendGridClientB14 {
  private readonly baseUrl = "https://api.sendgrid.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async getBounces(): Promise<SendGridBounce[]> {
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
      // BUG: Only returns the first page of bounces.
      // SendGrid paginates bounces with a Link header:
      //   Link: <https://api.sendgrid.com/v3/suppression/bounces?offset=500>; rel="next"
      // Manifest declares: pagination.cursor = header
      // This response may contain only 500 of 5,000 total bounces.
      // The Link header is completely ignored.
      const bounces: SendGridBounce[] = await response.json();
      return bounces;
    }

    throw new Error(`SendGrid error: ${response.status}`);
  }

  async getBouncedEmailCount(): Promise<number> {
    // BUG: Returns count of first page only, not total bounces.
    // If there are 5,000 bounces and the page size is 500,
    // this returns 500 instead of 5,000.
    const bounces = await this.getBounces();
    return bounces.length;
  }

  async isEmailBounced(email: string): Promise<boolean> {
    // BUG: Only checks first page. If the bounce is on page 2+,
    // this returns false even though the email has bounced.
    const bounces = await this.getBounces();
    return bounces.some((b) => b.email === email);
  }
}
```

**Expected violation:**
```
CTR-response-shape: GET /v3/suppression/bounces manifest declares pagination.cursor = header.
Client fetches a single page and does not check the Link header for subsequent pages.
Response data may be incomplete — only the first page of a multi-page result set is returned.
```

**Production impact:** The suppression list check (`isEmailBounced`) reports false negatives. An email address that bounced months ago is on page 3 of the bounces list. The check says it has not bounced. The system sends email to the address, it bounces again, and SendGrid marks the account as a repeat bouncer. Enough repeat bounces can trigger SendGrid account suspension for poor sender reputation.

---

## B15 -- Race Condition (Check-Then-Send Without Atomicity)

**Bug:** The code checks the suppression list for a recipient, then sends the email in a separate request. Between the check and the send, the recipient could be added to the suppression list (by another bounce, spam complaint, or manual addition). The email is sent to a suppressed address, damaging sender reputation.

**Stricture rule:** `CTR-request-shape`

```typescript
// email-service/src/sendgrid-client-b15.ts — Race condition: check then send.

export class SendGridClientB15 {
  private readonly baseUrl = "https://api.sendgrid.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async getBounces(): Promise<SendGridBounce[]> {
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
      return await response.json();
    }

    throw new Error(`SendGrid error: ${response.status}`);
  }

  async sendEmail(request: SendGridSendRequest): Promise<void> {
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
      throw new Error(`SendGrid network error: ${(err as Error).message}`);
    }

    if (response.status === 202) {
      return;
    }

    throw new Error(`SendGrid error: ${response.status}`);
  }

  async sendEmailIfNotBounced(
    recipientEmail: string,
    request: SendGridSendRequest,
  ): Promise<{ sent: boolean; reason?: string }> {
    // BUG: TOCTOU race condition (Time Of Check vs Time Of Use).
    //
    // Step 1: Check suppression list (takes ~200ms network round trip)
    const bounces = await this.getBounces();
    const isBounced = bounces.some((b) => b.email === recipientEmail);

    // RACE WINDOW: Between step 1 and step 2, another service or webhook
    // could add this email to the suppression list. Scenarios:
    //   - A simultaneous email send bounced and the bounce webhook fired
    //   - An admin manually added the email to suppressions
    //   - A spam complaint was processed
    //   - Another instance of this service processed a bounce

    if (isBounced) {
      return { sent: false, reason: "Email is on bounce suppression list" };
    }

    // Step 2: Send email (~200ms later — the suppression state may have changed)
    // BUG: The correct approach is to:
    // 1. Send the email optimistically (SendGrid handles suppression server-side)
    // 2. Check the Activity API afterward to confirm delivery status
    // 3. Or use SendGrid's server-side suppression which atomically checks
    //    suppression status at send time
    await this.sendEmail(request);

    return { sent: true };
  }

  async sendBulkIfNotBounced(
    recipients: Array<{ email: string; name?: string }>,
    from: { email: string },
    subject: string,
    content: string,
  ): Promise<{ sent: string[]; skipped: string[] }> {
    // BUG: Fetches bounces once, then iterates and sends sequentially.
    // By the time the last recipient is processed (could be minutes later
    // for a large list), the bounce list is stale.
    const bounces = await this.getBounces();
    const bouncedEmails = new Set(bounces.map((b) => b.email));

    const sent: string[] = [];
    const skipped: string[] = [];

    for (const recipient of recipients) {
      if (bouncedEmails.has(recipient.email)) {
        skipped.push(recipient.email);
        continue;
      }

      await this.sendEmail({
        personalizations: [{ to: [recipient] }],
        from,
        subject,
        content: [{ type: "text/html", value: content }],
      });
      sent.push(recipient.email);
    }

    return { sent, skipped };
  }
}
```

**Expected violation:**
```
CTR-request-shape: sendEmailIfNotBounced performs a read (GET /v3/suppression/bounces)
followed by a conditional write (POST /v3/mail/send) without atomicity guarantees.
The suppression state can change between the check and the send (TOCTOU race condition).
The send call does not include any mechanism to verify suppression status at execution time.
```

**Production impact:** In a high-volume system, the race window is significant. A bounce webhook and a re-send attempt can interleave: (1) Check bounces -- email not found; (2) Bounce webhook fires -- email added to suppressions; (3) Send email -- goes to a suppressed address. The email bounces again, incrementing the repeat-bounce counter. Enough repeat bounces trigger SendGrid's automated account review, which can suspend the entire sending domain. For bulk sends, the stale bounce list means hundreds of emails may be sent to addresses that bounced during the batch processing.
