package account

import (
	"context"
	"net/http"

	sdk "go.proteos.ai/sdk"
)

const meBasePath = "/accounts/v1/me"

type meOrganizationsResponse struct {
	Data []Organization `json:"data"`
}

// MeServiceAPI is the contract a MeService satisfies — queries scoped to the
// calling user (from the token), not an arbitrary id.
type MeServiceAPI interface {
	Get(ctx context.Context) (User, error)
	Organizations(ctx context.Context) ([]Organization, error)
}

// MeService answers "me"-scoped questions for the authenticated caller.
type MeService struct{ c *sdk.Client }

var _ MeServiceAPI = (*MeService)(nil)

// Get returns the caller's own account record, resolved from the token's user
// id. The canonical "who am I" — use this instead of filtering Users.List by
// email.
func (s *MeService) Get(ctx context.Context) (User, error) {
	var out User
	err := s.c.Do(ctx, http.MethodGet, meBasePath, nil, &out)
	return out, err
}

// Organizations returns the orgs the caller can act in (all orgs for a platform
// admin; otherwise the orgs they're a member of). Powers the org-switcher.
func (s *MeService) Organizations(ctx context.Context) ([]Organization, error) {
	var out meOrganizationsResponse
	err := s.c.Do(ctx, http.MethodGet, meBasePath+"/organizations", nil, &out)
	return out.Data, err
}
