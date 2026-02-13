// ListObjectsNoPagination returns only first page of results.
func (c *Client) ListObjectsNoPagination(ctx context.Context, bucket, prefix string) ([]ObjectMetadata, error) {
	if err := validateBucketName(bucket); err != nil {
		return nil, err
	}

	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		MaxKeys: aws.Int32(1000),
	}
	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}

	output, err := c.s3Client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("ListObjectsV2 failed: %w", err)
	}

	// BUG: Ignores IsTruncated and NextContinuationToken
	// If bucket has 5000 objects, this only returns first 1000
	var objects []ObjectMetadata
	for _, obj := range output.Contents {
		if obj.Key == nil {
			continue
		}
		objects = append(objects, ObjectMetadata{
			Key:  aws.ToString(obj.Key),
			Size: aws.ToInt64(obj.Size),
			ETag: aws.ToString(obj.ETag),
		})
	}

	// Should check: if aws.ToBool(output.IsTruncated) { /* fetch next page */ }
	return objects, nil
}
