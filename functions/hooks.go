package functions

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	sdk "go.proteos.ai/sdk"
	"go.proteos.ai/sdk/internal/httpx"
	functionsmodel "go.proteos.ai/model/functions"
)

const hooksBasePath = "/functions/v1/hooks"

// HookServiceAPI is the contract a HookService satisfies.
type HookServiceAPI interface {
	List(opts *ListHooksOptions) *sdk.PageIterator[functionsmodel.Hook, ListHooksOptions]
	ListPage(ctx context.Context, opts *ListHooksOptions) (sdk.ListResult[functionsmodel.Hook], error)
	Get(ctx context.Context, slug string) (functionsmodel.Hook, error)
	Deploy(ctx context.Context, req DeployHookRequest, wasm io.Reader, filename string) (functionsmodel.Hook, error)
	Update(ctx context.Context, slug string, patch PatchHookRequest) (functionsmodel.Hook, error)
	Activate(ctx context.Context, slug string) (functionsmodel.Hook, error)
	Deactivate(ctx context.Context, slug string) (functionsmodel.Hook, error)
	Delete(ctx context.Context, slug string) error
	Logs(ctx context.Context, slug string, opts *ListHookLogsOptions) ([]functionsmodel.LogEntry, error)
	TailLogs(ctx context.Context, slug string, opts *ListHookLogsOptions) (*LogStream, error)
}

// HookService manages hook lifecycle on function-service.
type HookService struct{ c *sdk.Client }

var _ HookServiceAPI = (*HookService)(nil)

func (s *HookService) List(opts *ListHooksOptions) *sdk.PageIterator[functionsmodel.Hook, ListHooksOptions] {
	o := ListHooksOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListHooksOptions) (sdk.ListResult[functionsmodel.Hook], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *HookService) ListPage(ctx context.Context, opts *ListHooksOptions) (sdk.ListResult[functionsmodel.Hook], error) {
	var out sdk.ListResult[functionsmodel.Hook]
	err := s.c.DoWithQuery(ctx, http.MethodGet, hooksBasePath, opts, nil, &out)
	return out, err
}

func (s *HookService) Get(ctx context.Context, slug string) (functionsmodel.Hook, error) {
	var out functionsmodel.Hook
	err := s.c.Do(ctx, http.MethodGet, hooksBasePath+"/"+slug, nil, &out)
	return out, err
}

// Deploy uploads a hook bundle. Routed as `PUT /api/v1/hooks/:slug` so
// the call is idempotent — re-deploying the same slug replaces the
// in-place row's FileId. `req.Slug` must be set; it's also the URL slug.
func (s *HookService) Deploy(ctx context.Context, req DeployHookRequest, wasm io.Reader, filename string) (functionsmodel.Hook, error) {
	var out functionsmodel.Hook
	metaJSON, err := json.Marshal(req)
	if err != nil {
		return out, err
	}
	err = s.c.DoMultipart(ctx, http.MethodPut, hooksBasePath+"/"+req.Slug,
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

// Update applies a partial PATCH (currently just `{active}` — though the
// dedicated Activate / Deactivate helpers are usually clearer).
func (s *HookService) Update(ctx context.Context, slug string, patch PatchHookRequest) (functionsmodel.Hook, error) {
	var out functionsmodel.Hook
	err := s.c.Do(ctx, http.MethodPatch, hooksBasePath+"/"+slug, patch, &out)
	return out, err
}

func (s *HookService) Activate(ctx context.Context, slug string) (functionsmodel.Hook, error) {
	var out functionsmodel.Hook
	err := s.c.Do(ctx, http.MethodPatch, hooksBasePath+"/"+slug+"/activate", nil, &out)
	return out, err
}

func (s *HookService) Deactivate(ctx context.Context, slug string) (functionsmodel.Hook, error) {
	var out functionsmodel.Hook
	err := s.c.Do(ctx, http.MethodPatch, hooksBasePath+"/"+slug+"/deactivate", nil, &out)
	return out, err
}

func (s *HookService) Delete(ctx context.Context, slug string) error {
	return s.c.Do(ctx, http.MethodDelete, hooksBasePath+"/"+slug, nil, nil)
}

// Logs fetches a snapshot of the hook's recent log lines (the server's
// `{data: [...]}` envelope is unwrapped). For a live stream, use
// `TailLogs` instead — passing `opts.Follow=true` here is rejected so
// callers don't accidentally over-fetch the streaming response into
// memory.
func (s *HookService) Logs(ctx context.Context, slug string, opts *ListHookLogsOptions) ([]functionsmodel.LogEntry, error) {
	if opts != nil && opts.Follow {
		return nil, errFollowOnSnapshot
	}
	path := hooksBasePath + "/" + slug + "/logs"
	if qs := perResourceOptsToQuery(opts, false); qs != "" {
		path += "?" + qs
	}
	return doLogsList(ctx, s.c, path)
}
