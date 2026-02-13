// src/services/s3-client.ts — B02: No status code check (raw fetch variant)

export class S3Service {
  private baseUrl: string;
  private signer: AwsV4Signer;

  constructor(region: string, bucket: string) {
    this.baseUrl = `https://${bucket}.s3.${region}.amazonaws.com`;
    this.signer = new AwsV4Signer(region);
  }

  async uploadObject(key: string, body: Buffer, contentType: string) {
    const url = `${this.baseUrl}/${encodeURIComponent(key)}`;
    const headers = await this.signer.sign("PUT", url, { "Content-Type": contentType });

    const response = await fetch(url, {
      method: "PUT",
      headers,
      body,
    });

    // BUG: No response.ok or status check. If S3 returns 403 or 500,
    // the XML error body is returned as if it were success data.
    const etag = response.headers.get("ETag");
    return { etag, status: "success" };
  }

  async downloadObject(key: string) {
    const url = `${this.baseUrl}/${encodeURIComponent(key)}`;
    const headers = await this.signer.sign("GET", url);

    const response = await fetch(url, { headers });

    // BUG: 404 response body is S3 XML error, not the object data.
    // Client returns XML error string as "file contents".
    const data = await response.arrayBuffer();
    return {
      body: data,
      contentType: response.headers.get("Content-Type"),
    };
  }
}
