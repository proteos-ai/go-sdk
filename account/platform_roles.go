package account

import (
	"context"
	"net/http"
	"time"

	"go.proteos.ai/model/common"
	sdk "go.proteos.ai/sdk"
)

const platformRolesBasePath = "/accounts/v1/platform-roles"

// PlatformRole is an org-independent role grant to a user (e.g. "admin").
type PlatformRole struct {
	UserID    string         `json:"user_id"`
	Role      string         `json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	CreatedBy common.UserRef `json:"created_by"`
}

// GrantPlatformRoleRequest is the body for granting a platform role. Role
// defaults to "admin" server-side when empty.
type GrantPlatformRoleRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role,omitempty"`
}

type platformRolesQuery struct {
	UserID string `query:"user_id,omitempty"`
}

type platformRolesResponse struct {
	Data []PlatformRole `json:"data"`
}

// PlatformRoleServiceAPI is the contract a PlatformRoleService satisfies. All
// endpoints require the caller to be a platform admin.
type PlatformRoleServiceAPI interface {
	Grant(ctx context.Context, req GrantPlatformRoleRequest) (PlatformRole, error)
	Revoke(ctx context.Context, userID, role string) error
	List(ctx context.Context) ([]PlatformRole, error)
	ListByUser(ctx context.Context, userID string) ([]PlatformRole, error)
}

// PlatformRoleService manages platform-level (org-independent) roles.
type PlatformRoleService struct{ c *sdk.Client }

var _ PlatformRoleServiceAPI = (*PlatformRoleService)(nil)

func (s *PlatformRoleService) Grant(ctx context.Context, req GrantPlatformRoleRequest) (PlatformRole, error) {
	var out PlatformRole
	err := s.c.Do(ctx, http.MethodPost, platformRolesBasePath, req, &out)
	return out, err
}

func (s *PlatformRoleService) Revoke(ctx context.Context, userID, role string) error {
	return s.c.Do(ctx, http.MethodDelete, platformRolesBasePath+"/"+userID+"/"+role, nil, nil)
}

func (s *PlatformRoleService) List(ctx context.Context) ([]PlatformRole, error) {
	var out platformRolesResponse
	err := s.c.Do(ctx, http.MethodGet, platformRolesBasePath, nil, &out)
	return out.Data, err
}

func (s *PlatformRoleService) ListByUser(ctx context.Context, userID string) ([]PlatformRole, error) {
	var out platformRolesResponse
	err := s.c.DoWithQuery(ctx, http.MethodGet, platformRolesBasePath, platformRolesQuery{UserID: userID}, nil, &out)
	return out.Data, err
}
