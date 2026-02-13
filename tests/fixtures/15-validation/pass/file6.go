// UploadTextFileBuggy uploads text without ContentType.
func (c *Client) UploadTextFileBuggy(ctx context.Context, bucket, key string, data io.Reader) error {
	if err := validateBucketName(bucket); err != nil {
		return err
	}
	if err := validateKey(key); err != nil {
		return err
	}

	// BUG: Missing ContentType field — should be "text/plain" for .txt files
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   data,
		// ContentType: aws.String("text/plain"), // MISSING
	}

	_, err := c.s3Client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("PutObject failed: %w", err)
	}

	return nil
}
