package knowledge

import (
	"context"
	"net/http"

	sdk "go.proteos.ai/sdk"
	knowledgemodel "go.proteos.ai/model/knowledge"
	knowledgeapi "go.proteos.ai/model/knowledge/api"
)

const nodesBasePath = "/knowledge/v1/nodes"

// NodeServiceAPI is the contract a NodeService satisfies.
type NodeServiceAPI interface {
	List(opts *ListNodesOptions) *sdk.PageIterator[knowledgemodel.KnowledgeNodeMetadata, ListNodesOptions]
	ListPage(ctx context.Context, opts *ListNodesOptions) (sdk.ListResult[knowledgemodel.KnowledgeNodeMetadata], error)
	Get(ctx context.Context, id string) (knowledgemodel.KnowledgeNodeMetadata, error)
	Create(ctx context.Context, request knowledgeapi.CreateKnowledgeNodeRequest) (knowledgemodel.KnowledgeNode, error)
	Update(ctx context.Context, id string, request knowledgeapi.UpdateKnowledgeNodeRequest) (knowledgemodel.KnowledgeNodeMetadata, error)
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, request knowledgeapi.SearchNodesRequest) (sdk.ListResult[knowledgemodel.KnowledgeNodeSearchResult], error)
	Neighbors(ctx context.Context, id string, opts *NeighborsOptions) (knowledgemodel.NodeNeighborhood, error)
	GetContent(ctx context.Context, id string) (knowledgeapi.ContentResponse, error)
	WriteContent(ctx context.Context, id, content string) (knowledgeapi.WriteContentResponse, error)
	EditContent(ctx context.Context, id string, request knowledgeapi.EditContentRequest) (knowledgeapi.EditContentResponse, error)
	ListLabels(ctx context.Context, id string) ([]knowledgemodel.KnowledgeLabel, error)
	AttachLabel(ctx context.Context, id, labelID string) (knowledgemodel.KnowledgeNodeLabel, error)
	DetachLabel(ctx context.Context, id, labelID string) error
}

// NodeService manages knowledge nodes and their content, search, graph neighbors
// and labels via the knowledge-service. Reads return node metadata (no body) by
// default; the body is fetched explicitly with GetContent.
type NodeService struct{ c *sdk.Client }

var _ NodeServiceAPI = (*NodeService)(nil)

