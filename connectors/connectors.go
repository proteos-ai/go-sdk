package connectors

import (
	"context"
	"net/http"

	connectormodel "go.proteos.ai/model/connector"
	connectorapi "go.proteos.ai/model/connector/api"
	sdk "go.proteos.ai/sdk"
)

const connectorsBasePath = "/connectors/v1/connectors"

// ListConnectorsOptions filters the manifest catalog.
type ListConnectorsOptions struct {
	ListOptions
	Status string `query:"status,omitempty"`
}

// ListOptions mirrors the shared pagination/sorting query params.
type ListOptions struct {
	Page     int    `query:"page,omitempty"`
	PageSize int    `query:"page_size,omitempty"`
	SortBy   string `query:"sort_by,omitempty"`
	Sort     string `query:"sort,omitempty"`
}

// ConnectorServiceAPI is the contract a ConnectorService satisfies.
type ConnectorServiceAPI interface {
	ListPage(ctx context.Context, opts *ListConnectorsOptions) (sdk.ListResult[connectormodel.ConnectorManifest], error)
	Get(ctx context.Context, key string) (connectormodel.ConnectorManifest, error)
	Upsert(ctx context.Context, key string, req connectorapi.UpsertConnectorRequest) (connectormodel.ConnectorManifest, error)
	Delete(ctx context.Context, key string) error
}

// ConnectorService reads the manifest catalog and (for module deploy)
// upserts custom manifests.
type ConnectorService struct{ c *sdk.Client }

var _ ConnectorServiceAPI = (*ConnectorService)(nil)

func (s *ConnectorService) ListPage(ctx context.Context, opts *ListConnectorsOptions) (sdk.ListResult[connectormodel.ConnectorManifest], error) {
	var out sdk.ListResult[connectormodel.ConnectorManifest]
	err := s.c.DoWithQuery(ctx, http.MethodGet, connectorsBasePath, opts, nil, &out)
	return out, err
}

func (s *ConnectorService) Get(ctx context.Context, key string) (connectormodel.ConnectorManifest, error) {
	var out connectormodel.ConnectorManifest
	err := s.c.Do(ctx, http.MethodGet, connectorsBasePath+"/"+key, nil, &out)
	return out, err
}

// Upsert creates-or-replaces a CUSTOM manifest by key (the module-deploy
// target; pre-built keys are rejected server-side).
func (s *ConnectorService) Upsert(ctx context.Context, key string, req connectorapi.UpsertConnectorRequest) (connectormodel.ConnectorManifest, error) {
	var out connectormodel.ConnectorManifest
	err := s.c.Do(ctx, http.MethodPut, connectorsBasePath+"/"+key, req, &out)
	return out, err
}

func (s *ConnectorService) Delete(ctx context.Context, key string) error {
	return s.c.Do(ctx, http.MethodDelete, connectorsBasePath+"/"+key, nil, nil)
}
