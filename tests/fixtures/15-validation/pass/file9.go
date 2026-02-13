// ValidateACLBuggy only handles subset of valid ACL values.
func ValidateACLBuggy(acl types.ObjectCannedACL) error {
	// BUG: Missing valid ACL values
	switch acl {
	case types.ObjectCannedACLPrivate:
		return nil
	case types.ObjectCannedACLPublicRead:
		return nil
	default:
		return fmt.Errorf("invalid ACL: %s", acl)
	}
	// Missing:
	// - ObjectCannedACLPublicReadWrite
	// - ObjectCannedACLAuthenticatedRead
	// - ObjectCannedACLAwsExecRead
	// - ObjectCannedACLBucketOwnerRead
	// - ObjectCannedACLBucketOwnerFullControl
}

// GeneratePresignedPutURLWithACL rejects valid ACL values.
func (c *Client) GeneratePresignedPutURLWithACL(ctx context.Context, bucket, key string, acl types.ObjectCannedACL) (*s3.PresignedHTTPRequest, error) {
	if err := ValidateACLBuggy(acl); err != nil {
		return nil, err
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		ACL:    acl,
	}

	return c.presignClient.PresignPutObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = 1 * time.Hour
	})
}