// List returns a PageIterator over node metadata.
func (s *NodeService) List(opts *ListNodesOptions) *sdk.PageIterator[knowledgemodel.KnowledgeNodeMetadata, ListNodesOptions] {
	o := ListNodesOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListNodesOptions) (sdk.ListResult[knowledgemodel.KnowledgeNodeMetadata], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

// ListPage fetches a single page of node metadata.
func (s *NodeService) ListPage(ctx context.Context, opts *ListNodesOptions) (sdk.ListResult[knowledgemodel.KnowledgeNodeMetadata], error) {
	var out sdk.ListResult[knowledgemodel.KnowledgeNodeMetadata]
	err := s.c.DoWithQuery(ctx, http.MethodGet, nodesBasePath, opts, nil, &out)
	return out, err
}

// Get returns a node's metadata (body excluded).
func (s *NodeService) Get(ctx context.Context, id string) (knowledgemodel.KnowledgeNodeMetadata, error) {
	var out knowledgemodel.KnowledgeNodeMetadata
	err := s.c.Do(ctx, http.MethodGet, nodesBasePath+"/"+id, nil, &out)
	return out, err
}

// Create posts a new node, optionally connecting it to labels/links in one call.
func (s *NodeService) Create(ctx context.Context, request knowledgeapi.CreateKnowledgeNodeRequest) (knowledgemodel.KnowledgeNode, error) {
	var out knowledgemodel.KnowledgeNode
	err := s.c.Do(ctx, http.MethodPost, nodesBasePath, request, &out)
	return out, err
}

// Update patches a node's metadata (title / status / summary).
func (s *NodeService) Update(ctx context.Context, id string, request knowledgeapi.UpdateKnowledgeNodeRequest) (knowledgemodel.KnowledgeNodeMetadata, error) {
	var out knowledgemodel.KnowledgeNodeMetadata
	err := s.c.Do(ctx, http.MethodPatch, nodesBasePath+"/"+id, request, &out)
	return out, err
}

// Delete hard-deletes a node; its links and label associations cascade.
func (s *NodeService) Delete(ctx context.Context, id string) error {
	return s.c.Do(ctx, http.MethodDelete, nodesBasePath+"/"+id, nil, nil)
}

// Search runs hybrid retrieval. Only keyword is wired today; semantic/hybrid
// return a 501 error until embeddings land.
func (s *NodeService) Search(ctx context.Context, request knowledgeapi.SearchNodesRequest) (sdk.ListResult[knowledgemodel.KnowledgeNodeSearchResult], error) {
	var out sdk.ListResult[knowledgemodel.KnowledgeNodeSearchResult]
	err := s.c.Do(ctx, http.MethodPost, nodesBasePath+"/actions/search", request, &out)
	return out, err
}

// Neighbors walks the graph outward from a node.
func (s *NodeService) Neighbors(ctx context.Context, id string, opts *NeighborsOptions) (knowledgemodel.NodeNeighborhood, error) {
	var out knowledgemodel.NodeNeighborhood
	err := s.c.DoWithQuery(ctx, http.MethodGet, nodesBasePath+"/"+id+"/neighbors", opts, nil, &out)
	return out, err
}

// GetContent reads a node's raw content body plus its line count.
func (s *NodeService) GetContent(ctx context.Context, id string) (knowledgeapi.ContentResponse, error) {
	var out knowledgeapi.ContentResponse
	err := s.c.Do(ctx, http.MethodGet, nodesBasePath+"/"+id+"/content", nil, &out)
	return out, err
}

// WriteContent overwrites a node's content body. Rejects derived (file/url) nodes.
func (s *NodeService) WriteContent(ctx context.Context, id, content string) (knowledgeapi.WriteContentResponse, error) {
	var out knowledgeapi.WriteContentResponse
	err := s.c.Do(ctx, http.MethodPut, nodesBasePath+"/"+id+"/content", knowledgeapi.WriteContentRequest{Content: content}, &out)
	return out, err
}

// EditContent applies a surgical anchored edit to a node's content body.
func (s *NodeService) EditContent(ctx context.Context, id string, request knowledgeapi.EditContentRequest) (knowledgeapi.EditContentResponse, error) {
	var out knowledgeapi.EditContentResponse
	err := s.c.Do(ctx, http.MethodPost, nodesBasePath+"/"+id+"/content/actions/edit", request, &out)
	return out, err
}

// ListLabels returns the labels attached to a node.
func (s *NodeService) ListLabels(ctx context.Context, id string) ([]knowledgemodel.KnowledgeLabel, error) {
	var out knowledgeapi.ListNodeLabelsResponse
	err := s.c.Do(ctx, http.MethodGet, nodesBasePath+"/"+id+"/labels", nil, &out)
	return out.Data, err
}

// AttachLabel attaches an existing label to a node.
func (s *NodeService) AttachLabel(ctx context.Context, id, labelID string) (knowledgemodel.KnowledgeNodeLabel, error) {
	var out knowledgemodel.KnowledgeNodeLabel
	err := s.c.Do(ctx, http.MethodPost, nodesBasePath+"/"+id+"/labels", knowledgeapi.AttachLabelRequest{LabelId: labelID}, &out)
	return out, err
}

// DetachLabel detaches a label from a node.
func (s *NodeService) DetachLabel(ctx context.Context, id, labelID string) error {
	return s.c.Do(ctx, http.MethodDelete, nodesBasePath+"/"+id+"/labels/"+labelID, nil, nil)
}
