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
