// Package functions provides services for managing function-service
// resources (hooks + actions) and dispatching them.
//
// Resource shapes (Hook, Action, FunctionSource, HookEvent, ActionScope,
// LogEntry, …) are imported from go.proteos.ai/model/functions.
// Wire-format request types live in go.proteos.ai/model/functions/api;
// they're re-exposed via this package's typed methods. Only the
// query-options types are defined locally.
//
// Usage:
//
//	f := functions.New(c)
//	hook, err := f.Hooks.Deploy(ctx, req, wasmReader, "validate-invoice.wasm")
//	out, err  := f.Dispatch.OnBeforeCreate(ctx, "invoice", recordJSON)
package functions

import (
	"errors"

	sdk "go.proteos.ai/sdk"
)

// Client groups the function-service services. Construct with New, then
// access via the public fields:
//
//	f := functions.New(c)
//	hook, err := f.Hooks.Get(ctx, "validate-invoice")
type Client struct {
	Hooks    *HookService
	Actions  *ActionService
	Dispatch *DispatchService
	Logs     *LogService
}

// New builds a Client backed by the given *sdk.Client.
func New(c *sdk.Client) *Client {
	return &Client{
		Hooks:    &HookService{c: c},
		Actions:  &ActionService{c: c},
		Dispatch: &DispatchService{c: c},
		Logs:     &LogService{c: c},
	}
}

// errFollowOnSnapshot guards callers from accidentally invoking the
// snapshot Logs method with Follow=true — that path would over-fetch
// the streaming response into memory before returning. Use TailLogs for
// follow.
var errFollowOnSnapshot = errors.New("functions: pass Follow=true to TailLogs; Logs returns a finite snapshot")
