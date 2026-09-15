package scheduling

import (
	"context"
	"net/http"
	"net/url"

	schedulingmodel "go.proteos.ai/model/scheduling"
	schedulingapi "go.proteos.ai/model/scheduling/api"
	sdk "go.proteos.ai/sdk"
)

// CalendarService reads and re-roles the provider calendars under a
// connection: is_synced, is_availability_source, is_target (one per user),
// visibility, color.
type CalendarService struct{ c *sdk.Client }

func mineCalendarsPath() string {
	return basePath + "/me/calendars"
}

func userCalendarsPath(userId string) string {
	return basePath + "/users/" + url.PathEscape(userId) + "/calendars"
}

func (s *CalendarService) list(path string, opts *ListCalendarsOptions) *sdk.PageIterator[schedulingmodel.Calendar, ListCalendarsOptions] {
	o := ListCalendarsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListCalendarsOptions) (sdk.ListResult[schedulingmodel.Calendar], error) {
		in.Page = page
		return s.listPage(ctx, path, &in)
	}, o)
}

func (s *CalendarService) listPage(ctx context.Context, path string, opts *ListCalendarsOptions) (sdk.ListResult[schedulingmodel.Calendar], error) {
	var out sdk.ListResult[schedulingmodel.Calendar]
	err := s.c.DoWithQuery(ctx, http.MethodGet, path, opts, nil, &out)
	return out, err
}

func (s *CalendarService) get(ctx context.Context, path string) (schedulingmodel.Calendar, error) {
	var out schedulingmodel.Calendar
	err := s.c.Do(ctx, http.MethodGet, path, nil, &out)
	return out, err
}

func (s *CalendarService) update(ctx context.Context, path string, req schedulingapi.UpdateCalendarRequest) (schedulingmodel.Calendar, error) {
	var out schedulingmodel.Calendar
	err := s.c.Do(ctx, http.MethodPatch, path, req, &out)
	return out, err
}

// ListMine iterates the caller's own calendars.
func (s *CalendarService) ListMine(opts *ListCalendarsOptions) *sdk.PageIterator[schedulingmodel.Calendar, ListCalendarsOptions] {
	return s.list(mineCalendarsPath(), opts)
}

// ListMinePage fetches one page of the caller's own calendars.
func (s *CalendarService) ListMinePage(ctx context.Context, opts *ListCalendarsOptions) (sdk.ListResult[schedulingmodel.Calendar], error) {
	return s.listPage(ctx, mineCalendarsPath(), opts)
}

func (s *CalendarService) GetMine(ctx context.Context, id string) (schedulingmodel.Calendar, error) {
	return s.get(ctx, mineCalendarsPath()+"/"+url.PathEscape(id))
}

// UpdateMine changes the platform roles of one of the caller's calendars.
// IsTarget=true demotes the current target in the same transaction.
func (s *CalendarService) UpdateMine(ctx context.Context, id string, req schedulingapi.UpdateCalendarRequest) (schedulingmodel.Calendar, error) {
	return s.update(ctx, mineCalendarsPath()+"/"+url.PathEscape(id), req)
}

// ListForUser iterates another user's calendars (admin path).
func (s *CalendarService) ListForUser(userId string, opts *ListCalendarsOptions) *sdk.PageIterator[schedulingmodel.Calendar, ListCalendarsOptions] {
	return s.list(userCalendarsPath(userId), opts)
}

// ListForUserPage fetches one page of another user's calendars.
func (s *CalendarService) ListForUserPage(ctx context.Context, userId string, opts *ListCalendarsOptions) (sdk.ListResult[schedulingmodel.Calendar], error) {
	return s.listPage(ctx, userCalendarsPath(userId), opts)
}

func (s *CalendarService) GetForUser(ctx context.Context, userId, id string) (schedulingmodel.Calendar, error) {
	return s.get(ctx, userCalendarsPath(userId)+"/"+url.PathEscape(id))
}

func (s *CalendarService) UpdateForUser(ctx context.Context, userId, id string, req schedulingapi.UpdateCalendarRequest) (schedulingmodel.Calendar, error) {
	return s.update(ctx, userCalendarsPath(userId)+"/"+url.PathEscape(id), req)
}
