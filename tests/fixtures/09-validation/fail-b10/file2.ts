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
