package functions_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	sdk "go.proteos.ai/sdk"
	"go.proteos.ai/sdk/functions"
	"go.proteos.ai/model/common"
	functionsmodel "go.proteos.ai/model/functions"
)

// validSource returns a non-empty audit UserRef for test fixtures.
func validSource() common.UserRef {
	return common.UserRef{Type: common.UserTypePerson, Id: "sys"}
}

// sampleHook returns a representative Hook for tests.
func sampleHook() functionsmodel.Hook {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	return functionsmodel.Hook{
		Slug:       "validate-invoice",
		OrgId:      "org-1",
		ModuleSlug: "ta-proteos-documents",
		EntitySlug: "invoice",
		Event:      functionsmodel.HookEventBeforeCreate,
		IsActive:   true,
		FileId:     "0001.aaa",
		CreatedAt:  now,
		CreatedBy:  validSource(),
		UpdatedAt:  now,
		UpdatedBy:  validSource(),
	}
}

// sampleAction returns a representative entity-scoped Action.
func sampleAction() functionsmodel.Action {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	entity := "invoice"
	return functionsmodel.Action{
		Slug:       "send-invoice",
		OrgId:      "org-1",
		ModuleSlug: "ta-proteos-documents",
		Scope:      functionsmodel.ActionScopeEntity,
		EntitySlug: &entity,
		Name:       "send-invoice",
		IsActive:   true,
		FileId:     "0002.bbb",
		CreatedAt:  now,
		CreatedBy:  validSource(),
		UpdatedAt:  now,
		UpdatedBy:  validSource(),
	}
}

// newClient starts a test server and returns a functions.Client pointed at it.
func newClient(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *functions.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := sdk.NewClient(sdk.WithBaseURL(srv.URL), sdk.WithToken("t"))
	require.NoError(t, err)
	return srv, functions.New(c)
}
