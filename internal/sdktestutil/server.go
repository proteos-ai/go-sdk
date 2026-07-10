// Package sdktestutil provides shared test helpers for the SDK and its
// sub-packages. Internal — not part of the public API.
package sdktestutil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	sdk "go.proteos.ai/sdk"
)

// NewServer starts an httptest.Server using handler and returns it together
// with an *sdk.Client pointed at its URL. The server is closed via
// t.Cleanup; callers don't need to defer Close.
func NewServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *sdk.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := sdk.NewClient(sdk.WithBaseURL(srv.URL), sdk.WithToken("test-token"))
	require.NoError(t, err)
	return srv, c
}
