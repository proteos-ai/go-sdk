// Package meta provides services for managing the platform's metadata
// (entities, modules, lists, list views, components, pages, variables,
// menu configurations).
//
// Resource shapes (Entity, Module, List, etc.) are imported from
// go.proteos.ai/model and used directly. Only the request/options
// types are defined here.
package meta

import (
	"go.proteos.ai/model/common"
	"go.proteos.ai/model/meta"
)

// ListOptions are the pagination + sort fields shared across every meta
// list endpoint.
type ListOptions struct {
	Page      int    `query:"page"`
	PageSize  int    `query:"page_size"`
	SortBy    string `query:"sort_by,omitempty"`
	SortOrder string `query:"sort_direction,omitempty"`
}

// ----------------------------------------------------------------------
// Entity

type ListEntitiesOptions struct {
	ListOptions
	Slug       string `query:"slug,omitempty"`
	Name       string `query:"name,omitempty"`
	IsRemote   *bool  `query:"is_remote,omitempty"`
	ModuleSlug string `query:"module_slug,omitempty"`
	WithSchema bool   `query:"with_schema,omitempty"`
}

type CreateEntityRequest struct {
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	IsRemote bool   `json:"is_remote"`
	// PublicRecordAccess opts all records of the entity into unauthenticated
	// access via /data/v1/public/... (only ["read"] honored today).
	// Full-replacement on upsert: an upsert without the field resets it to
	// private.
	PublicRecordAccess common.PublicAccess   `json:"public_record_access"`
	ModuleSlug         string                `json:"module_slug"`
	Description        string                `json:"description"`
	TitleTemplate      string                `json:"title_template"`
	Attributes         []metamodel.Attribute `json:"attributes"`
}

type UpdateEntityRequest struct {
	Name               *string                `json:"name,omitempty"`
	IsRemote           *bool                  `json:"is_remote,omitempty"`
	PublicRecordAccess *common.PublicAccess   `json:"public_record_access,omitempty"`
	ModuleSlug         *string                `json:"module_slug,omitempty"`
	Description        *string                `json:"description,omitempty"`
	TitleTemplate      *string                `json:"title_template,omitempty"`
	Attributes         *[]metamodel.Attribute `json:"attributes,omitempty"`
}

// ----------------------------------------------------------------------
// Module

type ListModulesOptions struct {
	ListOptions
	Slug          string `query:"slug,omitempty"`
	Name          string `query:"name,omitempty"`
	IsDeactivated *bool  `query:"is_deactivated,omitempty"`
	FileID        string `query:"file_id,omitempty"`
	Status        string `query:"status,omitempty"`
	Version       string `query:"version,omitempty"`
}

// DeployModuleRequest is the JSON metadata field of the multipart deploy.
type DeployModuleRequest struct {
	Slug        string `json:"slug"`
	Version     string `json:"version"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ----------------------------------------------------------------------
// Variable

type ListVariablesOptions struct {
	ListOptions
	ID       string `query:"id,omitempty"`
	Key      string `query:"key,omitempty"`
	IsSecret *bool  `query:"is_secret,omitempty"`
	Module   string `query:"module,omitempty"`
}

type CreateVariableRequest struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret"`
	Module   string `json:"module"`
}

type UpdateVariableRequest struct {
	Value *string `json:"value,omitempty"`
}

// ----------------------------------------------------------------------
// Component

type ListComponentsOptions struct {
	ListOptions
	Slug       string `query:"slug,omitempty"`
	Name       string `query:"name,omitempty"`
	ModuleSlug string `query:"module_slug,omitempty"`
}

type CreateComponentRequest struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	ModuleSlug  string `json:"module_slug"`
	Description string `json:"description"`
	// Storage file ids + props schema, populated by `pro module deploy` after
	// it bundles + uploads the component (LUM-75). omitempty so metadata-only
	// callers (e.g. a rename) don't blank them out on the server.
	BundleFileId string         `json:"bundle_file_id,omitempty"`
	SourceFileId string         `json:"source_file_id,omitempty"`
	PropsSchema  map[string]any `json:"props_schema,omitempty"`
	// IsPublic opts the compiled bundle into UNAUTHENTICATED serving (public
	// pages may only reference public components). Manifest-driven full
	// replacement: a deploy without the field sets false.
	IsPublic bool `json:"is_public"`
}

type UpdateComponentRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// ----------------------------------------------------------------------
// List

type ListListsOptions struct {
	ListOptions
	Slug       string `query:"slug,omitempty"`
	Name       string `query:"name,omitempty"`
	ModuleSlug string `query:"module_slug,omitempty"`
	EntitySlug string `query:"entity_slug,omitempty"`
}

