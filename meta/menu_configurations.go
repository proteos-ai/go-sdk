package meta

import (
	"context"
	"net/http"
	metamodel "go.proteos.ai/model/meta"
	sdk "go.proteos.ai/sdk"
)

const menuConfigurationsBasePath = "/meta/v1/menu-configurations"

// MenuConfigurationServiceAPI is the contract a MenuConfigurationService satisfies.
type MenuConfigurationServiceAPI interface {
	List(opts *ListMenuConfigurationsOptions) *sdk.PageIterator[metamodel.MenuConfiguration, ListMenuConfigurationsOptions]
	ListPage(ctx context.Context, opts *ListMenuConfigurationsOptions) (sdk.ListResult[metamodel.MenuConfiguration], error)
	Get(ctx context.Context, slug string) (metamodel.MenuConfiguration, error)
	Create(ctx context.Context, req CreateMenuConfigurationRequest) (metamodel.MenuConfiguration, error)
	Upsert(ctx context.Context, slug string, req CreateMenuConfigurationRequest) (metamodel.MenuConfiguration, error)
	Update(ctx context.Context, slug string, req UpdateMenuConfigurationRequest) (metamodel.MenuConfiguration, error)
	Delete(ctx context.Context, slug string) error
}

// MenuConfigurationService manages app menu configurations.
type MenuConfigurationService struct{ c *sdk.Client }

var _ MenuConfigurationServiceAPI = (*MenuConfigurationService)(nil)

func (s *MenuConfigurationService) List(opts *ListMenuConfigurationsOptions) *sdk.PageIterator[metamodel.MenuConfiguration, ListMenuConfigurationsOptions] {
	o := ListMenuConfigurationsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListMenuConfigurationsOptions) (sdk.ListResult[metamodel.MenuConfiguration], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *MenuConfigurationService) ListPage(ctx context.Context, opts *ListMenuConfigurationsOptions) (sdk.ListResult[metamodel.MenuConfiguration], error) {
	var out sdk.ListResult[metamodel.MenuConfiguration]
	err := s.c.DoWithQuery(ctx, http.MethodGet, menuConfigurationsBasePath, opts, nil, &out)
	return out, err
}

func (s *MenuConfigurationService) Get(ctx context.Context, slug string) (metamodel.MenuConfiguration, error) {
	var out metamodel.MenuConfiguration
	err := s.c.Do(ctx, http.MethodGet, menuConfigurationsBasePath+"/"+slug, nil, &out)
	return out, err
}

func (s *MenuConfigurationService) Create(ctx context.Context, req CreateMenuConfigurationRequest) (metamodel.MenuConfiguration, error) {
	var out metamodel.MenuConfiguration
	err := s.c.Do(ctx, http.MethodPost, menuConfigurationsBasePath, req, &out)
	return out, err
}

// Upsert calls `PUT /meta/v1/menu-configurations/:slug` — the idempotent
// upload path used by `pro module deploy`.
func (s *MenuConfigurationService) Upsert(ctx context.Context, slug string, req CreateMenuConfigurationRequest) (metamodel.MenuConfiguration, error) {
	var out metamodel.MenuConfiguration
	req.Slug = slug
	err := s.c.Do(ctx, http.MethodPut, menuConfigurationsBasePath+"/"+slug, req, &out)
	return out, err
}

func (s *MenuConfigurationService) Update(ctx context.Context, slug string, req UpdateMenuConfigurationRequest) (metamodel.MenuConfiguration, error) {
	var out metamodel.MenuConfiguration
	err := s.c.Do(ctx, http.MethodPatch, menuConfigurationsBasePath+"/"+slug, req, &out)
	return out, err
}

func (s *MenuConfigurationService) Delete(ctx context.Context, slug string) error {
	return s.c.Do(ctx, http.MethodDelete, menuConfigurationsBasePath+"/"+slug, nil, nil)
}
