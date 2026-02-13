// src/services/s3-client.ts — B11: Content-Length precision loss

interface UploadProgress {
  loaded: number;
  total: number;     // BUG: Stored in a context where 32-bit truncation occurs
  percentage: number;
}

export class S3Service {
  async uploadLargeFile(
    bucket: string,
    key: string,
    filePath: string,
    contentType: string,
    onProgress?: (progress: UploadProgress) => void
  ) {
    const stats = await fs.stat(filePath);

    // BUG: Using bitwise OR to "convert to integer" truncates to 32 bits.
    // For a 3GB file (3,221,225,472 bytes):
    //   3221225472 | 0 = -1073741824 (32-bit signed overflow!)
    // For a 5TB file (5,497,558,138,880 bytes):
    //   5497558138880 | 0 = 1202590592 (truncated to 32 bits)
    const totalBytes = stats.size | 0;

    if (totalBytes <= 0) {
      throw new Error("File is empty or size calculation failed.");
    }

    // BUG: totalBytes is wrong for files > 2GB.
    // A 3GB file shows totalBytes = -1073741824, which fails the > 0 check.
    // A 4.5GB file shows totalBytes = 205521408, which passes but is wrong.

    const command = new PutObjectCommand({
      Bucket: bucket,
      Key: key,
      Body: createReadStream(filePath),
      ContentType: contentType,
      ContentLength: totalBytes,  // BUG: Truncated value sent to S3
    });

    try {
      const response = await this.client.send(command);
      return { etag: response.ETag ?? "", size: totalBytes };
    } catch (error) {
      throw new Error(`Upload failed: ${(error as Error).message}`);
    }
  }
}