type CreateListRequest struct {
	Slug          string                  `json:"slug"`
	ModuleSlug    string                  `json:"module_slug"`
	EntitySlug    string                  `json:"entity_slug"`
	Name          string                  `json:"name"`
	Columns       []metamodel.Column      `json:"columns"`
	Actions       []metamodel.PageAction  `json:"actions,omitempty"`
	SelectionMode metamodel.SelectionMode `json:"selection_mode,omitempty"`
	Sorting       []metamodel.SortConfig  `json:"sorting"`
	Filters       []common.FilterGroup    `json:"filters"`
}

type UpdateListRequest struct {
	Name          *string                  `json:"name,omitempty"`
	Columns       *[]metamodel.Column      `json:"columns,omitempty"`
	Actions       *[]metamodel.PageAction  `json:"actions,omitempty"`
	SelectionMode *metamodel.SelectionMode `json:"selection_mode,omitempty"`
	Sorting       *[]metamodel.SortConfig  `json:"sorting,omitempty"`
	Filters       *[]common.FilterGroup    `json:"filters,omitempty"`
}

// ----------------------------------------------------------------------
// ListView

type ListListViewsOptions struct {
	ListOptions
	Slug       string `query:"slug,omitempty"`
	Name       string `query:"name,omitempty"`
	ModuleSlug string `query:"module_slug,omitempty"`
	ListSlug   string `query:"list_slug,omitempty"`
}

type CreateListViewRequest struct {
	Slug       string                 `json:"slug"`
	ModuleSlug string                 `json:"module_slug"`
	ListSlug   string                 `json:"list_slug"`
	Name       string                 `json:"name"`
	Columns    []metamodel.Column     `json:"columns"`
	Sorting    []metamodel.SortConfig `json:"sorting,omitempty"`
	Filters    []common.FilterGroup   `json:"filters,omitempty"`
}

type UpdateListViewRequest struct {
	Name    *string                 `json:"name,omitempty"`
	Columns *[]metamodel.Column     `json:"columns,omitempty"`
	Sorting *[]metamodel.SortConfig `json:"sorting,omitempty"`
	Filters *[]common.FilterGroup   `json:"filters,omitempty"`
}

// ----------------------------------------------------------------------
// Page

type ListPagesOptions struct {
	ListOptions
	ID         string `query:"id,omitempty"`
	Slug       string `query:"slug,omitempty"`
	Name       string `query:"name,omitempty"`
	ModuleSlug string `query:"module_slug,omitempty"`
	Type       string `query:"type,omitempty"`
	EntitySlug string `query:"entity_slug,omitempty"`
}

type CreatePageRequest struct {
	Slug       string                 `json:"slug"`
	Name       string                 `json:"name"`
	ModuleSlug string                 `json:"module_slug"`
	Type       metamodel.PageType     `json:"type,omitempty"`
	EntitySlug string                 `json:"entity_slug,omitempty"`
	Actions    []metamodel.PageAction `json:"actions"`
	Layout     metamodel.PageLayout   `json:"layout"`
}

type UpdatePageRequest struct {
	Name    *string                 `json:"name,omitempty"`
	Actions *[]metamodel.PageAction `json:"actions,omitempty"`
	Layout  *metamodel.PageLayout   `json:"layout,omitempty"`
}

// ----------------------------------------------------------------------
// Menu Configuration

type ListMenuConfigurationsOptions struct {
	ListOptions
	ID         string `query:"id,omitempty"`
	Slug       string `query:"slug,omitempty"`
	Name       string `query:"name,omitempty"`
	ModuleSlug string `query:"module_slug,omitempty"`
	AppSlug    string `query:"app_slug,omitempty"`
	IsDefault  *bool  `query:"is_default,omitempty"`
}

type CreateMenuConfigurationRequest struct {
	Slug       string               `json:"slug"`
	ModuleSlug string               `json:"module_slug"`
	Name       string               `json:"name"`
	AppSlug    string               `json:"app_slug"`
	Items      []metamodel.MenuItem `json:"items"`
	IsDefault  bool                 `json:"is_default"`
}

type UpdateMenuConfigurationRequest struct {
	Name      *string               `json:"name,omitempty"`
	Items     *[]metamodel.MenuItem `json:"items,omitempty"`
	IsDefault *bool                 `json:"is_default,omitempty"`
}

