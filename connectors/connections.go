package connectors

import (
	"context"
	"encoding/json"
	"net/http"

	connectormodel "go.proteos.ai/model/connector"
	connectorapi "go.proteos.ai/model/connector/api"
	sdk "go.proteos.ai/sdk"
)

const connectionsBasePath = "/connectors/v1/connections"

// ListConnectionsOptions filters an org's connections.
type ListConnectionsOptions struct {
	ListOptions
	ConnectorKey string `query:"connector_key,omitempty"`
	Scope        string `query:"scope,omitempty"`
	Status       string `query:"status,omitempty"`
}

// ConnectionServiceAPI is the contract a ConnectionService satisfies.
type ConnectionServiceAPI interface {
	ListPage(ctx context.Context, opts *ListConnectionsOptions) (sdk.ListResult[connectormodel.Connection], error)
	Get(ctx context.Context, id string) (connectormodel.Connection, error)
	Create(ctx context.Context, req connectorapi.CreateConnectionRequest) (connectormodel.Connection, error)
	Update(ctx context.Context, id string, req connectorapi.UpdateConnectionRequest) (connectormodel.Connection, error)
	Delete(ctx context.Context, id string) error
	Install(ctx context.Context, id string) (connectorapi.InstallConnectionResponse, error)
	Token(ctx context.Context, id string) (connectorapi.ConnectionTokenResponse, error)
	InvokeMethod(ctx context.Context, id string, method string, params json.RawMessage) (json.RawMessage, error)
}

// ConnectionService manages connections and fronts method invocation.
type ConnectionService struct{ c *sdk.Client }

var _ ConnectionServiceAPI = (*ConnectionService)(nil)

func (s *ConnectionService) ListPage(ctx context.Context, opts *ListConnectionsOptions) (sdk.ListResult[connectormodel.Connection], error) {
	var out sdk.ListResult[connectormodel.Connection]
	err := s.c.DoWithQuery(ctx, http.MethodGet, connectionsBasePath, opts, nil, &out)
	return out, err
}

func (s *ConnectionService) Get(ctx context.Context, id string) (connectormodel.Connection, error) {
	var out connectormodel.Connection
	err := s.c.Do(ctx, http.MethodGet, connectionsBasePath+"/"+id, nil, &out)
	return out, err
}

func (s *ConnectionService) Create(ctx context.Context, req connectorapi.CreateConnectionRequest) (connectormodel.Connection, error) {
	var out connectormodel.Connection
	err := s.c.Do(ctx, http.MethodPost, connectionsBasePath, req, &out)
	return out, err
}

func (s *ConnectionService) Update(ctx context.Context, id string, req connectorapi.UpdateConnectionRequest) (connectormodel.Connection, error) {
	var out connectormodel.Connection
	err := s.c.Do(ctx, http.MethodPatch, connectionsBasePath+"/"+id, req, &out)
	return out, err
}

func (s *ConnectionService) Delete(ctx context.Context, id string) error {
	return s.c.Do(ctx, http.MethodDelete, connectionsBasePath+"/"+id, nil, nil)
}

// Install starts the OAuth flow: the returned authorization URL opens in a
// browser popup; completion lands on the broker's per-environment callback.
func (s *ConnectionService) Install(ctx context.Context, id string) (connectorapi.InstallConnectionResponse, error) {
	var out connectorapi.InstallConnectionResponse
	err := s.c.Do(ctx, http.MethodPost, connectionsBasePath+"/"+id+"/install", nil, &out)
	return out, err
}

// Token releases the connection's usable credential material (never
// refresh_token / client_secret). Requires connections write permission.
func (s *ConnectionService) Token(ctx context.Context, id string) (connectorapi.ConnectionTokenResponse, error) {
	var out connectorapi.ConnectionTokenResponse
	err := s.c.Do(ctx, http.MethodPost, connectionsBasePath+"/"+id+"/token", nil, &out)
	return out, err
}

// InvokeMethod runs a connector method against a connection; the body is
// the raw params JSON, the result is the unwrapped `{result}` payload.
func (s *ConnectionService) InvokeMethod(ctx context.Context, id string, method string, params json.RawMessage) (json.RawMessage, error) {
	if len(params) == 0 {
		params = json.RawMessage("{}")
	}
	var out struct {
		Result json.RawMessage `json:"result"`
	}
	err := s.c.Do(ctx, http.MethodPost, connectionsBasePath+"/"+id+"/methods/"+method+"/invoke", params, &out)
	return out.Result, err
}
