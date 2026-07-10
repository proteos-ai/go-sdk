package account

import (
	"context"
	"net/http"

	sdk "go.proteos.ai/sdk"
)

const rolesBasePath = "/accounts/v1/roles"

// RoleServiceAPI is the contract a RoleService satisfies.
type RoleServiceAPI interface {
	List(opts *ListRolesOptions) *sdk.PageIterator[Role, ListRolesOptions]
	ListPage(ctx context.Context, opts *ListRolesOptions) (sdk.ListResult[Role], error)
	Get(ctx context.Context, slug string) (Role, error)
	Create(ctx context.Context, req CreateRoleRequest) (Role, error)
	Update(ctx context.Context, slug string, req UpdateRoleRequest) (Role, error)
	Delete(ctx context.Context, slug string) error
	GetPermissions(roleSlug string, opts *ListRolePermissionsOptions) *sdk.PageIterator[RoleEntityPermission, ListRolePermissionsOptions]
	GetPermissionsPage(ctx context.Context, roleSlug string, opts *ListRolePermissionsOptions) (sdk.ListResult[RoleEntityPermission], error)
	AssignPermission(ctx context.Context, roleSlug string, req AssignPermissionRequest) (RoleEntityPermission, error)
	RemovePermission(ctx context.Context, roleSlug, permissionID string) error
}

// RoleService manages roles and their entity permissions.
type RoleService struct{ c *sdk.Client }

var _ RoleServiceAPI = (*RoleService)(nil)

func (s *RoleService) List(opts *ListRolesOptions) *sdk.PageIterator[Role, ListRolesOptions] {
	o := ListRolesOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListRolesOptions) (sdk.ListResult[Role], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *RoleService) ListPage(ctx context.Context, opts *ListRolesOptions) (sdk.ListResult[Role], error) {
	var out sdk.ListResult[Role]
	err := s.c.DoWithQuery(ctx, http.MethodGet, rolesBasePath, opts, nil, &out)
	return out, err
}

func (s *RoleService) Get(ctx context.Context, slug string) (Role, error) {
	var out Role
	err := s.c.Do(ctx, http.MethodGet, rolesBasePath+"/"+slug, nil, &out)
	return out, err
}

func (s *RoleService) Create(ctx context.Context, req CreateRoleRequest) (Role, error) {
	var out Role
	err := s.c.Do(ctx, http.MethodPost, rolesBasePath, req, &out)
	return out, err
}

func (s *RoleService) Update(ctx context.Context, slug string, req UpdateRoleRequest) (Role, error) {
	var out Role
	err := s.c.Do(ctx, http.MethodPatch, rolesBasePath+"/"+slug, req, &out)
	return out, err
}

func (s *RoleService) Delete(ctx context.Context, slug string) error {
	return s.c.Do(ctx, http.MethodDelete, rolesBasePath+"/"+slug, nil, nil)
}

func (s *RoleService) GetPermissions(roleSlug string, opts *ListRolePermissionsOptions) *sdk.PageIterator[RoleEntityPermission, ListRolePermissionsOptions] {
	o := ListRolePermissionsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListRolePermissionsOptions) (sdk.ListResult[RoleEntityPermission], error) {
		in.Page = page
		return s.GetPermissionsPage(ctx, roleSlug, &in)
	}, o)
}

func (s *RoleService) GetPermissionsPage(ctx context.Context, roleSlug string, opts *ListRolePermissionsOptions) (sdk.ListResult[RoleEntityPermission], error) {
	var out sdk.ListResult[RoleEntityPermission]
	err := s.c.DoWithQuery(ctx, http.MethodGet, rolesBasePath+"/"+roleSlug+"/permissions", opts, nil, &out)
	return out, err
}

func (s *RoleService) AssignPermission(ctx context.Context, roleSlug string, req AssignPermissionRequest) (RoleEntityPermission, error) {
	var out RoleEntityPermission
	err := s.c.Do(ctx, http.MethodPost, rolesBasePath+"/"+roleSlug+"/permissions", req, &out)
	return out, err
}

func (s *RoleService) RemovePermission(ctx context.Context, roleSlug, permissionID string) error {
	return s.c.Do(ctx, http.MethodDelete, rolesBasePath+"/"+roleSlug+"/permissions/"+permissionID, nil, nil)
}
