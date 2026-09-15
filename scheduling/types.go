package scheduling

import "time"

// ListOptions are the pagination + sort fields of the scheduling-service list
// endpoints. Pages are 0-indexed.
type ListOptions struct {
	Page      int    `query:"page"`
	PageSize  int    `query:"page_size"`
	SortBy    string `query:"sort_by,omitempty"`
	SortOrder string `query:"sort_direction,omitempty"`
}

// ListCalendarConnectionsOptions mirrors schedulingapi.GetManyCalendarConnectionsQuery.
type ListCalendarConnectionsOptions struct {
	ListOptions
	// Provider filters by calendar provider (google | microsoft).
	Provider string `query:"provider,omitempty"`
	// Status filters by binding health (active | error | revoked).
	Status string `query:"status,omitempty"`
}

// ListCalendarsOptions mirrors schedulingapi.GetManyCalendarsQuery. The bool
// filters are tri-state: nil leaves the filter off.
type ListCalendarsOptions struct {
	ListOptions
	CalendarConnectionId string `query:"calendar_connection_id,omitempty"`
	IsSynced             *bool  `query:"is_synced"`
	IsAvailabilitySource *bool  `query:"is_availability_source"`
	IsTarget             *bool  `query:"is_target"`
}

// ListCalendarEventsOptions mirrors schedulingapi.GetManyCalendarEventsQuery.
// It is defined SDK-side because the query encoder skips time.Time (a struct):
// From / To are sent as RFC3339. Both are required; the service caps the
// window at 93 days. UserIds / CalendarIds pick whose calendars (default: the
// caller's); colleagues' events come back projected by their calendar's
// visibility.
type ListCalendarEventsOptions struct {
	UserIds           []string
	CalendarIds       []string
	From              time.Time
	To                time.Time
	Timezone          string
	TypeKey           string
	RecordEntitySlug  string
	RecordId          string
	ContactId         string
	IsDeletedIncluded bool
}

// listCalendarEventsQuery is the wire form of ListCalendarEventsOptions.
type listCalendarEventsQuery struct {
	UserIds           []string `query:"user_ids"`
	CalendarIds       []string `query:"calendar_ids"`
	From              string   `query:"from"`
	To                string   `query:"to"`
	Timezone          string   `query:"timezone,omitempty"`
	TypeKey           string   `query:"type_key,omitempty"`
	RecordEntitySlug  string   `query:"record_entity_slug,omitempty"`
	RecordId          string   `query:"record_id,omitempty"`
	ContactId         string   `query:"contact_id,omitempty"`
	IsDeletedIncluded bool     `query:"is_deleted_included,omitempty"`
}

func (o ListCalendarEventsOptions) wire() listCalendarEventsQuery {
	return listCalendarEventsQuery{
		UserIds:           o.UserIds,
		CalendarIds:       o.CalendarIds,
		From:              o.From.Format(time.RFC3339),
		To:                o.To.Format(time.RFC3339),
		Timezone:          o.Timezone,
		TypeKey:           o.TypeKey,
		RecordEntitySlug:  o.RecordEntitySlug,
		RecordId:          o.RecordId,
		ContactId:         o.ContactId,
		IsDeletedIncluded: o.IsDeletedIncluded,
	}
}

// scopeQuery carries ?scope=instance|series on recurring-event edits.
type scopeQuery struct {
	Scope string `query:"scope,omitempty"`
}

// ListCalendarEventTypesOptions mirrors schedulingapi.GetManyCalendarEventTypesQuery.
type ListCalendarEventTypesOptions struct {
	ListOptions
	IsEnabled *bool  `query:"is_enabled"`
	Search    string `query:"search,omitempty"`
}

// ListSchedulingLinksOptions mirrors schedulingapi.GetManySchedulingLinksQuery.
// The bool filters are tri-state: nil leaves the filter off.
type ListSchedulingLinksOptions struct {
	ListOptions
	// TypeKey keeps links offering that calendar event type.
	TypeKey   string `query:"type_key,omitempty"`
	IsPublic  *bool  `query:"is_public"`
	IsEnabled *bool  `query:"is_enabled"`
	// Search is a case-insensitive substring of key or name.
	Search string `query:"search,omitempty"`
}

// ListAvailabilityConstraintsOptions mirrors schedulingapi.GetManyAvailabilityConstraintsQuery.
type ListAvailabilityConstraintsOptions struct {
	ListOptions
	// Kind filters by available | unavailable.
	Kind string `query:"kind,omitempty"`
}
