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

func TestConversationTagSetService_Paths(t *testing.T) {
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
			_ = json.NewEncoder(w).Encode(conversationmodel.ConversationTagSet{Key: "support-signals", Name: "Support signals"})
		}
	})
	ctx := context.Background()

	_, err := s.ConversationTagSets.ListPage(ctx, &conversation.ListConversationTagSetsOptions{ModuleSlug: "crm", Search: "support"})
	require.NoError(t, err)
	require.Equal(t, http.MethodGet, method)
	require.Equal(t, "/conversations/v1/conversation-tag-sets", path)
	require.Equal(t, "crm", seen.Get("module_slug"))
	require.Equal(t, "support", seen.Get("search"))

	got, err := s.ConversationTagSets.Get(ctx, "support signals")
	require.NoError(t, err)
	require.Equal(t, http.MethodGet, method)
	require.Equal(t, "/conversations/v1/conversation-tag-sets/support%20signals", path)
	require.Equal(t, "Support signals", got.Name)

	quietWindow := 30
	_, err = s.ConversationTagSets.Create(ctx, conversationapi.CreateConversationTagSetRequest{
		Key:                "support-signals",
		Name:               "Support signals",
		DefinitionKeys:     []string{"churn-risk"},
		Channels:           []conversationmodel.Channel{"email"},
		Directions:         []conversationmodel.MessageDirection{"inbound"},
		EvaluateOn:         []conversationmodel.ConversationTagTrigger{conversationmodel.ConversationTagTriggerRealtime, conversationmodel.ConversationTagTriggerCompletion},
		QuietWindowSeconds: &quietWindow,
	})
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, method)
	require.Equal(t, "/conversations/v1/conversation-tag-sets", path)
	require.JSONEq(t, `{"key":"support-signals","name":"Support signals","definition_keys":["churn-risk"],"channels":["email"],"connection_ids":null,"type_keys":null,"directions":["inbound"],"evaluate_on":["realtime","completion"],"quiet_window_seconds":30,"module_slug":""}`, body)

	isEnabled, maxWait := false, 300
	_, err = s.ConversationTagSets.Update(ctx, "support-signals", conversationapi.UpdateConversationTagSetRequest{IsEnabled: &isEnabled, MaxWaitSeconds: &maxWait})
	require.NoError(t, err)
	require.Equal(t, http.MethodPatch, method)
	require.Equal(t, "/conversations/v1/conversation-tag-sets/support-signals", path)
	require.JSONEq(t, `{"max_wait_seconds":300,"is_enabled":false}`, body)

	_, err = s.ConversationTagSets.UpsertByKey(ctx, "support-signals", conversationapi.CreateConversationTagSetRequest{
		Name:           "Support signals",
		DefinitionKeys: []string{"churn-risk"},
		EvaluateOn:     []conversationmodel.ConversationTagTrigger{conversationmodel.ConversationTagTriggerCompletion},
		ModuleSlug:     "crm",
	})
	require.NoError(t, err)
	require.Equal(t, http.MethodPut, method)
	require.Equal(t, "/conversations/v1/conversation-tag-sets/support-signals", path)
	var sent map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &sent))
	require.Equal(t, "support-signals", sent["key"], "UpsertByKey stamps the path key into the body")
	require.Equal(t, "crm", sent["module_slug"])

	require.NoError(t, s.ConversationTagSets.Delete(ctx, "support-signals"))
	require.Equal(t, http.MethodDelete, method)
	require.Equal(t, "/conversations/v1/conversation-tag-sets/support-signals", path)
}

func TestConversationTagSetService_List_Iterates(t *testing.T) {
	s := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		require.Equal(t, "1", r.URL.Query().Get("page_size"))
		require.Equal(t, "crm", r.URL.Query().Get("module_slug"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"meta": map[string]any{"page": page, "page_size": 1, "items_total": 2, "pages_total": 2},
			"data": []conversationmodel.ConversationTagSet{{Key: fmt.Sprintf("set-%d", page)}},
		})
	})
	all, err := s.ConversationTagSets.List(&conversation.ListConversationTagSetsOptions{
		ListOptions: conversation.ListOptions{PageSize: 1},
		ModuleSlug:  "crm",
	}).All(context.Background())
	require.NoError(t, err)
	require.Equal(t, []string{"set-0", "set-1"}, []string{all[0].Key, all[1].Key})
}
