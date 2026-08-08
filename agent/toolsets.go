package agent

import (
	"context"
	"net/http"

	agentmodel "go.proteos.ai/model/agent"
	agentapi "go.proteos.ai/model/agent/api"
	sdk "go.proteos.ai/sdk"
)

const toolsetsBasePath = "/agents/v1/toolsets"

// ToolsetService manages toolsets — the hardcoded platform toolsets (read-only
// groups of the platform MCP server's tools) merged with the org's custom
// toolsets (groups of its own Tool keys). Writes apply to custom toolsets only;
// a platform key is rejected server-side with `toolset_read_only`.
type ToolsetService struct{ c *sdk.Client }

func (s *ToolsetService) List(opts *ListToolsetsOptions) *sdk.PageIterator[agentmodel.Toolset, ListToolsetsOptions] {
	o := ListToolsetsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListToolsetsOptions) (sdk.ListResult[agentmodel.Toolset], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *ToolsetService) ListPage(ctx context.Context, opts *ListToolsetsOptions) (sdk.ListResult[agentmodel.Toolset], error) {
	var out sdk.ListResult[agentmodel.Toolset]
	err := s.c.DoWithQuery(ctx, http.MethodGet, toolsetsBasePath, opts, nil, &out)
	return out, err
}

func (s *ToolsetService) Get(ctx context.Context, key string) (agentmodel.Toolset, error) {
	var out agentmodel.Toolset
	err := s.c.Do(ctx, http.MethodGet, toolsetsBasePath+"/"+key, nil, &out)
	return out, err
}

// ListTools returns the tools inside a toolset — platform members proxied from
// the platform MCP server, custom members summarized from the org's Tool rows.
func (s *ToolsetService) ListTools(ctx context.Context, key string) ([]agentmodel.ToolsetToolSummary, error) {
	var out agentapi.ListToolsetToolsResponse
	err := s.c.Do(ctx, http.MethodGet, toolsetsBasePath+"/"+key+"/tools", nil, &out)
	return out.Data, err
}

func (s *ToolsetService) Create(ctx context.Context, req agentapi.CreateToolsetRequest) (agentmodel.Toolset, error) {
	var out agentmodel.Toolset
	err := s.c.Do(ctx, http.MethodPost, toolsetsBasePath, req, &out)
	return out, err
}

func (s *ToolsetService) Update(ctx context.Context, key string, req agentapi.UpdateToolsetRequest) (agentmodel.Toolset, error) {
	var out agentmodel.Toolset
	err := s.c.Do(ctx, http.MethodPatch, toolsetsBasePath+"/"+key, req, &out)
	return out, err
}

// UpsertByKey calls `PUT /agents/v1/toolsets/:key` — idempotent create-or-replace
// used by `pro module deploy`.
func (s *ToolsetService) UpsertByKey(ctx context.Context, key string, req agentapi.CreateToolsetRequest) (agentmodel.Toolset, error) {
	var out agentmodel.Toolset
	req.Key = key
	err := s.c.Do(ctx, http.MethodPut, toolsetsBasePath+"/"+key, req, &out)
	return out, err
}

func (s *ToolsetService) Delete(ctx context.Context, key string) error {
	return s.c.Do(ctx, http.MethodDelete, toolsetsBasePath+"/"+key, nil, nil)
}
