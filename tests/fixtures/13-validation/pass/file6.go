// stripe/buggy_client_test_b04.go — Only happy path tests.
package stripe

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// BUG: No negative test cases for:
// - 402 Payment Required (card declined)
// - 404 Not Found
// - 400 Invalid Request (missing currency)
// - 429 Rate Limit
// - Network timeout
// - Malformed JSON response

func TestCreateCharge_HappyPath(t *testing.T) {
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
	require.NotNil(t, charge)
	assert.Equal(t, "ch_123", charge.ID)
	assert.Equal(t, int64(1000), charge.Amount)
	assert.Equal(t, "usd", charge.Currency)
	assert.Equal(t, "succeeded", charge.Status)
}

func TestGetCharge_HappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"ch_123","amount":1000,"currency":"usd","status":"succeeded","created":1234567890}`))
	}))
	defer server.Close()

	client := NewClient("test_key")
	client.baseURL = server.URL

	charge, err := client.GetCharge("ch_123")

	require.NoError(t, err)
	assert.Equal(t, "ch_123", charge.ID)
}
