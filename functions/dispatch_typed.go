package functions

import (
	"context"
	"encoding/json"
)

// Typed action invocation.
//
// These are package-level generic functions, not methods on DispatchService,
// because Go forbids generic methods. They wrap the untyped
// DispatchService.InvokeGlobalAction / InvokeEntityAction primitives: the
// params value P is marshaled as the request body (sent verbatim — the invoke
// endpoint takes the params object with no envelope), and the unwrapped
// `{result}` bytes are decoded into R.
//
// The pairing is intentional: dynamic callers (e.g. workflow-node-host, which
// invokes actions from runtime config with map[string]any) keep the raw
// json.RawMessage methods, while callers that know an action's shape at compile
// time — typically via structs generated from the deployed schema by
// `pro functions actions gen-types` — get end-to-end typing here.
//
// There is deliberately no per-call existence check: a missing slug still
// surfaces as the underlying 404 error, and the codegen path is the
// authoritative "action exists" guarantee (it fails at generation time). This
// keeps invocation a single round-trip.

// marshalParams encodes typed params into the request body. A nil pointer (or
// otherwise null-marshaling) P becomes `{}` rather than `null`: the invoke
// endpoint only normalizes an *empty* body to `{}`, so a literal `null` would
// reach the guest and break decoding of an otherwise parameterless call.
func marshalParams[P any](params P) (json.RawMessage, error) {
	raw, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return json.RawMessage("{}"), nil
	}
	return raw, nil
}

// InvokeGlobalActionTyped invokes a global action with typed params and result.
// P is marshaled as the request body; the unwrapped result is decoded into R.
func InvokeGlobalActionTyped[P any, R any](ctx context.Context, d *DispatchService, slug string, params P) (R, error) {
	var out R
	raw, err := marshalParams(params)
	if err != nil {
		return out, err
	}
	result, err := d.InvokeGlobalAction(ctx, slug, raw)
	if err != nil {
		return out, err
	}
	if len(result) == 0 {
		return out, nil
	}
	if err := json.Unmarshal(result, &out); err != nil {
		return out, err
	}
	return out, nil
}

// InvokeEntityActionTyped invokes an entity-scoped action against one record
// with typed params and result.
func InvokeEntityActionTyped[P any, R any](ctx context.Context, d *DispatchService, entity, recordId, slug string, params P) (R, error) {
	var out R
	raw, err := marshalParams(params)
	if err != nil {
		return out, err
	}
	result, err := d.InvokeEntityAction(ctx, entity, recordId, slug, raw)
	if err != nil {
		return out, err
	}
	if len(result) == 0 {
		return out, nil
	}
	if err := json.Unmarshal(result, &out); err != nil {
		return out, err
	}
	return out, nil
}
