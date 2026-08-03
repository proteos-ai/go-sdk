package workflow

import sdk "go.proteos.ai/sdk"

// Client groups the workflow-service resource services. Construct with New,
// then access them via the public fields:
//
//	w := workflow.New(c)
//	wf, err := w.Workflows.Get(ctx, "lead-intake")
type Client struct {
	Workflows *WorkflowService
}

// New builds a Client backed by the given *sdk.Client.
func New(c *sdk.Client) *Client {
	return &Client{
		Workflows: &WorkflowService{c: c},
	}
}
