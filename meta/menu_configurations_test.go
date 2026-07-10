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

func sampleMenu() metamodel.MenuConfiguration {
	return metamodel.MenuConfiguration{Slug: "mc1", Name: "Main", AppSlug: "console", IsDefault: true, CreatedBy: validSource(), UpdatedBy: validSource()}
}

func TestMenuConfigurationService_Get(t *testing.T) {
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/meta/v1/menu-configurations/mc1", r.URL.Path)
		_ = json.NewEncoder(w).Encode(sampleMenu())
	})
	got, err := m.MenuConfigurations.Get(context.Background(), "mc1")
	require.NoError(t, err)
	require.Equal(t, "Main", got.Name)
}

func TestMenuConfigurationService_ListPage_FlagsBoolean(t *testing.T) {
	var seen string
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.String()
		_, _ = w.Write([]byte(`{"meta":{"page":0,"page_size":10,"items_total":0,"pages_total":0},"data":[]}`))
	})
	yes := true
	_, err := m.MenuConfigurations.ListPage(context.Background(), &meta.ListMenuConfigurationsOptions{
		ListOptions: meta.ListOptions{PageSize: 10},
		AppSlug:     "console",
		IsDefault:   &yes,
	})
	require.NoError(t, err)
	require.Contains(t, seen, "app_slug=console")
	require.Contains(t, seen, "is_default=true")
}

func TestMenuConfigurationService_Create(t *testing.T) {
	var body []byte
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/meta/v1/menu-configurations", r.URL.Path)
		body, _ = io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(sampleMenu())
	})
	_, err := m.MenuConfigurations.Create(context.Background(), meta.CreateMenuConfigurationRequest{
		Name: "Main", AppSlug: "console", IsDefault: true, Items: []metamodel.MenuItem{},
	})
	require.NoError(t, err)
	require.Contains(t, string(body), `"app_slug":"console"`)
}

func TestMenuConfigurationService_Update(t *testing.T) {
	var body []byte
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		require.Equal(t, "/meta/v1/menu-configurations/mc1", r.URL.Path)
		body, _ = io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(sampleMenu())
	})
	name := "Main v2"
	_, err := m.MenuConfigurations.Update(context.Background(), "mc1", meta.UpdateMenuConfigurationRequest{Name: &name})
	require.NoError(t, err)
	require.Contains(t, string(body), `"name":"Main v2"`)
}

func TestMenuConfigurationService_Delete(t *testing.T) {
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		require.Equal(t, "/meta/v1/menu-configurations/mc1", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, m.MenuConfigurations.Delete(context.Background(), "mc1"))
}
