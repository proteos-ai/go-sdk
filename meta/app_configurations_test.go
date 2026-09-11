package meta_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.proteos.ai/model/common"
	metamodel "go.proteos.ai/model/meta"
	"go.proteos.ai/sdk/meta"
)

func sampleAppConfiguration() metamodel.AppConfiguration {
	return metamodel.AppConfiguration{Slug: "sales", AppSlug: "sales", MenuSlug: "sales-nav", Home: &metamodel.AppHome{Type: metamodel.AppHomeTypeList, Reference: "companies"}, CreatedBy: validSource(), UpdatedBy: validSource()}
}

func TestAppConfigurationService_Get(t *testing.T) {
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/meta/v1/app-configurations/sales", r.URL.Path)
		_ = json.NewEncoder(w).Encode(sampleAppConfiguration())
	})
	got, err := m.AppConfigurations.Get(context.Background(), "sales")
	require.NoError(t, err)
	require.Equal(t, "companies", got.Home.Reference)
}

func TestAppConfigurationService_ListPage_Filters(t *testing.T) {
	var seen string
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.String()
		_, _ = w.Write([]byte(`{"meta":{"page":0,"page_size":10,"items_total":0,"pages_total":0},"data":[]}`))
	})
	_, err := m.AppConfigurations.ListPage(context.Background(), &meta.ListAppConfigurationsOptions{
		ListOptions: meta.ListOptions{PageSize: 10},
		AppSlug:     "sales",
		ProfileSlug: "sales-rep",
	})
	require.NoError(t, err)
	require.Contains(t, seen, "app_slug=sales")
	require.Contains(t, seen, "profile_slug=sales-rep")
}

// An empty ProfileSlug is omitted (never sent as `profile_slug=`), so the
// default rows are selected through the explicit is_default flag.
func TestAppConfigurationService_ListPage_IsDefault(t *testing.T) {
	var seen string
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.String()
		_, _ = w.Write([]byte(`{"meta":{"page":0,"page_size":10,"items_total":0,"pages_total":0},"data":[]}`))
	})
	isDefault := true
	_, err := m.AppConfigurations.ListPage(context.Background(), &meta.ListAppConfigurationsOptions{
		ListOptions: meta.ListOptions{PageSize: 10},
		AppSlug:     "sales",
		IsDefault:   &isDefault,
	})
	require.NoError(t, err)
	require.Contains(t, seen, "is_default=true")
	require.NotContains(t, seen, "profile_slug")
}

func TestAppConfigurationService_Upsert_PinsSlugAndSendsHome(t *testing.T) {
	var body []byte
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPut, r.Method)
		require.Equal(t, "/meta/v1/app-configurations/sales", r.URL.Path)
		body, _ = io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(sampleAppConfiguration())
	})
	_, err := m.AppConfigurations.Upsert(context.Background(), "sales", meta.CreateAppConfigurationRequest{
		AppSlug: "sales", Home: &metamodel.AppHome{Type: metamodel.AppHomeTypeList, Reference: "companies"}, RecordPages: map[string]string{"company": "company-sales"},
	})
	require.NoError(t, err)
	require.Contains(t, string(body), `"slug":"sales"`)
	require.Contains(t, string(body), `"home":{"type":"list","reference":"companies"}`)
	require.Contains(t, string(body), `"record_pages":{"company":"company-sales"}`)
}

func TestAppConfigurationService_Update_ClearsHomeWithNull(t *testing.T) {
	var body []byte
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		body, _ = io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(sampleAppConfiguration())
	})
	_, err := m.AppConfigurations.Update(context.Background(), "sales", meta.UpdateAppConfigurationRequest{
		Home: common.Optional[metamodel.AppHome]{Present: true, Value: nil},
	})
	require.NoError(t, err)
	require.Contains(t, string(body), `"home":null`)
}

func TestAppConfigurationService_Delete(t *testing.T) {
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		require.Equal(t, "/meta/v1/app-configurations/sales", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, m.AppConfigurations.Delete(context.Background(), "sales"))
}
