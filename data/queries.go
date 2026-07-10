package data

import (
	"context"
	"net/http"

	sdk "go.proteos.ai/sdk"
)

const queryBasePath = "/data/v1/query"

// QueryServiceAPI is the contract a QueryService satisfies.
type QueryServiceAPI interface {
	Execute(ctx context.Context, sql string) (QueryExecuteResponse, error)
	Validate(ctx context.Context, sql string) (QueryValidateResponse, error)
}

// QueryService is the raw-SQL query endpoint of the data-service. The
// server rewrites bare attribute references into JSONB accessors and
// authorizes per-table.
type QueryService struct{ c *sdk.Client }

var _ QueryServiceAPI = (*QueryService)(nil)

type sqlBody struct {
	SQL string `json:"sql"`
}

// Execute runs the given SQL and returns the result rows + meta.
func (s *QueryService) Execute(ctx context.Context, sql string) (QueryExecuteResponse, error) {
	var out QueryExecuteResponse
	err := s.c.Do(ctx, http.MethodPost, queryBasePath+"/execute", sqlBody{SQL: sql}, &out)
	return out, err
}

// Validate parses + authorizes the SQL without running it.
func (s *QueryService) Validate(ctx context.Context, sql string) (QueryValidateResponse, error) {
	var out QueryValidateResponse
	err := s.c.Do(ctx, http.MethodPost, queryBasePath+"/validate", sqlBody{SQL: sql}, &out)
	return out, err
}
