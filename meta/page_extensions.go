package meta

import (
	"context"
	"net/http"

	metamodel "go.proteos.ai/model/meta"
	sdk "go.proteos.ai/sdk"
)

const pageExtensionsBasePath = "/meta/v1/page-extensions"

// PageExtensionServiceAPI is the contract an PageExtensionService satisfies.
type PageExtensionServiceAPI interface {
	List(opts *ListPageExtensionsOptions) *sdk.PageIterator[metamodel.PageExtension, ListPageExtensionsOptions]
	ListPage(ctx context.Context, opts *ListPageExtensionsOptions) (sdk.ListResult[metamodel.PageExtension], error)
	Get(ctx context.Context, key string) (metamodel.PageExtension, error)
	Create(ctx context.Context, req CreatePageExtensionRequest) (metamodel.PageExtension, error)
	UpsertByKey(ctx context.Context, key string, req CreatePageExtensionRequest) (metamodel.PageExtension, error)
	Update(ctx context.Context, key string, req UpdatePageExtensionRequest) (metamodel.PageExtension, error)
	Delete(ctx context.Context, key string) error
}

// PageExtensionService manages anchored layout blocks contributed to an existing host page by a resource that does not own it. The host's
// stored layout is the materialized merge (each contributed element and tab
// stamped with extension_key); every write here re-materializes the host.
type PageExtensionService struct{ c *sdk.Client }

var _ PageExtensionServiceAPI = (*PageExtensionService)(nil)

func (s *PageExtensionService) List(opts *ListPageExtensionsOptions) *sdk.PageIterator[metamodel.PageExtension, ListPageExtensionsOptions] {
	o := ListPageExtensionsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListPageExtensionsOptions) (sdk.ListResult[metamodel.PageExtension], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *PageExtensionService) ListPage(ctx context.Context, opts *ListPageExtensionsOptions) (sdk.ListResult[metamodel.PageExtension], error) {
	var out sdk.ListResult[metamodel.PageExtension]
	err := s.c.DoWithQuery(ctx, http.MethodGet, pageExtensionsBasePath, opts, nil, &out)
	return out, err
}

func (s *PageExtensionService) Get(ctx context.Context, key string) (metamodel.PageExtension, error) {
	var out metamodel.PageExtension
	err := s.c.Do(ctx, http.MethodGet, pageExtensionsBasePath+"/"+key, nil, &out)
	return out, err
}

func (s *PageExtensionService) Create(ctx context.Context, req CreatePageExtensionRequest) (metamodel.PageExtension, error) {
	var out metamodel.PageExtension
	err := s.c.Do(ctx, http.MethodPost, pageExtensionsBasePath, req, &out)
	return out, err
}

// UpsertByKey calls `PUT /meta/v1/page-extensions/:key` — the idempotent
// upload path used by `pro module deploy`. Creates the extension or replaces
// its whole definition; the host binding (page_slug) of an existing key is
// immutable.
func (s *PageExtensionService) UpsertByKey(ctx context.Context, key string, req CreatePageExtensionRequest) (metamodel.PageExtension, error) {
	var out metamodel.PageExtension
	req.Key = key
	err := s.c.Do(ctx, http.MethodPut, pageExtensionsBasePath+"/"+key, req, &out)
	return out, err
}

func (s *PageExtensionService) Update(ctx context.Context, key string, req UpdatePageExtensionRequest) (metamodel.PageExtension, error) {
	var out metamodel.PageExtension
	err := s.c.Do(ctx, http.MethodPatch, pageExtensionsBasePath+"/"+key, req, &out)
	return out, err
}

func (s *PageExtensionService) Delete(ctx context.Context, key string) error {
	return s.c.Do(ctx, http.MethodDelete, pageExtensionsBasePath+"/"+key, nil, nil)
}
