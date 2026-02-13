// stripe/buggy_client_test_b03.go — Shallow assertion in test.
package stripe

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCharge_Shallow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"ch_123","amount":1000,"currency":"usd","status":"succeeded","created":1234567890}`))
	}))
	defer server.Close()

	client := NewClient("test_key")
	client.baseURL = server.URL

	req := CreateChargeRequest{Amount: 1000, Currency: "usd", Source: "tok_visa"}
	charge, err := client.CreateCharge(req)

	require.NoError(t, err)
	assert.NotNil(t, charge) // BUG: Shallow assertion - doesn't validate fields

	// Missing: charge.ID == "ch_123", charge.Amount == 1000, etc.
}
