package knowledge

import (
	"context"
	"net/http"

	sdk "go.proteos.ai/sdk"
	knowledgemodel "go.proteos.ai/model/knowledge"
	knowledgeapi "go.proteos.ai/model/knowledge/api"
)

const linksBasePath = "/knowledge/v1/links"

// LinkServiceAPI is the contract a LinkService satisfies.
type LinkServiceAPI interface {
	List(opts *ListLinksOptions) *sdk.PageIterator[knowledgemodel.KnowledgeLink, ListLinksOptions]
	ListPage(ctx context.Context, opts *ListLinksOptions) (sdk.ListResult[knowledgemodel.KnowledgeLink], error)
	Get(ctx context.Context, id string) (knowledgemodel.KnowledgeLink, error)
	Create(ctx context.Context, request knowledgeapi.CreateKnowledgeLinkRequest) (knowledgemodel.KnowledgeLink, error)
	Update(ctx context.Context, id string, request knowledgeapi.UpdateKnowledgeLinkRequest) (knowledgemodel.KnowledgeLink, error)
	Delete(ctx context.Context, id string) error
}

// LinkService manages typed edges (links) between knowledge nodes.
type LinkService struct{ c *sdk.Client }

var _ LinkServiceAPI = (*LinkService)(nil)

// List returns a PageIterator over links.
func (s *LinkService) List(opts *ListLinksOptions) *sdk.PageIterator[knowledgemodel.KnowledgeLink, ListLinksOptions] {
	o := ListLinksOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListLinksOptions) (sdk.ListResult[knowledgemodel.KnowledgeLink], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

// ListPage fetches a single page of links.
func (s *LinkService) ListPage(ctx context.Context, opts *ListLinksOptions) (sdk.ListResult[knowledgemodel.KnowledgeLink], error) {
	var out sdk.ListResult[knowledgemodel.KnowledgeLink]
	err := s.c.DoWithQuery(ctx, http.MethodGet, linksBasePath, opts, nil, &out)
	return out, err
}

// Get returns a single link by id.
func (s *LinkService) Get(ctx context.Context, id string) (knowledgemodel.KnowledgeLink, error) {
	var out knowledgemodel.KnowledgeLink
	err := s.c.Do(ctx, http.MethodGet, linksBasePath+"/"+id, nil, &out)
	return out, err
}

// Create posts a typed edge between two nodes.
func (s *LinkService) Create(ctx context.Context, request knowledgeapi.CreateKnowledgeLinkRequest) (knowledgemodel.KnowledgeLink, error) {
	var out knowledgemodel.KnowledgeLink
	err := s.c.Do(ctx, http.MethodPost, linksBasePath, request, &out)
	return out, err
}

// Update patches a link's type / description.
func (s *LinkService) Update(ctx context.Context, id string, request knowledgeapi.UpdateKnowledgeLinkRequest) (knowledgemodel.KnowledgeLink, error) {
	var out knowledgemodel.KnowledgeLink
	err := s.c.Do(ctx, http.MethodPatch, linksBasePath+"/"+id, request, &out)
	return out, err
}

// Delete removes a link.
func (s *LinkService) Delete(ctx context.Context, id string) error {
	return s.c.Do(ctx, http.MethodDelete, linksBasePath+"/"+id, nil, nil)
}
