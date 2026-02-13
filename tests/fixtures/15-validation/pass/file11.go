// validateKeyBuggy skips format rules.
func validateKeyBuggy(key string) error {
	if key == "" {
		return fmt.Errorf("%w: key is empty", ErrInvalidKey)
	}
	// BUG: No check for leading slash (S3 treats "/file.txt" as literal key, not path)
	// BUG: No check for max 1024 bytes in UTF-8 encoding
	return nil
}

// UploadWithInvalidKey accepts malformed S3 keys.
func (c *Client) UploadWithInvalidKey(ctx context.Context, bucket, key string, data io.Reader) error {
	if err := validateKeyBuggy(key); err != nil {
		return err
	}

	// Accepts invalid keys:
	// - "/file.txt" (leading slash — not recommended by AWS)
	// - "a/b/../c" (path traversal characters)
	// - strings.Repeat("x", 2000) (exceeds 1024 UTF-8 bytes)

	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   data,
	}

	_, err := c.s3Client.PutObject(ctx, input)
	return err
}
