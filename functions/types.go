package functions

import (
	functionsmodel "go.proteos.ai/model/functions"
	metamodel "go.proteos.ai/model/meta"
)

// ListOptions are the pagination + sort fields shared across every
// list endpoint.
type ListOptions struct {
	Page      int    `query:"page"`
	PageSize  int    `query:"page_size"`
	SortBy    string `query:"sort_by,omitempty"`
	SortOrder string `query:"sort_direction,omitempty"`
}

// ----------------------------------------------------------------------
// Hook

type ListHooksOptions struct {
	ListOptions
	Slug       string                   `query:"slug,omitempty"`
	ModuleSlug string                   `query:"module_slug,omitempty"`
	EntitySlug string                   `query:"entity,omitempty"`
	Event      functionsmodel.HookEvent `query:"event,omitempty"`
	IsActive   *bool                    `query:"is_active,omitempty"`
}

// DeployHookRequest is the JSON metadata side of the multipart deploy.
// The entry-point Go file is, by convention, always `./main.go` next to
// the manifest — not carried on the wire.
type DeployHookRequest struct {
	Slug       string                   `json:"slug"`
	ModuleSlug string                   `json:"module_slug"`
	EntitySlug string                   `json:"entity"`
	Event      functionsmodel.HookEvent `json:"event"`
}

// PatchHookRequest is the JSON body for `PATCH /api/v1/hooks/:slug`.
// Lifecycle transitions (activate / deactivate) have dedicated methods.
type PatchHookRequest struct {
	IsActive *bool `json:"is_active,omitempty"`
}

type ListHookLogsOptions struct {
	Follow bool   `query:"follow,omitempty"`
	Since  string `query:"since,omitempty"`
	Level  string `query:"level,omitempty"`
}

// ----------------------------------------------------------------------
// Action

type ListActionsOptions struct {
	ListOptions
	Slug       string                     `query:"slug,omitempty"`
	ModuleSlug string                     `query:"module_slug,omitempty"`
	EntitySlug string                     `query:"entity,omitempty"`
	Scope      functionsmodel.ActionScope `query:"scope,omitempty"`
	IsActive   *bool                      `query:"is_active,omitempty"`
	IsPublic   *bool                      `query:"is_public,omitempty"`
}

// DeployActionRequest is the JSON metadata side of the multipart deploy.
// EntitySlug is required when Scope == ActionScopeEntity. The
// entry-point Go file is, by convention, always `./main.go` next to the
// manifest — not carried on the wire.
type DeployActionRequest struct {
	Slug       string                     `json:"slug"`
	ModuleSlug string                     `json:"module_slug"`
	Scope      functionsmodel.ActionScope `json:"scope"`
	EntitySlug *string                    `json:"entity,omitempty"`
	Name       string                     `json:"name"`
	// IsPublic opts a global action into public (unauthenticated) dispatch.
	IsPublic bool                  `json:"is_public,omitempty"`
	Params   []metamodel.Attribute `json:"params"`
	Returns  []metamodel.Attribute `json:"returns"`
}

// PatchActionRequest is the JSON body for `PATCH /api/v1/actions/:slug`.
type PatchActionRequest struct {
	IsActive *bool `json:"is_active,omitempty"`
	IsPublic *bool `json:"is_public,omitempty"`
}
