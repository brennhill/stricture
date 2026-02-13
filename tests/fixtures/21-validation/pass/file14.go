// Go sends int64 (safe for large monetary values)
type Order struct {
	TotalAmount int64 `json:"total_amount"`  // CORRECT: cents as int64
}

// Example: 9007199254740992 cents ($90,071,992,547,409.92)
