package account

import (
	"context"
	"net/http"

	sdk "go.proteos.ai/sdk"
)

const organizationsBasePath = "/accounts/v1/organizations"

// OrganizationServiceAPI is the contract an OrganizationService satisfies.
type OrganizationServiceAPI interface {
	List(opts *ListOrganizationsOptions) *sdk.PageIterator[Organization, ListOrganizationsOptions]
	ListPage(ctx context.Context, opts *ListOrganizationsOptions) (sdk.ListResult[Organization], error)
	Get(ctx context.Context, id string) (Organization, error)
	Create(ctx context.Context, req CreateOrganizationRequest) (Organization, error)
	Update(ctx context.Context, id string, req UpdateOrganizationRequest) (Organization, error)
	Delete(ctx context.Context, id string) error
}

// OrganizationService manages organizations.
type OrganizationService struct{ c *sdk.Client }

var _ OrganizationServiceAPI = (*OrganizationService)(nil)

func (s *OrganizationService) List(opts *ListOrganizationsOptions) *sdk.PageIterator[Organization, ListOrganizationsOptions] {
	o := ListOrganizationsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListOrganizationsOptions) (sdk.ListResult[Organization], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *OrganizationService) ListPage(ctx context.Context, opts *ListOrganizationsOptions) (sdk.ListResult[Organization], error) {
	var out sdk.ListResult[Organization]
	err := s.c.DoWithQuery(ctx, http.MethodGet, organizationsBasePath, opts, nil, &out)
	return out, err
}

func (s *OrganizationService) Get(ctx context.Context, id string) (Organization, error) {
	var out Organization
	err := s.c.Do(ctx, http.MethodGet, organizationsBasePath+"/"+id, nil, &out)
	return out, err
}

func (s *OrganizationService) Create(ctx context.Context, req CreateOrganizationRequest) (Organization, error) {
	var out Organization
	err := s.c.Do(ctx, http.MethodPost, organizationsBasePath, req, &out)
	return out, err
}

func (s *OrganizationService) Update(ctx context.Context, id string, req UpdateOrganizationRequest) (Organization, error) {
	var out Organization
	err := s.c.Do(ctx, http.MethodPatch, organizationsBasePath+"/"+id, req, &out)
	return out, err
}

func (s *OrganizationService) Delete(ctx context.Context, id string) error {
	return s.c.Do(ctx, http.MethodDelete, organizationsBasePath+"/"+id, nil, nil)
}
