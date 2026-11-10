package api

import (
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"testing"
)

func TestPendingRouteDoesNotResolveAsTransactionHash(t *testing.T) {
	server := NewServer(&ServerConfig{Chain: NewMockChain(), Wallet: NewMockWallet()})
	request := httptest.NewRequest("GET", "/api/v1/transactions/pending", nil)
	response := httptest.NewRecorder()
	server.router.ServeHTTP(response, request)
	require.Equal(t, 200, response.Code)
	require.Contains(t, response.Body.String(), "pending_transactions")
}

func TestRateLimitIdentityIgnoresUntrustedForwardingHeaders(t *testing.T) {
	for _, peer := range []string{"192.0.2.1:1234", "[2001:db8::1]:1234"} {
		request := httptest.NewRequest("GET", "/health", nil)
		request.RemoteAddr = peer
		request.Header.Set("X-Forwarded-For", "198.51.100.9")
		request.Header.Set("X-Real-IP", "198.51.100.10")
		expected := "192.0.2.1"
		if peer[0] == '[' {
			expected = "2001:db8::1"
		}
		require.Equal(t, expected, remoteClientIP(request))
	}
}
