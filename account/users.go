package account

import (
	"context"
	"net/http"

	sdk "go.proteos.ai/sdk"
)

const usersBasePath = "/accounts/v1/users"

// UserServiceAPI is the contract a UserService satisfies. Provided so callers
// can inject mocks.
type UserServiceAPI interface {
	List(opts *ListUsersOptions) *sdk.PageIterator[User, ListUsersOptions]
	ListPage(ctx context.Context, opts *ListUsersOptions) (sdk.ListResult[User], error)
	Get(ctx context.Context, id string) (User, error)
	Create(ctx context.Context, req CreateUserRequest) (User, error)
	Update(ctx context.Context, id string, req UpdateUserRequest) (User, error)
	GetRoles(userID string, opts *ListUserRoleAssignmentsOptions) *sdk.PageIterator[UserRoleAssignment, ListUserRoleAssignmentsOptions]
	GetRolesPage(ctx context.Context, userID string, opts *ListUserRoleAssignmentsOptions) (sdk.ListResult[UserRoleAssignment], error)
	AssignRole(ctx context.Context, userID string, req AssignRoleRequest) (UserRoleAssignment, error)
	UnassignRole(ctx context.Context, userID, roleSlug string) error
	ListApiKeys(ctx context.Context, userID string) ([]ApiKey, error)
	CreateApiKey(ctx context.Context, userID string, req CreateApiKeyRequest) (CreatedApiKey, error)
	DeleteApiKey(ctx context.Context, userID, keyID string) error
}

// UserService manages users and their role assignments.
type UserService struct{ c *sdk.Client }

var _ UserServiceAPI = (*UserService)(nil)

// List returns a PageIterator that walks every user matching opts.
func (s *UserService) List(opts *ListUsersOptions) *sdk.PageIterator[User, ListUsersOptions] {
	o := ListUsersOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListUsersOptions) (sdk.ListResult[User], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

// ListPage fetches a single page of users.
func (s *UserService) ListPage(ctx context.Context, opts *ListUsersOptions) (sdk.ListResult[User], error) {
	var out sdk.ListResult[User]
	err := s.c.DoWithQuery(ctx, http.MethodGet, usersBasePath, opts, nil, &out)
	return out, err
}

// Get returns the user with the given id.
func (s *UserService) Get(ctx context.Context, id string) (User, error) {
	var out User
	err := s.c.Do(ctx, http.MethodGet, usersBasePath+"/"+id, nil, &out)
	return out, err
}

// Create creates a user.
func (s *UserService) Create(ctx context.Context, req CreateUserRequest) (User, error) {
	var out User
	err := s.c.Do(ctx, http.MethodPost, usersBasePath, req, &out)
	return out, err
}

// Update partially updates a user.
func (s *UserService) Update(ctx context.Context, id string, req UpdateUserRequest) (User, error) {
	var out User
	err := s.c.Do(ctx, http.MethodPatch, usersBasePath+"/"+id, req, &out)
	return out, err
}

// GetRoles returns a PageIterator over a user's role assignments.
func (s *UserService) GetRoles(userID string, opts *ListUserRoleAssignmentsOptions) *sdk.PageIterator[UserRoleAssignment, ListUserRoleAssignmentsOptions] {
	o := ListUserRoleAssignmentsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListUserRoleAssignmentsOptions) (sdk.ListResult[UserRoleAssignment], error) {
		in.Page = page
		return s.GetRolesPage(ctx, userID, &in)
	}, o)
}

// GetRolesPage fetches a single page of role assignments for a user.
func (s *UserService) GetRolesPage(ctx context.Context, userID string, opts *ListUserRoleAssignmentsOptions) (sdk.ListResult[UserRoleAssignment], error) {
	var out sdk.ListResult[UserRoleAssignment]
	err := s.c.DoWithQuery(ctx, http.MethodGet, usersBasePath+"/"+userID+"/roles", opts, nil, &out)
	return out, err
}

// AssignRole grants the role to the user.
func (s *UserService) AssignRole(ctx context.Context, userID string, req AssignRoleRequest) (UserRoleAssignment, error) {
	var out UserRoleAssignment
	err := s.c.Do(ctx, http.MethodPost, usersBasePath+"/"+userID+"/roles", req, &out)
	return out, err
}

// UnassignRole revokes a role from the user.
func (s *UserService) UnassignRole(ctx context.Context, userID, roleSlug string) error {
	return s.c.Do(ctx, http.MethodDelete, usersBasePath+"/"+userID+"/roles/"+roleSlug, nil, nil)
}

type listApiKeysResponse struct {
	Data []ApiKey `json:"data"`
}

// ListApiKeys lists a user's api keys (display hints only — never tokens).
// Pass "me" as userID to target the calling user.
func (s *UserService) ListApiKeys(ctx context.Context, userID string) ([]ApiKey, error) {
	var out listApiKeysResponse
	err := s.c.Do(ctx, http.MethodGet, usersBasePath+"/"+userID+"/api-keys", nil, &out)
	return out.Data, err
}

// CreateApiKey mints a new api key for the user. The response carries the
// full token exactly once; it cannot be read again. Requires a JWT session —
// requests authenticated with an api key are rejected. Pass "me" as userID to
// target the calling user.
func (s *UserService) CreateApiKey(ctx context.Context, userID string, req CreateApiKeyRequest) (CreatedApiKey, error) {
	var out CreatedApiKey
	err := s.c.Do(ctx, http.MethodPost, usersBasePath+"/"+userID+"/api-keys", req, &out)
	return out, err
}

// DeleteApiKey revokes an api key. Other services may keep accepting the key
// for up to their verification-cache TTL (~60s). Pass "me" as userID to
// target the calling user.
func (s *UserService) DeleteApiKey(ctx context.Context, userID, keyID string) error {
	return s.c.Do(ctx, http.MethodDelete, usersBasePath+"/"+userID+"/api-keys/"+keyID, nil, nil)
}
