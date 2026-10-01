package conversation_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	sdk "go.proteos.ai/sdk"
	"go.proteos.ai/sdk/conversation"
)

// newClient starts a test server and returns a conversation.Client pointed at it.
func newClient(t *testing.T, handler http.HandlerFunc) *conversation.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := sdk.NewClient(sdk.WithBaseURL(srv.URL), sdk.WithToken("t"))
	require.NoError(t, err)
	return conversation.New(c)
}

const emptyPage = `{"meta":{"page":0,"page_size":100,"items_total":0,"pages_total":0},"data":[]}`
