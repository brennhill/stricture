// src/services/s3-client.ts — B07: Wrong field types

interface S3DownloadResult {
  body: ReadableStream;
  contentType: string;
  contentLength: string;    // BUG: manifest says integer, stored as string
  etag: string;
  versionId: string | null;
}

export class S3Service {
  async downloadObject(bucket: string, key: string): Promise<S3DownloadResult> {
    const url = `https://${bucket}.s3.amazonaws.com/${encodeURIComponent(key)}`;
    const headers = await this.signer.sign("GET", url);

    try {
      const response = await fetch(url, { headers });
      if (!response.ok) {
        throw new Error(`S3 GET failed with status ${response.status}`);
      }

      return {
        body: response.body!,
        contentType: response.headers.get("Content-Type") ?? "",
        // BUG: Content-Length from headers is always a string.
        // Manifest requires integer. Arithmetic breaks: "4096" + 1024 = "40961024"
        contentLength: response.headers.get("Content-Length") ?? "0",
        // BUG: Stripping quotes from ETag. Manifest format requires quotes.
        // "d41d8cd98..." becomes d41d8cd98... (no longer matches ^"[a-f0-9]{32}"$)
        etag: (response.headers.get("ETag") ?? "").replace(/"/g, ""),
        versionId: response.headers.get("x-amz-version-id") ?? null,
      };
    } catch (error) {
      throw new Error(`Download failed: ${(error as Error).message}`);
    }
  }
}
