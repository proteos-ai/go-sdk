package meta

import (
	"context"
	"net/http"

	metamodel "go.proteos.ai/model/meta"
	sdk "go.proteos.ai/sdk"
)

const entityExtensionsBasePath = "/meta/v1/entity-extensions"

// EntityExtensionServiceAPI is the contract an EntityExtensionService satisfies.
type EntityExtensionServiceAPI interface {
	List(opts *ListEntityExtensionsOptions) *sdk.PageIterator[metamodel.EntityExtension, ListEntityExtensionsOptions]
	ListPage(ctx context.Context, opts *ListEntityExtensionsOptions) (sdk.ListResult[metamodel.EntityExtension], error)
	Get(ctx context.Context, key string) (metamodel.EntityExtension, error)
	Create(ctx context.Context, req CreateEntityExtensionRequest) (metamodel.EntityExtension, error)
	UpsertByKey(ctx context.Context, key string, req CreateEntityExtensionRequest) (metamodel.EntityExtension, error)
	Update(ctx context.Context, key string, req UpdateEntityExtensionRequest) (metamodel.EntityExtension, error)
	Delete(ctx context.Context, key string) error
}

// EntityExtensionService manages entity extensions — attributes contributed to
// an existing host entity by a resource that does not own it. The host's
// stored attributes are the materialized merge (each contributed attribute
// stamped with extension_key); every write here re-materializes the host.
type EntityExtensionService struct{ c *sdk.Client }

var _ EntityExtensionServiceAPI = (*EntityExtensionService)(nil)

func (s *EntityExtensionService) List(opts *ListEntityExtensionsOptions) *sdk.PageIterator[metamodel.EntityExtension, ListEntityExtensionsOptions] {
	o := ListEntityExtensionsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListEntityExtensionsOptions) (sdk.ListResult[metamodel.EntityExtension], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *EntityExtensionService) ListPage(ctx context.Context, opts *ListEntityExtensionsOptions) (sdk.ListResult[metamodel.EntityExtension], error) {
	var out sdk.ListResult[metamodel.EntityExtension]
	err := s.c.DoWithQuery(ctx, http.MethodGet, entityExtensionsBasePath, opts, nil, &out)
	return out, err
}

func (s *EntityExtensionService) Get(ctx context.Context, key string) (metamodel.EntityExtension, error) {
	var out metamodel.EntityExtension
	err := s.c.Do(ctx, http.MethodGet, entityExtensionsBasePath+"/"+key, nil, &out)
	return out, err
}

func (s *EntityExtensionService) Create(ctx context.Context, req CreateEntityExtensionRequest) (metamodel.EntityExtension, error) {
	var out metamodel.EntityExtension
	err := s.c.Do(ctx, http.MethodPost, entityExtensionsBasePath, req, &out)
	return out, err
}

// UpsertByKey calls `PUT /meta/v1/entity-extensions/:key` — the idempotent
// upload path used by `pro module deploy`. Creates the extension or replaces
// its whole definition; the host binding (entity_slug) of an existing key is
// immutable.
func (s *EntityExtensionService) UpsertByKey(ctx context.Context, key string, req CreateEntityExtensionRequest) (metamodel.EntityExtension, error) {
	var out metamodel.EntityExtension
	req.Key = key
	err := s.c.Do(ctx, http.MethodPut, entityExtensionsBasePath+"/"+key, req, &out)
	return out, err
}

func (s *EntityExtensionService) Update(ctx context.Context, key string, req UpdateEntityExtensionRequest) (metamodel.EntityExtension, error) {
	var out metamodel.EntityExtension
	err := s.c.Do(ctx, http.MethodPatch, entityExtensionsBasePath+"/"+key, req, &out)
	return out, err
}

func (s *EntityExtensionService) Delete(ctx context.Context, key string) error {
	return s.c.Do(ctx, http.MethodDelete, entityExtensionsBasePath+"/"+key, nil, nil)
}
