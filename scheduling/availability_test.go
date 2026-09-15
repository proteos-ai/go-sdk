package scheduling_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	schedulingmodel "go.proteos.ai/model/scheduling"
	schedulingapi "go.proteos.ai/model/scheduling/api"
	"go.proteos.ai/sdk/scheduling"
)

func TestCalendarEventTypeService_Paths(t *testing.T) {
	var method, path, query, body string
	s := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path, query = r.Method, r.URL.Path, r.URL.RawQuery
		raw, _ := io.ReadAll(r.Body)
		body = string(raw)
		if r.Method == http.MethodGet && r.URL.Path == "/scheduling/v1/calendar-event-types" {
			_, _ = w.Write([]byte(emptyPage))
			return
		}
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		_ = json.NewEncoder(w).Encode(schedulingmodel.CalendarEventType{Key: "intro-call", Name: "Intro call", DurationMinutes: 30})
	})
	ctx := context.Background()

	isEnabled := true
	_, err := s.CalendarEventTypes.List(&scheduling.ListCalendarEventTypesOptions{IsEnabled: &isEnabled, Search: "intro"}).All(ctx)
	require.NoError(t, err)
	require.Equal(t, "/scheduling/v1/calendar-event-types", path)
	require.Contains(t, query, "is_enabled=true")
	require.Contains(t, query, "search=intro")

	got, err := s.CalendarEventTypes.Get(ctx, "intro-call")
	require.NoError(t, err)
	require.Equal(t, http.MethodGet, method)
	require.Equal(t, "/scheduling/v1/calendar-event-types/intro-call", path)
	require.Equal(t, 30, got.DurationMinutes)

	_, err = s.CalendarEventTypes.Put(ctx, "intro-call", schedulingapi.PutCalendarEventTypeRequest{Name: "Intro call", DurationMinutes: 30})
	require.NoError(t, err)
	require.Equal(t, http.MethodPut, method)
	require.Contains(t, body, `"duration_minutes":30`)

	maxPerDay := 2
	_, err = s.CalendarEventTypes.Update(ctx, "intro-call", schedulingapi.UpdateCalendarEventTypeRequest{MaxPerDay: &maxPerDay})
	require.NoError(t, err)
	require.Equal(t, http.MethodPatch, method)
	require.JSONEq(t, `{"max_per_day":2}`, body)

	require.NoError(t, s.CalendarEventTypes.Delete(ctx, "intro-call"))
	require.Equal(t, http.MethodDelete, method)
}

func TestAvailabilityConstraintService_Paths(t *testing.T) {
	var method, path, body string
	s := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		body = string(raw)
		if r.Method == http.MethodGet && (r.URL.Path == "/scheduling/v1/me/availability-constraints" || r.URL.Path == "/scheduling/v1/users/u1/availability-constraints") {
			_, _ = w.Write([]byte(emptyPage))
			return
		}
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		_ = json.NewEncoder(w).Encode(schedulingmodel.AvailabilityConstraint{Id: "ac1", UserId: "u1", Kind: schedulingmodel.AvailabilityKindAvailable})
	})
	ctx := context.Background()
	create := schedulingapi.CreateAvailabilityConstraintRequest{Kind: schedulingmodel.AvailabilityKindAvailable, RecurrenceRule: "DTSTART;TZID=UTC:20260105T090000\nRRULE:FREQ=WEEKLY;BYDAY=MO", DurationMinutes: 480}

	_, err := s.AvailabilityConstraints.ListMine(nil).All(ctx)
	require.NoError(t, err)
	require.Equal(t, "/scheduling/v1/me/availability-constraints", path)

	got, err := s.AvailabilityConstraints.CreateMine(ctx, create)
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, method)
	require.Equal(t, "/scheduling/v1/me/availability-constraints", path)
	require.Contains(t, body, `"recurrence_rule"`)
	require.Equal(t, "ac1", got.Id)

	label := "Hours"
	_, err = s.AvailabilityConstraints.UpdateMine(ctx, "ac1", schedulingapi.UpdateAvailabilityConstraintRequest{Label: &label})
	require.NoError(t, err)
	require.Equal(t, http.MethodPatch, method)
	require.Equal(t, "/scheduling/v1/me/availability-constraints/ac1", path)
	require.JSONEq(t, `{"label":"Hours"}`, body)

	require.NoError(t, s.AvailabilityConstraints.DeleteMine(ctx, "ac1"))
	require.Equal(t, "/scheduling/v1/me/availability-constraints/ac1", path)

	_, err = s.AvailabilityConstraints.ListForUser("u1", nil).All(ctx)
	require.NoError(t, err)
	require.Equal(t, "/scheduling/v1/users/u1/availability-constraints", path)
	_, err = s.AvailabilityConstraints.CreateForUser(ctx, "u1", create)
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, method)
	_, err = s.AvailabilityConstraints.GetForUser(ctx, "u1", "ac1")
	require.NoError(t, err)
	require.Equal(t, "/scheduling/v1/users/u1/availability-constraints/ac1", path)
	require.NoError(t, s.AvailabilityConstraints.DeleteForUser(ctx, "u1", "ac1"))
	require.Equal(t, http.MethodDelete, method)
}

func TestSlotService_Get(t *testing.T) {
	var method, path, body string
	s := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		body = string(raw)
		_, _ = w.Write([]byte(`{"slots":{"2026-01-05":[{"start_at":"2026-01-05T08:00:00Z","end_at":"2026-01-05T08:30:00Z","host_user_ids":["u1"]}]},"unavailability":[{"start_at":"2026-01-05T00:00:00Z","end_at":"2026-01-05T04:00:00Z","host_user_id":"u1","reason":"outside_hours"}]}`))
	})
	from := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	got, err := s.Slots.Get(context.Background(), schedulingapi.GetSlotsRequest{TypeKey: "intro-call", HostUserIds: []string{"u1"}, From: from, To: from.AddDate(0, 0, 1), Timezone: "Europe/Berlin", IsVerbose: true})
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, method)
	require.Equal(t, "/scheduling/v1/slots", path)
	require.Contains(t, body, `"type_key":"intro-call"`)
	require.Contains(t, body, `"is_verbose":true`)
	require.Len(t, got.Slots["2026-01-05"], 1)
	require.Equal(t, []string{"u1"}, got.Slots["2026-01-05"][0].HostUserIds)
	require.Equal(t, schedulingmodel.SlotUnavailabilityOutsideHours, got.Unavailability[0].Reason)
}
