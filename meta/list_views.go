package meta

import (
	"context"
	"net/http"
	metamodel "go.proteos.ai/model/meta"
	sdk "go.proteos.ai/sdk"
)

const listViewsBasePath = "/meta/v1/list-views"

// ListViewServiceAPI is the contract a ListViewService satisfies.
type ListViewServiceAPI interface {
	List(opts *ListListViewsOptions) *sdk.PageIterator[metamodel.ListView, ListListViewsOptions]
	ListPage(ctx context.Context, opts *ListListViewsOptions) (sdk.ListResult[metamodel.ListView], error)
	Get(ctx context.Context, slug string) (metamodel.ListView, error)
	Create(ctx context.Context, req CreateListViewRequest) (metamodel.ListView, error)
	Upsert(ctx context.Context, req CreateListViewRequest) (metamodel.ListView, error)
	UpsertBySlug(ctx context.Context, slug string, req CreateListViewRequest) (metamodel.ListView, error)
	Update(ctx context.Context, slug string, req UpdateListViewRequest) (metamodel.ListView, error)
	Delete(ctx context.Context, slug string) error
}

// ListViewService manages list view configurations.
type ListViewService struct{ c *sdk.Client }

var _ ListViewServiceAPI = (*ListViewService)(nil)

func (s *ListViewService) List(opts *ListListViewsOptions) *sdk.PageIterator[metamodel.ListView, ListListViewsOptions] {
	o := ListListViewsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListListViewsOptions) (sdk.ListResult[metamodel.ListView], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *ListViewService) ListPage(ctx context.Context, opts *ListListViewsOptions) (sdk.ListResult[metamodel.ListView], error) {
	var out sdk.ListResult[metamodel.ListView]
	err := s.c.DoWithQuery(ctx, http.MethodGet, listViewsBasePath, opts, nil, &out)
	return out, err
}

func (s *ListViewService) Get(ctx context.Context, slug string) (metamodel.ListView, error) {
	var out metamodel.ListView
	err := s.c.Do(ctx, http.MethodGet, listViewsBasePath+"/"+slug, nil, &out)
	return out, err
}

func (s *ListViewService) Create(ctx context.Context, req CreateListViewRequest) (metamodel.ListView, error) {
	var out metamodel.ListView
	err := s.c.Do(ctx, http.MethodPost, listViewsBasePath, req, &out)
	return out, err
}

func (s *ListViewService) Upsert(ctx context.Context, req CreateListViewRequest) (metamodel.ListView, error) {
	var out metamodel.ListView
	err := s.c.Do(ctx, http.MethodPost, listViewsBasePath+"/upsert", req, &out)
	return out, err
}

// UpsertBySlug calls `PUT /meta/v1/list-views/:slug`.
func (s *ListViewService) UpsertBySlug(ctx context.Context, slug string, req CreateListViewRequest) (metamodel.ListView, error) {
	var out metamodel.ListView
	req.Slug = slug
	err := s.c.Do(ctx, http.MethodPut, listViewsBasePath+"/"+slug, req, &out)
	return out, err
}

func (s *ListViewService) Update(ctx context.Context, slug string, req UpdateListViewRequest) (metamodel.ListView, error) {
	var out metamodel.ListView
	err := s.c.Do(ctx, http.MethodPatch, listViewsBasePath+"/"+slug, req, &out)
	return out, err
}

func (s *ListViewService) Delete(ctx context.Context, slug string) error {
	return s.c.Do(ctx, http.MethodDelete, listViewsBasePath+"/"+slug, nil, nil)
}
