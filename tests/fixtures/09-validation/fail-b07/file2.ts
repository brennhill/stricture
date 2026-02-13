// src/routes/download.ts — B07: Downstream code breaks on wrong types

import { s3Service } from "../services/s3-client";

export async function handleDownload(req: Request): Promise<Response> {
  const result = await s3Service.downloadObject("my-bucket", req.url);

  // BUG: contentLength is string. This calculates string concatenation, not addition.
  // remainingBytes = "4096" + 0 = "40960" (string), not 4096 (number).
  const remainingBytes = result.contentLength + 0;

  // BUG: etag has no quotes. If-None-Match requires quoted etag per HTTP spec.
  // Sends: If-None-Match: d41d8cd98... instead of "d41d8cd98..."
  const cacheHeaders = { "If-None-Match": result.etag };

  return new Response(result.body, {
    headers: {
      "Content-Length": String(remainingBytes),
      "Cache-Control": "max-age=3600",
    },
  });
}
