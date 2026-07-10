package meta

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	metamodel "go.proteos.ai/model/meta"
	sdk "go.proteos.ai/sdk"
	"go.proteos.ai/sdk/internal/httpx"
)

const modulesBasePath = "/meta/v1/modules"

// ModuleServiceAPI is the contract a ModuleService satisfies.
type ModuleServiceAPI interface {
	List(opts *ListModulesOptions) *sdk.PageIterator[metamodel.Module, ListModulesOptions]
	ListPage(ctx context.Context, opts *ListModulesOptions) (sdk.ListResult[metamodel.Module], error)
	Get(ctx context.Context, slug string) (metamodel.Module, error)
	Deploy(ctx context.Context, req DeployModuleRequest, file io.Reader, filename string) (metamodel.Module, error)
	UpsertBySlug(ctx context.Context, slug string, req DeployModuleRequest) (metamodel.Module, error)
	Activate(ctx context.Context, slug string) error
	Deactivate(ctx context.Context, slug string) error
	Delete(ctx context.Context, slug string) error
	Download(ctx context.Context, slug string) (io.ReadCloser, metamodel.Module, error)
}

// ModuleService manages module deployments.
type ModuleService struct{ c *sdk.Client }

var _ ModuleServiceAPI = (*ModuleService)(nil)

func (s *ModuleService) List(opts *ListModulesOptions) *sdk.PageIterator[metamodel.Module, ListModulesOptions] {
	o := ListModulesOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListModulesOptions) (sdk.ListResult[metamodel.Module], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *ModuleService) ListPage(ctx context.Context, opts *ListModulesOptions) (sdk.ListResult[metamodel.Module], error) {
	var out sdk.ListResult[metamodel.Module]
	err := s.c.DoWithQuery(ctx, http.MethodGet, modulesBasePath, opts, nil, &out)
	return out, err
}

func (s *ModuleService) Get(ctx context.Context, slug string) (metamodel.Module, error) {
	var out metamodel.Module
	err := s.c.Do(ctx, http.MethodGet, modulesBasePath+"/"+slug, nil, &out)
	return out, err
}

// Deploy uploads a module bundle and returns the resulting Module record.
// req describes the module; file streams the bundle bytes (e.g. a .wasm file).
func (s *ModuleService) Deploy(ctx context.Context, req DeployModuleRequest, file io.Reader, filename string) (metamodel.Module, error) {
	var out metamodel.Module
	metaJSON, err := json.Marshal(req)
	if err != nil {
		return out, err
	}
	err = s.c.DoMultipart(ctx, http.MethodPost, modulesBasePath+"/deploy",
		map[string]string{"metadata": string(metaJSON)},
		httpx.MultipartFile{
			FieldName:   "file",
			Filename:    filename,
			ContentType: "application/octet-stream",
			Reader:      file,
		},
		&out,
	)
	return out, err
}

// UpsertBySlug calls `PUT /meta/v1/modules/:slug` — the metadata-only
// upsert path used by `pro module deploy` (no tarball; wasm artifacts ship
// separately to function-service).
func (s *ModuleService) UpsertBySlug(ctx context.Context, slug string, req DeployModuleRequest) (metamodel.Module, error) {
	var out metamodel.Module
	req.Slug = slug
	err := s.c.Do(ctx, http.MethodPut, modulesBasePath+"/"+slug, req, &out)
	return out, err
}

func (s *ModuleService) Activate(ctx context.Context, slug string) error {
	return s.c.Do(ctx, http.MethodPatch, modulesBasePath+"/"+slug+"/activate", nil, nil)
}

func (s *ModuleService) Deactivate(ctx context.Context, slug string) error {
	return s.c.Do(ctx, http.MethodPatch, modulesBasePath+"/"+slug+"/deactivate", nil, nil)
}

func (s *ModuleService) Delete(ctx context.Context, slug string) error {
	return s.c.Do(ctx, http.MethodDelete, modulesBasePath+"/"+slug, nil, nil)
}

// Download streams the deployed module bundle for the given slug, returning
// the response body and the Module metadata. The caller MUST Close the body.
func (s *ModuleService) Download(ctx context.Context, slug string) (io.ReadCloser, metamodel.Module, error) {
	mod, err := s.Get(ctx, slug)
	if err != nil {
		return nil, metamodel.Module{}, err
	}
	body, _, err := s.c.DoRaw(ctx, http.MethodGet, modulesBasePath+"/"+slug+"/download", nil)
	if err != nil {
		return nil, metamodel.Module{}, err
	}
	return body, mod, nil
}
