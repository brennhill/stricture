// src/services/s3-client.ts — B01: No error handling

export class S3Service {
  private client: S3Client;

  constructor(region: string) {
    this.client = new S3Client({ region });
  }

  // BUG: No try/catch. S3 SDK throws on 403, 404, 500, etc.
  // These propagate as unhandled promise rejections.
  async uploadObject(bucket: string, key: string, body: Buffer, contentType: string) {
    const command = new PutObjectCommand({
      Bucket: bucket, Key: key, Body: body, ContentType: contentType,
    });
    const response = await this.client.send(command);
    return { etag: response.ETag };
  }

  async downloadObject(bucket: string, key: string) {
    const command = new GetObjectCommand({ Bucket: bucket, Key: key });
    const response = await this.client.send(command);
    return { body: response.Body, contentType: response.ContentType };
  }

  async deleteObject(bucket: string, key: string) {
    const command = new DeleteObjectCommand({ Bucket: bucket, Key: key });
    await this.client.send(command);
  }

  async generatePresignedUrl(bucket: string, key: string, expires: number) {
    const command = new GetObjectCommand({ Bucket: bucket, Key: key });
    return await getSignedUrl(this.client, command, { expiresIn: expires });
  }
}
