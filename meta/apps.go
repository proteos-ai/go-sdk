package meta

import (
	"context"
	"errors"
	"net/http"
	metamodel "go.proteos.ai/model/meta"
	sdk "go.proteos.ai/sdk"
)

const appsBasePath = "/meta/v1/apps"

// AppServiceAPI is the contract an AppService satisfies.
type AppServiceAPI interface {
	List(opts *ListAppsOptions) *sdk.PageIterator[metamodel.App, ListAppsOptions]
	ListPage(ctx context.Context, opts *ListAppsOptions) (sdk.ListResult[metamodel.App], error)
	Get(ctx context.Context, slug string) (metamodel.App, error)
	Create(ctx context.Context, req CreateAppRequest) (metamodel.App, error)
	Update(ctx context.Context, slug string, req UpdateAppRequest) (metamodel.App, error)
	UpsertBySlug(ctx context.Context, slug string, req CreateAppRequest) (metamodel.App, error)
	Delete(ctx context.Context, slug string) error
}

// AppService manages apps. Apps are addressed by slug, unique per org.
type AppService struct{ c *sdk.Client }

var _ AppServiceAPI = (*AppService)(nil)

func (s *AppService) List(opts *ListAppsOptions) *sdk.PageIterator[metamodel.App, ListAppsOptions] {
	o := ListAppsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListAppsOptions) (sdk.ListResult[metamodel.App], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *AppService) ListPage(ctx context.Context, opts *ListAppsOptions) (sdk.ListResult[metamodel.App], error) {
	var out sdk.ListResult[metamodel.App]
	err := s.c.DoWithQuery(ctx, http.MethodGet, appsBasePath, opts, nil, &out)
	return out, err
}

func (s *AppService) Get(ctx context.Context, slug string) (metamodel.App, error) {
	var out metamodel.App
	err := s.c.Do(ctx, http.MethodGet, appsBasePath+"/"+slug, nil, &out)
	return out, err
}

func (s *AppService) Create(ctx context.Context, req CreateAppRequest) (metamodel.App, error) {
	var out metamodel.App
	err := s.c.Do(ctx, http.MethodPost, appsBasePath, req, &out)
	return out, err
}

func (s *AppService) Update(ctx context.Context, slug string, req UpdateAppRequest) (metamodel.App, error) {
	var out metamodel.App
	err := s.c.Do(ctx, http.MethodPatch, appsBasePath+"/"+slug, req, &out)
	return out, err
}

func (s *AppService) Delete(ctx context.Context, slug string) error {
	return s.c.Do(ctx, http.MethodDelete, appsBasePath+"/"+slug, nil, nil)
}

// UpsertBySlug creates the app if it doesn't exist, else updates it.
// metadata-service exposes no native upsert route for apps (POST creates,
// PATCH updates by slug), so we probe with GET and dispatch accordingly.
// Idempotent across re-runs of `pro module deploy`.
func (s *AppService) UpsertBySlug(ctx context.Context, slug string, req CreateAppRequest) (metamodel.App, error) {
	if _, err := s.Get(ctx, slug); err != nil {
		var sdkErr *sdk.Error
		if errors.As(err, &sdkErr) && sdkErr.HTTPStatus == http.StatusNotFound {
			req.Slug = slug
			return s.Create(ctx, req)
		}
		return metamodel.App{}, err
	}
	name := req.Name
	desc := req.Description
	icon := req.IconSlug
	return s.Update(ctx, slug, UpdateAppRequest{
		Name:        &name,
		Description: &desc,
		IconSlug:    &icon,
	})
}
