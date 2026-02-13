// tests/s3-client.test.ts — B12: Tests always mock VersionId as present

describe("S3Service", () => {
  it("uploads with version info", async () => {
    s3Mock.on(PutObjectCommand).resolves({
      ETag: '"d41d8cd98f00b204e9800998ecf8427e"',
      // BUG: Test always provides VersionId. Never tests the unversioned case.
      VersionId: "v1.1707000000",
      $metadata: { requestId: "REQ-001" },
    });

    const result = await service.uploadObject(
      "my-bucket", "file.txt", Buffer.from("data"), "text/plain"
    );
    expect(result.version.id).toBe("v1.1707000000");
    // This test passes, masking the crash that occurs on unversioned buckets.
  });

  // Missing: test with VersionId = undefined (unversioned bucket)
});
