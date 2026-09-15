package scheduling

import (
	"context"
	"net/http"
	"net/url"

	schedulingmodel "go.proteos.ai/model/scheduling"
	schedulingapi "go.proteos.ai/model/scheduling/api"
	sdk "go.proteos.ai/sdk"
)

// AvailabilityConstraintService manages a user's hours and blocks. `*Mine`
// methods act on the caller's own rows, `*ForUser` on another user's (admin
// path).
type AvailabilityConstraintService struct{ c *sdk.Client }

func mineAvailabilityConstraintsPath() string {
	return basePath + "/me/availability-constraints"
}

func userAvailabilityConstraintsPath(userId string) string {
	return basePath + "/users/" + url.PathEscape(userId) + "/availability-constraints"
}

func (s *AvailabilityConstraintService) list(path string, opts *ListAvailabilityConstraintsOptions) *sdk.PageIterator[schedulingmodel.AvailabilityConstraint, ListAvailabilityConstraintsOptions] {
	o := ListAvailabilityConstraintsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListAvailabilityConstraintsOptions) (sdk.ListResult[schedulingmodel.AvailabilityConstraint], error) {
		in.Page = page
		return s.listPage(ctx, path, &in)
	}, o)
}

func (s *AvailabilityConstraintService) listPage(ctx context.Context, path string, opts *ListAvailabilityConstraintsOptions) (sdk.ListResult[schedulingmodel.AvailabilityConstraint], error) {
	var out sdk.ListResult[schedulingmodel.AvailabilityConstraint]
	err := s.c.DoWithQuery(ctx, http.MethodGet, path, opts, nil, &out)
	return out, err
}

func (s *AvailabilityConstraintService) get(ctx context.Context, path string) (schedulingmodel.AvailabilityConstraint, error) {
	var out schedulingmodel.AvailabilityConstraint
	err := s.c.Do(ctx, http.MethodGet, path, nil, &out)
	return out, err
}

func (s *AvailabilityConstraintService) create(ctx context.Context, path string, req schedulingapi.CreateAvailabilityConstraintRequest) (schedulingmodel.AvailabilityConstraint, error) {
	var out schedulingmodel.AvailabilityConstraint
	err := s.c.Do(ctx, http.MethodPost, path, req, &out)
	return out, err
}

func (s *AvailabilityConstraintService) update(ctx context.Context, path string, req schedulingapi.UpdateAvailabilityConstraintRequest) (schedulingmodel.AvailabilityConstraint, error) {
	var out schedulingmodel.AvailabilityConstraint
	err := s.c.Do(ctx, http.MethodPatch, path, req, &out)
	return out, err
}

// ListMine iterates the caller's own constraints.
func (s *AvailabilityConstraintService) ListMine(opts *ListAvailabilityConstraintsOptions) *sdk.PageIterator[schedulingmodel.AvailabilityConstraint, ListAvailabilityConstraintsOptions] {
	return s.list(mineAvailabilityConstraintsPath(), opts)
}

// ListMinePage fetches one page of the caller's own constraints.
func (s *AvailabilityConstraintService) ListMinePage(ctx context.Context, opts *ListAvailabilityConstraintsOptions) (sdk.ListResult[schedulingmodel.AvailabilityConstraint], error) {
	return s.listPage(ctx, mineAvailabilityConstraintsPath(), opts)
}

func (s *AvailabilityConstraintService) GetMine(ctx context.Context, id string) (schedulingmodel.AvailabilityConstraint, error) {
	return s.get(ctx, mineAvailabilityConstraintsPath()+"/"+url.PathEscape(id))
}

func (s *AvailabilityConstraintService) CreateMine(ctx context.Context, req schedulingapi.CreateAvailabilityConstraintRequest) (schedulingmodel.AvailabilityConstraint, error) {
	return s.create(ctx, mineAvailabilityConstraintsPath(), req)
}

func (s *AvailabilityConstraintService) UpdateMine(ctx context.Context, id string, req schedulingapi.UpdateAvailabilityConstraintRequest) (schedulingmodel.AvailabilityConstraint, error) {
	return s.update(ctx, mineAvailabilityConstraintsPath()+"/"+url.PathEscape(id), req)
}

func (s *AvailabilityConstraintService) DeleteMine(ctx context.Context, id string) error {
	return s.c.Do(ctx, http.MethodDelete, mineAvailabilityConstraintsPath()+"/"+url.PathEscape(id), nil, nil)
}

// ListForUser iterates another user's constraints (admin path).
func (s *AvailabilityConstraintService) ListForUser(userId string, opts *ListAvailabilityConstraintsOptions) *sdk.PageIterator[schedulingmodel.AvailabilityConstraint, ListAvailabilityConstraintsOptions] {
	return s.list(userAvailabilityConstraintsPath(userId), opts)
}

// ListForUserPage fetches one page of another user's constraints.
func (s *AvailabilityConstraintService) ListForUserPage(ctx context.Context, userId string, opts *ListAvailabilityConstraintsOptions) (sdk.ListResult[schedulingmodel.AvailabilityConstraint], error) {
	return s.listPage(ctx, userAvailabilityConstraintsPath(userId), opts)
}

func (s *AvailabilityConstraintService) GetForUser(ctx context.Context, userId string, id string) (schedulingmodel.AvailabilityConstraint, error) {
	return s.get(ctx, userAvailabilityConstraintsPath(userId)+"/"+url.PathEscape(id))
}

func (s *AvailabilityConstraintService) CreateForUser(ctx context.Context, userId string, req schedulingapi.CreateAvailabilityConstraintRequest) (schedulingmodel.AvailabilityConstraint, error) {
	return s.create(ctx, userAvailabilityConstraintsPath(userId), req)
}

func (s *AvailabilityConstraintService) UpdateForUser(ctx context.Context, userId string, id string, req schedulingapi.UpdateAvailabilityConstraintRequest) (schedulingmodel.AvailabilityConstraint, error) {
	return s.update(ctx, userAvailabilityConstraintsPath(userId)+"/"+url.PathEscape(id), req)
}

func (s *AvailabilityConstraintService) DeleteForUser(ctx context.Context, userId string, id string) error {
	return s.c.Do(ctx, http.MethodDelete, userAvailabilityConstraintsPath(userId)+"/"+url.PathEscape(id), nil, nil)
}
