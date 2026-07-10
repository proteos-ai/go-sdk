package agent

import (
	"context"
	"net/http"

	agentmodel "go.proteos.ai/model/agent"
	agentapi "go.proteos.ai/model/agent/api"
	sdk "go.proteos.ai/sdk"
)

const mcpServersBasePath = "/agents/v1/mcp-servers"

// McpServerService manages registered MCP servers an org's agents can call.
type McpServerService struct{ c *sdk.Client }

func (s *McpServerService) List(opts *ListMcpServersOptions) *sdk.PageIterator[agentmodel.McpServer, ListMcpServersOptions] {
	o := ListMcpServersOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListMcpServersOptions) (sdk.ListResult[agentmodel.McpServer], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *McpServerService) ListPage(ctx context.Context, opts *ListMcpServersOptions) (sdk.ListResult[agentmodel.McpServer], error) {
	var out sdk.ListResult[agentmodel.McpServer]
	err := s.c.DoWithQuery(ctx, http.MethodGet, mcpServersBasePath, opts, nil, &out)
	return out, err
}

func (s *McpServerService) Get(ctx context.Context, key string) (agentmodel.McpServer, error) {
	var out agentmodel.McpServer
	err := s.c.Do(ctx, http.MethodGet, mcpServersBasePath+"/"+key, nil, &out)
	return out, err
}

func (s *McpServerService) Create(ctx context.Context, req agentapi.CreateMcpServerRequest) (agentmodel.McpServer, error) {
	var out agentmodel.McpServer
	err := s.c.Do(ctx, http.MethodPost, mcpServersBasePath, req, &out)
	return out, err
}

func (s *McpServerService) Update(ctx context.Context, key string, req agentapi.UpdateMcpServerRequest) (agentmodel.McpServer, error) {
	var out agentmodel.McpServer
	err := s.c.Do(ctx, http.MethodPatch, mcpServersBasePath+"/"+key, req, &out)
	return out, err
}

// UpsertByKey calls `PUT /agents/v1/mcp-servers/:key` — idempotent create-or-update
// used by `pro module deploy`. Credential-preserving server-side: a re-deploy never
// wipes a bearer token / OAuth client connected out-of-band.
func (s *McpServerService) UpsertByKey(ctx context.Context, key string, req agentapi.CreateMcpServerRequest) (agentmodel.McpServer, error) {
	var out agentmodel.McpServer
	req.Key = key
	err := s.c.Do(ctx, http.MethodPut, mcpServersBasePath+"/"+key, req, &out)
	return out, err
}

func (s *McpServerService) Delete(ctx context.Context, key string) error {
	return s.c.Do(ctx, http.MethodDelete, mcpServersBasePath+"/"+key, nil, nil)
}

// ListTools returns the live tools/list of a (connected) MCP server.
func (s *McpServerService) ListTools(ctx context.Context, key string) ([]agentmodel.McpToolSummary, error) {
	var out agentapi.ListMcpServerToolsResponse
	err := s.c.Do(ctx, http.MethodGet, mcpServersBasePath+"/"+key+"/tools", nil, &out)
	return out.Data, err
}

// ConnectionStatus reports an MCP server's derived OAuth connection state.
func (s *McpServerService) ConnectionStatus(ctx context.Context, key string) (agentmodel.McpConnectionStatus, error) {
	var out agentapi.GetMcpConnectionStatusResponse
	err := s.c.Do(ctx, http.MethodGet, mcpServersBasePath+"/"+key+"/connection", nil, &out)
	return out.Data, err
}

// StartConnect kicks off the OAuth connect flow and returns the authorization URL.
func (s *McpServerService) StartConnect(ctx context.Context, key string) (agentapi.StartMcpOAuthResponse, error) {
	var out agentapi.StartMcpOAuthResponse
	err := s.c.Do(ctx, http.MethodPost, mcpServersBasePath+"/"+key+"/oauth/connect", nil, &out)
	return out, err
}

// Disconnect drops a server's stored OAuth tokens.
func (s *McpServerService) Disconnect(ctx context.Context, key string) error {
	return s.c.Do(ctx, http.MethodDelete, mcpServersBasePath+"/"+key+"/oauth/connect", nil, nil)
}
