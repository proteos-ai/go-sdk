package meta

import (
	"context"
	sdk "go.proteos.ai/sdk"
	metamodel "go.proteos.ai/model/meta"
	"net/http"
)

const entitiesBasePath = "/meta/v1/entities"

// EntityServiceAPI is the contract an EntityService satisfies.
type EntityServiceAPI interface {
	List(opts *ListEntitiesOptions) *sdk.PageIterator[metamodel.Entity, ListEntitiesOptions]
	ListPage(ctx context.Context, opts *ListEntitiesOptions) (sdk.ListResult[metamodel.Entity], error)
	ListWithSchema(opts *ListEntitiesOptions) *sdk.PageIterator[metamodel.EntityWithSchema, ListEntitiesOptions]
	ListPageWithSchema(ctx context.Context, opts *ListEntitiesOptions) (sdk.ListResult[metamodel.EntityWithSchema], error)
	Get(ctx context.Context, slug string) (metamodel.Entity, error)
	GetWithSchema(ctx context.Context, slug string) (metamodel.EntityWithSchema, error)
	Create(ctx context.Context, req CreateEntityRequest) (metamodel.Entity, error)
	UpsertBySlug(ctx context.Context, slug string, req CreateEntityRequest) (metamodel.Entity, error)
	Update(ctx context.Context, slug string, req UpdateEntityRequest) (metamodel.Entity, error)
	Delete(ctx context.Context, slug string) error
}

// EntityService manages metadata entities.
type EntityService struct{ c *sdk.Client }

var _ EntityServiceAPI = (*EntityService)(nil)

func (s *EntityService) List(opts *ListEntitiesOptions) *sdk.PageIterator[metamodel.Entity, ListEntitiesOptions] {
	o := pickListEntitiesOpts(opts, false)
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListEntitiesOptions) (sdk.ListResult[metamodel.Entity], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *EntityService) ListPage(ctx context.Context, opts *ListEntitiesOptions) (sdk.ListResult[metamodel.Entity], error) {
	var out sdk.ListResult[metamodel.Entity]
	err := s.c.DoWithQuery(ctx, http.MethodGet, entitiesBasePath, opts, nil, &out)
	return out, err
}

func (s *EntityService) ListWithSchema(opts *ListEntitiesOptions) *sdk.PageIterator[metamodel.EntityWithSchema, ListEntitiesOptions] {
	o := pickListEntitiesOpts(opts, true)
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListEntitiesOptions) (sdk.ListResult[metamodel.EntityWithSchema], error) {
		in.Page = page
		return s.ListPageWithSchema(ctx, &in)
	}, o)
}

func (s *EntityService) ListPageWithSchema(ctx context.Context, opts *ListEntitiesOptions) (sdk.ListResult[metamodel.EntityWithSchema], error) {
	var out sdk.ListResult[metamodel.EntityWithSchema]
	o := ListEntitiesOptions{}
	if opts != nil {
		o = *opts
	}
	o.WithSchema = true
	err := s.c.DoWithQuery(ctx, http.MethodGet, entitiesBasePath, &o, nil, &out)
	return out, err
}

func (s *EntityService) Get(ctx context.Context, slug string) (metamodel.Entity, error) {
	var out metamodel.Entity
	err := s.c.Do(ctx, http.MethodGet, entitiesBasePath+"/"+slug, nil, &out)
	return out, err
}

func (s *EntityService) GetWithSchema(ctx context.Context, slug string) (metamodel.EntityWithSchema, error) {
	var out metamodel.EntityWithSchema
	err := s.c.DoWithQuery(ctx, http.MethodGet, entitiesBasePath+"/"+slug, map[string]any{"with_schema": true}, nil, &out)
	return out, err
}

func (s *EntityService) Create(ctx context.Context, req CreateEntityRequest) (metamodel.Entity, error) {
	var out metamodel.Entity
	err := s.c.Do(ctx, http.MethodPost, entitiesBasePath, req, &out)
	return out, err
}

// UpsertBySlug calls `PUT /meta/v1/entities/:slug`, creating the entity if
// missing and updating it in place otherwise. Used by `pro module deploy`
// for idempotent per-resource upload.
func (s *EntityService) UpsertBySlug(ctx context.Context, slug string, req CreateEntityRequest) (metamodel.Entity, error) {
	var out metamodel.Entity
	req.Slug = slug
	err := s.c.Do(ctx, http.MethodPut, entitiesBasePath+"/"+slug, req, &out)
	return out, err
}

func (s *EntityService) Update(ctx context.Context, slug string, req UpdateEntityRequest) (metamodel.Entity, error) {
	var out metamodel.Entity
	err := s.c.Do(ctx, http.MethodPatch, entitiesBasePath+"/"+slug, req, &out)
	return out, err
}

func (s *EntityService) Delete(ctx context.Context, slug string) error {
	return s.c.Do(ctx, http.MethodDelete, entitiesBasePath+"/"+slug, nil, nil)
}

func pickListEntitiesOpts(opts *ListEntitiesOptions, withSchema bool) ListEntitiesOptions {
	o := ListEntitiesOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	o.WithSchema = withSchema
	return o
}
