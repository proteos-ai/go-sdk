package meta_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	sdk "go.proteos.ai/sdk"
	"go.proteos.ai/sdk/meta"
	metamodel "go.proteos.ai/model/meta"
)

func sampleList() metamodel.List {
	return metamodel.List{Slug: "customers-grid", Name: "Customers", EntitySlug: "customer", CreatedBy: validSource(), UpdatedBy: validSource()}
}

func TestListService_Get(t *testing.T) {
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/meta/v1/lists/customers-grid", r.URL.Path)
		_ = json.NewEncoder(w).Encode(sampleList())
	})
	got, err := m.Lists.Get(context.Background(), "customers-grid")
	require.NoError(t, err)
	require.Equal(t, "Customers", got.Name)
}

func TestListService_NotFound(t *testing.T) {
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	_, err := m.Lists.Get(context.Background(), "missing")
	require.True(t, sdk.IsNotFound(err))
}

func TestListService_ListPage(t *testing.T) {
	var seen string
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.String()
		_, _ = w.Write([]byte(`{"meta":{"page":0,"page_size":10,"items_total":0,"pages_total":0},"data":[]}`))
	})
	_, err := m.Lists.ListPage(context.Background(), &meta.ListListsOptions{
		ListOptions: meta.ListOptions{PageSize: 10},
		EntitySlug:  "customer",
	})
	require.NoError(t, err)
	require.Contains(t, seen, "entity_slug=customer")
}

func TestListService_Create(t *testing.T) {
	var body []byte
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/meta/v1/lists", r.URL.Path)
		body, _ = io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(sampleList())
	})
	_, err := m.Lists.Create(context.Background(), meta.CreateListRequest{
		Slug: "customers-grid", EntitySlug: "customer", Name: "Customers",
		Columns: []metamodel.Column{{Attribute: "name", Label: "Name", Width: 200}},
	})
	require.NoError(t, err)
	require.Contains(t, string(body), `"slug":"customers-grid"`)
	require.Contains(t, string(body), `"entity_slug":"customer"`)
}

func TestListService_Upsert_PostsToUpsertPath(t *testing.T) {
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/meta/v1/lists/upsert", r.URL.Path)
		_ = json.NewEncoder(w).Encode(sampleList())
	})
	_, err := m.Lists.Upsert(context.Background(), meta.CreateListRequest{Slug: "x", EntitySlug: "y", Name: "Z"})
	require.NoError(t, err)
}

func TestListService_Update(t *testing.T) {
	var body []byte
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		require.Equal(t, "/meta/v1/lists/customers-grid", r.URL.Path)
		body, _ = io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(sampleList())
	})
	name := "Customers v2"
	_, err := m.Lists.Update(context.Background(), "customers-grid", meta.UpdateListRequest{Name: &name})
	require.NoError(t, err)
	require.Contains(t, string(body), `"name":"Customers v2"`)
}

func TestListService_Delete(t *testing.T) {
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, m.Lists.Delete(context.Background(), "customers-grid"))
}
