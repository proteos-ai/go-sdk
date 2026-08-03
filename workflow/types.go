// Package workflow provides services for managing workflow-service resources
// (workflow definitions) over the platform API at /workflows/v1.
//
// Resource shapes (Workflow, WorkflowGraph, …) come from
// go.proteos.ai/model/workflow; the wire-format request types from
// go.proteos.ai/model/workflow/api are reused directly by this package's
// methods. Only the list-options types (query-tagged for the SDK's query
// encoder) are defined locally.
//
// workflow-service returns BARE single objects (no {data} envelope); only list
// responses wrap. The methods below decode accordingly.
package workflow

// ListOptions are the pagination + sort fields of the workflow list endpoint.
// Pages are 0-indexed.
type ListOptions struct {
	Page      int    `query:"page"`
	PageSize  int    `query:"page_size"`
	SortBy    string `query:"sort_by,omitempty"`
	SortOrder string `query:"sort_direction,omitempty"`
}

type ListWorkflowsOptions struct {
	ListOptions
	Name       string `query:"name,omitempty"`
	Status     string `query:"status,omitempty"`
	ModuleSlug string `query:"module_slug,omitempty"`
}
