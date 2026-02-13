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
