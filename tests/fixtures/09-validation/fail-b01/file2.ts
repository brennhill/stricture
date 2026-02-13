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
