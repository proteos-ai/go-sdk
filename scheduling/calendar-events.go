package scheduling

import (
	"context"
	"net/http"
	"net/url"

	schedulingmodel "go.proteos.ai/model/scheduling"
	schedulingapi "go.proteos.ai/model/scheduling/api"
	sdk "go.proteos.ai/sdk"
)

// CalendarEventService is the two-way calendar_event mirror: window reads
// (colleagues' events projected by their calendar's visibility) and
// write-through creates / updates / deletes / RSVPs on the caller's own
// calendars.
type CalendarEventService struct{ c *sdk.Client }

func calendarEventsPath() string {
	return basePath + "/calendar-events"
}

func calendarEventPath(id string) string {
	return calendarEventsPath() + "/" + url.PathEscape(id)
}

// List reads a window of events (From / To required, ≤ 93 days). The response
// is `{data}` without pagination meta, so it returns the slice directly.
func (s *CalendarEventService) List(ctx context.Context, opts ListCalendarEventsOptions) ([]schedulingmodel.CalendarEvent, error) {
	var out schedulingapi.GetManyCalendarEventsResponse
	err := s.c.DoWithQuery(ctx, http.MethodGet, calendarEventsPath(), opts.wire(), nil, &out)
	return out.Data, err
}

func (s *CalendarEventService) Get(ctx context.Context, id string) (schedulingmodel.CalendarEvent, error) {
	var out schedulingmodel.CalendarEvent
	err := s.c.Do(ctx, http.MethodGet, calendarEventPath(id), nil, &out)
	return out, err
}

// Create writes an event through to the provider (calendar_id = own
// calendar, or hosts).
func (s *CalendarEventService) Create(ctx context.Context, req schedulingapi.CreateCalendarEventRequest) (schedulingmodel.CalendarEvent, error) {
	var out schedulingmodel.CalendarEvent
	err := s.c.Do(ctx, http.MethodPost, calendarEventsPath(), req, &out)
	return out, err
}

// Update is a partial update; scope picks the instance or the whole series on
// a recurring event and is REQUIRED there (400 recurrence_scope_required when
// empty); omit it only on non-recurring events. Same on Delete.
func (s *CalendarEventService) Update(ctx context.Context, id string, scope schedulingmodel.RecurrenceScope, req schedulingapi.UpdateCalendarEventRequest) (schedulingmodel.CalendarEvent, error) {
	var out schedulingmodel.CalendarEvent
	err := s.c.DoWithQuery(ctx, http.MethodPatch, calendarEventPath(id), scopeQuery{Scope: string(scope)}, req, &out)
	return out, err
}

func (s *CalendarEventService) Delete(ctx context.Context, id string, scope schedulingmodel.RecurrenceScope) error {
	return s.c.DoWithQuery(ctx, http.MethodDelete, calendarEventPath(id), scopeQuery{Scope: string(scope)}, nil, nil)
}

// Respond records the caller's RSVP on an event they were invited to.
func (s *CalendarEventService) Respond(ctx context.Context, id string, req schedulingapi.RespondCalendarEventRequest) (schedulingmodel.CalendarEvent, error) {
	var out schedulingmodel.CalendarEvent
	err := s.c.Do(ctx, http.MethodPost, calendarEventPath(id)+"/respond", req, &out)
	return out, err
}
