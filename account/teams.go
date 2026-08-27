package account

import (
	"context"
	"net/http"

	sdk "go.proteos.ai/sdk"
)

const teamsBasePath = "/accounts/v1/teams"

// TeamServiceAPI is the contract a TeamService satisfies.
type TeamServiceAPI interface {
	List(opts *ListTeamsOptions) *sdk.PageIterator[Team, ListTeamsOptions]
	ListPage(ctx context.Context, opts *ListTeamsOptions) (sdk.ListResult[Team], error)
	Get(ctx context.Context, slug string) (Team, error)
	Create(ctx context.Context, req CreateTeamRequest) (Team, error)
	Update(ctx context.Context, slug string, req UpdateTeamRequest) (Team, error)
	Delete(ctx context.Context, slug string) error
	GetMembers(slug string, opts *ListTeamMembersOptions) *sdk.PageIterator[TeamMember, ListTeamMembersOptions]
	GetMembersPage(ctx context.Context, slug string, opts *ListTeamMembersOptions) (sdk.ListResult[TeamMember], error)
	AddMember(ctx context.Context, slug string, req AddTeamMemberRequest) (TeamMember, error)
	RemoveMember(ctx context.Context, slug, userID string) error
}

// TeamService manages teams — the org's structure, and a principal that can hold
// access anywhere a user can.
//
// Team hierarchy rolls UP: a member of a child team counts as a member of every
// ancestor FOR GRANTS MADE TO THAT ANCESTOR. Granting `sales` reaches the members
// of `sales-emea`; granting `sales-emea` never reaches someone only in `sales`.
type TeamService struct{ c *sdk.Client }

var _ TeamServiceAPI = (*TeamService)(nil)

func (s *TeamService) List(opts *ListTeamsOptions) *sdk.PageIterator[Team, ListTeamsOptions] {
	o := ListTeamsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListTeamsOptions) (sdk.ListResult[Team], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *TeamService) ListPage(ctx context.Context, opts *ListTeamsOptions) (sdk.ListResult[Team], error) {
	var out sdk.ListResult[Team]
	err := s.c.DoWithQuery(ctx, http.MethodGet, teamsBasePath, opts, nil, &out)
	return out, err
}

func (s *TeamService) Get(ctx context.Context, slug string) (Team, error) {
	var out Team
	err := s.c.Do(ctx, http.MethodGet, teamsBasePath+"/"+slug, nil, &out)
	return out, err
}

func (s *TeamService) Create(ctx context.Context, req CreateTeamRequest) (Team, error) {
	var out Team
	err := s.c.Do(ctx, http.MethodPost, teamsBasePath, req, &out)
	return out, err
}

// Update changes a team's name, description or parent. The slug cannot change —
// it is half the primary key and grants reference `team:<orgId>/<slug>`.
func (s *TeamService) Update(ctx context.Context, slug string, req UpdateTeamRequest) (Team, error) {
	var out Team
	err := s.c.Do(ctx, http.MethodPatch, teamsBasePath+"/"+slug, req, &out)
	return out, err
}

func (s *TeamService) Delete(ctx context.Context, slug string) error {
	return s.c.Do(ctx, http.MethodDelete, teamsBasePath+"/"+slug, nil, nil)
}

// GetMembers lists a team's DIRECT members. Membership of ancestor teams is
// derived rather than stored, so it does not appear here.
func (s *TeamService) GetMembers(slug string, opts *ListTeamMembersOptions) *sdk.PageIterator[TeamMember, ListTeamMembersOptions] {
	o := ListTeamMembersOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListTeamMembersOptions) (sdk.ListResult[TeamMember], error) {
		in.Page = page
		return s.GetMembersPage(ctx, slug, &in)
	}, o)
}

func (s *TeamService) GetMembersPage(ctx context.Context, slug string, opts *ListTeamMembersOptions) (sdk.ListResult[TeamMember], error) {
	var out sdk.ListResult[TeamMember]
	err := s.c.DoWithQuery(ctx, http.MethodGet, teamsBasePath+"/"+slug+"/members", opts, nil, &out)
	return out, err
}

func (s *TeamService) AddMember(ctx context.Context, slug string, req AddTeamMemberRequest) (TeamMember, error) {
	var out TeamMember
	err := s.c.Do(ctx, http.MethodPost, teamsBasePath+"/"+slug+"/members", req, &out)
	return out, err
}

func (s *TeamService) RemoveMember(ctx context.Context, slug, userID string) error {
	return s.c.Do(ctx, http.MethodDelete, teamsBasePath+"/"+slug+"/members/"+userID, nil, nil)
}
