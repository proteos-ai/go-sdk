package account_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	sdk "go.proteos.ai/sdk"
	"go.proteos.ai/sdk/account"
)

func sampleRole() account.Role {
	return account.Role{Slug: "admin", Name: "Administrator", OrgID: "org-1", Description: "full access"}
}

func TestRoleService_Get_Success(t *testing.T) {
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/accounts/v1/roles/admin", r.URL.Path)
		_ = json.NewEncoder(w).Encode(sampleRole())
	})
	got, err := a.Roles.Get(context.Background(), "admin")
	require.NoError(t, err)
	require.Equal(t, "admin", got.Slug)
}

func TestRoleService_Get_NotFound(t *testing.T) {
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":"not_found","message":"x"}`))
	})
	_, err := a.Roles.Get(context.Background(), "missing")
	require.True(t, sdk.IsNotFound(err))
}

func TestRoleService_ListPage_BuildsURL(t *testing.T) {
	var seen string
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.String()
		_, _ = w.Write([]byte(`{"meta":{"page":0,"page_size":10,"items_total":0,"pages_total":0},"data":[]}`))
	})
	_, err := a.Roles.ListPage(context.Background(), &account.ListRolesOptions{
		ListOptions:    account.ListOptions{Page: 0, PageSize: 10},
		OrgID: "org-1",
	})
	require.NoError(t, err)
	require.Contains(t, seen, "/accounts/v1/roles")
	require.Contains(t, seen, "org_id=org-1")
}

func TestRoleService_Create_PostsBody(t *testing.T) {
	var body []byte
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		body, _ = io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(sampleRole())
	})
	_, err := a.Roles.Create(context.Background(), account.CreateRoleRequest{
		OrgID: "org-1", Slug: "admin", Name: "Administrator", Description: "full",
	})
	require.NoError(t, err)
	require.Contains(t, string(body), `"slug":"admin"`)
	require.Contains(t, string(body), `"org_id":"org-1"`)
}

func TestRoleService_Update_PatchesBody(t *testing.T) {
	var body []byte
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		require.Equal(t, "/accounts/v1/roles/admin", r.URL.Path)
		body, _ = io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(sampleRole())
	})
	desc := "elevated access"
	_, err := a.Roles.Update(context.Background(), "admin", account.UpdateRoleRequest{Description: &desc})
	require.NoError(t, err)
	require.Contains(t, string(body), `"description":"elevated access"`)
	require.NotContains(t, string(body), `"name"`)
}

func TestRoleService_Delete(t *testing.T) {
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		require.Equal(t, "/accounts/v1/roles/admin", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, a.Roles.Delete(context.Background(), "admin"))
}

func TestRoleService_GetPermissions_Iterates(t *testing.T) {
	pages := [][]account.RoleEntityPermission{
		{{ID: "p1", RoleSlug: "admin", EntitySlug: "customer", Permission: account.PermissionWrite}},
		{{ID: "p2", RoleSlug: "admin", EntitySlug: "order", Permission: account.PermissionRead}},
	}
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Contains(t, r.URL.Path, "/accounts/v1/roles/admin/permissions")
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
	all, err := a.Roles.GetPermissions("admin", &account.ListRolePermissionsOptions{ListOptions: account.ListOptions{PageSize: 1}}).All(context.Background())
	require.NoError(t, err)
	require.Len(t, all, 2)
	require.Equal(t, account.PermissionWrite, all[0].Permission)
}

func TestRoleService_AssignPermission_PostsBody(t *testing.T) {
	var body []byte
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/accounts/v1/roles/admin/permissions", r.URL.Path)
		body, _ = io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(account.RoleEntityPermission{ID: "p1", RoleSlug: "admin", EntitySlug: "customer", Permission: account.PermissionWrite})
	})
	got, err := a.Roles.AssignPermission(context.Background(), "admin", account.AssignPermissionRequest{
		EntitySlug: "customer", Permission: account.PermissionWrite,
	})
	require.NoError(t, err)
	require.Equal(t, "p1", got.ID)
	require.Contains(t, string(body), `"entity_slug":"customer"`)
	require.Contains(t, string(body), `"permission":"write"`)
}

func TestRoleService_RemovePermission(t *testing.T) {
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		require.Equal(t, "/accounts/v1/roles/admin/permissions/p1", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, a.Roles.RemovePermission(context.Background(), "admin", "p1"))
}
