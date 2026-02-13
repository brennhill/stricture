// src/services/s3-client.ts — B14: Multipart upload never completed

const PART_SIZE = 5 * 1024 * 1024; // 5 MB

export class S3Service {
  async uploadLargeFile(
    bucket: string,
    key: string,
    fileBuffer: Buffer,
    contentType: string
  ): Promise<{ uploadId: string; partsUploaded: number }> {
    try {
      // Step 1: Initiate multipart upload
      const initCommand = new CreateMultipartUploadCommand({
        Bucket: bucket,
        Key: key,
        ContentType: contentType,
      });
      const initResponse = await this.client.send(initCommand);
      const uploadId = initResponse.UploadId!;

      // Step 2: Upload parts
      const partCount = Math.ceil(fileBuffer.length / PART_SIZE);
      const parts: Array<{ PartNumber: number; ETag: string }> = [];

      for (let i = 0; i < partCount; i++) {
        const start = i * PART_SIZE;
        const end = Math.min(start + PART_SIZE, fileBuffer.length);
        const partBody = fileBuffer.subarray(start, end);

        const uploadPartCommand = new UploadPartCommand({
          Bucket: bucket,
          Key: key,
          UploadId: uploadId,
          PartNumber: i + 1,
          Body: partBody,
          ContentLength: partBody.length,
        });

        const partResponse = await this.client.send(uploadPartCommand);
        parts.push({ PartNumber: i + 1, ETag: partResponse.ETag! });
      }

      // BUG: Step 3 (CompleteMultipartUpload) is MISSING.
      // The parts are uploaded but the object is never assembled.
      // S3 keeps the parts in storage, consuming space and cost.
      // The object does not appear in bucket listings.
      // No AbortMultipartUpload is called on error either.

      return { uploadId, partsUploaded: parts.length };
      // Should call CompleteMultipartUploadCommand here with the parts array
    } catch (error) {
      // BUG: On error, no AbortMultipartUpload is called.
      // The incomplete upload remains, consuming storage.
      throw new Error(`Upload failed: ${(error as Error).message}`);
    }
  }
}
