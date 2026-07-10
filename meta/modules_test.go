package meta_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.proteos.ai/sdk/meta"
	metamodel "go.proteos.ai/model/meta"
)

func sampleModule() metamodel.Module {
	return metamodel.Module{
		Slug:      "billing",
		Name:      "Billing",
		Version:   "1.0.0",
		FileId:    "f1",
		Status:    "active",
		CreatedBy: validSource(),
		UpdatedBy: validSource(),
	}
}

func TestModuleService_Get(t *testing.T) {
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/meta/v1/modules/billing", r.URL.Path)
		_ = json.NewEncoder(w).Encode(sampleModule())
	})
	got, err := m.Modules.Get(context.Background(), "billing")
	require.NoError(t, err)
	require.Equal(t, "Billing", got.Name)
}

func TestModuleService_ListPage(t *testing.T) {
	var seen string
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.String()
		_, _ = w.Write([]byte(`{"meta":{"page":0,"page_size":10,"items_total":1,"pages_total":1},"data":[]}`))
	})
	_, err := m.Modules.ListPage(context.Background(), &meta.ListModulesOptions{
		ListOptions: meta.ListOptions{Page: 0, PageSize: 10},
		Version:     "1.0.0",
	})
	require.NoError(t, err)
	require.Contains(t, seen, "/meta/v1/modules")
	require.Contains(t, seen, "version=1.0.0")
}

func TestModuleService_Activate_Patches(t *testing.T) {
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		require.Equal(t, "/meta/v1/modules/billing/activate", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, m.Modules.Activate(context.Background(), "billing"))
}

func TestModuleService_Deactivate_Patches(t *testing.T) {
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		require.Equal(t, "/meta/v1/modules/billing/deactivate", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, m.Modules.Deactivate(context.Background(), "billing"))
}

func TestModuleService_Delete(t *testing.T) {
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		require.Equal(t, "/meta/v1/modules/billing", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, m.Modules.Delete(context.Background(), "billing"))
}

func TestModuleService_Deploy_BuildsMultipart(t *testing.T) {
	var fields map[string][]string
	var fileBytes []byte
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/meta/v1/modules/deploy", r.URL.Path)
		require.True(t, strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data"))
		require.NoError(t, r.ParseMultipartForm(32<<20))
		fields = r.MultipartForm.Value
		hdr := r.MultipartForm.File["file"][0]
		f, err := hdr.Open()
		require.NoError(t, err)
		defer f.Close()
		fileBytes, _ = io.ReadAll(f)
		_ = json.NewEncoder(w).Encode(sampleModule())
	})

	req := meta.DeployModuleRequest{
		Slug: "billing", Version: "1.0.0", Name: "Billing",
		Description: "billing engine",
	}
	got, err := m.Modules.Deploy(context.Background(), req, strings.NewReader("WASM-BYTES"), "billing.wasm")
	require.NoError(t, err)
	require.Equal(t, "Billing", got.Name)
	require.Contains(t, fields["metadata"][0], `"slug":"billing"`)
	require.Equal(t, "WASM-BYTES", string(fileBytes))
}

func TestModuleService_Download_StreamsBody(t *testing.T) {
	payload := strings.Repeat("X", 256)
	calls := 0
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		switch r.URL.Path {
		case "/meta/v1/modules/billing":
			_ = json.NewEncoder(w).Encode(sampleModule())
		case "/meta/v1/modules/billing/download":
			_, _ = io.WriteString(w, payload)
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})
	body, mod, err := m.Modules.Download(context.Background(), "billing")
	require.NoError(t, err)
	defer body.Close()
	require.Equal(t, "Billing", mod.Name)
	got, err := io.ReadAll(body)
	require.NoError(t, err)
	require.Equal(t, payload, string(got))
	require.Equal(t, 2, calls, "Download fetches metadata then streams the bundle")
}
