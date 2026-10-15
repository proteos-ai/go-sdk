package conversation_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	conversationmodel "go.proteos.ai/model/conversation"
	conversationapi "go.proteos.ai/model/conversation/api"
	"go.proteos.ai/sdk/conversation"
)

func TestConversationObjectiveDefinitionService_Paths(t *testing.T) {
	var method, path, body string
	var seen url.Values
	s := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path, seen = r.Method, r.URL.EscapedPath(), r.URL.Query()
		raw, _ := io.ReadAll(r.Body)
		body = string(raw)
		switch {
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Query().Has("page"):
			_, _ = w.Write([]byte(emptyPage))
		default:
			_ = json.NewEncoder(w).Encode(conversationmodel.ConversationObjectiveDefinition{Key: "confirm-budget", Name: "Confirm budget"})
		}
	})
	ctx := context.Background()

	isEnabled := true
	_, err := s.ConversationObjectiveDefinitions.ListPage(ctx, &conversation.ListConversationObjectiveDefinitionsOptions{ModuleSlug: "crm", Search: "budget", TypeKey: "sales-discovery", PhaseKey: "qualify", IsEnabled: &isEnabled})
	require.NoError(t, err)
	require.Equal(t, http.MethodGet, method)
	require.Equal(t, "/conversations/v1/conversation-objective-definitions", path)
	require.Equal(t, "crm", seen.Get("module_slug"))
	require.Equal(t, "budget", seen.Get("search"))
	require.Equal(t, "sales-discovery", seen.Get("type_key"))
	require.Equal(t, "qualify", seen.Get("phase_key"))
	require.Equal(t, "true", seen.Get("is_enabled"))

	got, err := s.ConversationObjectiveDefinitions.Get(ctx, "confirm budget")
	require.NoError(t, err)
	require.Equal(t, http.MethodGet, method)
	require.Equal(t, "/conversations/v1/conversation-objective-definitions/confirm%20budget", path)
	require.Equal(t, "Confirm budget", got.Name)

	achieved := true
	_, err = s.ConversationObjectiveDefinitions.Create(ctx, conversationapi.CreateConversationObjectiveDefinitionRequest{
		Key:      "confirm-budget",
		Name:     "Confirm budget",
		TypeKey:  "sales-discovery",
		PhaseKey: "qualify",
		Measure: &conversationmodel.ConversationObjectiveMeasure{
			Kind:             conversationmodel.ConversationObjectiveMeasureTag,
			TagDefinitionKey: "budget-confirmed",
			AchievedWhen:     conversationmodel.ConversationObjectiveAchievedWhen{Boolean: &achieved},
		},
	})
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, method)
	require.Equal(t, "/conversations/v1/conversation-objective-definitions", path)
	require.JSONEq(t, `{"key":"confirm-budget","name":"Confirm budget","description":"","type_key":"sales-discovery","phase_key":"qualify","measure":{"kind":"tag","tag_definition_key":"budget-confirmed","achieved_when":{"boolean":true}},"module_slug":""}`, body)

	disabled, name := false, "Confirm the budget"
	_, err = s.ConversationObjectiveDefinitions.Update(ctx, "confirm-budget", conversationapi.UpdateConversationObjectiveDefinitionRequest{Name: &name, IsEnabled: &disabled})
	require.NoError(t, err)
	require.Equal(t, http.MethodPatch, method)
	require.Equal(t, "/conversations/v1/conversation-objective-definitions/confirm-budget", path)
	var patched map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &patched))
	require.Equal(t, "Confirm the budget", patched["name"])
	require.Equal(t, false, patched["is_enabled"])
	require.NotContains(t, patched, "type_key", "nil fields are omitted from a PATCH")

	_, err = s.ConversationObjectiveDefinitions.UpsertByKey(ctx, "confirm-budget", conversationapi.CreateConversationObjectiveDefinitionRequest{
		Name:       "Confirm budget",
		ModuleSlug: "crm",
	})
	require.NoError(t, err)
	require.Equal(t, http.MethodPut, method)
	require.Equal(t, "/conversations/v1/conversation-objective-definitions/confirm-budget", path)
	var sent map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &sent))
	require.Equal(t, "confirm-budget", sent["key"], "UpsertByKey stamps the path key into the body")
	require.Equal(t, "crm", sent["module_slug"])

	require.NoError(t, s.ConversationObjectiveDefinitions.Delete(ctx, "confirm-budget"))
	require.Equal(t, http.MethodDelete, method)
	require.Equal(t, "/conversations/v1/conversation-objective-definitions/confirm-budget", path)
}

func TestConversationObjectiveDefinitionService_List_Iterates(t *testing.T) {
	s := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		require.Equal(t, "1", r.URL.Query().Get("page_size"))
		require.Equal(t, "crm", r.URL.Query().Get("module_slug"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"meta": map[string]any{"page": page, "page_size": 1, "items_total": 2, "pages_total": 2},
			"data": []conversationmodel.ConversationObjectiveDefinition{{Key: fmt.Sprintf("objective-%d", page)}},
		})
	})
	all, err := s.ConversationObjectiveDefinitions.List(&conversation.ListConversationObjectiveDefinitionsOptions{
		ListOptions: conversation.ListOptions{PageSize: 1},
		ModuleSlug:  "crm",
	}).All(context.Background())
	require.NoError(t, err)
	require.Equal(t, []string{"objective-0", "objective-1"}, []string{all[0].Key, all[1].Key})
}
