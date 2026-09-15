package scheduling_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	schedulingmodel "go.proteos.ai/model/scheduling"
	schedulingapi "go.proteos.ai/model/scheduling/api"
	"go.proteos.ai/sdk/scheduling"
)

func TestSchedulingLinkService_Paths(t *testing.T) {
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
			_ = json.NewEncoder(w).Encode(schedulingmodel.SchedulingLink{Key: "intro-call", Name: "Intro call", Assignment: schedulingmodel.AssignmentRoundRobin})
		}
	})
	ctx := context.Background()
	isPublic, isEnabled := true, false

	_, err := s.SchedulingLinks.ListPage(ctx, &scheduling.ListSchedulingLinksOptions{TypeKey: "intro", IsPublic: &isPublic, Search: "call"})
	require.NoError(t, err)
	require.Equal(t, http.MethodGet, method)
	require.Equal(t, "/scheduling/v1/scheduling-links", path)
	require.Equal(t, "intro", seen.Get("type_key"))
	require.Equal(t, "true", seen.Get("is_public"))
	require.Equal(t, "call", seen.Get("search"))
	require.False(t, seen.Has("is_enabled"), "nil tri-state filter must stay off the wire")

	got, err := s.SchedulingLinks.Get(ctx, "intro call")
	require.NoError(t, err)
	require.Equal(t, http.MethodGet, method)
	require.Equal(t, "/scheduling/v1/scheduling-links/intro%20call", path)
	require.Equal(t, "Intro call", got.Name)

	_, err = s.SchedulingLinks.Create(ctx, schedulingapi.CreateSchedulingLinkRequest{
		Key:         "intro-call",
		Name:        "Intro call",
		CalendarEventTypeKeys:    []string{"intro"},
		HostUserIds: []string{"u1", "u2"},
		Assignment:  schedulingmodel.AssignmentRoundRobin,
		IsPublic:    true,
	})
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, method)
	require.Equal(t, "/scheduling/v1/scheduling-links", path)
	require.JSONEq(t, `{"key":"intro-call","name":"Intro call","calendar_event_type_keys":["intro"],"host_user_ids":["u1","u2"],"assignment":"round_robin","is_public":true,"is_host_selectable":false}`, body)

	rescheduleHost := schedulingmodel.RescheduleHostAny
	_, err = s.SchedulingLinks.Update(ctx, "intro-call", schedulingapi.UpdateSchedulingLinkRequest{IsEnabled: &isEnabled, RescheduleHost: &rescheduleHost})
	require.NoError(t, err)
	require.Equal(t, http.MethodPatch, method)
	require.Equal(t, "/scheduling/v1/scheduling-links/intro-call", path)
	require.JSONEq(t, `{"is_enabled":false,"reschedule_host":"any"}`, body)

	require.NoError(t, s.SchedulingLinks.Delete(ctx, "intro-call"))
	require.Equal(t, http.MethodDelete, method)
	require.Equal(t, "/scheduling/v1/scheduling-links/intro-call", path)
}

func TestSchedulingLinkService_List_Iterates(t *testing.T) {
	s := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		require.Equal(t, "1", r.URL.Query().Get("page_size"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"meta": map[string]any{"page": page, "page_size": 1, "items_total": 2, "pages_total": 2},
			"data": []schedulingmodel.SchedulingLink{{Key: fmt.Sprintf("link-%d", page)}},
		})
	})
	all, err := s.SchedulingLinks.List(&scheduling.ListSchedulingLinksOptions{ListOptions: scheduling.ListOptions{PageSize: 1}}).All(context.Background())
	require.NoError(t, err)
	require.Equal(t, []string{"link-0", "link-1"}, []string{all[0].Key, all[1].Key})
}

