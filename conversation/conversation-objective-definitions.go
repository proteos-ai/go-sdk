package conversation

import (
	"context"
	"net/http"

	conversationmodel "go.proteos.ai/model/conversation"
	conversationapi "go.proteos.ai/model/conversation/api"
	sdk "go.proteos.ai/sdk"
)

const conversationObjectiveDefinitionsBasePath = "/conversations/v1/conversation-objective-definitions"

// ConversationObjectiveDefinitionService manages the org's conversation
// objective definitions — WHAT a conversation (of one type, in one phase)
// should achieve and how the platform tells it was achieved (a tag or the
// outcome). Instances on a conversation are read through
// ConversationObjectiveService.
type ConversationObjectiveDefinitionService struct{ c *sdk.Client }

func (s *ConversationObjectiveDefinitionService) List(opts *ListConversationObjectiveDefinitionsOptions) *sdk.PageIterator[conversationmodel.ConversationObjectiveDefinition, ListConversationObjectiveDefinitionsOptions] {
	o := ListConversationObjectiveDefinitionsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListConversationObjectiveDefinitionsOptions) (sdk.ListResult[conversationmodel.ConversationObjectiveDefinition], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *ConversationObjectiveDefinitionService) ListPage(ctx context.Context, opts *ListConversationObjectiveDefinitionsOptions) (sdk.ListResult[conversationmodel.ConversationObjectiveDefinition], error) {
	var out sdk.ListResult[conversationmodel.ConversationObjectiveDefinition]
	err := s.c.DoWithQuery(ctx, http.MethodGet, conversationObjectiveDefinitionsBasePath, opts, nil, &out)
	return out, err
}

func (s *ConversationObjectiveDefinitionService) Get(ctx context.Context, key string) (conversationmodel.ConversationObjectiveDefinition, error) {
	var out conversationmodel.ConversationObjectiveDefinition
	err := s.c.Do(ctx, http.MethodGet, conversationObjectiveDefinitionsBasePath+"/"+key, nil, &out)
	return out, err
}

func (s *ConversationObjectiveDefinitionService) Create(ctx context.Context, req conversationapi.CreateConversationObjectiveDefinitionRequest) (conversationmodel.ConversationObjectiveDefinition, error) {
	var out conversationmodel.ConversationObjectiveDefinition
	err := s.c.Do(ctx, http.MethodPost, conversationObjectiveDefinitionsBasePath, req, &out)
	return out, err
}

func (s *ConversationObjectiveDefinitionService) Update(ctx context.Context, key string, req conversationapi.UpdateConversationObjectiveDefinitionRequest) (conversationmodel.ConversationObjectiveDefinition, error) {
	var out conversationmodel.ConversationObjectiveDefinition
	err := s.c.Do(ctx, http.MethodPatch, conversationObjectiveDefinitionsBasePath+"/"+key, req, &out)
	return out, err
}

// UpsertByKey calls `PUT /conversations/v1/conversation-objective-definitions/:key`
// — the idempotent create-or-update path used by `pro module deploy`. An
// equivalent body is a server-side no-op; an empty module_slug keeps the
// stored attribution.
func (s *ConversationObjectiveDefinitionService) UpsertByKey(ctx context.Context, key string, req conversationapi.CreateConversationObjectiveDefinitionRequest) (conversationmodel.ConversationObjectiveDefinition, error) {
	var out conversationmodel.ConversationObjectiveDefinition
	req.Key = key
	err := s.c.Do(ctx, http.MethodPut, conversationObjectiveDefinitionsBasePath+"/"+key, req, &out)
	return out, err
}

func (s *ConversationObjectiveDefinitionService) Delete(ctx context.Context, key string) error {
	return s.c.Do(ctx, http.MethodDelete, conversationObjectiveDefinitionsBasePath+"/"+key, nil, nil)
}
