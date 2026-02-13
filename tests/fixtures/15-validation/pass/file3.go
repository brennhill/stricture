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
