package functions

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	functionsapi "go.proteos.ai/model/functions/api"
	sdk "go.proteos.ai/sdk"
	"go.proteos.ai/sdk/internal/httpx"
)

const functionBundlesBasePath = "/functions/v1/function-bundles"

// FunctionBundleServiceAPI is the contract a FunctionBundleService satisfies.
type FunctionBundleServiceAPI interface {
	Deploy(ctx context.Context, req DeployFunctionBundleRequest, wasm io.Reader, filename string) (functionsapi.DeployFunctionBundleResponse, error)
}

// FunctionBundleService deploys one wasm serving every hook, action and
// connector method of a module.
type FunctionBundleService struct{ c *sdk.Client }

var _ FunctionBundleServiceAPI = (*FunctionBundleService)(nil)

// Deploy uploads a function bundle. Routed as
// `PUT /functions/v1/function-bundles/:module_slug` so the call is idempotent —
// re-deploying the module replaces every listed row's blob in place.
// `req.ModuleSlug` must be set; it is also the URL slug.
func (s *FunctionBundleService) Deploy(ctx context.Context, req DeployFunctionBundleRequest, wasm io.Reader, filename string) (functionsapi.DeployFunctionBundleResponse, error) {
	var out struct {
		Data functionsapi.DeployFunctionBundleResponse `json:"data"`
	}
	metaJSON, err := json.Marshal(req)
	if err != nil {
		return out.Data, err
	}
	err = s.c.DoMultipart(ctx, http.MethodPut, functionBundlesBasePath+"/"+req.ModuleSlug,
		map[string]string{"metadata": string(metaJSON)},
		httpx.MultipartFile{
			FieldName:   "wasm",
			Filename:    filename,
			ContentType: "application/wasm",
			Reader:      wasm,
		},
		&out,
	)
	return out.Data, err
}
