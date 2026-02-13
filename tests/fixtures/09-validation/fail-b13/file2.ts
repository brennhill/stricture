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
