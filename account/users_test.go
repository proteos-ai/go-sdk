package account_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	sdk "go.proteos.ai/sdk"
	"go.proteos.ai/sdk/account"
	"go.proteos.ai/model/common"
)

// queryPage parses ?page=N from the request, defaulting to 0 when absent.
func queryPage(r *http.Request) int {
	v := r.URL.Query().Get("page")
	if v == "" {
		return 0
	}
	n, _ := strconv.Atoi(v)
	return n
}

func newClient(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *account.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := sdk.NewClient(sdk.WithBaseURL(srv.URL), sdk.WithToken("t"))
	require.NoError(t, err)
	return srv, account.New(c)
}

func sampleUser() account.User {
	return account.User{
		AuditFields: account.AuditFields{
			CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
			CreatedBy: common.UserRef{Type: common.UserTypePerson, Id: "sys"},
			UpdatedBy: common.UserRef{Type: common.UserTypePerson, Id: "sys"},
		},
		ID:           "u1",
		Email:        "alice@example.com",
		GivenName:    "Alice",
		FamilyName:   "Smith",
		ExternalID:   "ext1",
		DefaultOrgID: "org-1",
	}
}

func TestUserService_Get_Success(t *testing.T) {
	want := sampleUser()
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/accounts/v1/users/u1", r.URL.Path)
		_ = json.NewEncoder(w).Encode(want)
	})
	got, err := a.Users.Get(context.Background(), "u1")
	require.NoError(t, err)
	require.Equal(t, "alice@example.com", got.Email)
	require.Equal(t, want.CreatedBy, got.CreatedBy)
}

func TestUserService_Get_NotFound(t *testing.T) {
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":"not_found","message":"missing"}`))
	})
	_, err := a.Users.Get(context.Background(), "missing")
	require.Error(t, err)
	require.True(t, sdk.IsNotFound(err))
}

func TestUserService_ListPage_BuildsQueryParams(t *testing.T) {
	var seenURL string
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		seenURL = r.URL.String()
		_, _ = w.Write([]byte(`{"meta":{"page":2,"page_size":50,"items_total":0,"pages_total":0},"data":[]}`))
	})
	_, err := a.Users.ListPage(context.Background(), &account.ListUsersOptions{
		ListOptions: account.ListOptions{Page: 2, PageSize: 50, SortBy: "email", SortOrder: "asc"},
		Email:       "alice@example.com",
	})
	require.NoError(t, err)
	require.Contains(t, seenURL, "page=2")
	require.Contains(t, seenURL, "page_size=50")
	require.Contains(t, seenURL, "sort_by=email")
	require.Contains(t, seenURL, "sort_direction=asc")
	require.Contains(t, seenURL, "email=alice%40example.com")
}

func TestUserService_List_IteratesAcrossPages(t *testing.T) {
	users := []account.User{sampleUser(), sampleUser(), sampleUser()}
	users[0].ID, users[1].ID, users[2].ID = "u1", "u2", "u3"
	pages := [][]account.User{{users[0], users[1]}, {users[2]}}

	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		page := queryPage(r)
		var data []account.User
		if page < len(pages) {
			data = pages[page]
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"meta": map[string]any{"page": page, "page_size": 2, "items_total": 3, "pages_total": 2},
			"data": data,
		})
	})

	all, err := a.Users.List(&account.ListUsersOptions{ListOptions: account.ListOptions{PageSize: 2}}).All(context.Background())
	require.NoError(t, err)
	require.Len(t, all, 3)
	require.Equal(t, []string{"u1", "u2", "u3"}, []string{all[0].ID, all[1].ID, all[2].ID})
}

func TestUserService_Create_PostsBody(t *testing.T) {
	want := sampleUser()
	var bodyBytes []byte
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/accounts/v1/users", r.URL.Path)
		bodyBytes, _ = io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(want)
	})
	got, err := a.Users.Create(context.Background(), account.CreateUserRequest{
		GivenName: "Alice", FamilyName: "Smith", Email: "alice@example.com",
	})
	require.NoError(t, err)
	require.Equal(t, "alice@example.com", got.Email)
	require.Contains(t, string(bodyBytes), `"given_name":"Alice"`)
	require.Contains(t, string(bodyBytes), `"email":"alice@example.com"`)
}

