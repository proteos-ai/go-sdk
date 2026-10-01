package conversation

import (
	"context"
	"net/http"

	conversationmodel "go.proteos.ai/model/conversation"
	conversationapi "go.proteos.ai/model/conversation/api"
	sdk "go.proteos.ai/sdk"
)

const conversationTagSetsBasePath = "/conversations/v1/conversation-tag-sets"

// ConversationTagSetService manages the org's conversation tag sets — WHERE
// (channels / connections / conversation types) and WHEN (realtime debounce,
// completion) a group of tag definitions is evaluated.
type ConversationTagSetService struct{ c *sdk.Client }

func (s *ConversationTagSetService) List(opts *ListConversationTagSetsOptions) *sdk.PageIterator[conversationmodel.ConversationTagSet, ListConversationTagSetsOptions] {
	o := ListConversationTagSetsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListConversationTagSetsOptions) (sdk.ListResult[conversationmodel.ConversationTagSet], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *ConversationTagSetService) ListPage(ctx context.Context, opts *ListConversationTagSetsOptions) (sdk.ListResult[conversationmodel.ConversationTagSet], error) {
	var out sdk.ListResult[conversationmodel.ConversationTagSet]
	err := s.c.DoWithQuery(ctx, http.MethodGet, conversationTagSetsBasePath, opts, nil, &out)
	return out, err
}

func (s *ConversationTagSetService) Get(ctx context.Context, key string) (conversationmodel.ConversationTagSet, error) {
	var out conversationmodel.ConversationTagSet
	err := s.c.Do(ctx, http.MethodGet, conversationTagSetsBasePath+"/"+key, nil, &out)
	return out, err
}

func (s *ConversationTagSetService) Create(ctx context.Context, req conversationapi.CreateConversationTagSetRequest) (conversationmodel.ConversationTagSet, error) {
	var out conversationmodel.ConversationTagSet
	err := s.c.Do(ctx, http.MethodPost, conversationTagSetsBasePath, req, &out)
	return out, err
}

func (s *ConversationTagSetService) Update(ctx context.Context, key string, req conversationapi.UpdateConversationTagSetRequest) (conversationmodel.ConversationTagSet, error) {
	var out conversationmodel.ConversationTagSet
	err := s.c.Do(ctx, http.MethodPatch, conversationTagSetsBasePath+"/"+key, req, &out)
	return out, err
}

// UpsertByKey calls `PUT /conversations/v1/conversation-tag-sets/:key` — the
// idempotent create-or-update path used by `pro module deploy`. An equivalent
// body is a server-side no-op; an empty module_slug keeps the stored
// attribution.
func (s *ConversationTagSetService) UpsertByKey(ctx context.Context, key string, req conversationapi.CreateConversationTagSetRequest) (conversationmodel.ConversationTagSet, error) {
	var out conversationmodel.ConversationTagSet
	req.Key = key
	err := s.c.Do(ctx, http.MethodPut, conversationTagSetsBasePath+"/"+key, req, &out)
	return out, err
}

func (s *ConversationTagSetService) Delete(ctx context.Context, key string) error {
	return s.c.Do(ctx, http.MethodDelete, conversationTagSetsBasePath+"/"+key, nil, nil)
}
