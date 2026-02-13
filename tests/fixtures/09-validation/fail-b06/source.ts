// src/services/s3-client.ts — B06: Response type doesn't match S3 API

// BUG: This type does not match the manifest's response shape.
// Missing: etag (required by manifest)
// Extra: url (S3 PUT does not return a url field)
// Wrong: requestId should come from $metadata, not top-level
interface S3UploadResult {
  url: string;              // EXTRA: S3 PUT does not return this
  versionId: string;        // WRONG: should be string | null (nullable per manifest)
  requestId: string;
  // MISSING: etag — required by manifest, needed for cache validation
}

// BUG: This type omits contentLength from S3 GET response.
interface S3DownloadResult {
  body: ReadableStream;
  contentType: string;
  etag: string;
  // MISSING: contentLength — required by manifest, needed for progress bars
}

export class S3Service {
  async uploadObject(bucket: string, key: string, body: Buffer, contentType: string): Promise<S3UploadResult> {
    try {
      const command = new PutObjectCommand({
        Bucket: bucket, Key: key, Body: body, ContentType: contentType,
      });
      const response = await this.client.send(command);

      return {
        url: `https://${bucket}.s3.amazonaws.com/${key}`,  // Fabricated, not from response
        versionId: response.VersionId ?? "",                // Empty string instead of null
        requestId: response.$metadata.requestId ?? "",
        // etag is not returned despite being required by manifest
      };
    } catch (error) {
      throw new Error(`Upload failed: ${(error as Error).message}`);
    }
  }

  async downloadObject(bucket: string, key: string): Promise<S3DownloadResult> {
    try {
      const command = new GetObjectCommand({ Bucket: bucket, Key: key });
      const response = await this.client.send(command);

      return {
        body: response.Body as unknown as ReadableStream,
        contentType: response.ContentType ?? "",
        etag: response.ETag ?? "",
        // contentLength is not returned despite being required by manifest
      };
    } catch (error) {
      throw new Error(`Download failed: ${(error as Error).message}`);
    }
  }
}
