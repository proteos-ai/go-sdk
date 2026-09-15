package scheduling

import (
	"context"
	"net/http"
	"net/url"

	schedulingmodel "go.proteos.ai/model/scheduling"
	schedulingapi "go.proteos.ai/model/scheduling/api"
	sdk "go.proteos.ai/sdk"
)

// CalendarEventTypeService is the slot-rules catalog: calendar event types
// keyed by (org, key), created or replaced with PUT semantics.
type CalendarEventTypeService struct{ c *sdk.Client }

func calendarEventTypesPath() string {
	return basePath + "/calendar-event-types"
}

func calendarEventTypePath(key string) string {
	return calendarEventTypesPath() + "/" + url.PathEscape(key)
}

// List iterates the org's calendar event types.
func (s *CalendarEventTypeService) List(opts *ListCalendarEventTypesOptions) *sdk.PageIterator[schedulingmodel.CalendarEventType, ListCalendarEventTypesOptions] {
	o := ListCalendarEventTypesOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListCalendarEventTypesOptions) (sdk.ListResult[schedulingmodel.CalendarEventType], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

// ListPage fetches one page of calendar event types.
func (s *CalendarEventTypeService) ListPage(ctx context.Context, opts *ListCalendarEventTypesOptions) (sdk.ListResult[schedulingmodel.CalendarEventType], error) {
	var out sdk.ListResult[schedulingmodel.CalendarEventType]
	err := s.c.DoWithQuery(ctx, http.MethodGet, calendarEventTypesPath(), opts, nil, &out)
	return out, err
}

func (s *CalendarEventTypeService) Get(ctx context.Context, key string) (schedulingmodel.CalendarEventType, error) {
	var out schedulingmodel.CalendarEventType
	err := s.c.Do(ctx, http.MethodGet, calendarEventTypePath(key), nil, &out)
	return out, err
}

// Put creates or replaces the type under key (every field is the new value).
func (s *CalendarEventTypeService) Put(ctx context.Context, key string, req schedulingapi.PutCalendarEventTypeRequest) (schedulingmodel.CalendarEventType, error) {
	var out schedulingmodel.CalendarEventType
	err := s.c.Do(ctx, http.MethodPut, calendarEventTypePath(key), req, &out)
	return out, err
}

// Update is a partial update; nil fields are left as they are.
func (s *CalendarEventTypeService) Update(ctx context.Context, key string, req schedulingapi.UpdateCalendarEventTypeRequest) (schedulingmodel.CalendarEventType, error) {
	var out schedulingmodel.CalendarEventType
	err := s.c.Do(ctx, http.MethodPatch, calendarEventTypePath(key), req, &out)
	return out, err
}

func (s *CalendarEventTypeService) Delete(ctx context.Context, key string) error {
	return s.c.Do(ctx, http.MethodDelete, calendarEventTypePath(key), nil, nil)
}
