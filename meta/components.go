package meta

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	sdk "go.proteos.ai/sdk"
	"go.proteos.ai/sdk/internal/httpx"
	metamodel "go.proteos.ai/model/meta"
)

const componentsBasePath = "/meta/v1/components"

// ComponentServiceAPI is the contract a ComponentService satisfies.
type ComponentServiceAPI interface {
	List(opts *ListComponentsOptions) *sdk.PageIterator[metamodel.Component, ListComponentsOptions]
	ListPage(ctx context.Context, opts *ListComponentsOptions) (sdk.ListResult[metamodel.Component], error)
	Get(ctx context.Context, slug string) (metamodel.Component, error)
	Create(ctx context.Context, req CreateComponentRequest) (metamodel.Component, error)
	Upsert(ctx context.Context, req CreateComponentRequest) (metamodel.Component, error)
	UpsertBySlug(ctx context.Context, slug string, req CreateComponentRequest) (metamodel.Component, error)
	Deploy(ctx context.Context, req CreateComponentRequest, bundle io.Reader, bundleName string, source io.Reader, sourceName string) (metamodel.Component, error)
	Update(ctx context.Context, slug string, req UpdateComponentRequest) (metamodel.Component, error)
	Delete(ctx context.Context, slug string) error
}

// ComponentService manages UI component definitions.
type ComponentService struct{ c *sdk.Client }

var _ ComponentServiceAPI = (*ComponentService)(nil)

func (s *ComponentService) List(opts *ListComponentsOptions) *sdk.PageIterator[metamodel.Component, ListComponentsOptions] {
	o := ListComponentsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListComponentsOptions) (sdk.ListResult[metamodel.Component], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *ComponentService) ListPage(ctx context.Context, opts *ListComponentsOptions) (sdk.ListResult[metamodel.Component], error) {
	var out sdk.ListResult[metamodel.Component]
	err := s.c.DoWithQuery(ctx, http.MethodGet, componentsBasePath, opts, nil, &out)
	return out, err
}

func (s *ComponentService) Get(ctx context.Context, slug string) (metamodel.Component, error) {
	var out metamodel.Component
	err := s.c.Do(ctx, http.MethodGet, componentsBasePath+"/"+slug, nil, &out)
	return out, err
}

func (s *ComponentService) Create(ctx context.Context, req CreateComponentRequest) (metamodel.Component, error) {
	var out metamodel.Component
	err := s.c.Do(ctx, http.MethodPost, componentsBasePath, req, &out)
	return out, err
}

func (s *ComponentService) Upsert(ctx context.Context, req CreateComponentRequest) (metamodel.Component, error) {
	var out metamodel.Component
	err := s.c.Do(ctx, http.MethodPost, componentsBasePath+"/upsert", req, &out)
	return out, err
}

// UpsertBySlug calls `PUT /meta/v1/components/by-slug/:slug`. The `/by-slug/`
// sub-path avoids a route-wildcard conflict with the existing `:id` GET path.
func (s *ComponentService) UpsertBySlug(ctx context.Context, slug string, req CreateComponentRequest) (metamodel.Component, error) {
	var out metamodel.Component
	req.Slug = slug
	err := s.c.Do(ctx, http.MethodPut, componentsBasePath+"/by-slug/"+slug, req, &out)
	return out, err
}

// Deploy uploads a component's compiled bundle + tar.gz'd source to
// metadata-service in one multipart request (JSON `metadata` part, then
// `bundle`, then `source`). metadata-service stores both in storage-service
// and stamps the resulting file ids onto the component — the CLI never talks
// to storage-service directly, mirroring how hooks ship through
// function-service. Idempotent: the server upserts by slug. The req's
// BundleFileId / SourceFileId are ignored (the server fills them); set Slug,
// ModuleSlug, Name, Description, PropsSchema.
func (s *ComponentService) Deploy(ctx context.Context, req CreateComponentRequest, bundle io.Reader, bundleName string, source io.Reader, sourceName string) (metamodel.Component, error) {
	var out metamodel.Component
	metaJSON, err := json.Marshal(req)
	if err != nil {
		return out, err
	}
	err = s.c.DoMultipartJSON(ctx, http.MethodPost, componentsBasePath+"/deploy", "metadata", string(metaJSON),
		[]httpx.MultipartFile{
			{FieldName: "bundle", Filename: bundleName, ContentType: "text/javascript", Reader: bundle},
			{FieldName: "source", Filename: sourceName, ContentType: "application/gzip", Reader: source},
		},
		&out,
	)
	return out, err
}

func (s *ComponentService) Update(ctx context.Context, slug string, req UpdateComponentRequest) (metamodel.Component, error) {
	var out metamodel.Component
	err := s.c.Do(ctx, http.MethodPatch, componentsBasePath+"/"+slug, req, &out)
	return out, err
}

func (s *ComponentService) Delete(ctx context.Context, slug string) error {
	return s.c.Do(ctx, http.MethodDelete, componentsBasePath+"/"+slug, nil, nil)
}
