package conversation_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.proteos.ai/model/common"
	conversationmodel "go.proteos.ai/model/conversation"
	conversationapi "go.proteos.ai/model/conversation/api"
)

func TestConversationObjectiveService_Paths(t *testing.T) {
	var method, path string
	var seen url.Values
	s := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path, seen = r.Method, r.URL.EscapedPath(), r.URL.Query()
		switch {
		case r.URL.Path == "/conversations/v1/conversation-objectives/counts":
			_ = json.NewEncoder(w).Encode(conversationapi.CountConversationObjectivesResponse{Data: []conversationmodel.ConversationObjectiveCount{
				{DefinitionKey: "confirm-budget", Status: conversationmodel.ConversationObjectiveStatusAchieved, Count: 3},
			}})
		case r.URL.Path == "/conversations/v1/conversation-objectives":
			_ = json.NewEncoder(w).Encode(conversationapi.GetManyConversationObjectivesResponse{
				Meta: common.ResponseMeta{Pagination: common.Pagination{Page: 0, PageSize: 10}, ItemsTotal: 1, PagesTotal: 1},
				Data: []conversationmodel.ConversationObjective{{ConversationId: "c1", DefinitionKey: "confirm-budget", Status: conversationmodel.ConversationObjectiveStatusOpen}},
			})
		default:
			_ = json.NewEncoder(w).Encode(conversationapi.GetConversationObjectiveProgressResponse{Data: conversationmodel.ConversationObjectiveProgress{
				PhaseKey:   "qualify",
				Phases:     []conversationmodel.ConversationPhaseProgress{{ConversationPhase: conversationmodel.ConversationPhase{Key: "qualify", Name: "Qualify"}, Status: conversationmodel.ConversationPhaseStatusCurrent}},
				Objectives: []conversationmodel.ConversationObjective{{ConversationId: "c 1", DefinitionKey: "confirm-budget", Status: conversationmodel.ConversationObjectiveStatusOpen}},
			}})
		}
	})
	ctx := context.Background()

	progress, err := s.ConversationObjectives.GetProgress(ctx, "c 1")
	require.NoError(t, err)
	require.Equal(t, http.MethodGet, method)
	require.Equal(t, "/conversations/v1/conversations/c%201/objective-progress", path)
	require.Equal(t, "qualify", progress.PhaseKey, "GetProgress unwraps the data envelope")
	require.Len(t, progress.Phases, 1)
	require.Len(t, progress.Objectives, 1)

	conversationId, status, isProvisional := "c1", "open", false
	since := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	sortBy, sort := "recorded_at", common.SortDirection("desc")
	page, err := s.ConversationObjectives.List(ctx, conversationapi.GetManyConversationObjectivesQuery{
		ConversationId: &conversationId,
		Status:         &status,
		IsProvisional:  &isProvisional,
		Since:          &since,
		Pagination:     common.Pagination{Page: 2, PageSize: 10},
		Sorting:        common.Sorting{SortBy: &sortBy, Sort: &sort},
	})
	require.NoError(t, err)
	require.Equal(t, http.MethodGet, method)
	require.Equal(t, "/conversations/v1/conversation-objectives", path)
	require.Equal(t, "c1", seen.Get("conversation_id"))
	require.Equal(t, "open", seen.Get("status"))
	require.Equal(t, "false", seen.Get("is_provisional"))
	require.Equal(t, "2026-10-01T12:00:00Z", seen.Get("since"), "time filters travel as RFC3339")
	require.False(t, seen.Has("until"), "nil filters are omitted")
	require.False(t, seen.Has("definition_key"), "nil filters are omitted")
	require.Equal(t, "2", seen.Get("page"))
	require.Equal(t, "10", seen.Get("page_size"))
	require.Equal(t, "recorded_at", seen.Get("sort_by"))
	require.Equal(t, "desc", seen.Get("sort_direction"))
	require.Equal(t, 1, page.Meta.ItemsTotal)
	require.Equal(t, "confirm-budget", page.Data[0].DefinitionKey)

	typeKey := "sales-discovery"
	counts, err := s.ConversationObjectives.Count(ctx, conversationapi.CountConversationObjectivesQuery{GroupBy: "definition_key,status", TypeKey: &typeKey})
	require.NoError(t, err)
	require.Equal(t, http.MethodGet, method)
	require.Equal(t, "/conversations/v1/conversation-objectives/counts", path)
	require.Equal(t, "definition_key,status", seen.Get("group_by"))
	require.Equal(t, "sales-discovery", seen.Get("type_key"))
	require.False(t, seen.Has("page"), "counts carry no pagination")
	require.Equal(t, []conversationmodel.ConversationObjectiveCount{{DefinitionKey: "confirm-budget", Status: conversationmodel.ConversationObjectiveStatusAchieved, Count: 3}}, counts)
}
