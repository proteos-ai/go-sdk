package account_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	sdk "go.proteos.ai/sdk"
	"go.proteos.ai/sdk/account"
)

func TestOrganizationService_Get(t *testing.T) {
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/accounts/v1/organizations/org-1", r.URL.Path)
		_ = json.NewEncoder(w).Encode(account.Organization{ID: "org-1", Name: "Acme"})
	})
	got, err := a.Organizations.Get(context.Background(), "org-1")
	require.NoError(t, err)
	require.Equal(t, "Acme", got.Name)
}

func TestOrganizationService_NotFound(t *testing.T) {
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	_, err := a.Organizations.Get(context.Background(), "missing")
	require.True(t, sdk.IsNotFound(err))
}

func TestOrganizationService_ListPage(t *testing.T) {
	var seen string
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.String()
		_, _ = w.Write([]byte(`{"meta":{"page":0,"page_size":10,"items_total":1,"pages_total":1},"data":[{"id":"org-1","name":"Acme"}]}`))
	})
	res, err := a.Organizations.ListPage(context.Background(), &account.ListOrganizationsOptions{
		ListOptions: account.ListOptions{Page: 0, PageSize: 10},
		Name:        "Acme",
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(res.Data))
	require.Contains(t, seen, "name=Acme")
}

func TestOrganizationService_Create(t *testing.T) {
	var body []byte
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		body, _ = io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(account.Organization{ID: "org-1", Name: "Acme"})
	})
	_, err := a.Organizations.Create(context.Background(), account.CreateOrganizationRequest{Name: "Acme"})
	require.NoError(t, err)
	require.Contains(t, string(body), `"name":"Acme"`)
}

func TestOrganizationService_Update(t *testing.T) {
	var body []byte
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		require.Equal(t, "/accounts/v1/organizations/org-1", r.URL.Path)
		body, _ = io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(account.Organization{ID: "org-1", Name: "Acme Inc"})
	})
	name := "Acme Inc"
	_, err := a.Organizations.Update(context.Background(), "org-1", account.UpdateOrganizationRequest{Name: &name})
	require.NoError(t, err)
	require.Contains(t, string(body), `"name":"Acme Inc"`)
	require.NotContains(t, string(body), `"description"`)
}

func TestOrganizationService_Delete(t *testing.T) {
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		require.Equal(t, "/accounts/v1/organizations/org-1", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, a.Organizations.Delete(context.Background(), "org-1"))
}

func TestOrganizationService_List_Iterates(t *testing.T) {
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		page := queryPage(r)
		body := fmt.Sprintf(`{"meta":{"page":%d,"page_size":1,"items_total":1,"pages_total":1},"data":[{"id":"org-1","name":"Acme"}]}`, page)
		_, _ = w.Write([]byte(body))
	})
	all, err := a.Organizations.List(nil).All(context.Background())
	require.NoError(t, err)
	require.Len(t, all, 1)
	require.Equal(t, "Acme", all[0].Name)
}
