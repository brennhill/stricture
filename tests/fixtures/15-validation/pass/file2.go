// UploadObject uploads data to S3 without error handling.
func (c *Client) UploadObject(ctx context.Context, bucket, key string, data io.Reader, contentType string) error {
	if err := validateBucketName(bucket); err != nil {
		return err
	}
	if err := validateKey(key); err != nil {
		return err
	}

	input := &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        data,
		ContentType: aws.String(contentType),
	}

	// BUG: No error handling on PutObject call
	c.s3Client.PutObject(ctx, input)

	return nil
}
