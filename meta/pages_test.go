package meta_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.proteos.ai/sdk/meta"
	metamodel "go.proteos.ai/model/meta"
)

func samplePage() metamodel.Page {
	return metamodel.Page{Slug: "p1", Name: "Customer detail", EntitySlug: "customer", CreatedBy: validSource(), UpdatedBy: validSource()}
}

func TestPageService_Get(t *testing.T) {
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/meta/v1/pages/p1", r.URL.Path)
		_ = json.NewEncoder(w).Encode(samplePage())
	})
	got, err := m.Pages.Get(context.Background(), "p1")
	require.NoError(t, err)
	require.Equal(t, "Customer detail", got.Name)
}

func TestPageService_ListPage(t *testing.T) {
	var seen string
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.String()
		_, _ = w.Write([]byte(`{"meta":{"page":0,"page_size":10,"items_total":0,"pages_total":0},"data":[]}`))
	})
	_, err := m.Pages.ListPage(context.Background(), &meta.ListPagesOptions{
		ListOptions: meta.ListOptions{PageSize: 10},
		EntitySlug:  "customer",
	})
	require.NoError(t, err)
	require.Contains(t, seen, "entity_slug=customer")
}

func TestPageService_Create(t *testing.T) {
	var body []byte
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/meta/v1/pages", r.URL.Path)
		body, _ = io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(samplePage())
	})
	_, err := m.Pages.Create(context.Background(), meta.CreatePageRequest{
		Name: "Customer detail", EntitySlug: "customer",
		Actions: []metamodel.PageAction{},
		Layout: metamodel.PageLayout{
			Version: 1,
			Main:    &metamodel.ColumnElement{Type: metamodel.LayoutElementTypeColumn, Children: []metamodel.LayoutElement{}},
		},
	})
	require.NoError(t, err)
	require.Contains(t, string(body), `"name":"Customer detail"`)
}

func TestPageService_Update(t *testing.T) {
	var body []byte
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		require.Equal(t, "/meta/v1/pages/p1", r.URL.Path)
		body, _ = io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(samplePage())
	})
	name := "Customer profile"
	_, err := m.Pages.Update(context.Background(), "p1", meta.UpdatePageRequest{Name: &name})
	require.NoError(t, err)
	require.Contains(t, string(body), `"name":"Customer profile"`)
}

func TestPageService_Delete(t *testing.T) {
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		require.Equal(t, "/meta/v1/pages/p1", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, m.Pages.Delete(context.Background(), "p1"))
}
