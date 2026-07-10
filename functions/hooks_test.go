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
)

func TestHookService_Get(t *testing.T) {
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/functions/v1/hooks/validate-invoice", r.URL.Path)
		_ = json.NewEncoder(w).Encode(sampleHook())
	})
	got, err := f.Hooks.Get(context.Background(), "validate-invoice")
	require.NoError(t, err)
	require.Equal(t, "validate-invoice", got.Slug)
	require.Equal(t, functionsmodel.HookEventBeforeCreate, got.Event)
}

func TestHookService_ListPage_FiltersInQuery(t *testing.T) {
	var seen string
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.String()
		_, _ = w.Write([]byte(`{"meta":{"page":0,"page_size":10,"items_total":1,"pages_total":1},"data":[]}`))
	})
	isActive := true
	_, err := f.Hooks.ListPage(context.Background(), &functions.ListHooksOptions{
		ListOptions: functions.ListOptions{Page: 0, PageSize: 10},
		EntitySlug:  "invoice",
		Event:       functionsmodel.HookEventBeforeCreate,
		IsActive:    &isActive,
	})
	require.NoError(t, err)
	require.Contains(t, seen, "/functions/v1/hooks")
	require.Contains(t, seen, "entity=invoice")
	require.Contains(t, seen, "event=before_create")
	require.Contains(t, seen, "is_active=true")
}

func TestHookService_Deploy_BuildsMultipartPutBySlug(t *testing.T) {
	var fields map[string][]string
	var wasmBytes []byte
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPut, r.Method, "Deploy upserts via PUT /:slug")
		require.Equal(t, "/functions/v1/hooks/validate-invoice", r.URL.Path)
		require.True(t, strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data"))
		require.NoError(t, r.ParseMultipartForm(32<<20))
		fields = r.MultipartForm.Value
		hdr := r.MultipartForm.File["wasm"][0]
		fp, err := hdr.Open()
		require.NoError(t, err)
		defer fp.Close()
		wasmBytes, _ = io.ReadAll(fp)
		_ = json.NewEncoder(w).Encode(sampleHook())
	})

	req := functions.DeployHookRequest{
		Slug:       "validate-invoice",
		ModuleSlug: "ta-proteos-documents",
		EntitySlug: "invoice",
		Event:      functionsmodel.HookEventBeforeCreate,
	}
	got, err := f.Hooks.Deploy(context.Background(), req, strings.NewReader("WASM-BYTES"), "validate-invoice.wasm")
	require.NoError(t, err)
	require.Equal(t, "validate-invoice", got.Slug)
	require.Contains(t, fields["metadata"][0], `"slug":"validate-invoice"`)
	require.Contains(t, fields["metadata"][0], `"event":"before_create"`)
	require.Equal(t, "WASM-BYTES", string(wasmBytes))
}

func TestHookService_Activate_Patches(t *testing.T) {
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		require.Equal(t, "/functions/v1/hooks/validate-invoice/activate", r.URL.Path)
		_ = json.NewEncoder(w).Encode(sampleHook())
	})
	got, err := f.Hooks.Activate(context.Background(), "validate-invoice")
	require.NoError(t, err)
	require.True(t, got.IsActive)
}

func TestHookService_Deactivate_Patches(t *testing.T) {
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		require.Equal(t, "/functions/v1/hooks/validate-invoice/deactivate", r.URL.Path)
		hook := sampleHook()
		hook.IsActive = false
		_ = json.NewEncoder(w).Encode(hook)
	})
	got, err := f.Hooks.Deactivate(context.Background(), "validate-invoice")
	require.NoError(t, err)
	require.False(t, got.IsActive)
}

func TestHookService_Update_Patch(t *testing.T) {
	var body map[string]any
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		require.Equal(t, "/functions/v1/hooks/validate-invoice", r.URL.Path)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		_ = json.NewEncoder(w).Encode(sampleHook())
	})
	isActive := false
	_, err := f.Hooks.Update(context.Background(), "validate-invoice", functions.PatchHookRequest{IsActive: &isActive})
	require.NoError(t, err)
	require.Equal(t, false, body["is_active"])
}

func TestHookService_Delete(t *testing.T) {
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		require.Equal(t, "/functions/v1/hooks/validate-invoice", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, f.Hooks.Delete(context.Background(), "validate-invoice"))
}

func TestHookService_Logs_PassesSinceAndLevelQuery(t *testing.T) {
	var seen string
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		seen = r.URL.String()
		_, _ = w.Write([]byte(`{"data":[{"timestamp":"2026-01-02T03:04:05Z","level":"info","message":"hi","resource_type":"hook","resource_slug":"validate-invoice"}]}`))
	})
	entries, err := f.Hooks.Logs(context.Background(), "validate-invoice", &functions.ListHookLogsOptions{
		Since: "10m",
		Level: "info",
	})
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, "hook", entries[0].ResourceType)
	require.Equal(t, "validate-invoice", entries[0].ResourceSlug)
	require.Contains(t, seen, "/functions/v1/hooks/validate-invoice/logs")
	require.Contains(t, seen, "since=10m")
	require.Contains(t, seen, "level=info")
}

func TestHookService_Logs_RejectsFollowOnSnapshot(t *testing.T) {
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("snapshot Logs must not hit the network when Follow=true")
	})
	_, err := f.Hooks.Logs(context.Background(), "validate-invoice", &functions.ListHookLogsOptions{
		Follow: true,
	})
	require.Error(t, err)
}

func TestHookService_TailLogs_StreamsNDJSON(t *testing.T) {
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Contains(t, r.URL.String(), "/functions/v1/hooks/validate-invoice/logs")
		require.Contains(t, r.URL.RawQuery, "follow=true")
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"timestamp":"2026-01-02T03:04:05Z","level":"info","message":"first","resource_type":"hook","resource_slug":"validate-invoice"}` + "\n"))
		_, _ = w.Write([]byte(`{"timestamp":"2026-01-02T03:04:06Z","level":"warn","message":"second","resource_type":"hook","resource_slug":"validate-invoice"}` + "\n"))
	})

	stream, err := f.Hooks.TailLogs(context.Background(), "validate-invoice", &functions.ListHookLogsOptions{
		Since: "10m",
	})
	require.NoError(t, err)
	defer stream.Close()

	e1, ok, err := stream.Next()
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "first", e1.Message)

	e2, ok, err := stream.Next()
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "second", e2.Message)

	_, ok, err = stream.Next()
	require.NoError(t, err)
	require.False(t, ok)
}
