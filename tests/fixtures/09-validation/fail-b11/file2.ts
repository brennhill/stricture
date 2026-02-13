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
