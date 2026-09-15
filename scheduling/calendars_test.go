package scheduling_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	schedulingmodel "go.proteos.ai/model/scheduling"
	schedulingapi "go.proteos.ai/model/scheduling/api"
	"go.proteos.ai/sdk/scheduling"
)

func TestCalendarService_Paths(t *testing.T) {
	var method, path, query, body string
	s := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path, query = r.Method, r.URL.Path, r.URL.RawQuery
		raw, _ := io.ReadAll(r.Body)
		body = string(raw)
		if r.Method == http.MethodGet && r.URL.Query().Has("page") {
			_, _ = w.Write([]byte(emptyPage))
			return
		}
		_ = json.NewEncoder(w).Encode(schedulingmodel.Calendar{Id: "cal1", Name: "Work", IsSynced: true})
	})
	ctx := context.Background()
	isTrue := true

	_, err := s.Calendars.ListMinePage(ctx, &scheduling.ListCalendarsOptions{CalendarConnectionId: "cc1", IsSynced: &isTrue})
	require.NoError(t, err)
	require.Equal(t, http.MethodGet, method)
	require.Equal(t, "/scheduling/v1/me/calendars", path)
	require.Contains(t, query, "calendar_connection_id=cc1")
	require.Contains(t, query, "is_synced=true")
	require.NotContains(t, query, "is_target", "nil tri-state filter must stay off the wire")

	got, err := s.Calendars.GetMine(ctx, "cal1")
	require.NoError(t, err)
	require.Equal(t, "/scheduling/v1/me/calendars/cal1", path)
	require.Equal(t, "Work", got.Name)

	visibility := schedulingmodel.CalendarVisibilityBusyOnly
	_, err = s.Calendars.UpdateMine(ctx, "cal1", schedulingapi.UpdateCalendarRequest{IsSynced: &isTrue, Visibility: &visibility})
	require.NoError(t, err)
	require.Equal(t, http.MethodPatch, method)
	require.Equal(t, "/scheduling/v1/me/calendars/cal1", path)
	require.JSONEq(t, `{"is_synced":true,"visibility":"busy_only"}`, body)

	_, err = s.Calendars.ListForUserPage(ctx, "u1", nil)
	require.NoError(t, err)
	require.Equal(t, "/scheduling/v1/users/u1/calendars", path)

	_, err = s.Calendars.GetForUser(ctx, "u1", "cal1")
	require.NoError(t, err)
	require.Equal(t, "/scheduling/v1/users/u1/calendars/cal1", path)

	_, err = s.Calendars.UpdateForUser(ctx, "u1", "cal1", schedulingapi.UpdateCalendarRequest{IsTarget: &isTrue})
	require.NoError(t, err)
	require.Equal(t, http.MethodPatch, method)
	require.Equal(t, "/scheduling/v1/users/u1/calendars/cal1", path)
	require.JSONEq(t, `{"is_target":true}`, body)
}
