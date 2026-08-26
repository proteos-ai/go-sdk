package conversation

import (
	"context"
	"net/http"

	conversationmodel "go.proteos.ai/model/conversation"
	conversationapi "go.proteos.ai/model/conversation/api"
	sdk "go.proteos.ai/sdk"
)

const contactGroupsBasePath = "/conversations/v1/contact-groups"

// ContactGroupService manages the org's contact groups — the generic audience
// taxonomy whose membership lives on the contact (one group per contact),
// deployable via `pro module deploy` (contact-groups/<key>.json).
type ContactGroupService struct{ c *sdk.Client }

func (s *ContactGroupService) List(opts *ListContactGroupsOptions) *sdk.PageIterator[conversationmodel.ContactGroup, ListContactGroupsOptions] {
	o := ListContactGroupsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListContactGroupsOptions) (sdk.ListResult[conversationmodel.ContactGroup], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *ContactGroupService) ListPage(ctx context.Context, opts *ListContactGroupsOptions) (sdk.ListResult[conversationmodel.ContactGroup], error) {
	var out sdk.ListResult[conversationmodel.ContactGroup]
	err := s.c.DoWithQuery(ctx, http.MethodGet, contactGroupsBasePath, opts, nil, &out)
	return out, err
}

func (s *ContactGroupService) Get(ctx context.Context, key string) (conversationmodel.ContactGroup, error) {
	var out conversationmodel.ContactGroup
	err := s.c.Do(ctx, http.MethodGet, contactGroupsBasePath+"/"+key, nil, &out)
	return out, err
}

func (s *ContactGroupService) Create(ctx context.Context, req conversationapi.CreateContactGroupRequest) (conversationmodel.ContactGroup, error) {
	var out conversationmodel.ContactGroup
	err := s.c.Do(ctx, http.MethodPost, contactGroupsBasePath, req, &out)
	return out, err
}

func (s *ContactGroupService) Update(ctx context.Context, key string, req conversationapi.UpdateContactGroupRequest) (conversationmodel.ContactGroup, error) {
	var out conversationmodel.ContactGroup
	err := s.c.Do(ctx, http.MethodPatch, contactGroupsBasePath+"/"+key, req, &out)
	return out, err
}

// UpsertByKey calls `PUT /conversations/v1/contact-groups/:key` — the
// idempotent create-or-update path used by `pro module deploy`. An equivalent
// body is a server-side no-op; an empty module_slug keeps the stored
// attribution.
func (s *ContactGroupService) UpsertByKey(ctx context.Context, key string, req conversationapi.CreateContactGroupRequest) (conversationmodel.ContactGroup, error) {
	var out conversationmodel.ContactGroup
	req.Key = key
	err := s.c.Do(ctx, http.MethodPut, contactGroupsBasePath+"/"+key, req, &out)
	return out, err
}

func (s *ContactGroupService) Delete(ctx context.Context, key string) error {
	return s.c.Do(ctx, http.MethodDelete, contactGroupsBasePath+"/"+key, nil, nil)
}
