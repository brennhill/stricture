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
