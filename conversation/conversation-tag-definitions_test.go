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

func TestConversationTagDefinitionService_Paths(t *testing.T) {
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
			_ = json.NewEncoder(w).Encode(conversationmodel.ConversationTagDefinition{Key: "churn-risk", Name: "Churn risk"})
		}
	})
	ctx := context.Background()

	_, err := s.ConversationTagDefinitions.ListPage(ctx, &conversation.ListConversationTagDefinitionsOptions{ModuleSlug: "crm", Search: "churn"})
	require.NoError(t, err)
	require.Equal(t, http.MethodGet, method)
	require.Equal(t, "/conversations/v1/conversation-tag-definitions", path)
	require.Equal(t, "crm", seen.Get("module_slug"))
	require.Equal(t, "churn", seen.Get("search"))

	got, err := s.ConversationTagDefinitions.Get(ctx, "churn risk")
	require.NoError(t, err)
	require.Equal(t, http.MethodGet, method)
	require.Equal(t, "/conversations/v1/conversation-tag-definitions/churn%20risk", path)
	require.Equal(t, "Churn risk", got.Name)

	_, err = s.ConversationTagDefinitions.Create(ctx, conversationapi.CreateConversationTagDefinitionRequest{
		Key:   "churn-risk",
		Name:  "Churn risk",
		Group: "signal",
		Question: conversationmodel.ConversationTagQuestion{
			Type:         conversationmodel.ConversationTagQuestionTypeBoolean,
			Instructions: "Does the customer signal intent to cancel?",
		},
		Scope:         conversationmodel.ConversationTagScopeConversation,
		Evaluator:     conversationmodel.ConversationTagEvaluator{Kind: conversationmodel.ConversationTagEvaluatorDecide},
		MinConfidence: 0.7,
	})
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, method)
	require.Equal(t, "/conversations/v1/conversation-tag-definitions", path)
	require.JSONEq(t, `{"key":"churn-risk","name":"Churn risk","group":"signal","question":{"type":"boolean","instructions":"Does the customer signal intent to cancel?","criteria":{}},"scope":"conversation","evaluator":{"kind":"decide"},"directions":null,"min_confidence":0.7,"color":"","module_slug":""}`, body)

	name, isEnabled := "Churn risk (v2)", false
	_, err = s.ConversationTagDefinitions.Update(ctx, "churn-risk", conversationapi.UpdateConversationTagDefinitionRequest{Name: &name, IsEnabled: &isEnabled})
	require.NoError(t, err)
	require.Equal(t, http.MethodPatch, method)
	require.Equal(t, "/conversations/v1/conversation-tag-definitions/churn-risk", path)
	require.JSONEq(t, `{"name":"Churn risk (v2)","is_enabled":false}`, body)

	_, err = s.ConversationTagDefinitions.UpsertByKey(ctx, "churn-risk", conversationapi.CreateConversationTagDefinitionRequest{
		Name:       "Churn risk",
		Scope:      conversationmodel.ConversationTagScopeConversation,
		ModuleSlug: "crm",
	})
	require.NoError(t, err)
	require.Equal(t, http.MethodPut, method)
	require.Equal(t, "/conversations/v1/conversation-tag-definitions/churn-risk", path)
	var sent map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &sent))
	require.Equal(t, "churn-risk", sent["key"], "UpsertByKey stamps the path key into the body")
	require.Equal(t, "crm", sent["module_slug"])

	require.NoError(t, s.ConversationTagDefinitions.Delete(ctx, "churn-risk"))
	require.Equal(t, http.MethodDelete, method)
	require.Equal(t, "/conversations/v1/conversation-tag-definitions/churn-risk", path)
}

func TestConversationTagDefinitionService_List_Iterates(t *testing.T) {
	s := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		require.Equal(t, "1", r.URL.Query().Get("page_size"))
		require.Equal(t, "crm", r.URL.Query().Get("module_slug"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"meta": map[string]any{"page": page, "page_size": 1, "items_total": 2, "pages_total": 2},
			"data": []conversationmodel.ConversationTagDefinition{{Key: fmt.Sprintf("definition-%d", page)}},
		})
	})
	all, err := s.ConversationTagDefinitions.List(&conversation.ListConversationTagDefinitionsOptions{
		ListOptions: conversation.ListOptions{PageSize: 1},
		ModuleSlug:  "crm",
	}).All(context.Background())
	require.NoError(t, err)
	require.Equal(t, []string{"definition-0", "definition-1"}, []string{all[0].Key, all[1].Key})
}