// ----------------------------------------------------------------------
// EntityExtension — attributes contributed to an existing host entity by a
// resource that does not own it. The host's stored attributes are the
// materialized merge (each contributed attribute stamped with extension_key).

type ListEntityExtensionsOptions struct {
	ListOptions
	Key        string `query:"key,omitempty"`
	Name       string `query:"name,omitempty"`
	EntitySlug string `query:"entity_slug,omitempty"`
	ModuleSlug string `query:"module_slug,omitempty"`
}

type CreateEntityExtensionRequest struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// EntitySlug is the host entity; immutable after create.
	EntitySlug string `json:"entity_slug"`
	ModuleSlug string `json:"module_slug"`
	// Attributes are the contributed definitions: at least one, no platform
	// names, no extension_key (the server stamps it on the host).
	Attributes []metamodel.Attribute `json:"attributes"`
}

// UpdateEntityExtensionRequest is a partial update; attributes, when present,
// replaces the whole contributed list. entity_slug is immutable.
type UpdateEntityExtensionRequest struct {
	Name        *string                `json:"name,omitempty"`
	Description *string                `json:"description,omitempty"`
	ModuleSlug  *string                `json:"module_slug,omitempty"`
	Attributes  *[]metamodel.Attribute `json:"attributes,omitempty"`
}

// ----------------------------------------------------------------------
// PageExtension — anchored layout blocks contributed to an existing host page
// by a resource that does not own it. The host's layout is the materialized
// merge (each contributed element and tab stamped with extension_key).

type ListPageExtensionsOptions struct {
	ListOptions
	Key        string `query:"key,omitempty"`
	Name       string `query:"name,omitempty"`
	PageSlug   string `query:"page_slug,omitempty"`
	ModuleSlug string `query:"module_slug,omitempty"`
}

type CreatePageExtensionRequest struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// PageSlug is the host page; immutable after create.
	PageSlug   string `json:"page_slug"`
	ModuleSlug string `json:"module_slug"`
	// Placements are the contributed blocks: at least one, every element and
	// tab with an authored id, no extension_key (the server stamps it on the
	// host).
	Placements []metamodel.PageExtensionPlacement `json:"placements"`
}

// UpdatePageExtensionRequest is a partial update; placements, when present,
// replaces the whole contributed list. page_slug is immutable.
type UpdatePageExtensionRequest struct {
	Name        *string                             `json:"name,omitempty"`
	Description *string                             `json:"description,omitempty"`
	ModuleSlug  *string                             `json:"module_slug,omitempty"`
	Placements  *[]metamodel.PageExtensionPlacement `json:"placements,omitempty"`
}

// ----------------------------------------------------------------------
// MenuConfigurationExtension — items contributed to an existing host menu
// configuration by a resource that does not own it. The host's items are the
// materialized merge (each contributed item stamped with extension_key; a
// replaced / removed host item kept under replaced_item).

type ListMenuConfigurationExtensionsOptions struct {
	ListOptions
	Key        string `query:"key,omitempty"`
	Name       string `query:"name,omitempty"`
	MenuSlug   string `query:"menu_slug,omitempty"`
	ModuleSlug string `query:"module_slug,omitempty"`
}

type CreateMenuConfigurationExtensionRequest struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// MenuSlug is the host menu configuration; immutable after create.
	MenuSlug   string `json:"menu_slug"`
	ModuleSlug string `json:"module_slug"`
	// Placements are the contributed blocks: at least one, every item with an
	// authored id, no extension_key (the server stamps it on the host).
	Placements []metamodel.MenuConfigurationExtensionPlacement `json:"placements"`
}

// UpdateMenuConfigurationExtensionRequest is a partial update; placements,
// when present, replaces the whole contributed list. menu_slug is immutable.
type UpdateMenuConfigurationExtensionRequest struct {
	Name        *string                                          `json:"name,omitempty"`
	Description *string                                          `json:"description,omitempty"`
	ModuleSlug  *string                                          `json:"module_slug,omitempty"`
	Placements  *[]metamodel.MenuConfigurationExtensionPlacement `json:"placements,omitempty"`
}

// ----------------------------------------------------------------------
// ListExtension — columns and toolbar actions contributed to an existing host
// list by a resource that does not own it. The host's columns / actions are
// the materialized merge (each contributed one stamped with extension_key).

type ListListExtensionsOptions struct {
	ListOptions
	Key        string `query:"key,omitempty"`
	Name       string `query:"name,omitempty"`
	ListSlug   string `query:"list_slug,omitempty"`
	ModuleSlug string `query:"module_slug,omitempty"`
}

