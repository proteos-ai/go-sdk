package agent

import (
	"context"
	"net/http"

	agentmodel "go.proteos.ai/model/agent"
	agentapi "go.proteos.ai/model/agent/api"
	sdk "go.proteos.ai/sdk"
)

const agentsBasePath = "/agents/v1/agents"

// AgentService manages configured agents (personas).
type AgentService struct{ c *sdk.Client }

func (s *AgentService) List(opts *ListAgentsOptions) *sdk.PageIterator[agentmodel.Agent, ListAgentsOptions] {
	o := ListAgentsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListAgentsOptions) (sdk.ListResult[agentmodel.Agent], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *AgentService) ListPage(ctx context.Context, opts *ListAgentsOptions) (sdk.ListResult[agentmodel.Agent], error) {
	var out sdk.ListResult[agentmodel.Agent]
	err := s.c.DoWithQuery(ctx, http.MethodGet, agentsBasePath, opts, nil, &out)
	return out, err
}

func (s *AgentService) Get(ctx context.Context, key string) (agentmodel.Agent, error) {
	var out agentmodel.Agent
	err := s.c.Do(ctx, http.MethodGet, agentsBasePath+"/"+key, nil, &out)
	return out, err
}

func (s *AgentService) Create(ctx context.Context, req agentapi.CreateAgentRequest) (agentmodel.Agent, error) {
	var out agentmodel.Agent
	err := s.c.Do(ctx, http.MethodPost, agentsBasePath, req, &out)
	return out, err
}

func (s *AgentService) Update(ctx context.Context, key string, req agentapi.UpdateAgentRequest) (agentmodel.Agent, error) {
	var out agentmodel.Agent
	err := s.c.Do(ctx, http.MethodPatch, agentsBasePath+"/"+key, req, &out)
	return out, err
}

// UpsertByKey calls `PUT /agents/v1/agents/:key` — the idempotent create-or-replace
// path used by `pro module deploy`.
func (s *AgentService) UpsertByKey(ctx context.Context, key string, req agentapi.CreateAgentRequest) (agentmodel.Agent, error) {
	var out agentmodel.Agent
	req.Key = key
	err := s.c.Do(ctx, http.MethodPut, agentsBasePath+"/"+key, req, &out)
	return out, err
}

func (s *AgentService) Delete(ctx context.Context, key string) error {
	return s.c.Do(ctx, http.MethodDelete, agentsBasePath+"/"+key, nil, nil)
}
