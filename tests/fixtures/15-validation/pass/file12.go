// ObjectSizeBuggy truncates ContentLength to int32.
type ObjectSizeBuggy struct {
	Key          string
	ContentLength int32 // BUG: Should be int64 (max S3 object: 5TB)
}

// HeadObjectTruncated loses precision for large files.
func (c *Client) HeadObjectTruncated(ctx context.Context, bucket, key string) (*ObjectSizeBuggy, error) {
	if err := validateBucketName(bucket); err != nil {
		return nil, err
	}
	if err := validateKey(key); err != nil {
		return nil, err
	}

	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	output, err := c.s3Client.HeadObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("HeadObject failed: %w", err)
	}

	// BUG: Truncates int64 to int32
	// For 3GB file (3221225472 bytes):
	//   int64: 3221225472
	//   int32: -1073741824 (overflow, becomes negative)
	return &ObjectSizeBuggy{
		Key:          key,
		ContentLength: int32(aws.ToInt64(output.ContentLength)),
	}, nil
}
