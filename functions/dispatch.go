package functions

import (
	"context"
	"encoding/json"
	"net/http"

	sdk "go.proteos.ai/sdk"
	functionsapi "go.proteos.ai/model/functions/api"
)

// DispatchServiceAPI invokes hook and action dispatch endpoints. Used
// server-to-server: data-service calls hook dispatch on every record
// CUD; the UI / external clients call action invocation.
//
// Hook before-* methods return the (possibly mutated) record; after-*
// methods return only an error. before-* fail-closed (data-service
// aborts the write on error); after-* fail-open per the LUM-56 contract.
//
// `record`, `currentRecord`, and `previousRecord` JSON keys are exact —
// they're the wire shape the function-service dispatcher decodes.
type DispatchServiceAPI interface {
	OnBeforeCreate(ctx context.Context, entity string, record json.RawMessage) (json.RawMessage, error)
	OnBeforeUpdate(ctx context.Context, entity string, record, currentRecord json.RawMessage) (json.RawMessage, error)
	OnBeforeDelete(ctx context.Context, entity string, record json.RawMessage) error
	OnAfterCreate(ctx context.Context, entity string, record json.RawMessage) error
	OnAfterUpdate(ctx context.Context, entity string, record, previousRecord json.RawMessage) error
	OnAfterDelete(ctx context.Context, entity string, record json.RawMessage) error

	InvokeEntityAction(ctx context.Context, entity, recordId, slug string, params json.RawMessage) (json.RawMessage, error)
	InvokeGlobalAction(ctx context.Context, slug string, params json.RawMessage) (json.RawMessage, error)
}

// DispatchService is the SDK client for function-service's dispatch
// endpoints. Auth + tracing follow the SDK's standard header rules
// (`Authorization: Bearer …`, `X-Trace-Id` / `traceparent` propagation).
type DispatchService struct{ c *sdk.Client }

var _ DispatchServiceAPI = (*DispatchService)(nil)

func hookDispatchPath(entity, event string) string {
	return "/functions/v1/entities/" + entity + "/hooks/" + event
}

func (s *DispatchService) OnBeforeCreate(ctx context.Context, entity string, record json.RawMessage) (json.RawMessage, error) {
	var out functionsapi.OnBeforeResponse
	err := s.c.Do(ctx, http.MethodPost, hookDispatchPath(entity, "on-before-create"),
		functionsapi.OnBeforeCreateRequest{Record: record}, &out)
	return out.Record, err
}

func (s *DispatchService) OnBeforeUpdate(ctx context.Context, entity string, record, currentRecord json.RawMessage) (json.RawMessage, error) {
	var out functionsapi.OnBeforeResponse
	err := s.c.Do(ctx, http.MethodPost, hookDispatchPath(entity, "on-before-update"),
		functionsapi.OnBeforeUpdateRequest{Record: record, CurrentRecord: currentRecord}, &out)
	return out.Record, err
}

func (s *DispatchService) OnBeforeDelete(ctx context.Context, entity string, record json.RawMessage) error {
	return s.c.Do(ctx, http.MethodPost, hookDispatchPath(entity, "on-before-delete"),
		functionsapi.OnBeforeDeleteRequest{Record: record}, nil)
}

func (s *DispatchService) OnAfterCreate(ctx context.Context, entity string, record json.RawMessage) error {
	return s.c.Do(ctx, http.MethodPost, hookDispatchPath(entity, "on-after-create"),
		functionsapi.OnAfterCreateRequest{Record: record}, nil)
}

func (s *DispatchService) OnAfterUpdate(ctx context.Context, entity string, record, previousRecord json.RawMessage) error {
	return s.c.Do(ctx, http.MethodPost, hookDispatchPath(entity, "on-after-update"),
		functionsapi.OnAfterUpdateRequest{Record: record, PreviousRecord: previousRecord}, nil)
}

func (s *DispatchService) OnAfterDelete(ctx context.Context, entity string, record json.RawMessage) error {
	return s.c.Do(ctx, http.MethodPost, hookDispatchPath(entity, "on-after-delete"),
		functionsapi.OnAfterDeleteRequest{Record: record}, nil)
}

// InvokeEntityAction calls
// `POST /functions/v1/entities/:entity/records/:recordId/actions/:slug/invoke`
// via the gateway.
// `params` is sent verbatim as the request body; the action's `result`
// is unwrapped from the `{result: …}` envelope before returning.
func (s *DispatchService) InvokeEntityAction(ctx context.Context, entity, recordId, slug string, params json.RawMessage) (json.RawMessage, error) {
	var out functionsapi.InvokeActionResponse
	path := "/functions/v1/entities/" + entity + "/records/" + recordId + "/actions/" + slug + "/invoke"
	err := s.c.Do(ctx, http.MethodPost, path, params, &out)
	return out.Result, err
}

// InvokeGlobalAction calls `POST /functions/v1/actions/:slug/invoke` via
// the gateway (traefik rewrites to `/api/v1/actions/:slug/invoke` on
// function-service).
func (s *DispatchService) InvokeGlobalAction(ctx context.Context, slug string, params json.RawMessage) (json.RawMessage, error) {
	var out functionsapi.InvokeActionResponse
	err := s.c.Do(ctx, http.MethodPost, "/functions/v1/actions/"+slug+"/invoke", params, &out)
	return out.Result, err
}
