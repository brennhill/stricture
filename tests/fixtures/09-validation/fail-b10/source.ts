// src/services/s3-client.ts — B10: No bucket name format validation

export class S3Service {
  // BUG: No validation at all on bucket name. Accepts:
  // - "My_Bucket" (uppercase + underscore — DNS-incompatible)
  // - "ab" (too short — min 3 chars)
  // - "a".repeat(100) (too long — max 63 chars)
  // - "bucket..name" (consecutive dots — invalid)
  // - "-bucket" (leading hyphen — invalid)

  async uploadObject(bucket: string, key: string, body: Buffer, contentType: string) {
    // No bucket name validation — goes straight to S3 API.
    // S3 returns different errors depending on the violation:
    // - Uppercase/underscores: DNS resolution fails (NXDOMAIN)
    // - Too short/long: 400 InvalidBucketName
    // - Leading hyphen: 400 InvalidBucketName

    const command = new PutObjectCommand({
      Bucket: bucket,
      Key: key,
      Body: body,
      ContentType: contentType,
    });

    try {
      const response = await this.client.send(command);
      return { etag: response.ETag ?? "" };
    } catch (error) {
      throw new Error(`Upload failed: ${(error as Error).message}`);
    }
  }

  async downloadObject(bucket: string, key: string) {
    // Same bug: no bucket validation.
    const command = new GetObjectCommand({ Bucket: bucket, Key: key });
    try {
      const response = await this.client.send(command);
      return { body: response.Body };
    } catch (error) {
      throw new Error(`Download failed: ${(error as Error).message}`);
    }
  }
}
