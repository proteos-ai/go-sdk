package conversation

import (
	"context"
	"net/http"

	conversationmodel "go.proteos.ai/model/conversation"
	conversationapi "go.proteos.ai/model/conversation/api"
	sdk "go.proteos.ai/sdk"
)

const conversationTagDefinitionsBasePath = "/conversations/v1/conversation-tag-definitions"

// ConversationTagDefinitionService manages the org's conversation tag
// definitions — WHAT the tag evaluator looks for in a conversation (one typed
// question, where its answers anchor and which model path answers it). WHERE
// and WHEN a definition is evaluated is a ConversationTagSet's business.
type ConversationTagDefinitionService struct{ c *sdk.Client }

func (s *ConversationTagDefinitionService) List(opts *ListConversationTagDefinitionsOptions) *sdk.PageIterator[conversationmodel.ConversationTagDefinition, ListConversationTagDefinitionsOptions] {
	o := ListConversationTagDefinitionsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListConversationTagDefinitionsOptions) (sdk.ListResult[conversationmodel.ConversationTagDefinition], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *ConversationTagDefinitionService) ListPage(ctx context.Context, opts *ListConversationTagDefinitionsOptions) (sdk.ListResult[conversationmodel.ConversationTagDefinition], error) {
	var out sdk.ListResult[conversationmodel.ConversationTagDefinition]
	err := s.c.DoWithQuery(ctx, http.MethodGet, conversationTagDefinitionsBasePath, opts, nil, &out)
	return out, err
}

func (s *ConversationTagDefinitionService) Get(ctx context.Context, key string) (conversationmodel.ConversationTagDefinition, error) {
	var out conversationmodel.ConversationTagDefinition
	err := s.c.Do(ctx, http.MethodGet, conversationTagDefinitionsBasePath+"/"+key, nil, &out)
	return out, err
}

func (s *ConversationTagDefinitionService) Create(ctx context.Context, req conversationapi.CreateConversationTagDefinitionRequest) (conversationmodel.ConversationTagDefinition, error) {
	var out conversationmodel.ConversationTagDefinition
	err := s.c.Do(ctx, http.MethodPost, conversationTagDefinitionsBasePath, req, &out)
	return out, err
}

func (s *ConversationTagDefinitionService) Update(ctx context.Context, key string, req conversationapi.UpdateConversationTagDefinitionRequest) (conversationmodel.ConversationTagDefinition, error) {
	var out conversationmodel.ConversationTagDefinition
	err := s.c.Do(ctx, http.MethodPatch, conversationTagDefinitionsBasePath+"/"+key, req, &out)
	return out, err
}

// UpsertByKey calls `PUT /conversations/v1/conversation-tag-definitions/:key`
// — the idempotent create-or-update path used by `pro module deploy`. An
// equivalent body is a server-side no-op; an empty module_slug keeps the
// stored attribution.
func (s *ConversationTagDefinitionService) UpsertByKey(ctx context.Context, key string, req conversationapi.CreateConversationTagDefinitionRequest) (conversationmodel.ConversationTagDefinition, error) {
	var out conversationmodel.ConversationTagDefinition
	req.Key = key
	err := s.c.Do(ctx, http.MethodPut, conversationTagDefinitionsBasePath+"/"+key, req, &out)
	return out, err
}

func (s *ConversationTagDefinitionService) Delete(ctx context.Context, key string) error {
	return s.c.Do(ctx, http.MethodDelete, conversationTagDefinitionsBasePath+"/"+key, nil, nil)
}
