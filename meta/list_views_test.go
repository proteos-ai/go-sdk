package meta_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.proteos.ai/sdk/meta"
	metamodel "go.proteos.ai/model/meta"
)

func sampleListView() metamodel.ListView {
	return metamodel.ListView{Slug: "vip", ListSlug: "customers-grid", Name: "VIP customers", CreatedBy: validSource(), UpdatedBy: validSource()}
}

func TestListViewService_Get(t *testing.T) {
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/meta/v1/list-views/vip", r.URL.Path)
		_ = json.NewEncoder(w).Encode(sampleListView())
	})
	got, err := m.ListViews.Get(context.Background(), "vip")
	require.NoError(t, err)
	require.Equal(t, "VIP customers", got.Name)
}

func TestListViewService_ListPage(t *testing.T) {
	var seen string
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.String()
		_, _ = w.Write([]byte(`{"meta":{"page":0,"page_size":10,"items_total":0,"pages_total":0},"data":[]}`))
	})
	_, err := m.ListViews.ListPage(context.Background(), &meta.ListListViewsOptions{
		ListOptions: meta.ListOptions{PageSize: 10},
		ListSlug:    "customers-grid",
	})
	require.NoError(t, err)
	require.Contains(t, seen, "list_slug=customers-grid")
}

func TestListViewService_Create_Upsert_Update_Delete(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPost, r.Method)
			require.Equal(t, "/meta/v1/list-views", r.URL.Path)
			_ = json.NewEncoder(w).Encode(sampleListView())
		})
		_, err := m.ListViews.Create(context.Background(), meta.CreateListViewRequest{Slug: "vip", ListSlug: "customers-grid", Name: "VIP"})
		require.NoError(t, err)
	})
	t.Run("upsert", func(t *testing.T) {
		_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/meta/v1/list-views/upsert", r.URL.Path)
			_ = json.NewEncoder(w).Encode(sampleListView())
		})
		_, err := m.ListViews.Upsert(context.Background(), meta.CreateListViewRequest{Slug: "vip", ListSlug: "customers-grid", Name: "VIP"})
		require.NoError(t, err)
	})
	t.Run("update", func(t *testing.T) {
		_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPatch, r.Method)
			require.Equal(t, "/meta/v1/list-views/vip", r.URL.Path)
			_ = json.NewEncoder(w).Encode(sampleListView())
		})
		name := "VIP-2"
		_, err := m.ListViews.Update(context.Background(), "vip", meta.UpdateListViewRequest{Name: &name})
		require.NoError(t, err)
	})
	t.Run("delete", func(t *testing.T) {
		_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodDelete, r.Method)
			require.Equal(t, "/meta/v1/list-views/vip", r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
		})
		require.NoError(t, m.ListViews.Delete(context.Background(), "vip"))
	})
}
