package meta

import (
	"context"
	"net/http"

	metamodel "go.proteos.ai/model/meta"
	sdk "go.proteos.ai/sdk"
)

const listExtensionsBasePath = "/meta/v1/list-extensions"

// ListExtensionServiceAPI is the contract an ListExtensionService satisfies.
type ListExtensionServiceAPI interface {
	List(opts *ListListExtensionsOptions) *sdk.PageIterator[metamodel.ListExtension, ListListExtensionsOptions]
	ListPage(ctx context.Context, opts *ListListExtensionsOptions) (sdk.ListResult[metamodel.ListExtension], error)
	Get(ctx context.Context, key string) (metamodel.ListExtension, error)
	Create(ctx context.Context, req CreateListExtensionRequest) (metamodel.ListExtension, error)
	UpsertByKey(ctx context.Context, key string, req CreateListExtensionRequest) (metamodel.ListExtension, error)
	Update(ctx context.Context, key string, req UpdateListExtensionRequest) (metamodel.ListExtension, error)
	Delete(ctx context.Context, key string) error
}

// ListExtensionService manages columns and toolbar actions contributed to an existing host list by a resource that does not own
// it. The host's stored columns / actions are the materialized merge (each
// contributed one stamped with extension_key); every write here
// re-materializes the host.
type ListExtensionService struct{ c *sdk.Client }

var _ ListExtensionServiceAPI = (*ListExtensionService)(nil)

func (s *ListExtensionService) List(opts *ListListExtensionsOptions) *sdk.PageIterator[metamodel.ListExtension, ListListExtensionsOptions] {
	o := ListListExtensionsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListListExtensionsOptions) (sdk.ListResult[metamodel.ListExtension], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *ListExtensionService) ListPage(ctx context.Context, opts *ListListExtensionsOptions) (sdk.ListResult[metamodel.ListExtension], error) {
	var out sdk.ListResult[metamodel.ListExtension]
	err := s.c.DoWithQuery(ctx, http.MethodGet, listExtensionsBasePath, opts, nil, &out)
	return out, err
}

func (s *ListExtensionService) Get(ctx context.Context, key string) (metamodel.ListExtension, error) {
	var out metamodel.ListExtension
	err := s.c.Do(ctx, http.MethodGet, listExtensionsBasePath+"/"+key, nil, &out)
	return out, err
}

func (s *ListExtensionService) Create(ctx context.Context, req CreateListExtensionRequest) (metamodel.ListExtension, error) {
	var out metamodel.ListExtension
	err := s.c.Do(ctx, http.MethodPost, listExtensionsBasePath, req, &out)
	return out, err
}

// UpsertByKey calls `PUT /meta/v1/list-extensions/:key` — the idempotent
// upload path used by `pro module deploy`. Creates the extension or replaces
// its whole definition; the host binding (list_slug) of an existing key is
// immutable.
func (s *ListExtensionService) UpsertByKey(ctx context.Context, key string, req CreateListExtensionRequest) (metamodel.ListExtension, error) {
	var out metamodel.ListExtension
	req.Key = key
	err := s.c.Do(ctx, http.MethodPut, listExtensionsBasePath+"/"+key, req, &out)
	return out, err
}

func (s *ListExtensionService) Update(ctx context.Context, key string, req UpdateListExtensionRequest) (metamodel.ListExtension, error) {
	var out metamodel.ListExtension
	err := s.c.Do(ctx, http.MethodPatch, listExtensionsBasePath+"/"+key, req, &out)
	return out, err
}

func (s *ListExtensionService) Delete(ctx context.Context, key string) error {
	return s.c.Do(ctx, http.MethodDelete, listExtensionsBasePath+"/"+key, nil, nil)
}
