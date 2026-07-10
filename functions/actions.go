package functions

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	sdk "go.proteos.ai/sdk"
	"go.proteos.ai/sdk/internal/httpx"
	functionsapi "go.proteos.ai/model/functions/api"
	functionsmodel "go.proteos.ai/model/functions"
)

const (
	actionsBasePath  = "/functions/v1/actions"
	entitiesBasePath = "/functions/v1/entities"
)

// ActionServiceAPI is the contract an ActionService satisfies.
type ActionServiceAPI interface {
	List(opts *ListActionsOptions) *sdk.PageIterator[functionsmodel.Action, ListActionsOptions]
	ListPage(ctx context.Context, opts *ListActionsOptions) (sdk.ListResult[functionsmodel.Action], error)
	Get(ctx context.Context, slug string) (functionsmodel.Action, error)
	Deploy(ctx context.Context, req DeployActionRequest, wasm io.Reader, filename string) (functionsmodel.Action, error)
	Update(ctx context.Context, slug string, patch PatchActionRequest) (functionsmodel.Action, error)
	Activate(ctx context.Context, slug string) (functionsmodel.Action, error)
	Deactivate(ctx context.Context, slug string) (functionsmodel.Action, error)
	Delete(ctx context.Context, slug string) error
	ListForEntity(ctx context.Context, entity string) ([]functionsapi.ActionSummary, error)
	Logs(ctx context.Context, slug string, opts *ListActionLogsOptions) ([]functionsmodel.LogEntry, error)
	TailLogs(ctx context.Context, slug string, opts *ListActionLogsOptions) (*LogStream, error)
}

// ActionService manages action lifecycle on function-service.
type ActionService struct{ c *sdk.Client }

var _ ActionServiceAPI = (*ActionService)(nil)

func (s *ActionService) List(opts *ListActionsOptions) *sdk.PageIterator[functionsmodel.Action, ListActionsOptions] {
	o := ListActionsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListActionsOptions) (sdk.ListResult[functionsmodel.Action], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *ActionService) ListPage(ctx context.Context, opts *ListActionsOptions) (sdk.ListResult[functionsmodel.Action], error) {
	var out sdk.ListResult[functionsmodel.Action]
	err := s.c.DoWithQuery(ctx, http.MethodGet, actionsBasePath, opts, nil, &out)
	return out, err
}

func (s *ActionService) Get(ctx context.Context, slug string) (functionsmodel.Action, error) {
	var out functionsmodel.Action
	err := s.c.Do(ctx, http.MethodGet, actionsBasePath+"/"+slug, nil, &out)
	return out, err
}

// Deploy uploads an action bundle as a multipart PUT, upserting by slug.
// `req.Slug` is also the URL slug.
func (s *ActionService) Deploy(ctx context.Context, req DeployActionRequest, wasm io.Reader, filename string) (functionsmodel.Action, error) {
	var out functionsmodel.Action
	metaJSON, err := json.Marshal(req)
	if err != nil {
		return out, err
	}
	err = s.c.DoMultipart(ctx, http.MethodPut, actionsBasePath+"/"+req.Slug,
		map[string]string{"metadata": string(metaJSON)},
		httpx.MultipartFile{
			FieldName:   "wasm",
			Filename:    filename,
			ContentType: "application/wasm",
			Reader:      wasm,
		},
		&out,
	)
	return out, err
}

func (s *ActionService) Update(ctx context.Context, slug string, patch PatchActionRequest) (functionsmodel.Action, error) {
	var out functionsmodel.Action
	err := s.c.Do(ctx, http.MethodPatch, actionsBasePath+"/"+slug, patch, &out)
	return out, err
}

func (s *ActionService) Activate(ctx context.Context, slug string) (functionsmodel.Action, error) {
	var out functionsmodel.Action
	err := s.c.Do(ctx, http.MethodPatch, actionsBasePath+"/"+slug+"/activate", nil, &out)
	return out, err
}

func (s *ActionService) Deactivate(ctx context.Context, slug string) (functionsmodel.Action, error) {
	var out functionsmodel.Action
	err := s.c.Do(ctx, http.MethodPatch, actionsBasePath+"/"+slug+"/deactivate", nil, &out)
	return out, err
}

func (s *ActionService) Delete(ctx context.Context, slug string) error {
	return s.c.Do(ctx, http.MethodDelete, actionsBasePath+"/"+slug, nil, nil)
}

// ListForEntity returns the entity-scoped + global actions visible for
// the given entity. Backs the UI's action picker.
func (s *ActionService) ListForEntity(ctx context.Context, entity string) ([]functionsapi.ActionSummary, error) {
	var out []functionsapi.ActionSummary
	err := s.c.Do(ctx, http.MethodGet, entitiesBasePath+"/"+entity+"/actions", nil, &out)
	return out, err
}
