package knowledge

import (
	"context"
	"net/http"

	sdk "go.proteos.ai/sdk"
	knowledgemodel "go.proteos.ai/model/knowledge"
	knowledgeapi "go.proteos.ai/model/knowledge/api"
)

const labelsBasePath = "/knowledge/v1/labels"

// LabelServiceAPI is the contract a LabelService satisfies.
type LabelServiceAPI interface {
	List(opts *ListLabelsOptions) *sdk.PageIterator[knowledgemodel.KnowledgeLabel, ListLabelsOptions]
	ListPage(ctx context.Context, opts *ListLabelsOptions) (sdk.ListResult[knowledgemodel.KnowledgeLabel], error)
	Get(ctx context.Context, id string) (knowledgemodel.KnowledgeLabel, error)
	Create(ctx context.Context, request knowledgeapi.CreateKnowledgeLabelRequest) (knowledgemodel.KnowledgeLabel, error)
	Update(ctx context.Context, id string, request knowledgeapi.UpdateKnowledgeLabelRequest) (knowledgemodel.KnowledgeLabel, error)
	Delete(ctx context.Context, id string) error
}

// LabelService manages knowledge labels. Attach/detach to a node via the node
// service's label methods.
type LabelService struct{ c *sdk.Client }

var _ LabelServiceAPI = (*LabelService)(nil)

// List returns a PageIterator over labels.
func (s *LabelService) List(opts *ListLabelsOptions) *sdk.PageIterator[knowledgemodel.KnowledgeLabel, ListLabelsOptions] {
	o := ListLabelsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListLabelsOptions) (sdk.ListResult[knowledgemodel.KnowledgeLabel], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

// ListPage fetches a single page of labels.
func (s *LabelService) ListPage(ctx context.Context, opts *ListLabelsOptions) (sdk.ListResult[knowledgemodel.KnowledgeLabel], error) {
	var out sdk.ListResult[knowledgemodel.KnowledgeLabel]
	err := s.c.DoWithQuery(ctx, http.MethodGet, labelsBasePath, opts, nil, &out)
	return out, err
}

// Get returns a single label by id.
func (s *LabelService) Get(ctx context.Context, id string) (knowledgemodel.KnowledgeLabel, error) {
	var out knowledgemodel.KnowledgeLabel
	err := s.c.Do(ctx, http.MethodGet, labelsBasePath+"/"+id, nil, &out)
	return out, err
}

// Create posts a new label.
func (s *LabelService) Create(ctx context.Context, request knowledgeapi.CreateKnowledgeLabelRequest) (knowledgemodel.KnowledgeLabel, error) {
	var out knowledgemodel.KnowledgeLabel
	err := s.c.Do(ctx, http.MethodPost, labelsBasePath, request, &out)
	return out, err
}

// Update patches a label.
func (s *LabelService) Update(ctx context.Context, id string, request knowledgeapi.UpdateKnowledgeLabelRequest) (knowledgemodel.KnowledgeLabel, error) {
	var out knowledgemodel.KnowledgeLabel
	err := s.c.Do(ctx, http.MethodPatch, labelsBasePath+"/"+id, request, &out)
	return out, err
}

// Delete removes a label; its node associations cascade.
func (s *LabelService) Delete(ctx context.Context, id string) error {
	return s.c.Do(ctx, http.MethodDelete, labelsBasePath+"/"+id, nil, nil)
}
