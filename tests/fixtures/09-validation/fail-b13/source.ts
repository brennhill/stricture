// src/services/s3-client.ts — B13: No credential check before presigning

export class S3Service {
  private client: S3Client;

  constructor(region: string) {
    // BUG: S3Client created without explicit credentials.
    // Falls back to environment variables / instance profile.
    // If neither is configured, client.send() fails, but getSignedUrl()
    // may succeed with empty credentials — producing an invalid URL.
    this.client = new S3Client({ region });
  }

  async generatePresignedUrl(
    bucket: string,
    key: string,
    method: "GET" | "PUT",
    expiresInSeconds: number
  ): Promise<{ url: string; expiresAt: string }> {
    // BUG: No check that credentials are actually configured.
    // getSignedUrl() computes the signature using whatever credentials
    // the SDK resolves. If no credentials are found:
    // - In SDK v3, it may throw CredentialsProviderError
    // - Or it may silently use empty credentials, producing a URL whose
    //   signature is invalid. The URL looks valid but S3 returns 403.

    const command = method === "GET"
      ? new GetObjectCommand({ Bucket: bucket, Key: key })
      : new PutObjectCommand({ Bucket: bucket, Key: key });

    const url = await getSignedUrl(this.client, command, { expiresIn: expiresInSeconds });
    const expiresAt = new Date(Date.now() + expiresInSeconds * 1000).toISOString();

    return { url, expiresAt };
  }
}
