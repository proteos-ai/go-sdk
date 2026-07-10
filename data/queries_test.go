package data_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.proteos.ai/sdk/data"
)

func TestQueryService_Execute_PostsSQL(t *testing.T) {
	var seen []byte
	_, d := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/data/v1/query/execute", r.URL.Path)
		seen, _ = io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(data.QueryExecuteResponse{
			Data: []data.QueryRow{{"id": "r1", "name": "Alice"}},
			Meta: &data.QueryExecuteMeta{Columns: []string{"id", "name"}, Items: 1, LimitApplied: 100, ExecutionTimeMs: 12},
		})
	})
	got, err := d.Queries.Execute(context.Background(), "SELECT * FROM customer")
	require.NoError(t, err)
	require.Contains(t, string(seen), `"sql":"SELECT * FROM customer"`)
	require.Equal(t, "Alice", got.Data[0]["name"])
	require.Equal(t, 12, got.Meta.ExecutionTimeMs)
	require.Equal(t, []string{"id", "name"}, got.Meta.Columns)
}

func TestQueryService_Validate_ReturnsRewritten(t *testing.T) {
	var seen []byte
	_, d := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/data/v1/query/validate", r.URL.Path)
		seen, _ = io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(data.QueryValidateResponse{
			Valid:        true,
			RewrittenSQL: "SELECT data->>'name' FROM org_x.customer",
			Tables:       []string{"customer"},
			Meta:         &data.QueryValidateMeta{LimitApplied: 100, WasDefaultLimitApplied: true},
		})
	})
	got, err := d.Queries.Validate(context.Background(), "SELECT name FROM customer")
	require.NoError(t, err)
	require.Contains(t, string(seen), `"sql":"SELECT name FROM customer"`)
	require.True(t, got.Valid)
	require.Equal(t, "SELECT data->>'name' FROM org_x.customer", got.RewrittenSQL)
	require.Equal(t, []string{"customer"}, got.Tables)
	require.True(t, got.Meta.WasDefaultLimitApplied)
}

func TestQueryService_Execute_BadRequest(t *testing.T) {
	_, d := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":"bad_request","message":"syntax error"}`))
	})
	_, err := d.Queries.Execute(context.Background(), "SELECT")
	require.Error(t, err)
	require.Contains(t, err.Error(), "syntax error")
}
