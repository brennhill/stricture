// src/services/s3-client.ts — B05: Missing required Content-Type header

export class S3Service {
  async uploadObject(bucket: string, key: string, body: Buffer) {
    // BUG: No contentType parameter. PutObjectCommand is created without
    // ContentType. S3 defaults to "application/octet-stream" for all uploads.
    const command = new PutObjectCommand({
      Bucket: bucket,
      Key: key,
      Body: body,
      // ContentType is MISSING — required by manifest
    });

    try {
      const response = await this.client.send(command);
      return { etag: response.ETag ?? "" };
    } catch (error) {
      throw new Error(`Upload failed: ${(error as Error).message}`);
    }
  }

  async uploadImage(bucket: string, key: string, imageBuffer: Buffer) {
    // BUG: Calls uploadObject without specifying content type.
    // A JPEG uploaded this way is stored as application/octet-stream.
    return this.uploadObject(bucket, key, imageBuffer);
  }
}
