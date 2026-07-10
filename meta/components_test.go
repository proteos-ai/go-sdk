package meta_test

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.proteos.ai/sdk/meta"
	metamodel "go.proteos.ai/model/meta"
)

func sampleComponent() metamodel.Component {
	return metamodel.Component{Slug: "btn", Name: "Button", ModuleSlug: "core", CreatedBy: validSource(), UpdatedBy: validSource()}
}

func TestComponentService_Get(t *testing.T) {
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/meta/v1/components/btn", r.URL.Path)
		_ = json.NewEncoder(w).Encode(sampleComponent())
	})
	got, err := m.Components.Get(context.Background(), "btn")
	require.NoError(t, err)
	require.Equal(t, "Button", got.Name)
}

func TestComponentService_ListPage(t *testing.T) {
	var seen string
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.String()
		_, _ = w.Write([]byte(`{"meta":{"page":0,"page_size":10,"items_total":0,"pages_total":0},"data":[]}`))
	})
	_, err := m.Components.ListPage(context.Background(), &meta.ListComponentsOptions{
		ListOptions: meta.ListOptions{PageSize: 10}, ModuleSlug: "core",
	})
	require.NoError(t, err)
	require.Contains(t, seen, "module_slug=core")
}

func TestComponentService_Deploy(t *testing.T) {
	type part struct {
		contentType string
		body        string
	}
	parts := map[string]part{}

	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/meta/v1/components/deploy", r.URL.Path)

		_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		require.NoError(t, err)
		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
			body, _ := io.ReadAll(p)
			parts[p.FormName()] = part{contentType: p.Header.Get("Content-Type"), body: string(body)}
		}
		_ = json.NewEncoder(w).Encode(sampleComponent())
	})

	_, err := m.Components.Deploy(context.Background(),
		meta.CreateComponentRequest{Slug: "btn", Name: "Button", ModuleSlug: "core", PropsSchema: map[string]any{"type": "object"}},
		strings.NewReader("export default function B(){}"), "btn.js",
		strings.NewReader("SOURCE-TARBALL"), "btn-source.tar.gz",
	)
	require.NoError(t, err)

	// metadata part: JSON content type + carries the request (incl. propsSchema).
	require.Equal(t, "application/json", parts["metadata"].contentType)
	require.Contains(t, parts["metadata"].body, `"slug":"btn"`)
	require.Contains(t, parts["metadata"].body, `"props_schema"`)
	// bundle + source parts present with their payloads.
	require.Equal(t, "export default function B(){}", parts["bundle"].body)
	require.Equal(t, "SOURCE-TARBALL", parts["source"].body)
}

func TestComponentService_Create_Upsert_Update_Delete(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPost, r.Method)
			require.Equal(t, "/meta/v1/components", r.URL.Path)
			_ = json.NewEncoder(w).Encode(sampleComponent())
		})
		_, err := m.Components.Create(context.Background(), meta.CreateComponentRequest{Slug: "btn", Name: "Button", ModuleSlug: "core"})
		require.NoError(t, err)
	})
	t.Run("upsert", func(t *testing.T) {
		_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/meta/v1/components/upsert", r.URL.Path)
			_ = json.NewEncoder(w).Encode(sampleComponent())
		})
		_, err := m.Components.Upsert(context.Background(), meta.CreateComponentRequest{Slug: "btn", Name: "Button", ModuleSlug: "core"})
		require.NoError(t, err)
	})
	t.Run("update", func(t *testing.T) {
		_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPatch, r.Method)
			require.Equal(t, "/meta/v1/components/btn", r.URL.Path)
			_ = json.NewEncoder(w).Encode(sampleComponent())
		})
		name := "BigButton"
		_, err := m.Components.Update(context.Background(), "btn", meta.UpdateComponentRequest{Name: &name})
		require.NoError(t, err)
	})
	t.Run("delete", func(t *testing.T) {
		_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodDelete, r.Method)
			require.Equal(t, "/meta/v1/components/btn", r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
		})
		require.NoError(t, m.Components.Delete(context.Background(), "btn"))
	})
}
