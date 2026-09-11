package meta

import (
	"context"
	"net/http"

	metamodel "go.proteos.ai/model/meta"
	sdk "go.proteos.ai/sdk"
)

const appConfigurationsBasePath = "/meta/v1/app-configurations"

// AppConfigurationServiceAPI is the contract an AppConfigurationService satisfies.
type AppConfigurationServiceAPI interface {
	List(opts *ListAppConfigurationsOptions) *sdk.PageIterator[metamodel.AppConfiguration, ListAppConfigurationsOptions]
	ListPage(ctx context.Context, opts *ListAppConfigurationsOptions) (sdk.ListResult[metamodel.AppConfiguration], error)
	Get(ctx context.Context, slug string) (metamodel.AppConfiguration, error)
	Create(ctx context.Context, req CreateAppConfigurationRequest) (metamodel.AppConfiguration, error)
	Upsert(ctx context.Context, slug string, req CreateAppConfigurationRequest) (metamodel.AppConfiguration, error)
	Update(ctx context.Context, slug string, req UpdateAppConfigurationRequest) (metamodel.AppConfiguration, error)
	Delete(ctx context.Context, slug string) error
}

// AppConfigurationService manages the typed (app × profile) binding rows —
// how an app presents itself (home, menu, agents, record pages) to everyone
// (profile_slug "") or to one profile.
type AppConfigurationService struct{ c *sdk.Client }

var _ AppConfigurationServiceAPI = (*AppConfigurationService)(nil)

func (s *AppConfigurationService) List(opts *ListAppConfigurationsOptions) *sdk.PageIterator[metamodel.AppConfiguration, ListAppConfigurationsOptions] {
	o := ListAppConfigurationsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListAppConfigurationsOptions) (sdk.ListResult[metamodel.AppConfiguration], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *AppConfigurationService) ListPage(ctx context.Context, opts *ListAppConfigurationsOptions) (sdk.ListResult[metamodel.AppConfiguration], error) {
	var out sdk.ListResult[metamodel.AppConfiguration]
	err := s.c.DoWithQuery(ctx, http.MethodGet, appConfigurationsBasePath, opts, nil, &out)
	return out, err
}

func (s *AppConfigurationService) Get(ctx context.Context, slug string) (metamodel.AppConfiguration, error) {
	var out metamodel.AppConfiguration
	err := s.c.Do(ctx, http.MethodGet, appConfigurationsBasePath+"/"+slug, nil, &out)
	return out, err
}

func (s *AppConfigurationService) Create(ctx context.Context, req CreateAppConfigurationRequest) (metamodel.AppConfiguration, error) {
	var out metamodel.AppConfiguration
	err := s.c.Do(ctx, http.MethodPost, appConfigurationsBasePath, req, &out)
	return out, err
}

// Upsert calls `PUT /meta/v1/app-configurations/:slug` — the idempotent
// upload path used by `pro module deploy`. Replaces the whole configuration;
// the (app_slug, profile_slug) binding of an existing slug is immutable.
func (s *AppConfigurationService) Upsert(ctx context.Context, slug string, req CreateAppConfigurationRequest) (metamodel.AppConfiguration, error) {
	var out metamodel.AppConfiguration
	req.Slug = slug
	err := s.c.Do(ctx, http.MethodPut, appConfigurationsBasePath+"/"+slug, req, &out)
	return out, err
}

func (s *AppConfigurationService) Update(ctx context.Context, slug string, req UpdateAppConfigurationRequest) (metamodel.AppConfiguration, error) {
	var out metamodel.AppConfiguration
	err := s.c.Do(ctx, http.MethodPatch, appConfigurationsBasePath+"/"+slug, req, &out)
	return out, err
}

func (s *AppConfigurationService) Delete(ctx context.Context, slug string) error {
	return s.c.Do(ctx, http.MethodDelete, appConfigurationsBasePath+"/"+slug, nil, nil)
}