func TestUserService_Update_OmitsNilFields(t *testing.T) {
	var seen string
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		b, _ := io.ReadAll(r.Body)
		seen = string(b)
		_ = json.NewEncoder(w).Encode(sampleUser())
	})
	given := "Alicia"
	_, err := a.Users.Update(context.Background(), "u1", account.UpdateUserRequest{GivenName: &given})
	require.NoError(t, err)
	require.Contains(t, seen, `"given_name":"Alicia"`)
	// Nil pointers must be omitted, not sent as null.
	require.NotContains(t, seen, `"family_name"`)
	require.NotContains(t, seen, `"default_org_id"`)
}

func TestUserService_GetRoles_Iterates(t *testing.T) {
	pages := [][]account.UserRoleAssignment{
		{{ID: "a1", UserID: "u1", RoleSlug: "admin"}},
		{{ID: "a2", UserID: "u1", RoleSlug: "viewer"}},
	}
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.True(t, strings.HasSuffix(r.URL.Path, "/accounts/v1/users/u1/roles"))
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
	all, err := a.Users.GetRoles("u1", &account.ListUserRoleAssignmentsOptions{ListOptions: account.ListOptions{PageSize: 1}}).All(context.Background())
	require.NoError(t, err)
	require.Equal(t, []string{"admin", "viewer"}, []string{all[0].RoleSlug, all[1].RoleSlug})
}

func TestUserService_AssignRole_PostsBody(t *testing.T) {
	var body []byte
	want := account.UserRoleAssignment{ID: "a1", UserID: "u1", RoleSlug: "admin"}
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/accounts/v1/users/u1/roles", r.URL.Path)
		body, _ = io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(want)
	})
	got, err := a.Users.AssignRole(context.Background(), "u1", account.AssignRoleRequest{RoleSlug: "admin"})
	require.NoError(t, err)
	require.Equal(t, want.ID, got.ID)
	require.Contains(t, string(body), `"role_slug":"admin"`)
}

func TestUserService_UnassignRole_Deletes(t *testing.T) {
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		require.Equal(t, "/accounts/v1/users/u1/roles/admin", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, a.Users.UnassignRole(context.Background(), "u1", "admin"))
}

func TestUserService_Interface_MatchesImpl(t *testing.T) {
	// Compile-time check via var statement in users.go.
	var _ account.UserServiceAPI = (*account.UserService)(nil)
}

func TestUserService_ListApiKeys_UnwrapsData(t *testing.T) {
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/accounts/v1/users/me/api-keys", r.URL.Path)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []account.ApiKey{{ID: "k1", Name: "ci", TokenPrefix: "pak_8f3kQ2x9", TokenSuffix: "Ab12"}},
		})
	})
	keys, err := a.Users.ListApiKeys(context.Background(), "me")
	require.NoError(t, err)
	require.Len(t, keys, 1)
	require.Equal(t, "k1", keys[0].ID)
	require.Equal(t, "pak_8f3kQ2x9", keys[0].TokenPrefix)
}

func TestUserService_CreateApiKey_ReturnsTokenOnce(t *testing.T) {
	var body []byte
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/accounts/v1/users/me/api-keys", r.URL.Path)
		body, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(account.CreatedApiKey{
			ApiKey: account.ApiKey{ID: "k1", Name: "ci", OrgID: "org-1"},
			Token:  "pak_fulltokenvalue",
		})
	})
	created, err := a.Users.CreateApiKey(context.Background(), "me", account.CreateApiKeyRequest{Name: "ci"})
	require.NoError(t, err)
	require.Equal(t, "pak_fulltokenvalue", created.Token)
	require.Contains(t, string(body), `"name":"ci"`)
	require.NotContains(t, string(body), "expires_at", "nil expiry must be omitted")
}

func TestUserService_DeleteApiKey_Deletes(t *testing.T) {
	_, a := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		require.Equal(t, "/accounts/v1/users/u1/api-keys/k1", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, a.Users.DeleteApiKey(context.Background(), "u1", "k1"))
}
