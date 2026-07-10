package functions_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.proteos.ai/sdk/functions"
	functionsmodel "go.proteos.ai/model/functions"
	functionsapi "go.proteos.ai/model/functions/api"
	metamodel "go.proteos.ai/model/meta"
)

func TestActionService_Get(t *testing.T) {
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/functions/v1/actions/send-invoice", r.URL.Path)
		_ = json.NewEncoder(w).Encode(sampleAction())
	})
	got, err := f.Actions.Get(context.Background(), "send-invoice")
	require.NoError(t, err)
	require.Equal(t, "send-invoice", got.Slug)
	require.Equal(t, functionsmodel.ActionScopeEntity, got.Scope)
}

func TestActionService_ListPage_FiltersInQuery(t *testing.T) {
	var seen string
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.String()
		_, _ = w.Write([]byte(`{"meta":{"page":0,"page_size":10,"items_total":1,"pages_total":1},"data":[]}`))
	})
	_, err := f.Actions.ListPage(context.Background(), &functions.ListActionsOptions{
		ListOptions: functions.ListOptions{Page: 0, PageSize: 10},
		Scope:       functionsmodel.ActionScopeGlobal,
	})
	require.NoError(t, err)
	require.Contains(t, seen, "/functions/v1/actions")
	require.Contains(t, seen, "scope=global")
}

func TestActionService_Deploy_EntityScoped_RoundTripsParamsSchema(t *testing.T) {
	var fields map[string][]string
	var wasmBytes []byte
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPut, r.Method)
		require.Equal(t, "/functions/v1/actions/send-invoice", r.URL.Path)
		require.True(t, strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data"))
		require.NoError(t, r.ParseMultipartForm(32<<20))
		fields = r.MultipartForm.Value
		hdr := r.MultipartForm.File["wasm"][0]
		fp, err := hdr.Open()
		require.NoError(t, err)
		defer fp.Close()
		wasmBytes, _ = io.ReadAll(fp)
		_ = json.NewEncoder(w).Encode(sampleAction())
	})

	entity := "invoice"
	req := functions.DeployActionRequest{
		Slug:       "send-invoice",
		ModuleSlug: "ta-proteos-documents",
		Scope:      functionsmodel.ActionScopeEntity,
		EntitySlug: &entity,
		Name:       "send-invoice",
		Params: []metamodel.Attribute{
			{Name: "recipientEmail", Type: metamodel.AttributeTypeString, IsRequired: true},
		},
		Returns: []metamodel.Attribute{
			{Name: "messageId", Type: metamodel.AttributeTypeString, IsRequired: true},
		},
	}
	got, err := f.Actions.Deploy(context.Background(), req, strings.NewReader("ACTION-BYTES"), "send-invoice.wasm")
	require.NoError(t, err)
	require.Equal(t, "send-invoice", got.Slug)
	require.Contains(t, fields["metadata"][0], `"scope":"entity"`)
	require.Contains(t, fields["metadata"][0], `"entity":"invoice"`)
	require.Contains(t, fields["metadata"][0], `"recipientEmail"`)
	require.Equal(t, "ACTION-BYTES", string(wasmBytes))
}

func TestActionService_Deploy_GlobalOmitsEntity(t *testing.T) {
	var fields map[string][]string
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPut, r.Method)
		require.Equal(t, "/functions/v1/actions/rebuild-search-index", r.URL.Path)
		require.NoError(t, r.ParseMultipartForm(32<<20))
		fields = r.MultipartForm.Value
		action := sampleAction()
		action.Slug = "rebuild-search-index"
		action.Scope = functionsmodel.ActionScopeGlobal
		action.EntitySlug = nil
		_ = json.NewEncoder(w).Encode(action)
	})

	req := functions.DeployActionRequest{
		Slug:       "rebuild-search-index",
		ModuleSlug: "ta-proteos-documents",
		Scope:      functionsmodel.ActionScopeGlobal,
		Name:       "rebuild-search-index",
	}
	_, err := f.Actions.Deploy(context.Background(), req, strings.NewReader("BYTES"), "rebuild.wasm")
	require.NoError(t, err)
	require.Contains(t, fields["metadata"][0], `"scope":"global"`)
	require.NotContains(t, fields["metadata"][0], `"entity"`, "global actions must not serialise an entity field")
}

func TestActionService_Activate_Deactivate_Patches(t *testing.T) {
	t.Run("activate", func(t *testing.T) {
		_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPatch, r.Method)
			require.Equal(t, "/functions/v1/actions/send-invoice/activate", r.URL.Path)
			_ = json.NewEncoder(w).Encode(sampleAction())
		})
		_, err := f.Actions.Activate(context.Background(), "send-invoice")
		require.NoError(t, err)
	})
	t.Run("deactivate", func(t *testing.T) {
		_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPatch, r.Method)
			require.Equal(t, "/functions/v1/actions/send-invoice/deactivate", r.URL.Path)
			_ = json.NewEncoder(w).Encode(sampleAction())
		})
		_, err := f.Actions.Deactivate(context.Background(), "send-invoice")
		require.NoError(t, err)
	})
}

func TestActionService_Delete(t *testing.T) {
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		require.Equal(t, "/functions/v1/actions/send-invoice", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, f.Actions.Delete(context.Background(), "send-invoice"))
}

func TestActionService_ListForEntity_HitsEntityScopedPath(t *testing.T) {
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/functions/v1/entities/invoice/actions", r.URL.Path)
		out := []functionsapi.ActionSummary{
			{Slug: "send-invoice", Name: "send-invoice", Scope: functionsmodel.ActionScopeEntity},
			{Slug: "rebuild-index", Name: "rebuild-index", Scope: functionsmodel.ActionScopeGlobal},
		}
		_ = json.NewEncoder(w).Encode(out)
	})
	got, err := f.Actions.ListForEntity(context.Background(), "invoice")
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, functionsmodel.ActionScopeEntity, got[0].Scope)
	require.Equal(t, functionsmodel.ActionScopeGlobal, got[1].Scope)
}
