// Package connectors provides services for connector-service resources:
// the connector manifest catalog and connections (install flow, token
// release, method invocation).
//
// Usage:
//
//	cn := connectors.New(c)
//	catalog, err := cn.Connectors.ListPage(ctx, nil)
//	result, err := cn.Connections.InvokeMethod(ctx, connectionId, "list_events", params)
package connectors

import (
	sdk "go.proteos.ai/sdk"
)

// Client groups the connector-service services.
type Client struct {
	Connectors  *ConnectorService
	Connections *ConnectionService
}

// New builds a Client backed by the given *sdk.Client.
func New(c *sdk.Client) *Client {
	return &Client{
		Connectors:  &ConnectorService{c: c},
		Connections: &ConnectionService{c: c},
	}
}
