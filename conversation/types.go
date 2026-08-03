// Package conversation provides services for managing conversation-service
// resources over the platform API at /conversations/v1. Currently: conversation
// types — the per-org taxonomy the pre-summary classifier reads, deployable via
// `pro module deploy` (conversation-types/<key>.json).
//
// Resource shapes come from go.proteos.ai/model/conversation; the wire-format
// request types from go.proteos.ai/model/conversation/api are reused directly.
// Only the list-options types (query-tagged for the SDK's query encoder) are
// defined locally.
//
// conversation-service returns BARE single objects (no {data} envelope); only
// list responses wrap. The methods below decode accordingly.
package conversation

// ListOptions are the pagination + sort fields of the conversation-service
// list endpoints. Pages are 0-indexed.
type ListOptions struct {
	Page      int    `query:"page"`
	PageSize  int    `query:"page_size"`
	SortBy    string `query:"sort_by,omitempty"`
	SortOrder string `query:"sort_direction,omitempty"`
}

type ListConversationTypesOptions struct {
	ListOptions
	Search     string `query:"search,omitempty"`
	ModuleSlug string `query:"module_slug,omitempty"`
}
