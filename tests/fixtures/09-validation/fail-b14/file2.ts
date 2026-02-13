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
