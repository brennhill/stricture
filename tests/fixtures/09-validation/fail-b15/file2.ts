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
