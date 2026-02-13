// src/services/s3-client.ts — B15: Race condition on conditional upload

export class S3Service {
  async safeUpload(
    bucket: string,
    key: string,
    body: Buffer,
    contentType: string
  ): Promise<{ etag: string; action: "created" | "skipped" }> {
    try {
      // Step 1: Check if object already exists
      const headCommand = new GetObjectCommand({ Bucket: bucket, Key: key });
      try {
        await this.client.send(headCommand);
        // Object exists — skip upload
        return { etag: "", action: "skipped" };
      } catch (error: unknown) {
        if ((error as Error).name !== "NoSuchKey") {
          throw error;
        }
        // Object does not exist — proceed to upload
      }

      // BUG: Race condition window between HEAD check and PUT.
      // Another process can upload the same key between these two calls.
      // This upload will silently overwrite the other process's data.
      //
      // CORRECT approach: Use If-None-Match: "*" header on PUT to make
      // S3 reject the upload with 412 Precondition Failed if the object
      // already exists. This is an atomic check-and-write.

      // Step 2: Upload (no conditional header)
      const putCommand = new PutObjectCommand({
        Bucket: bucket,
        Key: key,
        Body: body,
        ContentType: contentType,
        // MISSING: IfNoneMatch: "*"
        // Without this, concurrent uploads can overwrite each other.
      });

      const response = await this.client.send(putCommand);
      return { etag: response.ETag ?? "", action: "created" };
    } catch (error) {
      throw new Error(`Safe upload failed: ${(error as Error).message}`);
    }
  }

  async updateObject(
    bucket: string,
    key: string,
    body: Buffer,
    contentType: string,
    _expectedETag: string  // Parameter accepted but never used
  ): Promise<{ etag: string }> {
    // BUG: expectedETag is accepted as a parameter but never used.
    // Should be sent as If-Match header to prevent overwriting concurrent changes.
    // Without If-Match, this is a blind overwrite.

    const putCommand = new PutObjectCommand({
      Bucket: bucket,
      Key: key,
      Body: body,
      ContentType: contentType,
      // MISSING: IfMatch: expectedETag
    });

    try {
      const response = await this.client.send(putCommand);
      return { etag: response.ETag ?? "" };
    } catch (error) {
      throw new Error(`Update failed: ${(error as Error).message}`);
    }
  }
}
