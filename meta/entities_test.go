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

func sampleEntity() metamodel.Entity {
	return metamodel.Entity{
		Slug:        "customer",
		Name:        "Customer",
		Description: "people who buy stuff",
		IsRemote:    false,
		ModuleSlug:  "core",
		Attributes:  []metamodel.Attribute{},
		CreatedBy:   validSource(),
		UpdatedBy:   validSource(),
	}
}

func TestEntityService_Get_Success(t *testing.T) {
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/meta/v1/entities/customer", r.URL.Path)
		_ = json.NewEncoder(w).Encode(sampleEntity())
	})
	got, err := m.Entities.Get(context.Background(), "customer")
	require.NoError(t, err)
	require.Equal(t, "Customer", got.Name)
}

func TestEntityService_Get_NotFound(t *testing.T) {
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	_, err := m.Entities.Get(context.Background(), "missing")
	require.True(t, sdk.IsNotFound(err))
}

func TestEntityService_GetWithSchema_AppendsQuery(t *testing.T) {
	var seen string
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(metamodel.EntityWithSchema{Entity: sampleEntity(), Schema: map[string]any{"type": "object"}})
	})
	got, err := m.Entities.GetWithSchema(context.Background(), "customer")
	require.NoError(t, err)
	require.Contains(t, seen, "with_schema=true")
	require.NotNil(t, got.Schema)
}

func TestEntityService_ListPage_BuildsURL(t *testing.T) {
	var seenURL string
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		seenURL = r.URL.String()
		_, _ = w.Write([]byte(`{"meta":{"page":0,"page_size":10,"items_total":0,"pages_total":0},"data":[]}`))
	})
	_, err := m.Entities.ListPage(context.Background(), &meta.ListEntitiesOptions{
		ListOptions: meta.ListOptions{Page: 0, PageSize: 10},
		Slug:        "customer",
	})
	require.NoError(t, err)
	require.Contains(t, seenURL, "/meta/v1/entities")
	require.Contains(t, seenURL, "slug=customer")
	require.Contains(t, seenURL, "page=0")
	require.Contains(t, seenURL, "page_size=10")
}

func TestEntityService_ListPageWithSchema_AppendsWithSchema(t *testing.T) {
	var seen string
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"meta":{"page":0,"page_size":10,"items_total":0,"pages_total":0},"data":[]}`))
	})
	_, err := m.Entities.ListPageWithSchema(context.Background(), nil)
	require.NoError(t, err)
	require.Contains(t, seen, "with_schema=true")
}

func TestEntityService_List_Iterates(t *testing.T) {
	pages := [][]metamodel.Entity{
		{{Slug: "customer", Name: "Customer", CreatedBy: validSource(), UpdatedBy: validSource()}},
		{{Slug: "order", Name: "Order", CreatedBy: validSource(), UpdatedBy: validSource()}},
	}
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		page := queryPage(r)
		idx := page
		if idx >= len(pages) {
			idx = len(pages) - 1
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"meta": map[string]any{"page": page, "page_size": 1, "items_total": 2, "pages_total": 2},
			"data": pages[idx],
		})
	})
	all, err := m.Entities.List(&meta.ListEntitiesOptions{ListOptions: meta.ListOptions{PageSize: 1}}).All(context.Background())
	require.NoError(t, err)
	require.Equal(t, []string{"customer", "order"}, []string{all[0].Slug, all[1].Slug})
}

func TestEntityService_Create_PostsBody(t *testing.T) {
	var body []byte
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/meta/v1/entities", r.URL.Path)
		body, _ = io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(sampleEntity())
	})
	_, err := m.Entities.Create(context.Background(), meta.CreateEntityRequest{
		Slug: "customer", Name: "Customer", IsRemote: false, ModuleSlug: "core",
		Description: "people who buy stuff", Attributes: []metamodel.Attribute{},
	})
	require.NoError(t, err)
	require.Contains(t, string(body), `"slug":"customer"`)
	require.Contains(t, string(body), `"module_slug":"core"`)
}

func TestEntityService_Update_PatchesBody(t *testing.T) {
	var body []byte
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		require.Equal(t, "/meta/v1/entities/customer", r.URL.Path)
		body, _ = io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(sampleEntity())
	})
	desc := "new description"
	_, err := m.Entities.Update(context.Background(), "customer", meta.UpdateEntityRequest{Description: &desc})
	require.NoError(t, err)
	require.Contains(t, string(body), `"description":"new description"`)
	require.NotContains(t, string(body), `"name"`)
}

func TestEntityService_Delete(t *testing.T) {
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		require.Equal(t, "/meta/v1/entities/customer", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, m.Entities.Delete(context.Background(), "customer"))
}
