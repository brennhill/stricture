// s3_client.go — Production AWS S3 client with presigned URLs and pagination.
package s3client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3API defines the interface for S3 operations (testability).
type S3API interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
}

// PresignAPI defines presigning operations.
type PresignAPI interface {
	PresignPutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.PresignOptions)) (*s3.PresignedHTTPRequest, error)
	PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*s3.PresignedHTTPRequest, error)
}

// Client wraps AWS S3 client with presigning support.
type Client struct {
	s3Client     S3API
	presignClient PresignAPI
}

// Config holds S3 client configuration.
type Config struct {
	Region          string
	MaxRetries      int
	RequestTimeout  time.Duration
	PresignDuration time.Duration
}

// ObjectMetadata represents S3 object metadata.
type ObjectMetadata struct {
	Key          string
	Size         int64
	LastModified *time.Time
	ETag         string
	ContentType  string
	StorageClass types.StorageClass
}

// ListResult holds paginated list results.
type ListResult struct {
	Objects               []ObjectMetadata
	ContinuationToken     *string
	NextContinuationToken *string
	IsTruncated           bool
	KeyCount              int32
}

var (
	ErrInvalidBucket      = errors.New("invalid bucket name")
	ErrInvalidKey         = errors.New("invalid key")
	ErrPresignExpired     = errors.New("presigned URL has expired")
	ErrPresignInvalid     = errors.New("presigned URL is invalid")
	ErrObjectNotFound     = errors.New("object not found")
	ErrAccessDenied       = errors.New("access denied")
	ErrInvalidCredentials = errors.New("invalid AWS credentials")
)

const (
	MinPresignDuration = 1 * time.Second
	MaxPresignDuration = 7 * 24 * time.Hour // 7 days (AWS maximum)
	DefaultTimeout     = 30 * time.Second
)