func TestSchedulingLinkService_SlotsAndBook(t *testing.T) {
	var method, path, body string
	var seen url.Values
	s := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		method, path, seen = r.Method, r.URL.Path, r.URL.Query()
		raw, _ := io.ReadAll(r.Body)
		body = string(raw)
		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(schedulingapi.GetSlotsResponse{Slots: map[string][]schedulingmodel.Slot{
				"2026-09-16": {{HostUserIds: []string{"u1"}}},
			}})
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(schedulingapi.BookSchedulingLinkResponse{
			CalendarEvent: schedulingmodel.CalendarEvent{Id: "ev1", Title: "Intro call", LinkKey: "intro-call"},
			ManageToken:   "tok-1",
			ManageUrl:     "https://app.example/s/org1/events/ev1?token=tok-1",
		})
	})
	ctx := context.Background()
	from := time.Date(2026, 9, 16, 0, 0, 0, 0, time.UTC)

	slots, err := s.SchedulingLinks.GetSlots(ctx, "intro-call", schedulingapi.GetPublicSlotsQuery{
		From:     from,
		To:       from.Add(7 * 24 * time.Hour),
		Timezone: "Europe/Berlin",
		TypeKey:  "intro",
	})
	require.NoError(t, err)
	require.Equal(t, http.MethodGet, method)
	require.Equal(t, "/scheduling/v1/scheduling-links/intro-call/slots", path)
	require.Equal(t, "2026-09-16T00:00:00Z", seen.Get("from"))
	require.Equal(t, "2026-09-23T00:00:00Z", seen.Get("to"))
	require.Equal(t, "Europe/Berlin", seen.Get("timezone"))
	require.Equal(t, "intro", seen.Get("type_key"))
	require.False(t, seen.Has("host_user_id"), "empty host filter stays off the wire")
	require.Len(t, slots.Slots["2026-09-16"], 1)

	booked, err := s.SchedulingLinks.Book(ctx, "intro-call", schedulingapi.BookSchedulingLinkRequest{
		StartAt:    from.Add(9 * time.Hour),
		Timezone:   "Europe/Berlin",
		Contact:    schedulingapi.BookingContact{Name: "Ada", Email: "ada@example.com"},
		Notes:      "via sdk",
		HostUserId: "u1",
	})
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, method)
	require.Equal(t, "/scheduling/v1/scheduling-links/intro-call/calendar-events", path)
	require.JSONEq(t, `{"start_at":"2026-09-16T09:00:00Z","timezone":"Europe/Berlin","contact":{"name":"Ada","email":"ada@example.com"},"notes":"via sdk","host_user_id":"u1"}`, body)
	require.Equal(t, "ev1", booked.Id)
	require.Equal(t, "intro-call", booked.LinkKey)
	require.Equal(t, "tok-1", booked.ManageToken)
	require.Equal(t, "https://app.example/s/org1/events/ev1?token=tok-1", booked.ManageUrl)
}

