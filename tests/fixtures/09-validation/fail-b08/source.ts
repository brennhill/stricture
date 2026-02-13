// src/services/s3-client.ts — B08: Incomplete storage class enum

type StorageClass = "STANDARD" | "REDUCED_REDUNDANCY" | "STANDARD_IA";
// BUG: Missing "ONEZONE_IA", "INTELLIGENT_TIERING", "GLACIER", "DEEP_ARCHIVE"
// Manifest declares all 7. Client handles only 3.

export class S3Service {
  async downloadObject(bucket: string, key: string) {
    try {
      const command = new GetObjectCommand({ Bucket: bucket, Key: key });
      const response = await this.client.send(command);

      const storageClass = response.StorageClass as StorageClass;

      // BUG: Only handles 3 of 7 storage classes. When storageClass is
      // "GLACIER" or "DEEP_ARCHIVE", this switch falls through to default
      // and treats it as available for immediate download.
      switch (storageClass) {
        case "STANDARD":
        case "REDUCED_REDUNDANCY":
        case "STANDARD_IA":
          return { body: response.Body, available: true };
        default:
          // BUG: GLACIER objects are NOT immediately available.
          // They require a RestoreObject call and hours/days of wait time.
          // Treating them as available causes download to fail silently.
          return { body: response.Body, available: true };
      }
    } catch (error) {
      throw new Error(`Download failed: ${(error as Error).message}`);
    }
  }

  async setStorageClass(bucket: string, key: string, storageClass: StorageClass) {
    // BUG: Type only accepts 3 values. Cannot transition to GLACIER or
    // DEEP_ARCHIVE, even though the manifest and S3 API support it.
    const command = new PutObjectCommand({
      Bucket: bucket,
      Key: key,
      StorageClass: storageClass,
    });
    await this.client.send(command);
  }
}
