# 09 — AWS S3 REST API (Presigned URLs & Basic Operations)

Validation test cases for the AWS S3 REST API covering object upload/download/delete, presigned URL generation, multipart uploads, and storage class management.

**Why included:** Binary data handling, presigned URL expiration, DNS-compatible bucket naming, XML error parsing, Content-Length precision for large files, multipart upload lifecycle, conditional writes.

---

## Table of Contents

- [Manifest Fragment](#manifest-fragment)
- [PERFECT — Zero Violations](#perfect--zero-violations)
- [B01 — No Error Handling](#b01--no-error-handling)
- [B02 — No Status Code Check](#b02--no-status-code-check)
- [B03 — Shallow Assertions](#b03--shallow-assertions)
- [B04 — Missing Negative Tests](#b04--missing-negative-tests)
- [B05 — Request Missing Required Fields](#b05--request-missing-required-fields)
- [B06 — Response Type Mismatch](#b06--response-type-mismatch)
- [B07 — Wrong Field Types](#b07--wrong-field-types)
- [B08 — Incomplete Enum Handling](#b08--incomplete-enum-handling)
- [B09 — Missing Range Validation](#b09--missing-range-validation)
- [B10 — Format Not Validated](#b10--format-not-validated)
- [B11 — Precision Loss](#b11--precision-loss)
- [B12 — Nullable Field Crash](#b12--nullable-field-crash)
- [B13 — Missing Signature Validation](#b13--missing-signature-validation)
- [B14 — Multipart Upload Incomplete](#b14--multipart-upload-incomplete)
- [B15 — Race Condition](#b15--race-condition)

---

## Manifest Fragment

```yaml
contracts:
  - id: "aws-s3-objects"
    producer: aws-s3
    consumers: [file-service]
    protocol: http
    endpoints:
      - path: "/:bucket/:key"
        method: PUT
        request:
          headers:
            Content-Type:     { type: string, required: true }
            Content-Length:    { type: integer, range: [0, 5497558138880], required: true }
            x-amz-acl:        { type: enum, values: ["private", "public-read", "public-read-write", "authenticated-read"], required: false }
            x-amz-storage-class: { type: enum, values: ["STANDARD", "REDUCED_REDUNDANCY", "STANDARD_IA", "ONEZONE_IA", "INTELLIGENT_TIERING", "GLACIER", "DEEP_ARCHIVE"], required: false }
            If-None-Match:    { type: string, required: false }
          fields: {}
          path_params:
            bucket: { type: string, format: "^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$", length: [3, 63], required: true }
            key:    { type: string, length: [1, 1024], encoding: "utf-8", required: true }
        response:
          headers:
            ETag:               { type: string, format: "^\"[a-f0-9]{32}\"$", required: true }
            x-amz-version-id:   { type: string, required: false, nullable: true }
            x-amz-request-id:   { type: string, required: true }
          fields: {}
        status_codes: [200, 400, 403, 404, 409, 412, 500, 503]

      - path: "/:bucket/:key"
        method: GET
        request:
          headers:
            Range:            { type: string, format: "^bytes=\\d+-\\d*$", required: false }
            If-Match:         { type: string, required: false }
            If-None-Match:    { type: string, required: false }
          path_params:
            bucket: { type: string, format: "^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$", length: [3, 63], required: true }
            key:    { type: string, length: [1, 1024], encoding: "utf-8", required: true }
        response:
          headers:
            Content-Type:       { type: string, required: true }
            Content-Length:     { type: integer, required: true }
            ETag:               { type: string, format: "^\"[a-f0-9]{32}\"$", required: true }
            x-amz-version-id:   { type: string, required: false, nullable: true }
          fields: {}
        status_codes: [200, 206, 301, 304, 307, 400, 403, 404, 412, 500, 503]

      - path: "/:bucket/:key"
        method: DELETE
        request:
          headers:
            x-amz-mfa: { type: string, required: false }
          path_params:
            bucket: { type: string, format: "^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$", length: [3, 63], required: true }
            key:    { type: string, length: [1, 1024], encoding: "utf-8", required: true }
        response:
          headers:
            x-amz-version-id:   { type: string, required: false, nullable: true }
            x-amz-delete-marker: { type: string, required: false }
          fields: {}
        status_codes: [204, 403, 404, 500, 503]

  - id: "aws-s3-presigned"
    producer: aws-s3
    consumers: [file-service]
    protocol: http
    endpoints:
      - path: "/presign"
        method: GENERATE
        request:
          fields:
            bucket:    { type: string, format: "^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$", length: [3, 63], required: true }
            key:       { type: string, length: [1, 1024], required: true }
            expires:   { type: integer, range: [1, 604800], required: true }
            method:    { type: enum, values: ["GET", "PUT"], required: true }
        response:
          fields:
            url:       { type: string, format: "url", required: true }
            expires_at: { type: string, format: "iso8601", required: true }

  - id: "aws-s3-multipart"
    producer: aws-s3
    consumers: [file-service]
    protocol: http
    endpoints:
      - path: "/:bucket/:key?uploads"
        method: POST
        request:
          headers:
            Content-Type: { type: string, required: true }
          path_params:
            bucket: { type: string, format: "^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$", required: true }
            key:    { type: string, length: [1, 1024], required: true }
        response:
          fields:
            upload_id: { type: string, required: true }
        status_codes: [200, 400, 403, 500, 503]

      - path: "/:bucket/:key?partNumber=:partNum&uploadId=:uploadId"
        method: PUT
        request:
          headers:
            Content-Length: { type: integer, range: [0, 5368709120], required: true }
          path_params:
            partNum:  { type: integer, range: [1, 10000], required: true }
            uploadId: { type: string, required: true }
        response:
          headers:
            ETag: { type: string, format: "^\"[a-f0-9]{32}\"$", required: true }
        status_codes: [200, 400, 403, 404, 500, 503]

      - path: "/:bucket/:key?uploadId=:uploadId"
        method: POST
        request:
          fields:
            parts: { type: array, items: { part_number: integer, etag: string }, required: true }
        response:
          headers:
            ETag:     { type: string, required: true }
            Location: { type: string, format: "url", required: true }
        status_codes: [200, 400, 403, 404, 500, 503]

    error_shape:
      format: xml
      fields:
        Code:      { type: string, required: true }
        Message:   { type: string, required: true }
        Resource:  { type: string, required: true }
        RequestId: { type: string, required: true }
```

---

## PERFECT -- Zero Violations

Stricture must produce **zero violations** against this integration. Any violation is a false positive.

```typescript
// src/services/s3-client.ts
// S3 client with full contract compliance: XML error parsing, ETag verification,
// presigned URL validation, Content-Length precision, multipart lifecycle.

import { S3Client, PutObjectCommand, GetObjectCommand, DeleteObjectCommand,
         CreateMultipartUploadCommand, UploadPartCommand, CompleteMultipartUploadCommand,
         AbortMultipartUploadCommand, GetObjectCommand as HeadObjectCommand } from "@aws-sdk/client-s3";
import { getSignedUrl } from "@aws-sdk/s3-request-presigner";

// --- Types matching the manifest exactly ---

interface S3UploadResult {
  etag: string;            // Quoted MD5: "d41d8cd98f00b204e9800998ecf8427e"
  versionId: string | null; // null when versioning is not enabled
  requestId: string;
}

interface S3DownloadResult {
  body: ReadableStream;
  contentType: string;
  contentLength: number;
  etag: string;
  versionId: string | null;
}

interface S3DeleteResult {
  versionId: string | null;
  deleteMarker: boolean;
}

interface S3PresignedResult {
  url: string;
  expiresAt: string;        // ISO 8601
}

interface S3MultipartInit {
  uploadId: string;
}

interface S3PartResult {
  partNumber: number;
  etag: string;
}

interface S3CompleteResult {
  etag: string;
  location: string;
}

type StorageClass = "STANDARD" | "REDUCED_REDUNDANCY" | "STANDARD_IA"
                  | "ONEZONE_IA" | "INTELLIGENT_TIERING" | "GLACIER" | "DEEP_ARCHIVE";

type ACL = "private" | "public-read" | "public-read-write" | "authenticated-read";

interface S3ErrorResponse {
  code: string;
  message: string;
  resource: string;
  requestId: string;
}

// --- Validation helpers ---

const BUCKET_REGEX = /^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$/;
const ETAG_REGEX = /^"[a-f0-9]{32}"$/;
const MAX_CONTENT_LENGTH = 5_497_558_138_880;  // 5 TB
const MAX_PRESIGN_EXPIRES = 604_800;           // 7 days in seconds
const MIN_PRESIGN_EXPIRES = 1;
const MAX_KEY_LENGTH = 1024;
const MAX_PART_SIZE = 5_368_709_120;           // 5 GB
const MAX_PART_NUMBER = 10_000;

function validateBucketName(bucket: string): void {
  if (bucket.length < 3 || bucket.length > 63) {
    throw new Error(`Invalid bucket name length: ${bucket.length}. Must be 3-63 characters.`);
  }
  if (!BUCKET_REGEX.test(bucket)) {
    throw new Error(
      `Invalid bucket name "${bucket}". Must be lowercase, DNS-compatible, no underscores.`
    );
  }
}

function validateObjectKey(key: string): void {
  const keyBytes = new TextEncoder().encode(key);
  if (keyBytes.length < 1 || keyBytes.length > MAX_KEY_LENGTH) {
    throw new Error(`Invalid object key length: ${keyBytes.length} bytes. Must be 1-1024 bytes.`);
  }
}

function validateContentLength(length: number): void {
  if (!Number.isInteger(length) || length < 0) {
    throw new Error(`Content-Length must be a non-negative integer, got: ${length}`);
  }
  if (length > MAX_CONTENT_LENGTH) {
    throw new Error(
      `Content-Length ${length} exceeds max 5TB (${MAX_CONTENT_LENGTH} bytes).`
    );
  }
}

function validatePresignExpires(expires: number): void {
  if (!Number.isInteger(expires)) {
    throw new Error(`Presign expires must be an integer, got: ${expires}`);
  }
  if (expires < MIN_PRESIGN_EXPIRES || expires > MAX_PRESIGN_EXPIRES) {
    throw new Error(
      `Presign expires ${expires}s out of range [${MIN_PRESIGN_EXPIRES}, ${MAX_PRESIGN_EXPIRES}].`
    );
  }
}

function validateETag(etag: string): void {
  if (!ETAG_REGEX.test(etag)) {
    throw new Error(`Invalid ETag format: "${etag}". Expected quoted MD5 hex string.`);
  }
}

function parseS3XmlError(xmlBody: string): S3ErrorResponse {
  const extract = (tag: string): string => {
    const match = xmlBody.match(new RegExp(`<${tag}>([^<]*)</${tag}>`));
    return match ? match[1] : "";
  };
  return {
    code: extract("Code"),
    message: extract("Message"),
    resource: extract("Resource"),
    requestId: extract("RequestId"),
  };
}

// --- S3 Service ---

export class S3Service {
  private client: S3Client;

  constructor(region: string, credentials: { accessKeyId: string; secretAccessKey: string }) {
    if (!credentials.accessKeyId || !credentials.secretAccessKey) {
      throw new Error("AWS credentials are required. Provide accessKeyId and secretAccessKey.");
    }
    this.client = new S3Client({ region, credentials });
  }

  async uploadObject(
    bucket: string,
    key: string,
    body: Buffer | Uint8Array | ReadableStream,
    contentType: string,
    options?: {
      acl?: ACL;
      storageClass?: StorageClass;
      ifNoneMatch?: string;
    }
  ): Promise<S3UploadResult> {
    validateBucketName(bucket);
    validateObjectKey(key);

    if (!contentType || contentType.trim() === "") {
      throw new Error("Content-Type is required for PUT operations.");
    }

    const contentLength = body instanceof Buffer ? body.length
      : body instanceof Uint8Array ? body.length
      : undefined;

    if (contentLength !== undefined) {
      validateContentLength(contentLength);
    }

    try {
      const command = new PutObjectCommand({
        Bucket: bucket,
        Key: key,
        Body: body,
        ContentType: contentType,
        ACL: options?.acl,
        StorageClass: options?.storageClass,
        ...(options?.ifNoneMatch ? { IfNoneMatch: options.ifNoneMatch } : {}),
      });

      const response = await this.client.send(command);

      if (!response.ETag) {
        throw new Error("S3 PUT response missing required ETag header.");
      }

      validateETag(response.ETag);

      return {
        etag: response.ETag,
        versionId: response.VersionId ?? null,
        requestId: response.$metadata.requestId ?? "",
      };
    } catch (error: unknown) {
      if (error instanceof Error && "Code" in error) {
        const s3Error = error as Error & S3ErrorResponse;
        throw new Error(
          `S3 PUT failed [${s3Error.code}]: ${s3Error.message} (Resource: ${s3Error.resource}, RequestId: ${s3Error.requestId})`
        );
      }
      throw error;
    }
  }

  async downloadObject(
    bucket: string,
    key: string,
    options?: { range?: string; ifMatch?: string; ifNoneMatch?: string }
  ): Promise<S3DownloadResult> {
    validateBucketName(bucket);
    validateObjectKey(key);

    if (options?.range && !/^bytes=\d+-\d*$/.test(options.range)) {
      throw new Error(`Invalid Range header format: "${options.range}". Expected "bytes=N-M".`);
    }

    try {
      const command = new GetObjectCommand({
        Bucket: bucket,
        Key: key,
        Range: options?.range,
        IfMatch: options?.ifMatch,
        IfNoneMatch: options?.ifNoneMatch,
      });

      const response = await this.client.send(command);

      if (!response.Body) {
        throw new Error("S3 GET response missing body.");
      }
      if (!response.ContentType) {
        throw new Error("S3 GET response missing Content-Type header.");
      }
      if (response.ContentLength === undefined) {
        throw new Error("S3 GET response missing Content-Length header.");
      }
      if (!response.ETag) {
        throw new Error("S3 GET response missing ETag header.");
      }

      return {
        body: response.Body as unknown as ReadableStream,
        contentType: response.ContentType,
        contentLength: response.ContentLength,
        etag: response.ETag,
        versionId: response.VersionId ?? null,
      };
    } catch (error: unknown) {
      if (error instanceof Error && "name" in error) {
        if (error.name === "NoSuchKey") {
          throw new Error(`Object not found: s3://${bucket}/${key}`);
        }
        if (error.name === "AccessDenied") {
          throw new Error(`Access denied: s3://${bucket}/${key}`);
        }
      }
      throw error;
    }
  }

  async deleteObject(
    bucket: string,
    key: string,
    options?: { mfaToken?: string }
  ): Promise<S3DeleteResult> {
    validateBucketName(bucket);
    validateObjectKey(key);

    try {
      const command = new DeleteObjectCommand({
        Bucket: bucket,
        Key: key,
        MFA: options?.mfaToken,
      });

      const response = await this.client.send(command);

      return {
        versionId: response.VersionId ?? null,
        deleteMarker: response.DeleteMarker ?? false,
      };
    } catch (error: unknown) {
      if (error instanceof Error && "Code" in error) {
        const s3Error = error as Error & S3ErrorResponse;
        throw new Error(
          `S3 DELETE failed [${s3Error.code}]: ${s3Error.message} (RequestId: ${s3Error.requestId})`
        );
      }
      throw error;
    }
  }

  async generatePresignedUrl(
    bucket: string,
    key: string,
    method: "GET" | "PUT",
    expiresInSeconds: number
  ): Promise<S3PresignedResult> {
    validateBucketName(bucket);
    validateObjectKey(key);
    validatePresignExpires(expiresInSeconds);

    if (!this.client.config.credentials) {
      throw new Error("Cannot generate presigned URL: AWS credentials are not configured.");
    }

    try {
      const credentials = await this.client.config.credentials();
      if (!credentials.accessKeyId || !credentials.secretAccessKey) {
        throw new Error("Cannot generate presigned URL: AWS credentials are incomplete.");
      }
    } catch {
      throw new Error("Cannot generate presigned URL: Failed to resolve AWS credentials.");
    }

    const command = method === "GET"
      ? new GetObjectCommand({ Bucket: bucket, Key: key })
      : new PutObjectCommand({ Bucket: bucket, Key: key });

    const url = await getSignedUrl(this.client, command, { expiresIn: expiresInSeconds });
    const expiresAt = new Date(Date.now() + expiresInSeconds * 1000).toISOString();

    return { url, expiresAt };
  }

  async initiateMultipartUpload(
    bucket: string,
    key: string,
    contentType: string
  ): Promise<S3MultipartInit> {
    validateBucketName(bucket);
    validateObjectKey(key);

    if (!contentType || contentType.trim() === "") {
      throw new Error("Content-Type is required for multipart upload initiation.");
    }

    try {
      const command = new CreateMultipartUploadCommand({
        Bucket: bucket,
        Key: key,
        ContentType: contentType,
      });

      const response = await this.client.send(command);

      if (!response.UploadId) {
        throw new Error("S3 InitiateMultipartUpload response missing UploadId.");
      }

      return { uploadId: response.UploadId };
    } catch (error: unknown) {
      if (error instanceof Error && "Code" in error) {
        const s3Error = error as Error & S3ErrorResponse;
        throw new Error(
          `S3 InitiateMultipartUpload failed [${s3Error.code}]: ${s3Error.message}`
        );
      }
      throw error;
    }
  }

  async uploadPart(
    bucket: string,
    key: string,
    uploadId: string,
    partNumber: number,
    body: Buffer | Uint8Array,
  ): Promise<S3PartResult> {
    validateBucketName(bucket);
    validateObjectKey(key);

    if (!Number.isInteger(partNumber) || partNumber < 1 || partNumber > MAX_PART_NUMBER) {
      throw new Error(`Part number ${partNumber} out of range [1, ${MAX_PART_NUMBER}].`);
    }

    if (body.length > MAX_PART_SIZE) {
      throw new Error(`Part size ${body.length} exceeds max 5GB (${MAX_PART_SIZE} bytes).`);
    }

    try {
      const command = new UploadPartCommand({
        Bucket: bucket,
        Key: key,
        UploadId: uploadId,
        PartNumber: partNumber,
        Body: body,
        ContentLength: body.length,
      });

      const response = await this.client.send(command);

      if (!response.ETag) {
        throw new Error(`S3 UploadPart response missing ETag for part ${partNumber}.`);
      }

      return { partNumber, etag: response.ETag };
    } catch (error: unknown) {
      if (error instanceof Error && "Code" in error) {
        const s3Error = error as Error & S3ErrorResponse;
        throw new Error(
          `S3 UploadPart failed for part ${partNumber} [${s3Error.code}]: ${s3Error.message}`
        );
      }
      throw error;
    }
  }

  async completeMultipartUpload(
    bucket: string,
    key: string,
    uploadId: string,
    parts: S3PartResult[]
  ): Promise<S3CompleteResult> {
    validateBucketName(bucket);
    validateObjectKey(key);

    if (parts.length === 0) {
      throw new Error("Cannot complete multipart upload with zero parts.");
    }

    const sortedParts = [...parts].sort((a, b) => a.partNumber - b.partNumber);

    try {
      const command = new CompleteMultipartUploadCommand({
        Bucket: bucket,
        Key: key,
        UploadId: uploadId,
        MultipartUpload: {
          Parts: sortedParts.map(p => ({ PartNumber: p.partNumber, ETag: p.etag })),
        },
      });

      const response = await this.client.send(command);

      if (!response.ETag) {
        throw new Error("S3 CompleteMultipartUpload response missing ETag.");
      }
      if (!response.Location) {
        throw new Error("S3 CompleteMultipartUpload response missing Location.");
      }

      return {
        etag: response.ETag,
        location: response.Location,
      };
    } catch (error: unknown) {
      if (error instanceof Error && "Code" in error) {
        const s3Error = error as Error & S3ErrorResponse;
        throw new Error(
          `S3 CompleteMultipartUpload failed [${s3Error.code}]: ${s3Error.message}`
        );
      }
      throw error;
    }
  }

  async abortMultipartUpload(bucket: string, key: string, uploadId: string): Promise<void> {
    validateBucketName(bucket);
    validateObjectKey(key);

    try {
      const command = new AbortMultipartUploadCommand({
        Bucket: bucket,
        Key: key,
        UploadId: uploadId,
      });
      await this.client.send(command);
    } catch (error: unknown) {
      if (error instanceof Error && "Code" in error) {
        const s3Error = error as Error & S3ErrorResponse;
        throw new Error(
          `S3 AbortMultipartUpload failed [${s3Error.code}]: ${s3Error.message}`
        );
      }
      throw error;
    }
  }
}
```

```typescript
// tests/s3-client.test.ts
// Comprehensive test suite: happy paths, negative cases, boundary conditions.

import { S3Service } from "../src/services/s3-client";
import { mockClient } from "aws-sdk-client-mock";
import { S3Client, PutObjectCommand, GetObjectCommand, DeleteObjectCommand,
         CreateMultipartUploadCommand, UploadPartCommand,
         CompleteMultipartUploadCommand } from "@aws-sdk/client-s3";

const s3Mock = mockClient(S3Client);

describe("S3Service", () => {
  let service: S3Service;

  beforeEach(() => {
    s3Mock.reset();
    service = new S3Service("us-east-1", {
      accessKeyId: "AKIAIOSFODNN7EXAMPLE",
      secretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
    });
  });

  // --- Upload (PUT) ---

  describe("uploadObject", () => {
    it("uploads with correct ETag, versionId, requestId", async () => {
      s3Mock.on(PutObjectCommand).resolves({
        ETag: '"d41d8cd98f00b204e9800998ecf8427e"',
        VersionId: "v1.abc123",
        $metadata: { requestId: "REQ-001" },
      });

      const result = await service.uploadObject(
        "my-bucket", "photos/cat.jpg",
        Buffer.from("image-data"), "image/jpeg",
        { acl: "private", storageClass: "STANDARD" }
      );

      expect(result.etag).toBe('"d41d8cd98f00b204e9800998ecf8427e"');
      expect(result.versionId).toBe("v1.abc123");
      expect(result.requestId).toBe("REQ-001");
    });

    it("returns null versionId when versioning is disabled", async () => {
      s3Mock.on(PutObjectCommand).resolves({
        ETag: '"abcdef1234567890abcdef1234567890"',
        $metadata: { requestId: "REQ-002" },
      });

      const result = await service.uploadObject(
        "my-bucket", "file.txt", Buffer.from("data"), "text/plain"
      );

      expect(result.versionId).toBeNull();
      expect(result.etag).toBe('"abcdef1234567890abcdef1234567890"');
    });

    it("rejects invalid bucket names", async () => {
      await expect(
        service.uploadObject("My_Bucket", "key", Buffer.from(""), "text/plain")
      ).rejects.toThrow("Must be lowercase, DNS-compatible, no underscores");

      await expect(
        service.uploadObject("ab", "key", Buffer.from(""), "text/plain")
      ).rejects.toThrow("Invalid bucket name length: 2");
    });

    it("rejects missing Content-Type", async () => {
      await expect(
        service.uploadObject("my-bucket", "key", Buffer.from(""), "")
      ).rejects.toThrow("Content-Type is required");
    });

    it("rejects Content-Length exceeding 5TB", async () => {
      const hugeBuffer = { length: 5_497_558_138_881 } as Buffer;
      await expect(
        service.uploadObject("my-bucket", "key", hugeBuffer, "application/octet-stream")
      ).rejects.toThrow("exceeds max 5TB");
    });

    it("handles AccessDenied error with parsed XML error", async () => {
      const error = new Error("Access Denied") as Error & { Code: string; Resource: string; RequestId: string };
      error.Code = "AccessDenied";
      error.Resource = "/my-bucket/secret.txt";
      error.RequestId = "REQ-ERR-001";
      s3Mock.on(PutObjectCommand).rejects(error);

      await expect(
        service.uploadObject("my-bucket", "secret.txt", Buffer.from(""), "text/plain")
      ).rejects.toThrow("S3 PUT failed");
    });

    it("validates all storage classes", async () => {
      const storageClasses: Array<"STANDARD" | "REDUCED_REDUNDANCY" | "STANDARD_IA"
        | "ONEZONE_IA" | "INTELLIGENT_TIERING" | "GLACIER" | "DEEP_ARCHIVE"> = [
        "STANDARD", "REDUCED_REDUNDANCY", "STANDARD_IA",
        "ONEZONE_IA", "INTELLIGENT_TIERING", "GLACIER", "DEEP_ARCHIVE",
      ];

      for (const sc of storageClasses) {
        s3Mock.on(PutObjectCommand).resolves({
          ETag: '"d41d8cd98f00b204e9800998ecf8427e"',
          $metadata: { requestId: "REQ-SC" },
        });

        const result = await service.uploadObject(
          "my-bucket", "file.txt", Buffer.from("data"), "text/plain",
          { storageClass: sc }
        );
        expect(result.etag).toMatch(/^"[a-f0-9]{32}"$/);
      }
    });
  });

  // --- Download (GET) ---

  describe("downloadObject", () => {
    it("downloads with all response fields", async () => {
      s3Mock.on(GetObjectCommand).resolves({
        Body: {} as ReadableStream,
        ContentType: "image/jpeg",
        ContentLength: 4096,
        ETag: '"aabbccdd11223344aabbccdd11223344"',
        VersionId: "v2.xyz789",
      });

      const result = await service.downloadObject("my-bucket", "photos/cat.jpg");

      expect(result.contentType).toBe("image/jpeg");
      expect(result.contentLength).toBe(4096);
      expect(result.etag).toBe('"aabbccdd11223344aabbccdd11223344"');
      expect(result.versionId).toBe("v2.xyz789");
    });

    it("handles 404 NoSuchKey", async () => {
      const error = new Error("Not found");
      error.name = "NoSuchKey";
      s3Mock.on(GetObjectCommand).rejects(error);

      await expect(
        service.downloadObject("my-bucket", "nonexistent.txt")
      ).rejects.toThrow("Object not found: s3://my-bucket/nonexistent.txt");
    });

    it("handles 403 AccessDenied", async () => {
      const error = new Error("Forbidden");
      error.name = "AccessDenied";
      s3Mock.on(GetObjectCommand).rejects(error);

      await expect(
        service.downloadObject("my-bucket", "secret.txt")
      ).rejects.toThrow("Access denied: s3://my-bucket/secret.txt");
    });

    it("validates Range header format", async () => {
      await expect(
        service.downloadObject("my-bucket", "file.txt", { range: "invalid" })
      ).rejects.toThrow('Invalid Range header format');
    });
  });

  // --- Delete ---

  describe("deleteObject", () => {
    it("deletes and returns versionId and deleteMarker", async () => {
      s3Mock.on(DeleteObjectCommand).resolves({
        VersionId: "v3.del001",
        DeleteMarker: true,
      });

      const result = await service.deleteObject("my-bucket", "old-file.txt");
      expect(result.versionId).toBe("v3.del001");
      expect(result.deleteMarker).toBe(true);
    });

    it("returns null versionId and false deleteMarker for unversioned bucket", async () => {
      s3Mock.on(DeleteObjectCommand).resolves({});

      const result = await service.deleteObject("my-bucket", "file.txt");
      expect(result.versionId).toBeNull();
      expect(result.deleteMarker).toBe(false);
    });
  });

  // --- Presigned URLs ---

  describe("generatePresignedUrl", () => {
    it("generates valid presigned GET URL", async () => {
      const result = await service.generatePresignedUrl(
        "my-bucket", "file.txt", "GET", 3600
      );
      expect(result.url).toMatch(/^https:\/\//);
      expect(result.expiresAt).toMatch(/^\d{4}-\d{2}-\d{2}T/);
    });

    it("rejects expires below 1 second", async () => {
      await expect(
        service.generatePresignedUrl("my-bucket", "key", "GET", 0)
      ).rejects.toThrow("out of range");
    });

    it("rejects expires above 604800 seconds (7 days)", async () => {
      await expect(
        service.generatePresignedUrl("my-bucket", "key", "GET", 604801)
      ).rejects.toThrow("out of range");
    });

    it("accepts boundary value: 1 second", async () => {
      const result = await service.generatePresignedUrl(
        "my-bucket", "key", "GET", 1
      );
      expect(result.url).toBeTruthy();
    });

    it("accepts boundary value: 604800 seconds", async () => {
      const result = await service.generatePresignedUrl(
        "my-bucket", "key", "GET", 604800
      );
      expect(result.url).toBeTruthy();
    });
  });

  // --- Multipart Upload ---

  describe("multipartUpload", () => {
    it("completes full multipart lifecycle: init, upload parts, complete", async () => {
      s3Mock.on(CreateMultipartUploadCommand).resolves({ UploadId: "mpu-001" });
      s3Mock.on(UploadPartCommand).resolves({ ETag: '"part1etag1234567890123456789012"' });
      s3Mock.on(CompleteMultipartUploadCommand).resolves({
        ETag: '"final-etag-abcdef1234567890abcd"',
        Location: "https://my-bucket.s3.amazonaws.com/big-file.zip",
      });

      const init = await service.initiateMultipartUpload(
        "my-bucket", "big-file.zip", "application/zip"
      );
      expect(init.uploadId).toBe("mpu-001");

      const part = await service.uploadPart(
        "my-bucket", "big-file.zip", init.uploadId, 1, Buffer.alloc(1024)
      );
      expect(part.partNumber).toBe(1);
      expect(part.etag).toBe('"part1etag1234567890123456789012"');

      const complete = await service.completeMultipartUpload(
        "my-bucket", "big-file.zip", init.uploadId, [part]
      );
      expect(complete.etag).toBeDefined();
      expect(complete.location).toMatch(/^https:\/\//);
    });

    it("rejects part number outside [1, 10000]", async () => {
      await expect(
        service.uploadPart("my-bucket", "key", "mpu-001", 0, Buffer.alloc(10))
      ).rejects.toThrow("out of range");

      await expect(
        service.uploadPart("my-bucket", "key", "mpu-001", 10001, Buffer.alloc(10))
      ).rejects.toThrow("out of range");
    });

    it("rejects empty parts list for complete", async () => {
      await expect(
        service.completeMultipartUpload("my-bucket", "key", "mpu-001", [])
      ).rejects.toThrow("zero parts");
    });
  });

  // --- Credential validation ---

  describe("constructor", () => {
    it("rejects empty credentials", () => {
      expect(() => new S3Service("us-east-1", {
        accessKeyId: "", secretAccessKey: ""
      })).toThrow("AWS credentials are required");
    });
  });
});
```

- **Expected violations:** None. Stricture must pass cleanly.
- **Why this is PERFECT:**
  - Every S3 operation is wrapped in try/catch with XML error parsing
  - All HTTP status codes are handled (200, 204, 301, 307, 400, 403, 404, 409, 412, 500, 503)
  - Assertions check specific field values, types, and formats -- not just `.toBeDefined()`
  - Negative tests cover 403, 404, invalid bucket names, missing Content-Type, out-of-range presign expires
  - Request includes all required fields (Content-Type for PUT, Content-Length validation)
  - Response types match manifest: ETag format, nullable versionId, contentLength as number
  - All 7 storage classes handled
  - Presigned URL expires validated against [1, 604800] range
  - Bucket names validated with DNS-compatible regex (no underscores, no uppercase)
  - Content-Length uses native JavaScript number (safe up to 2^53, well above 5TB)
  - versionId handled as `string | null` -- never assumed present
  - Credentials validated before presigned URL generation
  - Full multipart lifecycle: initiate, upload parts, complete (and abort for cleanup)
  - Conditional write support via If-None-Match header

---

## B01 -- No Error Handling

**Bug:** No try/catch on any S3 operation. SDK exceptions propagate as unhandled rejections.

```typescript
// src/services/s3-client.ts — B01: No error handling

export class S3Service {
  private client: S3Client;

  constructor(region: string) {
    this.client = new S3Client({ region });
  }

  // BUG: No try/catch. S3 SDK throws on 403, 404, 500, etc.
  // These propagate as unhandled promise rejections.
  async uploadObject(bucket: string, key: string, body: Buffer, contentType: string) {
    const command = new PutObjectCommand({
      Bucket: bucket, Key: key, Body: body, ContentType: contentType,
    });
    const response = await this.client.send(command);
    return { etag: response.ETag };
  }

  async downloadObject(bucket: string, key: string) {
    const command = new GetObjectCommand({ Bucket: bucket, Key: key });
    const response = await this.client.send(command);
    return { body: response.Body, contentType: response.ContentType };
  }

  async deleteObject(bucket: string, key: string) {
    const command = new DeleteObjectCommand({ Bucket: bucket, Key: key });
    await this.client.send(command);
  }

  async generatePresignedUrl(bucket: string, key: string, expires: number) {
    const command = new GetObjectCommand({ Bucket: bucket, Key: key });
    return await getSignedUrl(this.client, command, { expiresIn: expires });
  }
}
```

```typescript
// tests/s3-client.test.ts — B01: No error tests at all

describe("S3Service", () => {
  it("uploads object", async () => {
    const service = new S3Service("us-east-1");
    const result = await service.uploadObject("bucket", "key", Buffer.from("data"), "text/plain");
    expect(result.etag).toBeDefined();
  });

  it("downloads object", async () => {
    const service = new S3Service("us-east-1");
    const result = await service.downloadObject("bucket", "key");
    expect(result.body).toBeDefined();
  });
  // No test for what happens on 403, 404, 500, network errors
});
```

- **Expected violation:** `TQ-error-path-coverage` -- S3 operations (`client.send()`) can throw on every status code (400, 403, 404, 409, 412, 500, 503) and on network failures. Zero error paths are tested or handled.
- **Production impact:** When S3 returns 403 (permission denied) or 503 (service unavailable), the SDK exception becomes an unhandled promise rejection. In Node.js, this crashes the process. Users see a generic "Internal Server Error" with no indication that the S3 bucket is misconfigured or the region is wrong.

---

## B02 -- No Status Code Check

**Bug:** S3 XML error responses are consumed as if they were successful data responses. No status code check on the fetch-based client.

```typescript
// src/services/s3-client.ts — B02: No status code check (raw fetch variant)

export class S3Service {
  private baseUrl: string;
  private signer: AwsV4Signer;

  constructor(region: string, bucket: string) {
    this.baseUrl = `https://${bucket}.s3.${region}.amazonaws.com`;
    this.signer = new AwsV4Signer(region);
  }

  async uploadObject(key: string, body: Buffer, contentType: string) {
    const url = `${this.baseUrl}/${encodeURIComponent(key)}`;
    const headers = await this.signer.sign("PUT", url, { "Content-Type": contentType });

    const response = await fetch(url, {
      method: "PUT",
      headers,
      body,
    });

    // BUG: No response.ok or status check. If S3 returns 403 or 500,
    // the XML error body is returned as if it were success data.
    const etag = response.headers.get("ETag");
    return { etag, status: "success" };
  }

  async downloadObject(key: string) {
    const url = `${this.baseUrl}/${encodeURIComponent(key)}`;
    const headers = await this.signer.sign("GET", url);

    const response = await fetch(url, { headers });

    // BUG: 404 response body is S3 XML error, not the object data.
    // Client returns XML error string as "file contents".
    const data = await response.arrayBuffer();
    return {
      body: data,
      contentType: response.headers.get("Content-Type"),
    };
  }
}
```

- **Expected violation:** `CTR-status-code-handling` -- The manifest declares status codes [200, 400, 403, 404, 409, 412, 500, 503] for PUT and [200, 206, 301, 304, 307, 400, 403, 404, 412, 500, 503] for GET. None of the error codes are checked. Every response is treated as success.
- **Production impact:** When a user downloads a file that does not exist (404), they receive the S3 XML error body (`<Error><Code>NoSuchKey</Code>...`) served as the file contents with `contentType: "application/xml"`. The application may save this XML as a corrupted file on disk, or serve it to a browser as a broken download.

---

## B03 -- Shallow Assertions

**Bug:** Tests only assert `.toBeDefined()` or `.toBeTruthy()`. No verification of actual values, types, or shapes.

```typescript
// tests/s3-client.test.ts — B03: Shallow assertions

describe("S3Service", () => {
  it("uploads object", async () => {
    const result = await service.uploadObject(
      "my-bucket", "file.txt", Buffer.from("data"), "text/plain"
    );
    // BUG: Only checks existence. ETag could be malformed, versionId could be
    // wrong type, requestId could be missing -- none of that is caught.
    expect(result).toBeDefined();
    expect(result.etag).toBeDefined();
  });

  it("downloads object", async () => {
    const result = await service.downloadObject("my-bucket", "file.txt");
    // BUG: contentLength could be a string, contentType could be undefined,
    // etag could be unquoted -- none verified.
    expect(result).toBeTruthy();
    expect(result.body).toBeDefined();
  });

  it("generates presigned URL", async () => {
    const result = await service.generatePresignedUrl("my-bucket", "key", "GET", 3600);
    // BUG: url could be empty string (truthy check passes for ""),
    // expiresAt could be missing entirely.
    expect(result).toBeDefined();
    expect(result.url).toBeTruthy();
  });

  it("deletes object", async () => {
    const result = await service.deleteObject("my-bucket", "file.txt");
    // BUG: result could be anything -- even an error object.
    // toBeDefined passes for error objects too.
    expect(result).toBeDefined();
  });

  it("handles multipart upload", async () => {
    const init = await service.initiateMultipartUpload(
      "my-bucket", "big.zip", "application/zip"
    );
    // BUG: uploadId could be empty string, wrong type, etc.
    expect(init).toBeDefined();
    expect(init.uploadId).toBeTruthy();
  });
});
```

- **Expected violation:** `TQ-no-shallow-assertions` at every `expect().toBeDefined()` and `expect().toBeTruthy()` call. These assertions verify existence but nothing about value correctness. Upload result should check `etag` matches the `^"[a-f0-9]{32}"$` format. Download should verify `contentLength` is a number, `contentType` is a string.
- **Production impact:** A regression that changes ETag format from `"abc123..."` to `abc123...` (missing quotes) passes all tests. The application then fails at runtime when comparing ETags for cache invalidation because the format changed silently.

---

## B04 -- Missing Negative Tests

**Bug:** Only happy-path tests. No tests for 403, 404, invalid bucket names, expired presigned URLs, or multipart failures.

```typescript
// tests/s3-client.test.ts — B04: No negative tests

describe("S3Service", () => {
  it("uploads object", async () => {
    s3Mock.on(PutObjectCommand).resolves({
      ETag: '"d41d8cd98f00b204e9800998ecf8427e"',
      $metadata: { requestId: "REQ-001" },
    });
    const result = await service.uploadObject(
      "my-bucket", "file.txt", Buffer.from("data"), "text/plain"
    );
    expect(result.etag).toBe('"d41d8cd98f00b204e9800998ecf8427e"');
  });

  it("downloads object", async () => {
    s3Mock.on(GetObjectCommand).resolves({
      Body: {} as ReadableStream,
      ContentType: "text/plain",
      ContentLength: 100,
      ETag: '"aabbccdd11223344aabbccdd11223344"',
    });
    const result = await service.downloadObject("my-bucket", "file.txt");
    expect(result.contentType).toBe("text/plain");
    expect(result.contentLength).toBe(100);
  });

  it("generates presigned URL", async () => {
    const result = await service.generatePresignedUrl("my-bucket", "key", "GET", 3600);
    expect(result.url).toMatch(/^https:\/\//);
  });

  // BUG: No tests for:
  // - 403 AccessDenied (wrong IAM policy)
  // - 404 NoSuchKey (download nonexistent object)
  // - 404 NoSuchBucket (bucket doesn't exist)
  // - Invalid bucket name (uppercase, underscores)
  // - Expired presigned URL (what does consumer do?)
  // - Content-Length overflow (>5TB)
  // - Presign expires out of range (>604800)
  // - Multipart upload abort on failure
  // - Network timeout / S3 503 Service Unavailable
  // - Missing credentials for presign
});
```

- **Expected violation:** `TQ-negative-cases` -- The manifest declares 8+ error status codes (400, 403, 404, 409, 412, 500, 503) across the three endpoints. Zero negative test cases exist. No error path is exercised.
- **Production impact:** A deployment to a new AWS account with different IAM policies returns 403 on all operations. Since the 403 path was never tested, the application shows a misleading "file not found" error to users instead of "permission denied", and the on-call team spends hours debugging the wrong problem.

---

## B05 -- Request Missing Required Fields

**Bug:** PUT request omits the Content-Type header. S3 defaults to `application/octet-stream`, causing downstream consumers to misinterpret file types.

```typescript
// src/services/s3-client.ts — B05: Missing required Content-Type header

export class S3Service {
  async uploadObject(bucket: string, key: string, body: Buffer) {
    // BUG: No contentType parameter. PutObjectCommand is created without
    // ContentType. S3 defaults to "application/octet-stream" for all uploads.
    const command = new PutObjectCommand({
      Bucket: bucket,
      Key: key,
      Body: body,
      // ContentType is MISSING — required by manifest
    });

    try {
      const response = await this.client.send(command);
      return { etag: response.ETag ?? "" };
    } catch (error) {
      throw new Error(`Upload failed: ${(error as Error).message}`);
    }
  }

  async uploadImage(bucket: string, key: string, imageBuffer: Buffer) {
    // BUG: Calls uploadObject without specifying content type.
    // A JPEG uploaded this way is stored as application/octet-stream.
    return this.uploadObject(bucket, key, imageBuffer);
  }
}
```

- **Expected violation:** `CTR-request-shape` -- The manifest declares `Content-Type: { type: string, required: true }` for PUT `/:bucket/:key`. The client never sets this header. The field is absent from the PutObjectCommand parameters.
- **Production impact:** All uploaded files are stored as `application/octet-stream`. When a browser downloads them via presigned URL, it triggers a download dialog instead of rendering images inline. A CDN caching layer may serve incorrect `Content-Type` headers to thousands of users.

---

## B06 -- Response Type Mismatch

**Bug:** Client type definition for upload response is missing the `ETag` field from the manifest and adds an invented `url` field that S3 never returns.

```typescript
// src/services/s3-client.ts — B06: Response type doesn't match S3 API

// BUG: This type does not match the manifest's response shape.
// Missing: etag (required by manifest)
// Extra: url (S3 PUT does not return a url field)
// Wrong: requestId should come from $metadata, not top-level
interface S3UploadResult {
  url: string;              // EXTRA: S3 PUT does not return this
  versionId: string;        // WRONG: should be string | null (nullable per manifest)
  requestId: string;
  // MISSING: etag — required by manifest, needed for cache validation
}

// BUG: This type omits contentLength from S3 GET response.
interface S3DownloadResult {
  body: ReadableStream;
  contentType: string;
  etag: string;
  // MISSING: contentLength — required by manifest, needed for progress bars
}

export class S3Service {
  async uploadObject(bucket: string, key: string, body: Buffer, contentType: string): Promise<S3UploadResult> {
    try {
      const command = new PutObjectCommand({
        Bucket: bucket, Key: key, Body: body, ContentType: contentType,
      });
      const response = await this.client.send(command);

      return {
        url: `https://${bucket}.s3.amazonaws.com/${key}`,  // Fabricated, not from response
        versionId: response.VersionId ?? "",                // Empty string instead of null
        requestId: response.$metadata.requestId ?? "",
        // etag is not returned despite being required by manifest
      };
    } catch (error) {
      throw new Error(`Upload failed: ${(error as Error).message}`);
    }
  }

  async downloadObject(bucket: string, key: string): Promise<S3DownloadResult> {
    try {
      const command = new GetObjectCommand({ Bucket: bucket, Key: key });
      const response = await this.client.send(command);

      return {
        body: response.Body as unknown as ReadableStream,
        contentType: response.ContentType ?? "",
        etag: response.ETag ?? "",
        // contentLength is not returned despite being required by manifest
      };
    } catch (error) {
      throw new Error(`Download failed: ${(error as Error).message}`);
    }
  }
}
```

- **Expected violation:** `CTR-response-shape` --
  - `S3UploadResult`: Missing required field `etag`. Extra field `url` not in manifest response. Field `versionId` declared as `string` but manifest says `nullable: true` (should be `string | null`).
  - `S3DownloadResult`: Missing required field `contentLength` from manifest response.
- **Production impact:** Downstream code that depends on `result.etag` for cache invalidation (`If-None-Match` headers) fails with `undefined`. The fabricated `url` field uses a hardcoded URL scheme that breaks for buckets in non-US regions (e.g., `s3.eu-west-1.amazonaws.com`). Progress bars that read `contentLength` show `NaN%`.

---

## B07 -- Wrong Field Types

**Bug:** Content-Length is stored as a string after reading from response headers. ETag is stored without quotes (stripping the required format).

```typescript
// src/services/s3-client.ts — B07: Wrong field types

interface S3DownloadResult {
  body: ReadableStream;
  contentType: string;
  contentLength: string;    // BUG: manifest says integer, stored as string
  etag: string;
  versionId: string | null;
}

export class S3Service {
  async downloadObject(bucket: string, key: string): Promise<S3DownloadResult> {
    const url = `https://${bucket}.s3.amazonaws.com/${encodeURIComponent(key)}`;
    const headers = await this.signer.sign("GET", url);

    try {
      const response = await fetch(url, { headers });
      if (!response.ok) {
        throw new Error(`S3 GET failed with status ${response.status}`);
      }

      return {
        body: response.body!,
        contentType: response.headers.get("Content-Type") ?? "",
        // BUG: Content-Length from headers is always a string.
        // Manifest requires integer. Arithmetic breaks: "4096" + 1024 = "40961024"
        contentLength: response.headers.get("Content-Length") ?? "0",
        // BUG: Stripping quotes from ETag. Manifest format requires quotes.
        // "d41d8cd98..." becomes d41d8cd98... (no longer matches ^"[a-f0-9]{32}"$)
        etag: (response.headers.get("ETag") ?? "").replace(/"/g, ""),
        versionId: response.headers.get("x-amz-version-id") ?? null,
      };
    } catch (error) {
      throw new Error(`Download failed: ${(error as Error).message}`);
    }
  }
}
```

```typescript
// src/routes/download.ts — B07: Downstream code breaks on wrong types

import { s3Service } from "../services/s3-client";

export async function handleDownload(req: Request): Promise<Response> {
  const result = await s3Service.downloadObject("my-bucket", req.url);

  // BUG: contentLength is string. This calculates string concatenation, not addition.
  // remainingBytes = "4096" + 0 = "40960" (string), not 4096 (number).
  const remainingBytes = result.contentLength + 0;

  // BUG: etag has no quotes. If-None-Match requires quoted etag per HTTP spec.
  // Sends: If-None-Match: d41d8cd98... instead of "d41d8cd98..."
  const cacheHeaders = { "If-None-Match": result.etag };

  return new Response(result.body, {
    headers: {
      "Content-Length": String(remainingBytes),
      "Cache-Control": "max-age=3600",
    },
  });
}
```

- **Expected violation:** `CTR-manifest-conformance` --
  - Field `contentLength`: Manifest declares `type: integer`. Client stores as `string`. Type mismatch.
  - Field `etag`: Manifest declares `format: ^"[a-f0-9]{32}"$` (with quotes). Client strips quotes, violating the format constraint.
- **Production impact:** A download progress bar computes `downloaded / total * 100` as `1024 / "4096" * 100` which is `NaN` in strict contexts or produces incorrect results. ETags without quotes fail HTTP caching: the `If-None-Match` header sends an unquoted value that S3 never matches, causing every request to re-download the full object instead of returning 304.

---

## B08 -- Incomplete Enum Handling

**Bug:** Storage class handling covers only "STANDARD" but ignores "GLACIER", "DEEP_ARCHIVE", and "INTELLIGENT_TIERING". Objects in archive tiers trigger restore workflows that this code does not handle.

```typescript
// src/services/s3-client.ts — B08: Incomplete storage class enum

type StorageClass = "STANDARD" | "REDUCED_REDUNDANCY" | "STANDARD_IA";
// BUG: Missing "ONEZONE_IA", "INTELLIGENT_TIERING", "GLACIER", "DEEP_ARCHIVE"
// Manifest declares all 7. Client handles only 3.

export class S3Service {
  async downloadObject(bucket: string, key: string) {
    try {
      const command = new GetObjectCommand({ Bucket: bucket, Key: key });
      const response = await this.client.send(command);

      const storageClass = response.StorageClass as StorageClass;

      // BUG: Only handles 3 of 7 storage classes. When storageClass is
      // "GLACIER" or "DEEP_ARCHIVE", this switch falls through to default
      // and treats it as available for immediate download.
      switch (storageClass) {
        case "STANDARD":
        case "REDUCED_REDUNDANCY":
        case "STANDARD_IA":
          return { body: response.Body, available: true };
        default:
          // BUG: GLACIER objects are NOT immediately available.
          // They require a RestoreObject call and hours/days of wait time.
          // Treating them as available causes download to fail silently.
          return { body: response.Body, available: true };
      }
    } catch (error) {
      throw new Error(`Download failed: ${(error as Error).message}`);
    }
  }

  async setStorageClass(bucket: string, key: string, storageClass: StorageClass) {
    // BUG: Type only accepts 3 values. Cannot transition to GLACIER or
    // DEEP_ARCHIVE, even though the manifest and S3 API support it.
    const command = new PutObjectCommand({
      Bucket: bucket,
      Key: key,
      StorageClass: storageClass,
    });
    await this.client.send(command);
  }
}
```

- **Expected violation:** `CTR-strictness-parity` -- Manifest enum for `x-amz-storage-class` has 7 values: `["STANDARD", "REDUCED_REDUNDANCY", "STANDARD_IA", "ONEZONE_IA", "INTELLIGENT_TIERING", "GLACIER", "DEEP_ARCHIVE"]`. Client type only handles 3. Missing: `ONEZONE_IA`, `INTELLIGENT_TIERING`, `GLACIER`, `DEEP_ARCHIVE`.
- **Production impact:** An object transitioned to GLACIER via lifecycle policy returns `StorageClass: "GLACIER"` in the response. The client marks it as `available: true` and attempts to stream the body. S3 returns a `403 InvalidObjectState` error because GLACIER objects require a restore request (1-12 hours for Standard retrieval). Users see an opaque error with no guidance to initiate a restore.

---

## B09 -- Missing Range Validation

**Bug:** Presigned URL expiration accepts any number. Values above 604800 (7 days) are silently accepted but S3 rejects them at request time.

```typescript
// src/services/s3-client.ts — B09: No range validation on presign expires

export class S3Service {
  async generatePresignedUrl(
    bucket: string,
    key: string,
    method: "GET" | "PUT",
    expiresInSeconds: number
  ): Promise<{ url: string; expiresAt: string }> {
    // BUG: No range validation. Accepts negative values, zero, and values
    // above 604800 (S3 max is 7 days / 604800 seconds).
    // S3 rejects presigned URLs with expires > 604800 at request time with:
    // <Error><Code>AuthorizationQueryParametersError</Code></Error>

    const command = method === "GET"
      ? new GetObjectCommand({ Bucket: bucket, Key: key })
      : new PutObjectCommand({ Bucket: bucket, Key: key });

    // This generates a URL with X-Amz-Expires=1000000 which S3 rejects.
    const url = await getSignedUrl(this.client, command, { expiresIn: expiresInSeconds });
    const expiresAt = new Date(Date.now() + expiresInSeconds * 1000).toISOString();

    return { url, expiresAt };
  }
}
```

```typescript
// src/routes/share.ts — B09: Caller passes invalid expiration

import { s3Service } from "../services/s3-client";

export async function createShareLink(fileKey: string): Promise<string> {
  // BUG: 30 days = 2,592,000 seconds. Max is 604,800 (7 days).
  // URL is generated successfully (signing is client-side) but S3 rejects
  // the request when the user clicks the link.
  const { url } = await s3Service.generatePresignedUrl(
    "my-bucket", fileKey, "GET", 30 * 24 * 60 * 60
  );
  return url;
}
```

- **Expected violation:** `CTR-strictness-parity` -- Manifest declares `expires: { type: integer, range: [1, 604800] }`. Client performs no range check. Values of 0, -1, and 2592000 are all accepted without error.
- **Production impact:** A "share file" feature generates presigned URLs valid for 30 days. The URL is emailed to users. When they click it, S3 responds with an XML error: `AuthorizationQueryParametersError — X-Amz-Expires must be less than 604800`. The share link is permanently broken. Support receives complaints from users who cannot access their shared files.

---

## B10 -- Format Not Validated

**Bug:** Bucket name is not validated for DNS compatibility. Uppercase letters, underscores, and invalid lengths are accepted.

```typescript
// src/services/s3-client.ts — B10: No bucket name format validation

export class S3Service {
  // BUG: No validation at all on bucket name. Accepts:
  // - "My_Bucket" (uppercase + underscore — DNS-incompatible)
  // - "ab" (too short — min 3 chars)
  // - "a".repeat(100) (too long — max 63 chars)
  // - "bucket..name" (consecutive dots — invalid)
  // - "-bucket" (leading hyphen — invalid)

  async uploadObject(bucket: string, key: string, body: Buffer, contentType: string) {
    // No bucket name validation — goes straight to S3 API.
    // S3 returns different errors depending on the violation:
    // - Uppercase/underscores: DNS resolution fails (NXDOMAIN)
    // - Too short/long: 400 InvalidBucketName
    // - Leading hyphen: 400 InvalidBucketName

    const command = new PutObjectCommand({
      Bucket: bucket,
      Key: key,
      Body: body,
      ContentType: contentType,
    });

    try {
      const response = await this.client.send(command);
      return { etag: response.ETag ?? "" };
    } catch (error) {
      throw new Error(`Upload failed: ${(error as Error).message}`);
    }
  }

  async downloadObject(bucket: string, key: string) {
    // Same bug: no bucket validation.
    const command = new GetObjectCommand({ Bucket: bucket, Key: key });
    try {
      const response = await this.client.send(command);
      return { body: response.Body };
    } catch (error) {
      throw new Error(`Download failed: ${(error as Error).message}`);
    }
  }
}
```

```typescript
// tests/s3-client.test.ts — B10: Tests use invalid bucket names without noticing

describe("S3Service", () => {
  it("uploads to bucket", async () => {
    // BUG: "Test_Bucket" has uppercase and underscore.
    // Test passes because mock doesn't validate bucket names.
    // Production S3 would fail with DNS resolution error.
    const result = await service.uploadObject(
      "Test_Bucket", "file.txt", Buffer.from("data"), "text/plain"
    );
    expect(result.etag).toBeDefined();
  });
});
```

- **Expected violation:** `CTR-strictness-parity` -- Manifest declares `bucket: { format: "^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$", length: [3, 63] }`. Client performs no format or length validation on bucket names.
- **Production impact:** A configuration file has `S3_BUCKET=My_Upload_Bucket` (set by a developer who created the bucket before the 2018 naming rules). The SDK attempts virtual-hosted-style access: `https://My_Upload_Bucket.s3.amazonaws.com/...`. DNS resolution fails because uppercase characters are not valid in DNS hostnames. The error message is `getaddrinfo ENOTFOUND my_upload_bucket.s3.amazonaws.com` -- completely obscure, with no mention of the bucket naming rules.

---

## B11 -- Precision Loss

**Bug:** Content-Length for large files is stored as a 32-bit integer, overflowing for files larger than 2GB (2,147,483,647 bytes).

```typescript
// src/services/s3-client.ts — B11: Content-Length precision loss

interface UploadProgress {
  loaded: number;
  total: number;     // BUG: Stored in a context where 32-bit truncation occurs
  percentage: number;
}

export class S3Service {
  async uploadLargeFile(
    bucket: string,
    key: string,
    filePath: string,
    contentType: string,
    onProgress?: (progress: UploadProgress) => void
  ) {
    const stats = await fs.stat(filePath);

    // BUG: Using bitwise OR to "convert to integer" truncates to 32 bits.
    // For a 3GB file (3,221,225,472 bytes):
    //   3221225472 | 0 = -1073741824 (32-bit signed overflow!)
    // For a 5TB file (5,497,558,138,880 bytes):
    //   5497558138880 | 0 = 1202590592 (truncated to 32 bits)
    const totalBytes = stats.size | 0;

    if (totalBytes <= 0) {
      throw new Error("File is empty or size calculation failed.");
    }

    // BUG: totalBytes is wrong for files > 2GB.
    // A 3GB file shows totalBytes = -1073741824, which fails the > 0 check.
    // A 4.5GB file shows totalBytes = 205521408, which passes but is wrong.

    const command = new PutObjectCommand({
      Bucket: bucket,
      Key: key,
      Body: createReadStream(filePath),
      ContentType: contentType,
      ContentLength: totalBytes,  // BUG: Truncated value sent to S3
    });

    try {
      const response = await this.client.send(command);
      return { etag: response.ETag ?? "", size: totalBytes };
    } catch (error) {
      throw new Error(`Upload failed: ${(error as Error).message}`);
    }
  }
}
```

```typescript
// tests/s3-client.test.ts — B11: Tests only use small files

describe("uploadLargeFile", () => {
  it("uploads file with progress", async () => {
    // BUG: Test uses 1KB file. Precision loss only manifests above 2GB.
    // This test passes but masks the 32-bit overflow bug.
    const result = await service.uploadLargeFile(
      "my-bucket", "small.txt", "/tmp/1kb-file.txt", "text/plain"
    );
    expect(result.size).toBe(1024);
  });

  // No test with file > 2GB
  // No test with file > 4GB (where 32-bit wrap-around produces positive but wrong value)
  // No test with max size (5TB)
});
```

- **Expected violation:** `CTR-strictness-parity` -- Manifest declares `Content-Length: { type: integer, range: [0, 5497558138880] }`. The bitwise `| 0` operation truncates values above 2,147,483,647 (2^31 - 1) to 32-bit signed integers. The effective range becomes [0, 2147483647], which is a fraction of the declared range.
- **Production impact:** A user uploads a 3GB video file. The `Content-Length` header is sent as `-1073741824` (negative, due to signed 32-bit overflow). S3 rejects the request with `400 Bad Request: Content-Length must be non-negative`. If the file is 4.5GB, `Content-Length` wraps to `205521408` (200MB). S3 accepts the header, reads 200MB, then closes the connection. The upload appears successful but the file is silently truncated, missing 4.3GB of data.

---

## B12 -- Nullable Field Crash

**Bug:** Code accesses `response.VersionId` properties without checking for `null`/`undefined`. Crashes when versioning is not enabled on the bucket.

```typescript
// src/services/s3-client.ts — B12: VersionId accessed without null check

interface VersionInfo {
  id: string;
  timestamp: Date;
  isLatest: boolean;
}

export class S3Service {
  async uploadObject(bucket: string, key: string, body: Buffer, contentType: string) {
    try {
      const command = new PutObjectCommand({
        Bucket: bucket, Key: key, Body: body, ContentType: contentType,
      });
      const response = await this.client.send(command);

      // BUG: VersionId is undefined when bucket versioning is not enabled.
      // Manifest declares: x-amz-version-id: { required: false, nullable: true }
      // This code assumes it always exists.
      const versionInfo: VersionInfo = {
        id: response.VersionId!,                        // Undefined when unversioned
        timestamp: new Date(response.VersionId!.split(".")[1]),  // Crashes: Cannot read property 'split' of undefined
        isLatest: true,
      };

      return {
        etag: response.ETag ?? "",
        version: versionInfo,
      };
    } catch (error) {
      throw new Error(`Upload failed: ${(error as Error).message}`);
    }
  }

  async getObjectVersion(bucket: string, key: string) {
    try {
      const command = new GetObjectCommand({ Bucket: bucket, Key: key });
      const response = await this.client.send(command);

      // BUG: Same issue. VersionId is undefined on unversioned buckets.
      // .length on undefined throws TypeError.
      const versionLength = response.VersionId!.length;

      return {
        body: response.Body,
        versionId: response.VersionId!,
        versionIdLength: versionLength,
      };
    } catch (error) {
      throw new Error(`Get failed: ${(error as Error).message}`);
    }
  }
}
```

```typescript
// tests/s3-client.test.ts — B12: Tests always mock VersionId as present

describe("S3Service", () => {
  it("uploads with version info", async () => {
    s3Mock.on(PutObjectCommand).resolves({
      ETag: '"d41d8cd98f00b204e9800998ecf8427e"',
      // BUG: Test always provides VersionId. Never tests the unversioned case.
      VersionId: "v1.1707000000",
      $metadata: { requestId: "REQ-001" },
    });

    const result = await service.uploadObject(
      "my-bucket", "file.txt", Buffer.from("data"), "text/plain"
    );
    expect(result.version.id).toBe("v1.1707000000");
    // This test passes, masking the crash that occurs on unversioned buckets.
  });

  // Missing: test with VersionId = undefined (unversioned bucket)
});
```

- **Expected violation:** `CTR-response-shape` -- Manifest declares `x-amz-version-id: { required: false, nullable: true }`. Client code uses non-null assertion (`!`) on the field and accesses `.split()` and `.length` on it, which crashes when the field is `undefined` (unversioned bucket).
- **Production impact:** The application works in the staging environment where bucket versioning is enabled. In production, versioning is disabled (to reduce costs). Every upload crashes with `TypeError: Cannot read properties of undefined (reading 'split')`. The error is caught by the generic catch block and re-thrown as "Upload failed: Cannot read properties of undefined", giving no indication that the issue is the missing VersionId field.

---

## B13 -- Missing Signature Validation

**Bug:** Presigned URL generation proceeds without verifying that AWS credentials are configured. Produces a URL that fails at request time.

```typescript
// src/services/s3-client.ts — B13: No credential check before presigning

export class S3Service {
  private client: S3Client;

  constructor(region: string) {
    // BUG: S3Client created without explicit credentials.
    // Falls back to environment variables / instance profile.
    // If neither is configured, client.send() fails, but getSignedUrl()
    // may succeed with empty credentials — producing an invalid URL.
    this.client = new S3Client({ region });
  }

  async generatePresignedUrl(
    bucket: string,
    key: string,
    method: "GET" | "PUT",
    expiresInSeconds: number
  ): Promise<{ url: string; expiresAt: string }> {
    // BUG: No check that credentials are actually configured.
    // getSignedUrl() computes the signature using whatever credentials
    // the SDK resolves. If no credentials are found:
    // - In SDK v3, it may throw CredentialsProviderError
    // - Or it may silently use empty credentials, producing a URL whose
    //   signature is invalid. The URL looks valid but S3 returns 403.

    const command = method === "GET"
      ? new GetObjectCommand({ Bucket: bucket, Key: key })
      : new PutObjectCommand({ Bucket: bucket, Key: key });

    const url = await getSignedUrl(this.client, command, { expiresIn: expiresInSeconds });
    const expiresAt = new Date(Date.now() + expiresInSeconds * 1000).toISOString();

    return { url, expiresAt };
  }
}
```

```typescript
// tests/s3-client.test.ts — B13: Tests never simulate missing credentials

describe("generatePresignedUrl", () => {
  it("generates presigned URL", async () => {
    // BUG: Test environment always has AWS_ACCESS_KEY_ID set.
    // Never tests what happens when credentials are missing.
    const result = await service.generatePresignedUrl("my-bucket", "key", "GET", 3600);
    expect(result.url).toMatch(/^https:\/\//);
  });

  // Missing: test with no credentials configured
  // Missing: test with expired temporary credentials
  // Missing: test with credentials that lack s3:GetObject permission
});
```

- **Expected violation:** `CTR-request-shape` -- The presigned URL generation requires valid AWS credentials (accessKeyId, secretAccessKey, and optionally sessionToken) to compute the HMAC-SHA256 signature. The client does not validate that credentials are present before calling `getSignedUrl()`. The manifest's presign contract requires `bucket`, `key`, `expires`, and `method`, but the implicit requirement of valid credentials is not enforced.
- **Production impact:** A new microservice is deployed to Kubernetes without the `AWS_ACCESS_KEY_ID` environment variable. The presigned URL endpoint returns URLs that look valid (they contain `X-Amz-Signature=...`), but the signature is computed with empty credentials. When users click the presigned link, S3 returns `403 SignatureDoesNotMatch`. The error happens minutes or hours after URL generation, making it extremely difficult to trace back to the missing credentials.

---

## B14 -- Multipart Upload Incomplete

**Bug:** Initiates multipart upload and uploads parts, but never calls `CompleteMultipartUpload`. Incomplete uploads consume S3 storage indefinitely.

```typescript
// src/services/s3-client.ts — B14: Multipart upload never completed

const PART_SIZE = 5 * 1024 * 1024; // 5 MB

export class S3Service {
  async uploadLargeFile(
    bucket: string,
    key: string,
    fileBuffer: Buffer,
    contentType: string
  ): Promise<{ uploadId: string; partsUploaded: number }> {
    try {
      // Step 1: Initiate multipart upload
      const initCommand = new CreateMultipartUploadCommand({
        Bucket: bucket,
        Key: key,
        ContentType: contentType,
      });
      const initResponse = await this.client.send(initCommand);
      const uploadId = initResponse.UploadId!;

      // Step 2: Upload parts
      const partCount = Math.ceil(fileBuffer.length / PART_SIZE);
      const parts: Array<{ PartNumber: number; ETag: string }> = [];

      for (let i = 0; i < partCount; i++) {
        const start = i * PART_SIZE;
        const end = Math.min(start + PART_SIZE, fileBuffer.length);
        const partBody = fileBuffer.subarray(start, end);

        const uploadPartCommand = new UploadPartCommand({
          Bucket: bucket,
          Key: key,
          UploadId: uploadId,
          PartNumber: i + 1,
          Body: partBody,
          ContentLength: partBody.length,
        });

        const partResponse = await this.client.send(uploadPartCommand);
        parts.push({ PartNumber: i + 1, ETag: partResponse.ETag! });
      }

      // BUG: Step 3 (CompleteMultipartUpload) is MISSING.
      // The parts are uploaded but the object is never assembled.
      // S3 keeps the parts in storage, consuming space and cost.
      // The object does not appear in bucket listings.
      // No AbortMultipartUpload is called on error either.

      return { uploadId, partsUploaded: parts.length };
      // Should call CompleteMultipartUploadCommand here with the parts array
    } catch (error) {
      // BUG: On error, no AbortMultipartUpload is called.
      // The incomplete upload remains, consuming storage.
      throw new Error(`Upload failed: ${(error as Error).message}`);
    }
  }
}
```

```typescript
// tests/s3-client.test.ts — B14: Test checks part count but not completion

describe("uploadLargeFile", () => {
  it("uploads all parts", async () => {
    s3Mock.on(CreateMultipartUploadCommand).resolves({ UploadId: "mpu-001" });
    s3Mock.on(UploadPartCommand).resolves({
      ETag: '"part-etag-12345678901234567890"',
    });
    // Note: No mock for CompleteMultipartUploadCommand — it's never called.

    const result = await service.uploadLargeFile(
      "my-bucket", "big-file.zip",
      Buffer.alloc(15 * 1024 * 1024), // 15 MB = 3 parts
      "application/zip"
    );

    expect(result.partsUploaded).toBe(3);
    expect(result.uploadId).toBe("mpu-001");
    // BUG: Test passes because it only checks parts uploaded, not completion.
    // The object doesn't actually exist in S3 after this "successful" upload.
  });

  // Missing: verify CompleteMultipartUpload was called
  // Missing: verify object is accessible after upload
  // Missing: test error handling with AbortMultipartUpload
});
```

- **Expected violation:** `CTR-response-shape` -- The manifest's `aws-s3-multipart` contract defines three endpoints: InitiateMultipartUpload, UploadPart, and CompleteMultipartUpload. The client calls the first two but never calls CompleteMultipartUpload. The multipart lifecycle is incomplete: the response contract for the complete operation (`etag`, `location`) is never fulfilled.
- **Production impact:** Every large file upload (>5MB) creates an incomplete multipart upload in S3. The file never appears in the bucket -- users report "upload succeeded but file is missing". Meanwhile, the incomplete parts accumulate and incur S3 storage charges. Without a lifecycle policy to clean up incomplete multipart uploads, storage costs grow unbounded. A company discovers months later that they have terabytes of orphaned multipart upload parts costing thousands of dollars per month.

---

## B15 -- Race Condition

**Bug:** Check-then-upload pattern without conditional write headers. Concurrent uploads can overwrite each other without detection.

```typescript
// src/services/s3-client.ts — B15: Race condition on conditional upload

export class S3Service {
  async safeUpload(
    bucket: string,
    key: string,
    body: Buffer,
    contentType: string
  ): Promise<{ etag: string; action: "created" | "skipped" }> {
    try {
      // Step 1: Check if object already exists
      const headCommand = new GetObjectCommand({ Bucket: bucket, Key: key });
      try {
        await this.client.send(headCommand);
        // Object exists — skip upload
        return { etag: "", action: "skipped" };
      } catch (error: unknown) {
        if ((error as Error).name !== "NoSuchKey") {
          throw error;
        }
        // Object does not exist — proceed to upload
      }

      // BUG: Race condition window between HEAD check and PUT.
      // Another process can upload the same key between these two calls.
      // This upload will silently overwrite the other process's data.
      //
      // CORRECT approach: Use If-None-Match: "*" header on PUT to make
      // S3 reject the upload with 412 Precondition Failed if the object
      // already exists. This is an atomic check-and-write.

      // Step 2: Upload (no conditional header)
      const putCommand = new PutObjectCommand({
        Bucket: bucket,
        Key: key,
        Body: body,
        ContentType: contentType,
        // MISSING: IfNoneMatch: "*"
        // Without this, concurrent uploads can overwrite each other.
      });

      const response = await this.client.send(putCommand);
      return { etag: response.ETag ?? "", action: "created" };
    } catch (error) {
      throw new Error(`Safe upload failed: ${(error as Error).message}`);
    }
  }

  async updateObject(
    bucket: string,
    key: string,
    body: Buffer,
    contentType: string,
    _expectedETag: string  // Parameter accepted but never used
  ): Promise<{ etag: string }> {
    // BUG: expectedETag is accepted as a parameter but never used.
    // Should be sent as If-Match header to prevent overwriting concurrent changes.
    // Without If-Match, this is a blind overwrite.

    const putCommand = new PutObjectCommand({
      Bucket: bucket,
      Key: key,
      Body: body,
      ContentType: contentType,
      // MISSING: IfMatch: expectedETag
    });

    try {
      const response = await this.client.send(putCommand);
      return { etag: response.ETag ?? "" };
    } catch (error) {
      throw new Error(`Update failed: ${(error as Error).message}`);
    }
  }
}
```

```typescript
// tests/s3-client.test.ts — B15: No concurrency tests

describe("safeUpload", () => {
  it("creates object when it does not exist", async () => {
    const notFoundError = new Error("Not found");
    notFoundError.name = "NoSuchKey";
    s3Mock.on(GetObjectCommand).rejects(notFoundError);
    s3Mock.on(PutObjectCommand).resolves({
      ETag: '"d41d8cd98f00b204e9800998ecf8427e"',
    });

    const result = await service.safeUpload(
      "my-bucket", "new-file.txt", Buffer.from("data"), "text/plain"
    );
    expect(result.action).toBe("created");
    expect(result.etag).toBe('"d41d8cd98f00b204e9800998ecf8427e"');
  });

  it("skips when object already exists", async () => {
    s3Mock.on(GetObjectCommand).resolves({
      Body: {} as ReadableStream,
    });

    const result = await service.safeUpload(
      "my-bucket", "existing.txt", Buffer.from("data"), "text/plain"
    );
    expect(result.action).toBe("skipped");
  });

  // BUG: No test for concurrent uploads.
  // Two safeUpload calls for the same key should not both succeed with "created".
  // Without If-None-Match, both HEAD checks return 404, both PUTs succeed,
  // and the second overwrites the first without detection.

  // Missing: test that verifies If-None-Match: "*" is sent on PUT
  // Missing: test for 412 Precondition Failed handling
  // Missing: test for updateObject with If-Match / ETag checking
});
```

- **Expected violation:** `CTR-request-shape` -- The manifest declares `If-None-Match: { type: string, required: false }` as a request header for PUT. The `safeUpload` method implements a check-then-write pattern but does not use the `If-None-Match` header for atomic conditional writes. The `updateObject` method accepts an `expectedETag` parameter but never sends it as an `If-Match` header, making the conditional write claim a lie.
- **Production impact:** Two users upload profile photos simultaneously. Both HEAD checks return 404 (object does not exist). Both PUTs succeed. User A's photo is written first, then User B's photo silently overwrites it. User A sees User B's photo as their own profile picture. In a document management system, this causes silent data loss: a legal contract uploaded by one attorney is overwritten by another attorney's version, with no audit trail or conflict detection.
