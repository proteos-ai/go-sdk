package knowledge

import (
	"context"
	"net/http"

	knowledgemodel "go.proteos.ai/model/knowledge"
	knowledgeapi "go.proteos.ai/model/knowledge/api"
	sdk "go.proteos.ai/sdk"
)

const recordLinksBasePath = "/knowledge/v1/record-links"

// RecordLinkServiceAPI is the contract a RecordLinkService satisfies.
type RecordLinkServiceAPI interface {
	List(opts *ListRecordLinksOptions) *sdk.PageIterator[knowledgemodel.KnowledgeRecordLink, ListRecordLinksOptions]
	ListPage(ctx context.Context, opts *ListRecordLinksOptions) (sdk.ListResult[knowledgemodel.KnowledgeRecordLink], error)
	Get(ctx context.Context, id string) (knowledgemodel.KnowledgeRecordLink, error)
	Create(ctx context.Context, request knowledgeapi.CreateKnowledgeRecordLinkRequest) (knowledgemodel.KnowledgeRecordLink, error)
	Delete(ctx context.Context, id string) error
}

// RecordLinkService manages directed node→record links — the edge from a
// knowledge node to the data-service record it describes. Links are
// immutable: create and delete only, no update.
type RecordLinkService struct{ c *sdk.Client }

var _ RecordLinkServiceAPI = (*RecordLinkService)(nil)

// List returns a PageIterator over record links.
func (s *RecordLinkService) List(opts *ListRecordLinksOptions) *sdk.PageIterator[knowledgemodel.KnowledgeRecordLink, ListRecordLinksOptions] {
	o := ListRecordLinksOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListRecordLinksOptions) (sdk.ListResult[knowledgemodel.KnowledgeRecordLink], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

// ListPage fetches a single page of record links.
func (s *RecordLinkService) ListPage(ctx context.Context, opts *ListRecordLinksOptions) (sdk.ListResult[knowledgemodel.KnowledgeRecordLink], error) {
	var out sdk.ListResult[knowledgemodel.KnowledgeRecordLink]
	err := s.c.DoWithQuery(ctx, http.MethodGet, recordLinksBasePath, opts, nil, &out)
	return out, err
}

// Get returns a single record link by id.
func (s *RecordLinkService) Get(ctx context.Context, id string) (knowledgemodel.KnowledgeRecordLink, error) {
	var out knowledgemodel.KnowledgeRecordLink
	err := s.c.Do(ctx, http.MethodGet, recordLinksBasePath+"/"+id, nil, &out)
	return out, err
}

// Create posts a directed node→record link.
func (s *RecordLinkService) Create(ctx context.Context, request knowledgeapi.CreateKnowledgeRecordLinkRequest) (knowledgemodel.KnowledgeRecordLink, error) {
	var out knowledgemodel.KnowledgeRecordLink
	err := s.c.Do(ctx, http.MethodPost, recordLinksBasePath, request, &out)
	return out, err
}

// Delete removes a record link.
func (s *RecordLinkService) Delete(ctx context.Context, id string) error {
	return s.c.Do(ctx, http.MethodDelete, recordLinksBasePath+"/"+id, nil, nil)
}
