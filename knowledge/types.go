// Package knowledge provides the nodes, links and labels services backed by the
// knowledge-service (graph-native knowledge base).
//
// Resource shapes (KnowledgeNode, KnowledgeLink, KnowledgeLabel, …) are imported
// from go.proteos.ai/model/knowledge and request bodies from
// go.proteos.ai/model/knowledge/api and used directly. Only the list/query
// option types (which need `query:` tags) are defined here.
package knowledge

// ListOptions are the pagination + sort fields shared across every knowledge
// list endpoint.
type ListOptions struct {
	Page          int    `query:"page"`
	PageSize      int    `query:"page_size"`
	SortBy        string `query:"sort_by,omitempty"`
	SortDirection string `query:"sort_direction,omitempty"`
}

// ListNodesOptions filters and paginates GET /knowledge/v1/nodes.
type ListNodesOptions struct {
	ListOptions
	Type          string `query:"type,omitempty"`
	Status        string `query:"status,omitempty"`
	Title         string `query:"title,omitempty"`
	TitleContains string `query:"title[contains],omitempty"`
}

// ListLinksOptions filters and paginates GET /knowledge/v1/links.
type ListLinksOptions struct {
	ListOptions
	FromID string `query:"from_id,omitempty"`
	ToID   string `query:"to_id,omitempty"`
	Type   string `query:"type,omitempty"`
}

// ListRecordLinksOptions filters and paginates GET /knowledge/v1/record-links.
type ListRecordLinksOptions struct {
	ListOptions
	NodeID     string `query:"node_id,omitempty"`
	EntitySlug string `query:"entity_slug,omitempty"`
	RecordID   string `query:"record_id,omitempty"`
}

// ListLabelsOptions filters and paginates GET /knowledge/v1/labels.
type ListLabelsOptions struct {
	ListOptions
	Slug         string `query:"slug,omitempty"`
	Name         string `query:"name,omitempty"`
	NameContains string `query:"name[contains],omitempty"`
}

// NeighborsOptions is the query for GET /knowledge/v1/nodes/{id}/neighbors.
// Direction is out|in|both (default both); LinkTypes is a comma-separated list
// (empty = any); Depth defaults to 1 and is clamped server-side.
type NeighborsOptions struct {
	Direction string `query:"direction,omitempty"`
	LinkTypes string `query:"link_types,omitempty"`
	Depth     int    `query:"depth,omitempty"`
}
