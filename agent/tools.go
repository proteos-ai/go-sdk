package agent

import (
	"context"
	"net/http"

	agentmodel "go.proteos.ai/model/agent"
	agentapi "go.proteos.ai/model/agent/api"
	sdk "go.proteos.ai/sdk"
)

const toolsBasePath = "/agents/v1/tools"

// ToolService manages tools (registry entries over action / mcp / client bindings).
type ToolService struct{ c *sdk.Client }

func (s *ToolService) List(opts *ListToolsOptions) *sdk.PageIterator[agentmodel.Tool, ListToolsOptions] {
	o := ListToolsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListToolsOptions) (sdk.ListResult[agentmodel.Tool], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *ToolService) ListPage(ctx context.Context, opts *ListToolsOptions) (sdk.ListResult[agentmodel.Tool], error) {
	var out sdk.ListResult[agentmodel.Tool]
	err := s.c.DoWithQuery(ctx, http.MethodGet, toolsBasePath, opts, nil, &out)
	return out, err
}

func (s *ToolService) Get(ctx context.Context, key string) (agentmodel.Tool, error) {
	var out agentmodel.Tool
	err := s.c.Do(ctx, http.MethodGet, toolsBasePath+"/"+key, nil, &out)
	return out, err
}

func (s *ToolService) Create(ctx context.Context, req agentapi.CreateToolRequest) (agentmodel.Tool, error) {
	var out agentmodel.Tool
	err := s.c.Do(ctx, http.MethodPost, toolsBasePath, req, &out)
	return out, err
}

func (s *ToolService) Update(ctx context.Context, key string, req agentapi.UpdateToolRequest) (agentmodel.Tool, error) {
	var out agentmodel.Tool
	err := s.c.Do(ctx, http.MethodPatch, toolsBasePath+"/"+key, req, &out)
	return out, err
}

// UpsertByKey calls `PUT /agents/v1/tools/:key` — idempotent create-or-replace used
// by `pro module deploy`.
func (s *ToolService) UpsertByKey(ctx context.Context, key string, req agentapi.CreateToolRequest) (agentmodel.Tool, error) {
	var out agentmodel.Tool
	req.Key = key
	err := s.c.Do(ctx, http.MethodPut, toolsBasePath+"/"+key, req, &out)
	return out, err
}

func (s *ToolService) Delete(ctx context.Context, key string) error {
	return s.c.Do(ctx, http.MethodDelete, toolsBasePath+"/"+key, nil, nil)
}
