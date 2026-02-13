// ObjectMetadataBuggy uses wrong type for ContentLength.
type ObjectMetadataBuggy struct {
	Key          string
	Size         int64
	LastModified *time.Time
	ETag         string
	ContentType  string
	ContentLength string // BUG: Should be int64, not string
}

// HeadObjectBuggy returns metadata with wrong ContentLength type.
func (c *Client) HeadObjectBuggy(ctx context.Context, bucket, key string) (*ObjectMetadataBuggy, error) {
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

	// BUG: Converting int64 to string — caller must parse it back
	return &ObjectMetadataBuggy{
		Key:          key,
		Size:         aws.ToInt64(output.ContentLength),
		ETag:         aws.ToString(output.ETag),
		ContentType:  aws.ToString(output.ContentType),
		ContentLength: fmt.Sprintf("%d", aws.ToInt64(output.ContentLength)),
	}, nil
}
