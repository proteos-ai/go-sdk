package functions_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.proteos.ai/sdk/functions"
)

func TestActionService_Logs_UnwrapsDataEnvelope(t *testing.T) {
	var seen string
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		seen = r.URL.String()
		_, _ = w.Write([]byte(`{"data":[{"timestamp":"2026-01-02T03:04:05Z","level":"info","message":"hi","resource_type":"action","resource_slug":"send-invoice"}]}`))
	})
	entries, err := f.Actions.Logs(context.Background(), "send-invoice", &functions.ListActionLogsOptions{
		Since: "10m",
		Level: "warn",
	})
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, "action", entries[0].ResourceType)
	require.Contains(t, seen, "/functions/v1/actions/send-invoice/logs")
	require.Contains(t, seen, "since=10m")
	require.Contains(t, seen, "level=warn")
}

func TestActionService_TailLogs_StreamsNDJSON(t *testing.T) {
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Contains(t, r.URL.RawQuery, "follow=true")
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"timestamp":"2026-01-02T03:04:05Z","level":"info","message":"a","resource_type":"action","resource_slug":"send-invoice"}` + "\n"))
	})
	stream, err := f.Actions.TailLogs(context.Background(), "send-invoice", nil)
	require.NoError(t, err)
	defer stream.Close()
	e, ok, err := stream.Next()
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "a", e.Message)
}

func TestLogService_Query_PassesHookAndActionRepeats(t *testing.T) {
	var seenQuery string
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		seenQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	_, err := f.Logs.Query(context.Background(), &functions.ListLogsOptions{
		Hooks:   []string{"a", "b"},
		Actions: []string{"c"},
		Since:   "1h",
		Level:   "info",
	})
	require.NoError(t, err)
	// Repeatable params get separate keys so the server sees ["a","b"]
	// for hook, not "a,b".
	require.Equal(t, 2, strings.Count(seenQuery, "hook="))
	require.Contains(t, seenQuery, "hook=a")
	require.Contains(t, seenQuery, "hook=b")
	require.Contains(t, seenQuery, "action=c")
	require.Contains(t, seenQuery, "since=1h")
	require.Contains(t, seenQuery, "level=info")
}

func TestLogService_Tail_AlwaysSetsFollow(t *testing.T) {
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Contains(t, r.URL.RawQuery, "follow=true")
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"timestamp":"2026-01-02T03:04:05Z","level":"info","message":"x","resource_type":"hook","resource_slug":"foo"}` + "\n"))
	})
	stream, err := f.Logs.Tail(context.Background(), &functions.ListLogsOptions{Hooks: []string{"foo"}})
	require.NoError(t, err)
	defer stream.Close()
	e, ok, err := stream.Next()
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "x", e.Message)
}