type CreateListExtensionRequest struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// ListSlug is the host list; immutable after create.
	ListSlug   string `json:"list_slug"`
	ModuleSlug string `json:"module_slug"`
	// Columns are the anchored column placements, Actions the appended
	// toolbar buttons — at least one of the two; no extension_key (the server
	// stamps it on the host).
	Columns []metamodel.ListExtensionColumnPlacement `json:"columns"`
	Actions []metamodel.PageAction                   `json:"actions"`
}

// UpdateListExtensionRequest is a partial update; columns / actions, when
// present, each replace their whole list. list_slug is immutable.
type UpdateListExtensionRequest struct {
	Name        *string                                   `json:"name,omitempty"`
	Description *string                                   `json:"description,omitempty"`
	ModuleSlug  *string                                   `json:"module_slug,omitempty"`
	Columns     *[]metamodel.ListExtensionColumnPlacement `json:"columns,omitempty"`
	Actions     *[]metamodel.PageAction                   `json:"actions,omitempty"`
}

// ----------------------------------------------------------------------
// AppConfiguration — the typed (app × profile) binding: home, menu, agents,
// record pages. profile_slug "" = the app's default configuration.

type ListAppConfigurationsOptions struct {
	ListOptions
	Slug        string `query:"slug,omitempty"`
	ModuleSlug  string `query:"module_slug,omitempty"`
	AppSlug     string `query:"app_slug,omitempty"`
	ProfileSlug string `query:"profile_slug,omitempty"`
	// IsDefault: true = only the default rows (no profile), false = only profile
	// overrides. Needed because an empty ProfileSlug is omitted, not sent.
	IsDefault *bool  `query:"is_default,omitempty"`
	MenuSlug  string `query:"menu_slug,omitempty"`
}

type CreateAppConfigurationRequest struct {
	Slug            string             `json:"slug"`
	ModuleSlug      string             `json:"module_slug"`
	AppSlug         string             `json:"app_slug"`
	ProfileSlug     string             `json:"profile_slug"`
	Home            *metamodel.AppHome `json:"home,omitempty"`
	MenuSlug        string             `json:"menu_slug,omitempty"`
	DefaultAgentKey string             `json:"default_agent_key,omitempty"`
	AgentKeys       []string           `json:"agent_keys,omitempty"`
	RecordPages     map[string]string  `json:"record_pages,omitempty"`
}

// UpdateAppConfigurationRequest is a partial update. Home is tri-state
// (absent = unchanged, null = clear, object = set) — send it through
// common.Optional.
type UpdateAppConfigurationRequest struct {
	ModuleSlug      *string                            `json:"module_slug,omitempty"`
	Home            common.Optional[metamodel.AppHome] `json:"home"`
	MenuSlug        *string                            `json:"menu_slug,omitempty"`
	DefaultAgentKey *string                            `json:"default_agent_key,omitempty"`
	AgentKeys       *[]string                          `json:"agent_keys,omitempty"`
	RecordPages     *map[string]string                 `json:"record_pages,omitempty"`
}

// ----------------------------------------------------------------------
// App

type ListAppsOptions struct {
	ListOptions
	Slug       string `query:"slug,omitempty"`
	Name       string `query:"name,omitempty"`
	ModuleSlug string `query:"module_slug,omitempty"`
	IconSlug   string `query:"icon_slug,omitempty"`
}

type CreateAppRequest struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	ModuleSlug  string `json:"module_slug"`
	IconSlug    string `json:"icon_slug"`
	Description string `json:"description,omitempty"`
}

type UpdateAppRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	IconSlug    *string `json:"icon_slug,omitempty"`
}

// ----------------------------------------------------------------------
// DesignReference

type ListDesignReferencesOptions struct {
	ListOptions
	ID          string `query:"id,omitempty"`
	Slug        string `query:"slug,omitempty"`
	Name        string `query:"name,omitempty"`
	Description string `query:"description,omitempty"`
}

type CreateDesignReferenceRequest struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// Content optionally seeds the DESIGN.md body at create time; omitempty so a
	// metadata-only create doesn't send an empty body. Edit later via SetContent.
	Content string `json:"content,omitempty"`
}

type UpdateDesignReferenceRequest struct {
	Slug        *string `json:"slug,omitempty"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// SetDesignReferenceContentRequest is the body of PUT /design-references/:id/content.
type SetDesignReferenceContentRequest struct {
	Content string `json:"content"`
}

// DesignReferenceContentResponse is the body of GET /design-references/:id/content.
type DesignReferenceContentResponse struct {
	Content string `json:"content"`
}
