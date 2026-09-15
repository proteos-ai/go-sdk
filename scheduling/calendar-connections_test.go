package scheduling_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	schedulingmodel "go.proteos.ai/model/scheduling"
	schedulingapi "go.proteos.ai/model/scheduling/api"
	"go.proteos.ai/sdk/scheduling"
)

func TestCalendarConnectionService_Paths(t *testing.T) {
	var method, path, query, body string
	s := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path, query = r.Method, r.URL.Path, r.URL.RawQuery
		raw, _ := io.ReadAll(r.Body)
		body = string(raw)
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(emptyPage))
			return
		}
		_ = json.NewEncoder(w).Encode(schedulingmodel.CalendarConnection{Id: "cc1", Provider: schedulingmodel.CalendarProviderGoogle})
	})
	ctx := context.Background()
	active := schedulingmodel.CalendarConnectionStatusActive

	_, err := s.CalendarConnections.ListMinePage(ctx, &scheduling.ListCalendarConnectionsOptions{Provider: "google", Status: "active"})
	require.NoError(t, err)
	require.Equal(t, http.MethodGet, method)
	require.Equal(t, "/scheduling/v1/me/calendar-connections", path)
	require.Contains(t, query, "provider=google")
	require.Contains(t, query, "status=active")

	got, err := s.CalendarConnections.Connect(ctx, schedulingapi.CreateCalendarConnectionRequest{ConnectionId: "conn-1"})
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, method)
	require.Equal(t, "/scheduling/v1/me/calendar-connections", path)
	require.JSONEq(t, `{"connection_id":"conn-1"}`, body)
	require.Equal(t, "cc1", got.Id)

	_, err = s.CalendarConnections.UpdateMine(ctx, "cc1", schedulingapi.UpdateCalendarConnectionRequest{Status: &active})
	require.NoError(t, err)
	require.Equal(t, http.MethodPatch, method)
	require.Equal(t, "/scheduling/v1/me/calendar-connections/cc1", path)
	require.JSONEq(t, `{"status":"active"}`, body)

	_, err = s.CalendarConnections.SyncMine(ctx, "cc1", schedulingapi.UpdateCalendarSyncRequest{})
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, method)
	require.Equal(t, "/scheduling/v1/me/calendar-connections/cc1/sync", path)

	require.NoError(t, s.CalendarConnections.DisconnectMine(ctx, "cc1"))
	require.Equal(t, http.MethodDelete, method)
	require.Equal(t, "/scheduling/v1/me/calendar-connections/cc1", path)

	_, err = s.CalendarConnections.ListForUserPage(ctx, "u1", nil)
	require.NoError(t, err)
	require.Equal(t, "/scheduling/v1/users/u1/calendar-connections", path)

	_, err = s.CalendarConnections.UpdateForUser(ctx, "u1", "cc1", schedulingapi.UpdateCalendarConnectionRequest{Status: &active})
	require.NoError(t, err)
	require.Equal(t, http.MethodPatch, method)
	require.Equal(t, "/scheduling/v1/users/u1/calendar-connections/cc1", path)

	require.NoError(t, s.CalendarConnections.DisconnectForUser(ctx, "u1", "cc1"))
	require.Equal(t, http.MethodDelete, method)
	require.Equal(t, "/scheduling/v1/users/u1/calendar-connections/cc1", path)
}

func TestCalendarConnectionService_ListMine_Iterates(t *testing.T) {
	s := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		id := fmt.Sprintf("cc-%d", page)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"meta": map[string]any{"page": page, "page_size": 1, "items_total": 2, "pages_total": 2},
			"data": []schedulingmodel.CalendarConnection{{Id: id}},
		})
	})
	all, err := s.CalendarConnections.ListMine(&scheduling.ListCalendarConnectionsOptions{ListOptions: scheduling.ListOptions{PageSize: 1}}).All(context.Background())
	require.NoError(t, err)
	require.Equal(t, []string{"cc-0", "cc-1"}, []string{all[0].Id, all[1].Id})
}
