package workflow

import (
	"context"
	"net/http"

	workflowmodel "go.proteos.ai/model/workflow"
	workflowapi "go.proteos.ai/model/workflow/api"
	sdk "go.proteos.ai/sdk"
)

const workflowsBasePath = "/workflows/v1/workflows"

// WorkflowService manages workflow definitions (automation graphs).
type WorkflowService struct{ c *sdk.Client }

func (s *WorkflowService) List(opts *ListWorkflowsOptions) *sdk.PageIterator[workflowmodel.Workflow, ListWorkflowsOptions] {
	o := ListWorkflowsOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListWorkflowsOptions) (sdk.ListResult[workflowmodel.Workflow], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

func (s *WorkflowService) ListPage(ctx context.Context, opts *ListWorkflowsOptions) (sdk.ListResult[workflowmodel.Workflow], error) {
	var out sdk.ListResult[workflowmodel.Workflow]
	err := s.c.DoWithQuery(ctx, http.MethodGet, workflowsBasePath, opts, nil, &out)
	return out, err
}

func (s *WorkflowService) Get(ctx context.Context, key string) (workflowmodel.Workflow, error) {
	var out workflowmodel.Workflow
	err := s.c.Do(ctx, http.MethodGet, workflowsBasePath+"/"+key, nil, &out)
	return out, err
}

func (s *WorkflowService) Create(ctx context.Context, req workflowapi.CreateWorkflowRequest) (workflowmodel.Workflow, error) {
	var out workflowmodel.Workflow
	err := s.c.Do(ctx, http.MethodPost, workflowsBasePath, req, &out)
	return out, err
}

func (s *WorkflowService) Update(ctx context.Context, key string, req workflowapi.UpdateWorkflowRequest) (workflowmodel.Workflow, error) {
	var out workflowmodel.Workflow
	err := s.c.Do(ctx, http.MethodPatch, workflowsBasePath+"/"+key, req, &out)
	return out, err
}

// UpsertByKey calls `PUT /workflows/v1/workflows/:key` — the idempotent
// create-or-update path used by `pro module deploy`. An equivalent body is a
// server-side no-op (no version bump); redeploys preserve runtime status and
// carry existing webhook tokens forward.
func (s *WorkflowService) UpsertByKey(ctx context.Context, key string, req workflowapi.CreateWorkflowRequest) (workflowmodel.Workflow, error) {
	var out workflowmodel.Workflow
	req.Key = key
	err := s.c.Do(ctx, http.MethodPut, workflowsBasePath+"/"+key, req, &out)
	return out, err
}

func (s *WorkflowService) Delete(ctx context.Context, key string) error {
	return s.c.Do(ctx, http.MethodDelete, workflowsBasePath+"/"+key, nil, nil)
}

// Pause suppresses scheduled/triggered firings; the definition stays intact.
func (s *WorkflowService) Pause(ctx context.Context, key string) (workflowmodel.Workflow, error) {
	var out workflowmodel.Workflow
	err := s.c.Do(ctx, http.MethodPost, workflowsBasePath+"/"+key+"/pause", nil, &out)
	return out, err
}

// Unpause re-activates a paused workflow.
func (s *WorkflowService) Unpause(ctx context.Context, key string) (workflowmodel.Workflow, error) {
	var out workflowmodel.Workflow
	err := s.c.Do(ctx, http.MethodPost, workflowsBasePath+"/"+key+"/unpause", nil, &out)
	return out, err
}
