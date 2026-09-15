package scheduling_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	schedulingmodel "go.proteos.ai/model/scheduling"
	schedulingapi "go.proteos.ai/model/scheduling/api"
	"go.proteos.ai/sdk/scheduling"
)

func TestCalendarEventService_List_EncodesWindow(t *testing.T) {
	var seen url.Values
	s := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/scheduling/v1/calendar-events", r.URL.Path)
		seen = r.URL.Query()
		_ = json.NewEncoder(w).Encode(schedulingapi.GetManyCalendarEventsResponse{Data: []schedulingmodel.CalendarEvent{{Id: "ev1", Title: "Standup"}}})
	})
	from := time.Date(2026, 9, 15, 8, 0, 0, 0, time.UTC)
	to := from.Add(48 * time.Hour)
	got, err := s.CalendarEvents.List(context.Background(), scheduling.ListCalendarEventsOptions{
		UserIds:     []string{"u1", "u2"},
		CalendarIds: []string{"cal1"},
		From:        from,
		To:          to,
		Timezone:    "Europe/Berlin",
		TypeKey:     "intro-call",
		ContactId:   "c1",
	})
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "Standup", got[0].Title)
	require.Equal(t, []string{"u1", "u2"}, seen["user_ids"])
	require.Equal(t, []string{"cal1"}, seen["calendar_ids"])
	require.Equal(t, "2026-09-15T08:00:00Z", seen.Get("from"))
	require.Equal(t, "2026-09-17T08:00:00Z", seen.Get("to"))
	require.Equal(t, "Europe/Berlin", seen.Get("timezone"))
	require.Equal(t, "intro-call", seen.Get("type_key"))
	require.Equal(t, "c1", seen.Get("contact_id"))
	require.False(t, seen.Has("record_id"), "empty optional filters stay off the wire")
	require.False(t, seen.Has("is_deleted_included"))
}

func TestCalendarEventService_WritePaths(t *testing.T) {
	var method, path, query, body string
	s := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path, query = r.Method, r.URL.Path, r.URL.RawQuery
		raw, _ := io.ReadAll(r.Body)
		body = string(raw)
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		_ = json.NewEncoder(w).Encode(schedulingmodel.CalendarEvent{Id: "ev1", Title: "Intro"})
	})
	ctx := context.Background()
	start := time.Date(2026, 9, 16, 9, 0, 0, 0, time.UTC)

	got, err := s.CalendarEvents.Get(ctx, "ev1")
	require.NoError(t, err)
	require.Equal(t, http.MethodGet, method)
	require.Equal(t, "/scheduling/v1/calendar-events/ev1", path)
	require.Equal(t, "Intro", got.Title)

	_, err = s.CalendarEvents.Create(ctx, schedulingapi.CreateCalendarEventRequest{
		CalendarId:  "cal1",
		Title:       "Intro",
		StartAt:     start,
		EndAt:       start.Add(30 * time.Minute),
		Attendees:   []schedulingapi.CalendarAttendeeRequest{{Email: "a@example.com"}},
		IsExclusive: true,
		Record:      &schedulingmodel.RecordRef{EntitySlug: "appointment", RecordId: "r1"},
	})
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, method)
	require.Equal(t, "/scheduling/v1/calendar-events", path)
	require.Contains(t, body, `"calendar_id":"cal1"`)
	require.Contains(t, body, `"start_at":"2026-09-16T09:00:00Z"`)
	require.Contains(t, body, `"attendees":[{"email":"a@example.com","name":"","is_optional":false}]`)
	require.Contains(t, body, `"is_exclusive":true`)
	require.Contains(t, body, `"record":{"entity_slug":"appointment","record_id":"r1"}`)

	title := "Intro (moved)"
	_, err = s.CalendarEvents.Update(ctx, "ev1", schedulingmodel.RecurrenceScopeSeries, schedulingapi.UpdateCalendarEventRequest{Title: &title})
	require.NoError(t, err)
	require.Equal(t, http.MethodPatch, method)
	require.Equal(t, "/scheduling/v1/calendar-events/ev1", path)
	require.Equal(t, "scope=series", query)
	require.JSONEq(t, `{"title":"Intro (moved)"}`, body)

	_, err = s.CalendarEvents.Update(ctx, "ev1", "", schedulingapi.UpdateCalendarEventRequest{Title: &title})
	require.NoError(t, err)
	require.Empty(t, query, "empty scope is not sent")

	require.NoError(t, s.CalendarEvents.Delete(ctx, "ev1", schedulingmodel.RecurrenceScopeInstance))
	require.Equal(t, http.MethodDelete, method)
	require.Equal(t, "/scheduling/v1/calendar-events/ev1", path)
	require.Equal(t, "scope=instance", query)

	_, err = s.CalendarEvents.Respond(ctx, "ev1", schedulingapi.RespondCalendarEventRequest{Status: schedulingmodel.ParticipationStatusAccepted, Comment: "see you"})
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, method)
	require.Equal(t, "/scheduling/v1/calendar-events/ev1/respond", path)
	require.JSONEq(t, `{"status":"accepted","comment":"see you"}`, body)
}
