package meta

import (
	"context"
	"net/http"
	metamodel "go.proteos.ai/model/meta"
	sdk "go.proteos.ai/sdk"
)

const pagesBasePath = "/meta/v1/pages"

// PageServiceAPI is the contract a PageService satisfies.
type PageServiceAPI interface {
	List(opts *ListPagesOptions) *sdk.PageIterator[metamodel.Page, ListPagesOptions]
	ListPage(ctx context.Context, opts *ListPagesOptions) (sdk.ListResult[metamodel.Page], error)
	Get(ctx context.Context, slug string) (metamodel.Page, error)
	Create(ctx context.Context, req CreatePageRequest) (metamodel.Page, error)
	Upsert(ctx context.Context, slug string, req CreatePageRequest) (metamodel.Page, error)
	Update(ctx context.Context, slug string, req UpdatePageRequest) (metamodel.Page, error)
	Delete(ctx context.Context, slug string) error
}

// PageService manages page configurations.
type PageService struct{ c *sdk.Client }

var _ PageServiceAPI = (*PageService)(nil)

func (s *PageService) List(opts *ListPagesOptions) *sdk.PageIterator[metamodel.Page, ListPagesOptions] {
	o := ListPagesOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListPagesOptions) (sdk.ListResult[metamodel.Page], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *PageService) ListPage(ctx context.Context, opts *ListPagesOptions) (sdk.ListResult[metamodel.Page], error) {
	var out sdk.ListResult[metamodel.Page]
	err := s.c.DoWithQuery(ctx, http.MethodGet, pagesBasePath, opts, nil, &out)
	return out, err
}

func (s *PageService) Get(ctx context.Context, slug string) (metamodel.Page, error) {
	var out metamodel.Page
	err := s.c.Do(ctx, http.MethodGet, pagesBasePath+"/"+slug, nil, &out)
	return out, err
}

func (s *PageService) Create(ctx context.Context, req CreatePageRequest) (metamodel.Page, error) {
	var out metamodel.Page
	err := s.c.Do(ctx, http.MethodPost, pagesBasePath, req, &out)
	return out, err
}

// Upsert calls `PUT /meta/v1/pages/:slug` — the idempotent upload path
// used by `pro module deploy`.
func (s *PageService) Upsert(ctx context.Context, slug string, req CreatePageRequest) (metamodel.Page, error) {
	var out metamodel.Page
	req.Slug = slug
	err := s.c.Do(ctx, http.MethodPut, pagesBasePath+"/"+slug, req, &out)
	return out, err
}

func (s *PageService) Update(ctx context.Context, slug string, req UpdatePageRequest) (metamodel.Page, error) {
	var out metamodel.Page
	err := s.c.Do(ctx, http.MethodPatch, pagesBasePath+"/"+slug, req, &out)
	return out, err
}

func (s *PageService) Delete(ctx context.Context, slug string) error {
	return s.c.Do(ctx, http.MethodDelete, pagesBasePath+"/"+slug, nil, nil)
}
