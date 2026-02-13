# 15 — AWS S3 Presigned URLs (Go)

**Why included:** Go AWS SDK v2 patterns, presigned URLs, interface-based testing, XML error parsing, pagination handling.

## Manifest Fragment

```yaml
contracts:
  - id: "15"
    name: "AWS S3 Presigned URLs (Go)"
    language: "go"
    framework: "aws-sdk-go-v2"
    operations:
      - PutObject presigned URL generation with expiration
      - GetObject presigned URL with validation
      - ListObjectsV2 with ContinuationToken pagination
      - DeleteObject with error handling
      - HeadObject for metadata retrieval
    patterns:
      - Interface-based testability (S3API interface)
      - Presigned URL expiration validation
      - Full pagination handling (IsTruncated + NextContinuationToken)
      - XML error response parsing
      - Table-driven tests with mock S3 service
```

---

## PERFECT — Full AWS S3 Implementation

```go
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
```

---

## BUG CASES

### B01 — No Error Handling on PutObject (TQ-error-path-coverage)

**Bug:** Missing error check on `s3Client.PutObject` call.
**Expected violation:** `TQ-error-path-coverage`

```go
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
```

**Why Stricture catches this:** `TQ-error-path-coverage` requires all external API calls to have error handling. The AWS SDK `PutObject` returns `(*s3.PutObjectOutput, error)`, but the error is ignored, creating a silent failure path.

---

### B02 — No Status Code Check on Presigned URL Response (CTR-status-code-handling)

**Bug:** HTTP response status not validated when using presigned URL.
**Expected violation:** `CTR-status-code-handling`

```go
// UploadViaPresignedURLBuggy uploads without checking HTTP status.
func UploadViaPresignedURLBuggy(ctx context.Context, presignedReq *s3.PresignedHTTPRequest, data io.Reader) error {
	if presignedReq == nil {
		return ErrPresignInvalid
	}

	req, err := http.NewRequestWithContext(ctx, presignedReq.Method, presignedReq.URL, data)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	for key, values := range presignedReq.SignedHeader {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	client := &http.Client{Timeout: DefaultTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload: %w", err)
	}
	defer resp.Body.Close()

	// BUG: No status code check — assumes success
	return nil
}
```

**Why Stricture catches this:** `CTR-status-code-handling` enforces that all HTTP responses must check `StatusCode`. The perfect version checks `resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent`, but this buggy version blindly assumes success.

---

### B03 — Shallow Test Assertions (TQ-no-shallow-assertions)

**Bug:** Test only checks `assert.NotNil(url)` instead of validating URL structure.
**Expected violation:** `TQ-no-shallow-assertions`

```go
// TestGeneratePresignedURLShallow uses weak assertions.
func TestGeneratePresignedURLShallow(t *testing.T) {
	ctx := context.Background()
	mockS3 := &mockS3Client{}
	mockPresign := &mockPresignClient{
		presignedURL: "https://bucket.s3.amazonaws.com/key?X-Amz-Signature=abc123",
	}
	client := &Client{s3Client: mockS3, presignClient: mockPresign}

	url, err := client.GeneratePresignedPutURL(ctx, "my-bucket", "test.txt", 1*time.Hour, "text/plain", types.ObjectCannedACLPrivate)

	// BUG: Shallow assertion — only checks non-nil
	assert.NoError(t, err)
	assert.NotNil(t, url)
	// Missing: URL structure validation, signature presence, expiration parameter
}
```

**Why Stricture catches this:** `TQ-no-shallow-assertions` flags assertions on existence (`NotNil`, `Len() > 0`) without validating structure. The perfect test would check `url.URL` contains `X-Amz-Signature`, `X-Amz-Expires`, and `X-Amz-Credential`.

---

### B04 — Missing Negative Tests (TQ-negative-cases)

**Bug:** Test suite lacks negative cases for expired presigned URLs and missing buckets.
**Expected violation:** `TQ-negative-cases`

