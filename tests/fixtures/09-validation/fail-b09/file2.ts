// src/routes/share.ts — B09: Caller passes invalid expiration

import { s3Service } from "../services/s3-client";

export async function createShareLink(fileKey: string): Promise<string> {
  // BUG: 30 days = 2,592,000 seconds. Max is 604,800 (7 days).
  // URL is generated successfully (signing is client-side) but S3 rejects
  // the request when the user clicks the link.
  const { url } = await s3Service.generatePresignedUrl(
    "my-bucket", fileKey, "GET", 30 * 24 * 60 * 60
  );
  return url;
}