func TestPublicSchedulingService_Paths(t *testing.T) {
	var method, path, body string
	var seen url.Values
	s := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		// EscapedPath keeps the %20 the SDK put in the org segment (Path decodes it).
		method, path, seen = r.Method, r.URL.EscapedPath(), r.URL.Query()
		raw, _ := io.ReadAll(r.Body)
		body = string(raw)
		switch {
		case path == "/scheduling/v1/public/orgs/org%201/links/intro-call":
			_ = json.NewEncoder(w).Encode(schedulingapi.PublicSchedulingLink{OrgId: "org 1", Key: "intro-call", IsContactBound: r.URL.Query().Has("contact_id")})
		case path == "/scheduling/v1/public/orgs/org%201/links/intro-call/slots":
			_ = json.NewEncoder(w).Encode(schedulingapi.GetSlotsResponse{Slots: map[string][]schedulingmodel.Slot{}})
		case r.Method == http.MethodPost && path == "/scheduling/v1/public/orgs/org%201/links/intro-call/calendar-events":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(schedulingapi.PublicBookSchedulingLinkResponse{PublicCalendarEvent: schedulingapi.PublicCalendarEvent{Id: "ev1"}, ManageToken: "tok-1"})
		case path == "/scheduling/v1/public/orgs/org%201/calendar-events/ev1/reschedule":
			_ = json.NewEncoder(w).Encode(schedulingapi.PublicBookSchedulingLinkResponse{PublicCalendarEvent: schedulingapi.PublicCalendarEvent{Id: "ev1"}, ManageToken: "tok-1"})
		default:
			_ = json.NewEncoder(w).Encode(schedulingapi.PublicCalendarEvent{Id: "ev1", IsCancelled: path == "/scheduling/v1/public/orgs/org%201/calendar-events/ev1/cancel"})
		}
	})
	ctx := context.Background()
	at := time.Date(2026, 9, 16, 9, 0, 0, 0, time.UTC)

	link, err := s.Public.GetLink(ctx, "org 1", "intro-call", "")
	require.NoError(t, err)
	require.Equal(t, http.MethodGet, method)
	require.Equal(t, "/scheduling/v1/public/orgs/org%201/links/intro-call", path)
	require.False(t, seen.Has("contact_id"), "empty contact stays off the wire")
	require.False(t, link.IsContactBound)

	link, err = s.Public.GetLink(ctx, "org 1", "intro-call", "c1")
	require.NoError(t, err)
	require.Equal(t, "c1", seen.Get("contact_id"))
	require.True(t, link.IsContactBound)

	_, err = s.Public.GetSlots(ctx, "org 1", "intro-call", schedulingapi.GetPublicSlotsQuery{From: at, To: at.Add(24 * time.Hour), Timezone: "UTC", HostUserId: "u1"})
	require.NoError(t, err)
	require.Equal(t, "/scheduling/v1/public/orgs/org%201/links/intro-call/slots", path)
	require.Equal(t, "2026-09-16T09:00:00Z", seen.Get("from"))
	require.Equal(t, "2026-09-17T09:00:00Z", seen.Get("to"))
	require.Equal(t, "UTC", seen.Get("timezone"))
	require.Equal(t, "u1", seen.Get("host_user_id"))
	require.False(t, seen.Has("type_key"))

	booked, err := s.Public.Book(ctx, "org 1", "intro-call", schedulingapi.BookSchedulingLinkRequest{StartAt: at, Timezone: "UTC", Contact: schedulingapi.BookingContact{Id: "c1"}})
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, method)
	require.Equal(t, "/scheduling/v1/public/orgs/org%201/links/intro-call/calendar-events", path)
	require.JSONEq(t, `{"start_at":"2026-09-16T09:00:00Z","timezone":"UTC","contact":{"id":"c1"},"notes":""}`, body)
	require.Equal(t, "tok-1", booked.ManageToken)

	event, err := s.Public.GetEvent(ctx, "org 1", "ev1", "tok-1")
	require.NoError(t, err)
	require.Equal(t, http.MethodGet, method)
	require.Equal(t, "/scheduling/v1/public/orgs/org%201/calendar-events/ev1", path)
	require.Equal(t, "tok-1", seen.Get("token"))
	require.Equal(t, "ev1", event.Id)

	cancelled, err := s.Public.Cancel(ctx, "org 1", "ev1", schedulingapi.CancelPublicCalendarEventRequest{Token: "tok-1", Reason: "conflict"})
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, method)
	require.Equal(t, "/scheduling/v1/public/orgs/org%201/calendar-events/ev1/cancel", path)
	require.JSONEq(t, `{"token":"tok-1","reason":"conflict"}`, body)
	require.True(t, cancelled.IsCancelled)

	moved, err := s.Public.Reschedule(ctx, "org 1", "ev1", schedulingapi.ReschedulePublicCalendarEventRequest{Token: "tok-1", StartAt: at.Add(time.Hour), Timezone: "UTC"})
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, method)
	require.Equal(t, "/scheduling/v1/public/orgs/org%201/calendar-events/ev1/reschedule", path)
	require.JSONEq(t, `{"token":"tok-1","start_at":"2026-09-16T10:00:00Z","timezone":"UTC"}`, body)
	require.Equal(t, "tok-1", moved.ManageToken)
}