```go
// TestPresignedURLsPositiveOnly only tests happy paths.
func TestPresignedURLsPositiveOnly(t *testing.T) {
	ctx := context.Background()
	mockS3 := &mockS3Client{}
	mockPresign := &mockPresignClient{
		presignedURL: "https://bucket.s3.amazonaws.com/key?X-Amz-Signature=valid",
	}
	client := &Client{s3Client: mockS3, presignClient: mockPresign}

	// BUG: Only positive test cases
	t.Run("generate PUT URL", func(t *testing.T) {
		url, err := client.GeneratePresignedPutURL(ctx, "my-bucket", "key", 1*time.Hour, "text/plain", "")
		assert.NoError(t, err)
		assert.Contains(t, url.URL, "X-Amz-Signature")
	})

	t.Run("generate GET URL", func(t *testing.T) {
		url, err := client.GeneratePresignedGetURL(ctx, "my-bucket", "key", 1*time.Hour)
		assert.NoError(t, err)
		assert.Contains(t, url.URL, "X-Amz-Signature")
	})

	// Missing negative tests:
	// - Expired presigned URL (403 Forbidden)
	// - Invalid bucket name
	// - Presign duration < MinPresignDuration or > MaxPresignDuration
	// - Missing object (404 on GET)
}
```

**Why Stricture catches this:** `TQ-negative-cases` requires tests for error paths. The perfect test suite includes cases for invalid inputs (bucket name, key length), expired URLs, and AWS error responses (403, 404).

---

### B05 — Request Missing Required Fields (CTR-request-shape)

**Bug:** `PutObjectInput` missing `ContentType` for text file upload.
**Expected violation:** `CTR-request-shape`

```go
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
```

**Why Stricture catches this:** `CTR-request-shape` validates that S3 `PutObject` requests for known file types include `ContentType`. Omitting this causes browsers to download files as `application/octet-stream` instead of rendering them inline.

---

### B06 — Response Type Mismatch (CTR-response-shape)

**Bug:** Missing `VersionId` field from response struct when versioning is enabled.
**Expected violation:** `CTR-response-shape`

```go
// PutObjectResponse is incomplete (missing VersionId).
type PutObjectResponse struct {
	ETag         string
	Expiration   string
	ServerSideEncryption string
	// BUG: Missing VersionId field when bucket has versioning enabled
}

// UploadWithVersioningBuggy returns incomplete response.
func (c *Client) UploadWithVersioningBuggy(ctx context.Context, bucket, key string, data io.Reader) (*PutObjectResponse, error) {
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   data,
	}

	output, err := c.s3Client.PutObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("PutObject failed: %w", err)
	}

	// BUG: VersionId is missing from response struct
	return &PutObjectResponse{
		ETag:         aws.ToString(output.ETag),
		Expiration:   aws.ToString(output.Expiration),
		ServerSideEncryption: string(output.ServerSideEncryption),
		// VersionId: aws.ToString(output.VersionId), // MISSING
	}, nil
}
```

**Why Stricture catches this:** `CTR-response-shape` compares response structs against AWS SDK manifest. `PutObjectOutput` includes `VersionId *string`, which is required when S3 bucket versioning is enabled. Missing this field breaks version tracking.

---

### B07 — Wrong Field Types (CTR-manifest-conformance)

**Bug:** `ContentLength` stored as `string` instead of `int64`.
**Expected violation:** `CTR-manifest-conformance`

```go
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
```

**Why Stricture catches this:** `CTR-manifest-conformance` enforces type safety against AWS SDK manifest. `HeadObjectOutput.ContentLength` is `*int64`, so storing it as `string` breaks arithmetic operations (range requests, progress bars) and adds unnecessary parsing overhead.

---

### B08 — Incomplete Enum Handling (CTR-strictness-parity)

**Bug:** ACL validation only handles `private` and `public-read`, missing other valid values.
**Expected violation:** `CTR-strictness-parity`

```go
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
```

**Why Stricture catches this:** `CTR-strictness-parity` compares enum validation against AWS SDK enum definitions. `ObjectCannedACL` has 7 valid values, but this implementation only accepts 2, rejecting legitimate ACL configurations like `bucket-owner-full-control`.

---

### B09 — Missing Range Validation (CTR-strictness-parity)

**Bug:** Presigned URL expiry accepts 0 seconds or 8 days (exceeds AWS 7-day maximum).
**Expected violation:** `CTR-strictness-parity`

```go
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
```

**Why Stricture catches this:** `CTR-strictness-parity` enforces AWS API constraints. The AWS SDK allows presigned URL expiration between 1 second and 7 days. This implementation accepts `duration=0` (unusable URL) or `duration > 7 days` (AWS API rejects it with "InvalidArgument").

---

### B10 — Format Not Validated (CTR-strictness-parity)

**Bug:** No S3 key format validation (leading slash, max 1024 bytes UTF-8).
**Expected violation:** `CTR-strictness-parity`

```go
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
```

