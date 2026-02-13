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