// bucketNameRegex validates S3 bucket naming rules.
var bucketNameRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9\-]{1,61}[a-z0-9]$`)

// NewClient creates a new S3 client with configuration.
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.Region == "" {
		return nil, errors.New("region is required")
	}
	if cfg.PresignDuration < MinPresignDuration || cfg.PresignDuration > MaxPresignDuration {
		return nil, fmt.Errorf("presign duration must be between %v and %v", MinPresignDuration, MaxPresignDuration)
	}
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = DefaultTimeout
	}

	// Load AWS configuration with retry settings.
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithRetryMaxAttempts(cfg.MaxRetries),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Validate credentials are present.
	creds, err := awsCfg.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidCredentials, err)
	}
	if creds.AccessKeyID == "" || creds.SecretAccessKey == "" {
		return nil, ErrInvalidCredentials
	}

	s3Client := s3.NewFromConfig(awsCfg)
	presignClient := s3.NewPresignClient(s3Client)

	return &Client{
		s3Client:     s3Client,
		presignClient: presignClient,
	}, nil
}

// validateBucketName ensures bucket name follows S3 naming rules.
func validateBucketName(bucket string) error {
	if bucket == "" {
		return fmt.Errorf("%w: bucket name is empty", ErrInvalidBucket)
	}
	if len(bucket) < 3 || len(bucket) > 63 {
		return fmt.Errorf("%w: bucket name must be 3-63 characters", ErrInvalidBucket)
	}
	if !bucketNameRegex.MatchString(bucket) {
		return fmt.Errorf("%w: bucket name contains invalid characters", ErrInvalidBucket)
	}
	if strings.Contains(bucket, "..") {
		return fmt.Errorf("%w: bucket name cannot contain consecutive dots", ErrInvalidBucket)
	}
	return nil
}

// validateKey ensures S3 key is valid.
func validateKey(key string) error {
	if key == "" {
		return fmt.Errorf("%w: key is empty", ErrInvalidKey)
	}
	if len(key) > 1024 {
		return fmt.Errorf("%w: key exceeds 1024 characters", ErrInvalidKey)
	}
	return nil
}

// GeneratePresignedPutURL creates a presigned URL for uploading an object.
func (c *Client) GeneratePresignedPutURL(ctx context.Context, bucket, key string, duration time.Duration, contentType string, acl types.ObjectCannedACL) (*s3.PresignedHTTPRequest, error) {
	if err := validateBucketName(bucket); err != nil {
		return nil, err
	}
	if err := validateKey(key); err != nil {
		return nil, err
	}
	if duration < MinPresignDuration || duration > MaxPresignDuration {
		return nil, fmt.Errorf("presign duration must be between %v and %v", MinPresignDuration, MaxPresignDuration)
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}
	if acl != "" {
		input.ACL = acl
	}

	presigned, err := c.presignClient.PresignPutObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = duration
	})
	if err != nil {
		return nil, fmt.Errorf("failed to presign PutObject: %w", err)
	}

	return presigned, nil
}

// GeneratePresignedGetURL creates a presigned URL for downloading an object.
func (c *Client) GeneratePresignedGetURL(ctx context.Context, bucket, key string, duration time.Duration) (*s3.PresignedHTTPRequest, error) {
	if err := validateBucketName(bucket); err != nil {
		return nil, err
	}
	if err := validateKey(key); err != nil {
		return nil, err
	}
	if duration < MinPresignDuration || duration > MaxPresignDuration {
		return nil, fmt.Errorf("presign duration must be between %v and %v", MinPresignDuration, MaxPresignDuration)
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	presigned, err := c.presignClient.PresignGetObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = duration
	})
	if err != nil {
		return nil, fmt.Errorf("failed to presign GetObject: %w", err)
	}

	return presigned, nil
}

// UploadViaPresignedURL uploads data using a presigned PUT URL.
func UploadViaPresignedURL(ctx context.Context, presignedReq *s3.PresignedHTTPRequest, data io.Reader, contentLength int64) error {
	if presignedReq == nil {
		return ErrPresignInvalid
	}

	req, err := http.NewRequestWithContext(ctx, presignedReq.Method, presignedReq.URL, data)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Copy headers from presigned request.
	for key, values := range presignedReq.SignedHeader {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	if contentLength > 0 {
		req.ContentLength = contentLength
	}

	client := &http.Client{Timeout: DefaultTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusForbidden {
			return fmt.Errorf("%w: %s", ErrPresignExpired, string(body))
		}
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DownloadViaPresignedURL downloads data using a presigned GET URL.
func DownloadViaPresignedURL(ctx context.Context, presignedReq *s3.PresignedHTTPRequest) ([]byte, error) {
	if presignedReq == nil {
		return nil, ErrPresignInvalid
	}

	req, err := http.NewRequestWithContext(ctx, presignedReq.Method, presignedReq.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	client := &http.Client{Timeout: DefaultTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusForbidden {
			return nil, fmt.Errorf("%w: %s", ErrPresignExpired, string(body))
		}
		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrObjectNotFound
		}
		return nil, fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

// ListObjects lists all objects in a bucket with full pagination support.
func (c *Client) ListObjects(ctx context.Context, bucket, prefix string, maxKeys int32, continuationToken *string) (*ListResult, error) {
	if err := validateBucketName(bucket); err != nil {
		return nil, err
	}
	if maxKeys <= 0 {
		maxKeys = 1000 // AWS default
	}
	if maxKeys > 1000 {
		maxKeys = 1000 // AWS maximum
	}

	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		MaxKeys: aws.Int32(maxKeys),
	}
	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}
	if continuationToken != nil && *continuationToken != "" {
		input.ContinuationToken = continuationToken
	}

	output, err := c.s3Client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("ListObjectsV2 failed: %w", err)
	}

	result := &ListResult{
		Objects:               make([]ObjectMetadata, 0, len(output.Contents)),
		ContinuationToken:     continuationToken,
		NextContinuationToken: output.NextContinuationToken,
		IsTruncated:           aws.ToBool(output.IsTruncated),
		KeyCount:              aws.ToInt32(output.KeyCount),
	}

	for _, obj := range output.Contents {
		if obj.Key == nil {
			continue // Skip malformed entries
		}
		metadata := ObjectMetadata{
			Key:          aws.ToString(obj.Key),
			Size:         aws.ToInt64(obj.Size),
			LastModified: obj.LastModified,
			ETag:         aws.ToString(obj.ETag),
			StorageClass: obj.StorageClass,
		}
		result.Objects = append(result.Objects, metadata)
	}

	return result, nil
}

// ListAllObjects lists all objects with automatic pagination.
func (c *Client) ListAllObjects(ctx context.Context, bucket, prefix string) ([]ObjectMetadata, error) {
	var allObjects []ObjectMetadata
	var continuationToken *string

	for {
		result, err := c.ListObjects(ctx, bucket, prefix, 1000, continuationToken)
		if err != nil {
			return nil, err
		}

		allObjects = append(allObjects, result.Objects...)

		if !result.IsTruncated {
			break
		}

		continuationToken = result.NextContinuationToken
		if continuationToken == nil || *continuationToken == "" {
			// Safeguard against infinite loop
			break
		}
	}

	return allObjects, nil
}

// DeleteObject deletes an object from S3.
func (c *Client) DeleteObject(ctx context.Context, bucket, key string) error {
	if err := validateBucketName(bucket); err != nil {
		return err
	}
	if err := validateKey(key); err != nil {
		return err
	}

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := c.s3Client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("DeleteObject failed: %w", err)
	}

	return nil
}

// HeadObject retrieves object metadata without downloading content.
func (c *Client) HeadObject(ctx context.Context, bucket, key string) (*ObjectMetadata, error) {
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
		if strings.Contains(err.Error(), "NotFound") {
			return nil, ErrObjectNotFound
		}
		return nil, fmt.Errorf("HeadObject failed: %w", err)
	}

	metadata := &ObjectMetadata{
		Key:          key,
		Size:         aws.ToInt64(output.ContentLength),
		LastModified: output.LastModified,
		ETag:         aws.ToString(output.ETag),
		ContentType:  aws.ToString(output.ContentType),
	}

	return metadata, nil
}
