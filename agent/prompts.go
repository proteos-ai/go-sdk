package agent

import (
	"context"
	"net/http"
	"strconv"

	agentmodel "go.proteos.ai/model/agent"
	agentapi "go.proteos.ai/model/agent/api"
	sdk "go.proteos.ai/sdk"
)

const promptsBasePath = "/agents/v1/prompts"

// PromptService manages versioned, Liquid-templated prompts.
type PromptService struct{ c *sdk.Client }

func (s *PromptService) List(opts *ListPromptsOptions) *sdk.PageIterator[agentmodel.Prompt, ListPromptsOptions] {
	o := ListPromptsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListPromptsOptions) (sdk.ListResult[agentmodel.Prompt], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *PromptService) ListPage(ctx context.Context, opts *ListPromptsOptions) (sdk.ListResult[agentmodel.Prompt], error) {
	var out sdk.ListResult[agentmodel.Prompt]
	err := s.c.DoWithQuery(ctx, http.MethodGet, promptsBasePath, opts, nil, &out)
	return out, err
}

func (s *PromptService) Get(ctx context.Context, key string) (agentmodel.Prompt, error) {
	var out agentmodel.Prompt
	err := s.c.Do(ctx, http.MethodGet, promptsBasePath+"/"+key, nil, &out)
	return out, err
}

func (s *PromptService) Create(ctx context.Context, req agentapi.CreatePromptRequest) (agentmodel.Prompt, error) {
	var out agentmodel.Prompt
	err := s.c.Do(ctx, http.MethodPost, promptsBasePath, req, &out)
	return out, err
}

func (s *PromptService) Update(ctx context.Context, key string, req agentapi.UpdatePromptRequest) (agentmodel.Prompt, error) {
	var out agentmodel.Prompt
	err := s.c.Do(ctx, http.MethodPatch, promptsBasePath+"/"+key, req, &out)
	return out, err
}

// UpsertByKey calls `PUT /agents/v1/prompts/:key` — idempotent create-or-update used
// by `pro module deploy`. A body identical to the current version forks no new
// version (server-side hash dedup).
func (s *PromptService) UpsertByKey(ctx context.Context, key string, req agentapi.CreatePromptRequest) (agentmodel.Prompt, error) {
	var out agentmodel.Prompt
	req.Key = key
	err := s.c.Do(ctx, http.MethodPut, promptsBasePath+"/"+key, req, &out)
	return out, err
}

func (s *PromptService) Delete(ctx context.Context, key string) error {
	return s.c.Do(ctx, http.MethodDelete, promptsBasePath+"/"+key, nil, nil)
}

// Versions returns every immutable version of a prompt (newest first).
func (s *PromptService) Versions(ctx context.Context, key string) ([]agentmodel.PromptVersion, error) {
	var out struct {
		Data []agentmodel.PromptVersion `json:"data"`
	}
	err := s.c.Do(ctx, http.MethodGet, promptsBasePath+"/"+key+"/versions", nil, &out)
	return out.Data, err
}

// Version returns a single prompt version by number.
func (s *PromptService) Version(ctx context.Context, key string, number uint32) (agentmodel.PromptVersion, error) {
	var out agentmodel.PromptVersion
	err := s.c.Do(ctx, http.MethodGet, promptsBasePath+"/"+key+"/versions/"+strconv.FormatUint(uint64(number), 10), nil, &out)
	return out, err
}
