package meta

import (
	"context"
	"net/http"
	metamodel "go.proteos.ai/model/meta"
	sdk "go.proteos.ai/sdk"
)

const listsBasePath = "/meta/v1/lists"

// ListServiceAPI is the contract a ListService satisfies.
type ListServiceAPI interface {
	List(opts *ListListsOptions) *sdk.PageIterator[metamodel.List, ListListsOptions]
	ListPage(ctx context.Context, opts *ListListsOptions) (sdk.ListResult[metamodel.List], error)
	Get(ctx context.Context, slug string) (metamodel.List, error)
	Create(ctx context.Context, req CreateListRequest) (metamodel.List, error)
	Upsert(ctx context.Context, req CreateListRequest) (metamodel.List, error)
	UpsertBySlug(ctx context.Context, slug string, req CreateListRequest) (metamodel.List, error)
	Update(ctx context.Context, slug string, req UpdateListRequest) (metamodel.List, error)
	Delete(ctx context.Context, slug string) error
}

// ListService manages list configurations.
type ListService struct{ c *sdk.Client }

var _ ListServiceAPI = (*ListService)(nil)

func (s *ListService) List(opts *ListListsOptions) *sdk.PageIterator[metamodel.List, ListListsOptions] {
	o := ListListsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListListsOptions) (sdk.ListResult[metamodel.List], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *ListService) ListPage(ctx context.Context, opts *ListListsOptions) (sdk.ListResult[metamodel.List], error) {
	var out sdk.ListResult[metamodel.List]
	err := s.c.DoWithQuery(ctx, http.MethodGet, listsBasePath, opts, nil, &out)
	return out, err
}

func (s *ListService) Get(ctx context.Context, slug string) (metamodel.List, error) {
	var out metamodel.List
	err := s.c.Do(ctx, http.MethodGet, listsBasePath+"/"+slug, nil, &out)
	return out, err
}

func (s *ListService) Create(ctx context.Context, req CreateListRequest) (metamodel.List, error) {
	var out metamodel.List
	err := s.c.Do(ctx, http.MethodPost, listsBasePath, req, &out)
	return out, err
}

func (s *ListService) Upsert(ctx context.Context, req CreateListRequest) (metamodel.List, error) {
	var out metamodel.List
	err := s.c.Do(ctx, http.MethodPost, listsBasePath+"/upsert", req, &out)
	return out, err
}

// UpsertBySlug calls `PUT /meta/v1/lists/:slug` — the idempotent upload path
// used by `pro module deploy`.
func (s *ListService) UpsertBySlug(ctx context.Context, slug string, req CreateListRequest) (metamodel.List, error) {
	var out metamodel.List
	req.Slug = slug
	err := s.c.Do(ctx, http.MethodPut, listsBasePath+"/"+slug, req, &out)
	return out, err
}

func (s *ListService) Update(ctx context.Context, slug string, req UpdateListRequest) (metamodel.List, error) {
	var out metamodel.List
	err := s.c.Do(ctx, http.MethodPatch, listsBasePath+"/"+slug, req, &out)
	return out, err
}

func (s *ListService) Delete(ctx context.Context, slug string) error {
	return s.c.Do(ctx, http.MethodDelete, listsBasePath+"/"+slug, nil, nil)
}
