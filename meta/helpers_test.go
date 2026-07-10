package meta_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	sdk "go.proteos.ai/sdk"
	"go.proteos.ai/sdk/meta"
	"go.proteos.ai/model/common"
)

// validSource returns a non-empty audit UserRef for test fixtures.
func validSource() common.UserRef {
	return common.UserRef{Type: common.UserTypePerson, Id: "sys"}
}

// auditAt returns audit fields populated with the given timestamp and a
// valid Source.
func auditAt(when time.Time) (time.Time, time.Time, common.UserRef, common.UserRef) {
	s := validSource()
	return when, when, s, s
}

// newClient starts a test server and returns a meta.Client pointed at it.
func newClient(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *meta.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := sdk.NewClient(sdk.WithBaseURL(srv.URL), sdk.WithToken("t"))
	require.NoError(t, err)
	return srv, meta.New(c)
}

// queryPage parses the "page" query param into an int (default 0).
func queryPage(r *http.Request) int {
	v := r.URL.Query().Get("page")
	if v == "" {
		return 0
	}
	n, _ := strconv.Atoi(v)
	return n
}
