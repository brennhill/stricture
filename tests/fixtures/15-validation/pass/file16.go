// UploadIfNotExistsBuggy has race condition on concurrent uploads.
func (c *Client) UploadIfNotExistsBuggy(ctx context.Context, bucket, key string, data io.Reader) error {
	if err := validateBucketName(bucket); err != nil {
		return err
	}
	if err := validateKey(key); err != nil {
		return err
	}

	// BUG: Race condition
	// 1. Check if object exists
	_, err := c.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err == nil {
		// Object exists, abort
		return fmt.Errorf("object already exists: %s", key)
	}

	// 2. Upload object (RACE: another process might upload between check and put)
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   data,
		// MISSING: IfNoneMatch: aws.String("*"), // Atomic check-and-put
	}

	_, err = c.s3Client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("PutObject failed: %w", err)
	}

	return nil
}
