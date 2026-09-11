package account

import (
	"context"
	"net/http"

	sdk "go.proteos.ai/sdk"
)

const (
	profilesBasePath               = "/accounts/v1/profiles"
	userProfileAssignmentsBasePath = "/accounts/v1/user-profile-assignments"
)

// ProfileServiceAPI is the contract a ProfileService satisfies.
type ProfileServiceAPI interface {
	List(opts *ListProfilesOptions) *sdk.PageIterator[Profile, ListProfilesOptions]
	ListPage(ctx context.Context, opts *ListProfilesOptions) (sdk.ListResult[Profile], error)
	Get(ctx context.Context, slug string) (Profile, error)
	Create(ctx context.Context, req CreateProfileRequest) (Profile, error)
	Upsert(ctx context.Context, slug string, req CreateProfileRequest) (Profile, error)
	Update(ctx context.Context, slug string, req UpdateProfileRequest) (Profile, error)
	Delete(ctx context.Context, slug string) error
	ListAssignments(opts *ListUserProfileAssignmentsOptions) *sdk.PageIterator[UserProfileAssignment, ListUserProfileAssignmentsOptions]
	ListAssignmentsPage(ctx context.Context, opts *ListUserProfileAssignmentsOptions) (sdk.ListResult[UserProfileAssignment], error)
}

// ProfileService manages profiles (the org's user-types) and answers the
// org-wide "who holds which profile" listing. Per-user reads and writes live
// on UserService (GetProfile / SetProfile / ClearProfile).
type ProfileService struct{ c *sdk.Client }

var _ ProfileServiceAPI = (*ProfileService)(nil)

func (s *ProfileService) List(opts *ListProfilesOptions) *sdk.PageIterator[Profile, ListProfilesOptions] {
	o := ListProfilesOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListProfilesOptions) (sdk.ListResult[Profile], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *ProfileService) ListPage(ctx context.Context, opts *ListProfilesOptions) (sdk.ListResult[Profile], error) {
	var out sdk.ListResult[Profile]
	err := s.c.DoWithQuery(ctx, http.MethodGet, profilesBasePath, opts, nil, &out)
	return out, err
}

func (s *ProfileService) Get(ctx context.Context, slug string) (Profile, error) {
	var out Profile
	err := s.c.Do(ctx, http.MethodGet, profilesBasePath+"/"+slug, nil, &out)
	return out, err
}

func (s *ProfileService) Create(ctx context.Context, req CreateProfileRequest) (Profile, error) {
	var out Profile
	err := s.c.Do(ctx, http.MethodPost, profilesBasePath, req, &out)
	return out, err
}

// Upsert creates or updates the profile by slug (PUT) — the idempotent deploy
// path. The URL slug wins over the body's.
func (s *ProfileService) Upsert(ctx context.Context, slug string, req CreateProfileRequest) (Profile, error) {
	req.Slug = slug
	var out Profile
	err := s.c.Do(ctx, http.MethodPut, profilesBasePath+"/"+slug, req, &out)
	return out, err
}

func (s *ProfileService) Update(ctx context.Context, slug string, req UpdateProfileRequest) (Profile, error) {
	var out Profile
	err := s.c.Do(ctx, http.MethodPatch, profilesBasePath+"/"+slug, req, &out)
	return out, err
}

// Delete removes a profile. A profile users still hold is refused with 409
// profile_in_use — move them first.
func (s *ProfileService) Delete(ctx context.Context, slug string) error {
	return s.c.Do(ctx, http.MethodDelete, profilesBasePath+"/"+slug, nil, nil)
}

func (s *ProfileService) ListAssignments(opts *ListUserProfileAssignmentsOptions) *sdk.PageIterator[UserProfileAssignment, ListUserProfileAssignmentsOptions] {
	o := ListUserProfileAssignmentsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListUserProfileAssignmentsOptions) (sdk.ListResult[UserProfileAssignment], error) {
		in.Page = page
		return s.ListAssignmentsPage(ctx, &in)
	}, o)
}

func (s *ProfileService) ListAssignmentsPage(ctx context.Context, opts *ListUserProfileAssignmentsOptions) (sdk.ListResult[UserProfileAssignment], error) {
	var out sdk.ListResult[UserProfileAssignment]
	err := s.c.DoWithQuery(ctx, http.MethodGet, userProfileAssignmentsBasePath, opts, nil, &out)
	return out, err
}
