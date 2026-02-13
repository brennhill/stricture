// GeneratePresignedURLNoValidation allows invalid expiration durations.
func (c *Client) GeneratePresignedURLNoValidation(ctx context.Context, bucket, key string, duration time.Duration) (*s3.PresignedHTTPRequest, error) {
	if err := validateBucketName(bucket); err != nil {
		return nil, err
	}
	if err := validateKey(key); err != nil {
		return nil, err
	}

	// BUG: No range validation on duration
	// AWS maximum: 7 days (604800 seconds) for IAM users
	// This accepts duration=0 (instant expiry) or duration=8*24*time.Hour (rejected by AWS)

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	presigned, err := c.presignClient.PresignGetObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = duration // BUG: No bounds checking
	})
	if err != nil {
		return nil, fmt.Errorf("failed to presign GetObject: %w", err)
	}

	return presigned, nil
}
