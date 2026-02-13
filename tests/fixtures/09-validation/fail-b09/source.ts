// src/services/s3-client.ts — B09: No range validation on presign expires

export class S3Service {
  async generatePresignedUrl(
    bucket: string,
    key: string,
    method: "GET" | "PUT",
    expiresInSeconds: number
  ): Promise<{ url: string; expiresAt: string }> {
    // BUG: No range validation. Accepts negative values, zero, and values
    // above 604800 (S3 max is 7 days / 604800 seconds).
    // S3 rejects presigned URLs with expires > 604800 at request time with:
    // <Error><Code>AuthorizationQueryParametersError</Code></Error>

    const command = method === "GET"
      ? new GetObjectCommand({ Bucket: bucket, Key: key })
      : new PutObjectCommand({ Bucket: bucket, Key: key });

    // This generates a URL with X-Amz-Expires=1000000 which S3 rejects.
    const url = await getSignedUrl(this.client, command, { expiresIn: expiresInSeconds });
    const expiresAt = new Date(Date.now() + expiresInSeconds * 1000).toISOString();

    return { url, expiresAt };
  }
}
