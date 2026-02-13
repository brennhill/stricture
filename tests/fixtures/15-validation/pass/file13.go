// GetObjectMetadataBuggy crashes on nil ETag.
func (c *Client) GetObjectMetadataBuggy(ctx context.Context, bucket, key string) (string, error) {
	if err := validateBucketName(bucket); err != nil {
		return "", err
	}
	if err := validateKey(key); err != nil {
		return "", err
	}

	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	output, err := c.s3Client.HeadObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("HeadObject failed: %w", err)
	}

	// BUG: ETag is *string in AWS SDK (nullable)
	// Dereferences without nil check → panic if ETag is nil
	etag := *output.ETag // CRASH: panic: invalid memory address or nil pointer dereference

	return etag, nil
}
