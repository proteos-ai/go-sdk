// Package scheduling provides services for scheduling-service resources over
// the platform API at /scheduling/v1: calendar connections (bindings to
// connector-service google-calendar / microsoft-calendar grants), the
// calendars under them with their platform roles, per-user scheduling
// profiles, the two-way calendar_event mirror, the slot engine and its
// catalog (calendar event types, availability constraints), and scheduling
// links with their hosted bookings. Public is the unauthenticated scheduling
// page surface (/public/orgs/:orgId/...).
//
// `*Mine` methods act on the caller's own rows (`/me/...`); `*ForUser`
// methods are the admin path (`/users/:userId/...`).
//
// Resource shapes come from go.proteos.ai/model/scheduling; the wire-format
// request types from go.proteos.ai/model/scheduling/api are reused directly.
// Only the list-options types (query-tagged for the SDK's query encoder) are
// defined locally.
//
// scheduling-service returns BARE single objects (no {data} envelope); list
// responses wrap in {meta, data}, except the calendar-events window which is
// {data} only.
//
//	s := scheduling.New(client)
//	connection, err := s.CalendarConnections.Connect(ctx, schedulingapi.CreateCalendarConnectionRequest{ConnectionId: id})
//	calendars, err := s.Calendars.ListMine(nil).All(ctx)
//	events, err := s.CalendarEvents.List(ctx, scheduling.ListCalendarEventsOptions{From: from, To: to})
//	booked, err := s.SchedulingLinks.Book(ctx, "intro-call", schedulingapi.BookSchedulingLinkRequest{StartAt: at, Timezone: "Europe/Berlin"})
//	link, err := s.Public.GetLink(ctx, orgId, "intro-call", "")
package scheduling

import sdk "go.proteos.ai/sdk"

const basePath = "/scheduling/v1"

// Client groups the scheduling-service resource services. Construct with New,
// then access them via the public fields.
type Client struct {
	CalendarConnections     *CalendarConnectionService
	Calendars               *CalendarService
	SchedulingProfiles      *SchedulingProfileService
	CalendarEvents          *CalendarEventService
	CalendarEventTypes      *CalendarEventTypeService
	AvailabilityConstraints *AvailabilityConstraintService
	Slots                   *SlotService
	SchedulingLinks         *SchedulingLinkService
	Public                  *PublicSchedulingService
}

// New builds a Client backed by the given *sdk.Client.
func New(c *sdk.Client) *Client {
	return &Client{
		CalendarConnections:     &CalendarConnectionService{c: c},
		Calendars:               &CalendarService{c: c},
		SchedulingProfiles:      &SchedulingProfileService{c: c},
		CalendarEvents:          &CalendarEventService{c: c},
		CalendarEventTypes:      &CalendarEventTypeService{c: c},
		AvailabilityConstraints: &AvailabilityConstraintService{c: c},
		Slots:                   &SlotService{c: c},
		SchedulingLinks:         &SchedulingLinkService{c: c},
		Public:                  &PublicSchedulingService{c: c},
	}
}
