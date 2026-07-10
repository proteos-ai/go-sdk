// Package agent provides services for managing agent-service configuration
// resources (agents, prompts, skills, tools, mcp-servers) over the platform API
// at /agents/v1.
//
// Resource shapes (Agent, Prompt, PromptVersion, Tool, McpServer, Skill, …) come
// from go.proteos.ai/model/agent; the wire-format request types from
// go.proteos.ai/model/agent/api are reused directly by this package's methods.
// Only the list-options types (query-tagged for the SDK's query encoder) are
// defined locally.
//
// agent-service returns BARE single objects (no {data} envelope); only list and
// version-list responses wrap. The methods below decode accordingly.
package agent

// ListOptions are the pagination + sort fields shared across the agent list
// endpoints. Pages are 0-indexed.
type ListOptions struct {
	Page      int    `query:"page"`
	PageSize  int    `query:"page_size"`
	SortBy    string `query:"sort_by,omitempty"`
	SortOrder string `query:"sort_direction,omitempty"`
}

type ListAgentsOptions struct {
	ListOptions
	Key        string `query:"key,omitempty"`
	Name       string `query:"name,omitempty"`
	ModuleSlug string `query:"module_slug,omitempty"`
}

type ListPromptsOptions struct {
	ListOptions
	Key        string `query:"key,omitempty"`
	Name       string `query:"name,omitempty"`
	ModuleSlug string `query:"module_slug,omitempty"`
}

type ListToolsOptions struct {
	ListOptions
	Key        string `query:"key,omitempty"`
	Name       string `query:"name,omitempty"`
	ModuleSlug string `query:"module_slug,omitempty"`
	Kind       string `query:"kind,omitempty"`
	// Expand opts a list into read-time schema resolution per tool (expand=schema).
	Expand string `query:"expand,omitempty"`
}

type ListMcpServersOptions struct {
	ListOptions
	Key        string `query:"key,omitempty"`
	Name       string `query:"name,omitempty"`
	ModuleSlug string `query:"module_slug,omitempty"`
}

type ListSkillsOptions struct {
	ListOptions
	Key        string `query:"key,omitempty"`
	Name       string `query:"name,omitempty"`
	ModuleSlug string `query:"module_slug,omitempty"`
}
