package conversation

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"go.proteos.ai/model/common"
	conversationmodel "go.proteos.ai/model/conversation"
	conversationapi "go.proteos.ai/model/conversation/api"
	sdk "go.proteos.ai/sdk"
)

const (
	conversationObjectivesBasePath = "/conversations/v1/conversation-objectives"
	conversationsBasePath          = "/conversations/v1/conversations"
)

// ConversationObjectiveService reads conversation objectives — one
// definition's standing (open / achieved) on one conversation. They are
// resolved by the service from tags and outcomes and never user-written:
// GetProgress is the per-conversation checklist read, List + Count the
// org-wide ledger.
type ConversationObjectiveService struct{ c *sdk.Client }

// GetProgress calls `GET /conversations/v1/conversations/:id/objective-progress`
// — where one conversation stands: current phase, every phase with its
// status, every applicable objective and the definitions they reference.
func (s *ConversationObjectiveService) GetProgress(ctx context.Context, conversationId string) (conversationmodel.ConversationObjectiveProgress, error) {
	var out conversationapi.GetConversationObjectiveProgressResponse
	err := s.c.Do(ctx, http.MethodGet, conversationsBasePath+"/"+url.PathEscape(conversationId)+"/objective-progress", nil, &out)
	return out.Data, err
}

// List calls `GET /conversations/v1/conversation-objectives` — one page of
// the org-wide ledger, filtered by the query (every field optional).
func (s *ConversationObjectiveService) List(ctx context.Context, query conversationapi.GetManyConversationObjectivesQuery) (sdk.ListResult[conversationmodel.ConversationObjective], error) {
	var out sdk.ListResult[conversationmodel.ConversationObjective]
	err := s.c.DoWithQuery(ctx, http.MethodGet, conversationObjectivesBasePath, listObjectivesQueryParams(query), nil, &out)
	return out, err
}

// Count calls `GET /conversations/v1/conversation-objectives/counts` — the
// ledger grouped by query.GroupBy (default definition_key,status) under the
// same filters as List.
func (s *ConversationObjectiveService) Count(ctx context.Context, query conversationapi.CountConversationObjectivesQuery) ([]conversationmodel.ConversationObjectiveCount, error) {
	var out conversationapi.CountConversationObjectivesResponse
	err := s.c.DoWithQuery(ctx, http.MethodGet, conversationObjectivesBasePath+"/counts", countObjectivesQueryParams(query), nil, &out)
	return out.Data, err
}

// conversationObjectivesQueryParams is the query-tagged projection of the
// ledger queries. It is defined SDK-side because the query encoder skips
// time.Time (a struct): Since / Until travel as RFC3339 strings. Pointer
// filters are left as-is — nil ones are skipped by the encoder.
type conversationObjectivesQueryParams struct {
	ConversationId      *string               `query:"conversation_id"`
	DefinitionKey       *string               `query:"definition_key"`
	TypeKey             *string               `query:"type_key"`
	PhaseKey            *string               `query:"phase_key"`
	Channel             *string               `query:"channel"`
	Status              *string               `query:"status"`
	Source              *string               `query:"source"`
	IsProvisional       *bool                 `query:"is_provisional"`
	IsConversationEnded *bool                 `query:"is_conversation_ended"`
	Since               string                `query:"since,omitempty"`
	Until               string                `query:"until,omitempty"`
	GroupBy             string                `query:"group_by,omitempty"`
	Page                int                   `query:"page,omitempty"`
	PageSize            int                   `query:"page_size,omitempty"`
	SortBy              *string               `query:"sort_by"`
	Sort                *common.SortDirection `query:"sort_direction"`
}

func listObjectivesQueryParams(query conversationapi.GetManyConversationObjectivesQuery) conversationObjectivesQueryParams {
	return conversationObjectivesQueryParams{
		ConversationId:      query.ConversationId,
		DefinitionKey:       query.DefinitionKey,
		TypeKey:             query.TypeKey,
		PhaseKey:            query.PhaseKey,
		Channel:             query.Channel,
		Status:              query.Status,
		Source:              query.Source,
		IsProvisional:       query.IsProvisional,
		IsConversationEnded: query.IsConversationEnded,
		Since:               formatQueryTime(query.Since),
		Until:               formatQueryTime(query.Until),
		Page:                query.Page,
		PageSize:            query.PageSize,
		SortBy:              query.SortBy,
		Sort:                query.Sort,
	}
}

func countObjectivesQueryParams(query conversationapi.CountConversationObjectivesQuery) conversationObjectivesQueryParams {
	return conversationObjectivesQueryParams{
		DefinitionKey:       query.DefinitionKey,
		TypeKey:             query.TypeKey,
		PhaseKey:            query.PhaseKey,
		Channel:             query.Channel,
		Status:              query.Status,
		Source:              query.Source,
		IsProvisional:       query.IsProvisional,
		IsConversationEnded: query.IsConversationEnded,
		Since:               formatQueryTime(query.Since),
		Until:               formatQueryTime(query.Until),
		GroupBy:             query.GroupBy,
	}
}

func formatQueryTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