**Why Stricture catches this:** `CTR-strictness-parity` validates field formats against AWS documentation. S3 keys must not start with `/`, must be max 1024 bytes in UTF-8, and should avoid path traversal sequences. The perfect version checks `len(key) > 1024` and rejects leading slashes.

---

### B11 — Precision Loss on Large Numbers (CTR-strictness-parity)

**Bug:** `ContentLength` truncated from `int64` to `int32` (loses precision for files > 2GB).
**Expected violation:** `CTR-strictness-parity`

```go
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
```

**Why Stricture catches this:** `CTR-strictness-parity` enforces numeric precision rules. S3 supports objects up to 5TB (5,497,558,138,880 bytes), requiring `int64`. Truncating to `int32` (max: 2,147,483,647 bytes) causes overflow for files > 2GB, resulting in negative sizes or corruption.

---

### B12 — Nullable Field Crashes (CTR-response-shape)

**Bug:** Dereferences `*string ETag` without nil check.
**Expected violation:** `CTR-response-shape`

```go
// GetObjectMetadataBuggy crashes on nil ETag.
func (c *Client) GetObjectMetadataBuggy(ctx context.Context, bucket, key string) (string, error) {
	if err := validateBucketName(bucket); err != nil {
		return "", err
	}
	if err := validateKey(key); err != nil {
		return "", err
	}

	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	output, err := c.s3Client.HeadObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("HeadObject failed: %w", err)
	}

	// BUG: ETag is *string in AWS SDK (nullable)
	// Dereferences without nil check → panic if ETag is nil
	etag := *output.ETag // CRASH: panic: invalid memory address or nil pointer dereference

	return etag, nil
}
```

**Why Stricture catches this:** `CTR-response-shape` validates nullable field handling. AWS SDK uses `*string` for optional fields like `ETag`. While S3 typically returns ETags, certain edge cases (multipart upload in progress, server error) can result in nil. The perfect version uses `aws.ToString(output.ETag)`.

---

### B13 — Missing S3 Event Notification Signature Verification (CTR-request-shape)

**Bug:** No signature verification on S3 event notifications received via SNS/SQS webhook.
**Expected violation:** `CTR-request-shape`

```go
// S3Event represents S3 event notification structure.
type S3Event struct {
	Records []struct {
		EventName string `json:"eventName"`
		S3        struct {
			Bucket struct {
				Name string `json:"name"`
			} `json:"bucket"`
			Object struct {
				Key  string `json:"key"`
				Size int64  `json:"size"`
			} `json:"object"`
		} `json:"s3"`
	} `json:"Records"`
}

// HandleS3EventBuggy processes S3 event without signature verification.
func HandleS3EventBuggy(w http.ResponseWriter, r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	// BUG: No SNS signature verification
	// If S3 → SNS → HTTP endpoint, must verify:
	// - x-amz-sns-message-type header
	// - SignatureVersion, Signature, SigningCertURL fields
	// - Download cert from SigningCertURL, verify signature

	var event S3Event
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	// Process event without authentication — vulnerable to spoofing
	for _, record := range event.Records {
		fmt.Printf("Event: %s, Bucket: %s, Key: %s\n",
			record.EventName, record.S3.Bucket.Name, record.S3.Object.Key)
	}

	w.WriteHeader(http.StatusOK)
	return nil
}
```

**Why Stricture catches this:** `CTR-request-shape` enforces webhook signature verification. S3 event notifications via SNS include `Signature` and `SigningCertURL` fields. Without verification, attackers can forge events to trigger unauthorized operations (delete objects, start workflows).

---

### B14 — Pagination Terminated Early (CTR-response-shape)

**Bug:** `ListObjectsV2` ignores `IsTruncated` and `ContinuationToken`, stops after first page.
**Expected violation:** `CTR-response-shape`

```go
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
```

**Why Stricture catches this:** `CTR-response-shape` validates pagination field usage. `ListObjectsV2Output` includes `IsTruncated *bool` and `NextContinuationToken *string` to signal more results. Ignoring these fields causes silent data loss when buckets contain > 1000 objects.

---

### B15 — Race Condition on Concurrent Uploads (CTR-request-shape)

**Bug:** No conditional put (`If-None-Match`) to prevent overwriting concurrent uploads.
**Expected violation:** `CTR-request-shape`

```go
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
```

**Why Stricture catches this:** `CTR-request-shape` detects check-then-act patterns without atomic operations. S3 supports conditional puts via `If-None-Match: "*"` (fails if object exists), but this implementation uses separate `HeadObject` → `PutObject` calls. Two concurrent uploads both pass the existence check and overwrite each other.

---
