package scheduling

import (
	"context"
	"net/http"
	"net/url"

	schedulingmodel "go.proteos.ai/model/scheduling"
	schedulingapi "go.proteos.ai/model/scheduling/api"
	sdk "go.proteos.ai/sdk"
)

// CalendarConnectionService manages bindings to connector-service calendar
// grants. A user may hold several (work Google + personal Gmail + Outlook).
type CalendarConnectionService struct{ c *sdk.Client }

func mineCalendarConnectionsPath() string {
	return basePath + "/me/calendar-connections"
}

func userCalendarConnectionsPath(userId string) string {
	return basePath + "/users/" + url.PathEscape(userId) + "/calendar-connections"
}

func (s *CalendarConnectionService) list(path string, opts *ListCalendarConnectionsOptions) *sdk.PageIterator[schedulingmodel.CalendarConnection, ListCalendarConnectionsOptions] {
	o := ListCalendarConnectionsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListCalendarConnectionsOptions) (sdk.ListResult[schedulingmodel.CalendarConnection], error) {
		in.Page = page
		return s.listPage(ctx, path, &in)
	}, o)
}

func (s *CalendarConnectionService) listPage(ctx context.Context, path string, opts *ListCalendarConnectionsOptions) (sdk.ListResult[schedulingmodel.CalendarConnection], error) {
	var out sdk.ListResult[schedulingmodel.CalendarConnection]
	err := s.c.DoWithQuery(ctx, http.MethodGet, path, opts, nil, &out)
	return out, err
}

// ListMine iterates the caller's own calendar connections.
func (s *CalendarConnectionService) ListMine(opts *ListCalendarConnectionsOptions) *sdk.PageIterator[schedulingmodel.CalendarConnection, ListCalendarConnectionsOptions] {
	return s.list(mineCalendarConnectionsPath(), opts)
}

// ListMinePage fetches one page of the caller's own calendar connections.
func (s *CalendarConnectionService) ListMinePage(ctx context.Context, opts *ListCalendarConnectionsOptions) (sdk.ListResult[schedulingmodel.CalendarConnection], error) {
	return s.listPage(ctx, mineCalendarConnectionsPath(), opts)
}

// Connect binds one of the caller's user-scope calendar connector connections
// to the mirror; the service discovers its calendars.
func (s *CalendarConnectionService) Connect(ctx context.Context, req schedulingapi.CreateCalendarConnectionRequest) (schedulingmodel.CalendarConnection, error) {
	var out schedulingmodel.CalendarConnection
	err := s.c.Do(ctx, http.MethodPost, mineCalendarConnectionsPath(), req, &out)
	return out, err
}

func (s *CalendarConnectionService) UpdateMine(ctx context.Context, id string, req schedulingapi.UpdateCalendarConnectionRequest) (schedulingmodel.CalendarConnection, error) {
	var out schedulingmodel.CalendarConnection
	err := s.c.Do(ctx, http.MethodPatch, mineCalendarConnectionsPath()+"/"+url.PathEscape(id), req, &out)
	return out, err
}

func (s *CalendarConnectionService) DisconnectMine(ctx context.Context, id string) error {
	return s.c.Do(ctx, http.MethodDelete, mineCalendarConnectionsPath()+"/"+url.PathEscape(id), nil, nil)
}

// SyncMine is "sync now": re-lists the account's calendars and enqueues a
// mirror sync of every synced calendar under the connection; IsRebaseline
// drops the change cursors so the worker re-reads the whole window. Returns
// the refreshed connection.
func (s *CalendarConnectionService) SyncMine(ctx context.Context, id string, req schedulingapi.UpdateCalendarSyncRequest) (schedulingmodel.CalendarConnection, error) {
	var out schedulingmodel.CalendarConnection
	err := s.c.Do(ctx, http.MethodPost, mineCalendarConnectionsPath()+"/"+url.PathEscape(id)+"/sync", req, &out)
	return out, err
}

// ListForUser iterates another user's calendar connections (admin path).
func (s *CalendarConnectionService) ListForUser(userId string, opts *ListCalendarConnectionsOptions) *sdk.PageIterator[schedulingmodel.CalendarConnection, ListCalendarConnectionsOptions] {
	return s.list(userCalendarConnectionsPath(userId), opts)
}

// ListForUserPage fetches one page of another user's calendar connections.
func (s *CalendarConnectionService) ListForUserPage(ctx context.Context, userId string, opts *ListCalendarConnectionsOptions) (sdk.ListResult[schedulingmodel.CalendarConnection], error) {
	return s.listPage(ctx, userCalendarConnectionsPath(userId), opts)
}

func (s *CalendarConnectionService) UpdateForUser(ctx context.Context, userId, id string, req schedulingapi.UpdateCalendarConnectionRequest) (schedulingmodel.CalendarConnection, error) {
	var out schedulingmodel.CalendarConnection
	err := s.c.Do(ctx, http.MethodPatch, userCalendarConnectionsPath(userId)+"/"+url.PathEscape(id), req, &out)
	return out, err
}

func (s *CalendarConnectionService) DisconnectForUser(ctx context.Context, userId, id string) error {
	return s.c.Do(ctx, http.MethodDelete, userCalendarConnectionsPath(userId)+"/"+url.PathEscape(id), nil, nil)
}
