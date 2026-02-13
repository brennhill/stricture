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
