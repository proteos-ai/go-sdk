package scheduling

import (
	"context"
	"net/http"
	"net/url"
	"time"

	schedulingmodel "go.proteos.ai/model/scheduling"
	schedulingapi "go.proteos.ai/model/scheduling/api"
	sdk "go.proteos.ai/sdk"
)

// SchedulingLinkService is the booking entry points: scheduling links keyed
// by (org, key) that bind hosts + an assignment to one or more calendar event
// types. List / Get / Create / Update / Delete are the config surface
// (scheduling-links:read|write|delete); GetSlots / Book read the link's
// offer and take one slot on behalf of a contact (scheduling-links:read).
type SchedulingLinkService struct{ c *sdk.Client }

func schedulingLinksPath() string {
	return basePath + "/scheduling-links"
}

func schedulingLinkPath(key string) string {
	return schedulingLinksPath() + "/" + url.PathEscape(key)
}

// List iterates the org's scheduling links.
func (s *SchedulingLinkService) List(opts *ListSchedulingLinksOptions) *sdk.PageIterator[schedulingmodel.SchedulingLink, ListSchedulingLinksOptions] {
	o := ListSchedulingLinksOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListSchedulingLinksOptions) (sdk.ListResult[schedulingmodel.SchedulingLink], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

// ListPage fetches one page of scheduling links.
func (s *SchedulingLinkService) ListPage(ctx context.Context, opts *ListSchedulingLinksOptions) (sdk.ListResult[schedulingmodel.SchedulingLink], error) {
	var out sdk.ListResult[schedulingmodel.SchedulingLink]
	err := s.c.DoWithQuery(ctx, http.MethodGet, schedulingLinksPath(), opts, nil, &out)
	return out, err
}

func (s *SchedulingLinkService) Get(ctx context.Context, key string) (schedulingmodel.SchedulingLink, error) {
	var out schedulingmodel.SchedulingLink
	err := s.c.Do(ctx, http.MethodGet, schedulingLinkPath(key), nil, &out)
	return out, err
}

// Create adds a link under req.Key (409 scheduling_link_exists when taken).
func (s *SchedulingLinkService) Create(ctx context.Context, req schedulingapi.CreateSchedulingLinkRequest) (schedulingmodel.SchedulingLink, error) {
	var out schedulingmodel.SchedulingLink
	err := s.c.Do(ctx, http.MethodPost, schedulingLinksPath(), req, &out)
	return out, err
}

// Update is a partial update; nil fields are left as they are.
func (s *SchedulingLinkService) Update(ctx context.Context, key string, req schedulingapi.UpdateSchedulingLinkRequest) (schedulingmodel.SchedulingLink, error) {
	var out schedulingmodel.SchedulingLink
	err := s.c.Do(ctx, http.MethodPatch, schedulingLinkPath(key), req, &out)
	return out, err
}

func (s *SchedulingLinkService) Delete(ctx context.Context, key string) error {
	return s.c.Do(ctx, http.MethodDelete, schedulingLinkPath(key), nil, nil)
}

// GetSlots reads the link's offer in [query.From, query.To) (at most 31
// days) as the requester's local days. TypeKey is required on a multi-type
// link; HostUserId narrows to one of the link's hosts.
func (s *SchedulingLinkService) GetSlots(ctx context.Context, key string, query schedulingapi.GetPublicSlotsQuery) (schedulingapi.GetSlotsResponse, error) {
	var out schedulingapi.GetSlotsResponse
	err := s.c.DoWithQuery(ctx, http.MethodGet, schedulingLinkPath(key)+"/slots", publicSlotsQueryOf(query), nil, &out)
	return out, err
}

// Book takes one offered slot through the link for req.Contact (an empty
// contact books for the caller's own platform contact). The answer carries
// the one-time manage token — it is never readable again.
func (s *SchedulingLinkService) Book(ctx context.Context, key string, req schedulingapi.BookSchedulingLinkRequest) (schedulingapi.BookSchedulingLinkResponse, error) {
	var out schedulingapi.BookSchedulingLinkResponse
	err := s.c.Do(ctx, http.MethodPost, schedulingLinkPath(key)+"/calendar-events", req, &out)
	return out, err
}

// PublicSchedulingService is the unauthenticated scheduling surface behind
// the /s/ pages: a public + enabled link, its slots, taking one, and the
// booked event by its manage token. The routes need no token and are
// rate-limited (429 rate_limited). The *sdk.Client still sends whatever
// Authorization it was built with — the service ignores the header on these
// routes — so a client without a token works just as well here.
type PublicSchedulingService struct{ c *sdk.Client }

func publicOrgPath(orgId string) string {
	return basePath + "/public/orgs/" + url.PathEscape(orgId)
}

func publicLinkPath(orgId, key string) string {
	return publicOrgPath(orgId) + "/links/" + url.PathEscape(key)
}

func publicCalendarEventPath(orgId, id string) string {
	return publicOrgPath(orgId) + "/calendar-events/" + url.PathEscape(id)
}

// GetLink resolves a public link as the page sees it. contactId (optional)
// binds the page to a contact the org knows: no name / email asked on book.
// A private or unknown link answers 404 scheduling_link_not_found.
func (s *PublicSchedulingService) GetLink(ctx context.Context, orgId, key, contactId string) (schedulingapi.PublicSchedulingLink, error) {
	var out schedulingapi.PublicSchedulingLink
	err := s.c.DoWithQuery(ctx, http.MethodGet, publicLinkPath(orgId, key), publicLinkQuery{ContactId: contactId}, nil, &out)
	return out, err
}

// GetSlots reads the public link's offer (see SchedulingLinkService.GetSlots).
func (s *PublicSchedulingService) GetSlots(ctx context.Context, orgId, key string, query schedulingapi.GetPublicSlotsQuery) (schedulingapi.GetSlotsResponse, error) {
	var out schedulingapi.GetSlotsResponse
	err := s.c.DoWithQuery(ctx, http.MethodGet, publicLinkPath(orgId, key)+"/slots", publicSlotsQueryOf(query), nil, &out)
	return out, err
}

// Book takes one offered slot through the public link. req.Contact.Id when
// the page is contact-bound, otherwise Email is required
// (400 booking_contact_required).
func (s *PublicSchedulingService) Book(ctx context.Context, orgId, key string, req schedulingapi.BookSchedulingLinkRequest) (schedulingapi.PublicBookSchedulingLinkResponse, error) {
	var out schedulingapi.PublicBookSchedulingLinkResponse
	err := s.c.Do(ctx, http.MethodPost, publicLinkPath(orgId, key)+"/calendar-events", req, &out)
	return out, err
}

// GetEvent reads a booked event by its manage token (403 manage_token_invalid).
func (s *PublicSchedulingService) GetEvent(ctx context.Context, orgId, id, token string) (schedulingapi.PublicCalendarEvent, error) {
	var out schedulingapi.PublicCalendarEvent
	query := schedulingapi.GetPublicCalendarEventQuery{Token: token}
	err := s.c.DoWithQuery(ctx, http.MethodGet, publicCalendarEventPath(orgId, id), query, nil, &out)
	return out, err
}

// Cancel cancels a booked event by its manage token (409 cancel_not_allowed /
// cancel_notice_passed under the type's policy).
func (s *PublicSchedulingService) Cancel(ctx context.Context, orgId, id string, req schedulingapi.CancelPublicCalendarEventRequest) (schedulingapi.PublicCalendarEvent, error) {
	var out schedulingapi.PublicCalendarEvent
	err := s.c.Do(ctx, http.MethodPost, publicCalendarEventPath(orgId, id)+"/cancel", req, &out)
	return out, err
}

// Reschedule moves a booked event to another offered slot by its manage
// token; the same token stays valid (409 reschedule_not_allowed under the
// type's policy).
func (s *PublicSchedulingService) Reschedule(ctx context.Context, orgId, id string, req schedulingapi.ReschedulePublicCalendarEventRequest) (schedulingapi.PublicBookSchedulingLinkResponse, error) {
	var out schedulingapi.PublicBookSchedulingLinkResponse
	err := s.c.Do(ctx, http.MethodPost, publicCalendarEventPath(orgId, id)+"/reschedule", req, &out)
	return out, err
}

// publicSlotsQuery is the wire form of schedulingapi.GetPublicSlotsQuery: the
// SDK's query encoder skips time.Time (a struct), so From / To travel as
// RFC3339 strings.
type publicSlotsQuery struct {
	TypeKey    string `query:"type_key,omitempty"`
	From       string `query:"from"`
	To         string `query:"to"`
	Timezone   string `query:"timezone"`
	HostUserId string `query:"host_user_id,omitempty"`
}

func publicSlotsQueryOf(query schedulingapi.GetPublicSlotsQuery) publicSlotsQuery {
	return publicSlotsQuery{
		TypeKey:    query.TypeKey,
		From:       query.From.Format(time.RFC3339),
		To:         query.To.Format(time.RFC3339),
		Timezone:   query.Timezone,
		HostUserId: query.HostUserId,
	}
}

// publicLinkQuery is the wire form of schedulingapi.GetPublicSchedulingLinkQuery
// with an empty contact_id kept off the wire.
type publicLinkQuery struct {
	ContactId string `query:"contact_id,omitempty"`
}
