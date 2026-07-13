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
	Slug          string                `json:"slug"`
	Name          string                `json:"name"`
	IsRemote      bool                  `json:"is_remote"`
	ModuleSlug    string                `json:"module_slug"`
	Description   string                `json:"description"`
	TitleTemplate string                `json:"title_template"`
	Attributes    []metamodel.Attribute `json:"attributes"`
}

type UpdateEntityRequest struct {
	Name          *string                `json:"name,omitempty"`
	IsRemote      *bool                  `json:"is_remote,omitempty"`
	ModuleSlug    *string                `json:"module_slug,omitempty"`
	Description   *string                `json:"description,omitempty"`
	TitleTemplate *string                `json:"title_template,omitempty"`
	Attributes    *[]metamodel.Attribute `json:"attributes,omitempty"`
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
	Slug       string                 `json:"slug"`
	ModuleSlug string                 `json:"module_slug"`
	EntitySlug string                 `json:"entity_slug"`
	Name       string                 `json:"name"`
	Columns    []metamodel.Column     `json:"columns"`
	Sorting    []metamodel.SortConfig `json:"sorting"`
	Filters    []common.FilterGroup   `json:"filters"`
}

type UpdateListRequest struct {
	Name    *string                 `json:"name,omitempty"`
	Columns *[]metamodel.Column     `json:"columns,omitempty"`
	Sorting *[]metamodel.SortConfig `json:"sorting,omitempty"`
	Filters *[]common.FilterGroup   `json:"filters,omitempty"`
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
