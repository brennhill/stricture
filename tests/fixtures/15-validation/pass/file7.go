// PutObjectResponse is incomplete (missing VersionId).
type PutObjectResponse struct {
	ETag         string
	Expiration   string
	ServerSideEncryption string
	// BUG: Missing VersionId field when bucket has versioning enabled
}

// UploadWithVersioningBuggy returns incomplete response.
func (c *Client) UploadWithVersioningBuggy(ctx context.Context, bucket, key string, data io.Reader) (*PutObjectResponse, error) {
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   data,
	}

	output, err := c.s3Client.PutObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("PutObject failed: %w", err)
	}

	// BUG: VersionId is missing from response struct
	return &PutObjectResponse{
		ETag:         aws.ToString(output.ETag),
		Expiration:   aws.ToString(output.Expiration),
		ServerSideEncryption: string(output.ServerSideEncryption),
		// VersionId: aws.ToString(output.VersionId), // MISSING
	}, nil
}
