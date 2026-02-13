// src/services/s3-client.ts — B12: VersionId accessed without null check

interface VersionInfo {
  id: string;
  timestamp: Date;
  isLatest: boolean;
}

export class S3Service {
  async uploadObject(bucket: string, key: string, body: Buffer, contentType: string) {
    try {
      const command = new PutObjectCommand({
        Bucket: bucket, Key: key, Body: body, ContentType: contentType,
      });
      const response = await this.client.send(command);

      // BUG: VersionId is undefined when bucket versioning is not enabled.
      // Manifest declares: x-amz-version-id: { required: false, nullable: true }
      // This code assumes it always exists.
      const versionInfo: VersionInfo = {
        id: response.VersionId!,                        // Undefined when unversioned
        timestamp: new Date(response.VersionId!.split(".")[1]),  // Crashes: Cannot read property 'split' of undefined
        isLatest: true,
      };

      return {
        etag: response.ETag ?? "",
        version: versionInfo,
      };
    } catch (error) {
      throw new Error(`Upload failed: ${(error as Error).message}`);
    }
  }

  async getObjectVersion(bucket: string, key: string) {
    try {
      const command = new GetObjectCommand({ Bucket: bucket, Key: key });
      const response = await this.client.send(command);

      // BUG: Same issue. VersionId is undefined on unversioned buckets.
      // .length on undefined throws TypeError.
      const versionLength = response.VersionId!.length;

      return {
        body: response.Body,
        versionId: response.VersionId!,
        versionIdLength: versionLength,
      };
    } catch (error) {
      throw new Error(`Get failed: ${(error as Error).message}`);
    }
  }
}
