package meta_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.proteos.ai/sdk/meta"
	metamodel "go.proteos.ai/model/meta"
)

func sampleVariable() metamodel.Variable {
	return metamodel.Variable{Id: "v1", Key: "API_KEY", Value: "secret", IsSecret: true, Module: "billing", CreatedBy: validSource(), UpdatedBy: validSource()}
}

func TestVariableService_Get(t *testing.T) {
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/meta/v1/variables/v1", r.URL.Path)
		_ = json.NewEncoder(w).Encode(sampleVariable())
	})
	got, err := m.Variables.Get(context.Background(), "v1")
	require.NoError(t, err)
	require.Equal(t, "API_KEY", got.Key)
	require.True(t, got.IsSecret)
}

func TestVariableService_ListPage_FlagsBoolean(t *testing.T) {
	var seen string
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.String()
		_, _ = w.Write([]byte(`{"meta":{"page":0,"page_size":10,"items_total":0,"pages_total":0},"data":[]}`))
	})
	yes := true
	_, err := m.Variables.ListPage(context.Background(), &meta.ListVariablesOptions{
		ListOptions: meta.ListOptions{PageSize: 10},
		Module:      "billing",
		IsSecret:    &yes,
	})
	require.NoError(t, err)
	require.Contains(t, seen, "module=billing")
	require.Contains(t, seen, "is_secret=true")
}

func TestVariableService_Create(t *testing.T) {
	var body []byte
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/meta/v1/variables", r.URL.Path)
		body, _ = io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(sampleVariable())
	})
	_, err := m.Variables.Create(context.Background(), meta.CreateVariableRequest{
		Key: "API_KEY", Value: "secret", IsSecret: true, Module: "billing",
	})
	require.NoError(t, err)
	require.Contains(t, string(body), `"is_secret":true`)
}

func TestVariableService_Update_OnlyValue(t *testing.T) {
	var body []byte
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		require.Equal(t, "/meta/v1/variables/v1", r.URL.Path)
		body, _ = io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(sampleVariable())
	})
	v := "rotated"
	_, err := m.Variables.Update(context.Background(), "v1", meta.UpdateVariableRequest{Value: &v})
	require.NoError(t, err)
	require.Contains(t, string(body), `"value":"rotated"`)
}

func TestVariableService_Delete(t *testing.T) {
	_, m := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, m.Variables.Delete(context.Background(), "v1"))
}
