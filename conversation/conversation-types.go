package conversation

import (
	"context"
	"net/http"

	conversationmodel "go.proteos.ai/model/conversation"
	conversationapi "go.proteos.ai/model/conversation/api"
	sdk "go.proteos.ai/sdk"
)

const conversationTypesBasePath = "/conversations/v1/conversation-types"

// ConversationTypeService manages the org's conversation taxonomy — the types
// the pre-summary classifier assigns and their per-type summary prompts.
type ConversationTypeService struct{ c *sdk.Client }

func (s *ConversationTypeService) List(opts *ListConversationTypesOptions) *sdk.PageIterator[conversationmodel.ConversationType, ListConversationTypesOptions] {
	o := ListConversationTypesOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListConversationTypesOptions) (sdk.ListResult[conversationmodel.ConversationType], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *ConversationTypeService) ListPage(ctx context.Context, opts *ListConversationTypesOptions) (sdk.ListResult[conversationmodel.ConversationType], error) {
	var out sdk.ListResult[conversationmodel.ConversationType]
	err := s.c.DoWithQuery(ctx, http.MethodGet, conversationTypesBasePath, opts, nil, &out)
	return out, err
}

func (s *ConversationTypeService) Get(ctx context.Context, key string) (conversationmodel.ConversationType, error) {
	var out conversationmodel.ConversationType
	err := s.c.Do(ctx, http.MethodGet, conversationTypesBasePath+"/"+key, nil, &out)
	return out, err
}

func (s *ConversationTypeService) Create(ctx context.Context, req conversationapi.CreateConversationTypeRequest) (conversationmodel.ConversationType, error) {
	var out conversationmodel.ConversationType
	err := s.c.Do(ctx, http.MethodPost, conversationTypesBasePath, req, &out)
	return out, err
}

func (s *ConversationTypeService) Update(ctx context.Context, key string, req conversationapi.UpdateConversationTypeRequest) (conversationmodel.ConversationType, error) {
	var out conversationmodel.ConversationType
	err := s.c.Do(ctx, http.MethodPatch, conversationTypesBasePath+"/"+key, req, &out)
	return out, err
}

// UpsertByKey calls `PUT /conversations/v1/conversation-types/:key` — the
// idempotent create-or-update path used by `pro module deploy`. An equivalent
// body is a server-side no-op; an empty module_slug keeps the stored
// attribution.
func (s *ConversationTypeService) UpsertByKey(ctx context.Context, key string, req conversationapi.CreateConversationTypeRequest) (conversationmodel.ConversationType, error) {
	var out conversationmodel.ConversationType
	req.Key = key
	err := s.c.Do(ctx, http.MethodPut, conversationTypesBasePath+"/"+key, req, &out)
	return out, err
}

func (s *ConversationTypeService) Delete(ctx context.Context, key string) error {
	return s.c.Do(ctx, http.MethodDelete, conversationTypesBasePath+"/"+key, nil, nil)
}
