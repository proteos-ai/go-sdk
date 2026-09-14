package meta

import (
	"context"
	"net/http"

	metamodel "go.proteos.ai/model/meta"
	sdk "go.proteos.ai/sdk"
)

const menuConfigurationExtensionsBasePath = "/meta/v1/menu-configuration-extensions"

// MenuConfigurationExtensionServiceAPI is the contract a
// MenuConfigurationExtensionService satisfies.
type MenuConfigurationExtensionServiceAPI interface {
	List(opts *ListMenuConfigurationExtensionsOptions) *sdk.PageIterator[metamodel.MenuConfigurationExtension, ListMenuConfigurationExtensionsOptions]
	ListPage(ctx context.Context, opts *ListMenuConfigurationExtensionsOptions) (sdk.ListResult[metamodel.MenuConfigurationExtension], error)
	Get(ctx context.Context, key string) (metamodel.MenuConfigurationExtension, error)
	Create(ctx context.Context, req CreateMenuConfigurationExtensionRequest) (metamodel.MenuConfigurationExtension, error)
	UpsertByKey(ctx context.Context, key string, req CreateMenuConfigurationExtensionRequest) (metamodel.MenuConfigurationExtension, error)
	Update(ctx context.Context, key string, req UpdateMenuConfigurationExtensionRequest) (metamodel.MenuConfigurationExtension, error)
	Delete(ctx context.Context, key string) error
}

// MenuConfigurationExtensionService manages menu-configuration extensions —
// items contributed to an existing host menu by a resource that does not own
// it. The host's stored items are the materialized merge (each contributed
// item stamped with extension_key; a replaced / removed host item kept under
// replaced_item); every write here re-materializes the host.
type MenuConfigurationExtensionService struct{ c *sdk.Client }

var _ MenuConfigurationExtensionServiceAPI = (*MenuConfigurationExtensionService)(nil)

func (s *MenuConfigurationExtensionService) List(opts *ListMenuConfigurationExtensionsOptions) *sdk.PageIterator[metamodel.MenuConfigurationExtension, ListMenuConfigurationExtensionsOptions] {
	o := ListMenuConfigurationExtensionsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListMenuConfigurationExtensionsOptions) (sdk.ListResult[metamodel.MenuConfigurationExtension], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *MenuConfigurationExtensionService) ListPage(ctx context.Context, opts *ListMenuConfigurationExtensionsOptions) (sdk.ListResult[metamodel.MenuConfigurationExtension], error) {
	var out sdk.ListResult[metamodel.MenuConfigurationExtension]
	err := s.c.DoWithQuery(ctx, http.MethodGet, menuConfigurationExtensionsBasePath, opts, nil, &out)
	return out, err
}

func (s *MenuConfigurationExtensionService) Get(ctx context.Context, key string) (metamodel.MenuConfigurationExtension, error) {
	var out metamodel.MenuConfigurationExtension
	err := s.c.Do(ctx, http.MethodGet, menuConfigurationExtensionsBasePath+"/"+key, nil, &out)
	return out, err
}

func (s *MenuConfigurationExtensionService) Create(ctx context.Context, req CreateMenuConfigurationExtensionRequest) (metamodel.MenuConfigurationExtension, error) {
	var out metamodel.MenuConfigurationExtension
	err := s.c.Do(ctx, http.MethodPost, menuConfigurationExtensionsBasePath, req, &out)
	return out, err
}

// UpsertByKey calls `PUT /meta/v1/menu-configuration-extensions/:key` — the idempotent
// upload path used by `pro module deploy`. Creates the extension or replaces
// its whole definition; the host binding (menu_slug) of an existing key is
// immutable.
func (s *MenuConfigurationExtensionService) UpsertByKey(ctx context.Context, key string, req CreateMenuConfigurationExtensionRequest) (metamodel.MenuConfigurationExtension, error) {
	var out metamodel.MenuConfigurationExtension
	req.Key = key
	err := s.c.Do(ctx, http.MethodPut, menuConfigurationExtensionsBasePath+"/"+key, req, &out)
	return out, err
}

func (s *MenuConfigurationExtensionService) Update(ctx context.Context, key string, req UpdateMenuConfigurationExtensionRequest) (metamodel.MenuConfigurationExtension, error) {
	var out metamodel.MenuConfigurationExtension
	err := s.c.Do(ctx, http.MethodPatch, menuConfigurationExtensionsBasePath+"/"+key, req, &out)
	return out, err
}

func (s *MenuConfigurationExtensionService) Delete(ctx context.Context, key string) error {
	return s.c.Do(ctx, http.MethodDelete, menuConfigurationExtensionsBasePath+"/"+key, nil, nil)
}
