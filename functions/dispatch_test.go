package functions_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// readBody decodes the request body into a generic map. Tests assert
// against the wire-format field names — they must match the dispatch
// envelope exactly (record / currentRecord / previousRecord).
func readBody(t *testing.T, r *http.Request) map[string]json.RawMessage {
	t.Helper()
	raw, err := io.ReadAll(r.Body)
	require.NoError(t, err)
	var body map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &body))
	return body
}

func TestDispatch_OnBeforeCreate_HitsEntityPathAndReturnsMutatedRecord(t *testing.T) {
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/functions/v1/entities/invoice/hooks/on-before-create", r.URL.Path)
		body := readBody(t, r)
		require.JSONEq(t, `{"amount":100}`, string(body["record"]))
		_, _ = w.Write([]byte(`{"record":{"amount":100,"normalized":true}}`))
	})
	got, err := f.Dispatch.OnBeforeCreate(context.Background(), "invoice", json.RawMessage(`{"amount":100}`))
	require.NoError(t, err)
	require.JSONEq(t, `{"amount":100,"normalized":true}`, string(got))
}

func TestDispatch_OnBeforeUpdate_CarriesCurrentRecordOnTheWire(t *testing.T) {
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/functions/v1/entities/invoice/hooks/on-before-update", r.URL.Path)
		body := readBody(t, r)
		require.JSONEq(t, `{"amount":200}`, string(body["record"]))
		require.JSONEq(t, `{"amount":100}`, string(body["current_record"]))
		_, _ = w.Write([]byte(`{"record":{"amount":200}}`))
	})
	_, err := f.Dispatch.OnBeforeUpdate(context.Background(), "invoice",
		json.RawMessage(`{"amount":200}`),
		json.RawMessage(`{"amount":100}`),
	)
	require.NoError(t, err)
}

func TestDispatch_OnBeforeDelete_ReturnsNoContent(t *testing.T) {
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/functions/v1/entities/invoice/hooks/on-before-delete", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})
	err := f.Dispatch.OnBeforeDelete(context.Background(), "invoice", json.RawMessage(`{"id":"r1"}`))
	require.NoError(t, err)
}

func TestDispatch_OnAfterCreate(t *testing.T) {
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/functions/v1/entities/invoice/hooks/on-after-create", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, f.Dispatch.OnAfterCreate(context.Background(), "invoice", json.RawMessage(`{"id":"r1"}`)))
}

func TestDispatch_OnAfterUpdate_CarriesPreviousRecordOnTheWire(t *testing.T) {
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/functions/v1/entities/invoice/hooks/on-after-update", r.URL.Path)
		body := readBody(t, r)
		require.JSONEq(t, `{"amount":200}`, string(body["record"]))
		require.JSONEq(t, `{"amount":100}`, string(body["previous_record"]))
		w.WriteHeader(http.StatusNoContent)
	})
	err := f.Dispatch.OnAfterUpdate(context.Background(), "invoice",
		json.RawMessage(`{"amount":200}`),
		json.RawMessage(`{"amount":100}`),
	)
	require.NoError(t, err)
}

func TestDispatch_OnAfterDelete(t *testing.T) {
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/functions/v1/entities/invoice/hooks/on-after-delete", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, f.Dispatch.OnAfterDelete(context.Background(), "invoice", json.RawMessage(`{"id":"r1"}`)))
}

func TestDispatch_InvokeEntityAction_PathAndResultUnwrap(t *testing.T) {
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/functions/v1/entities/invoice/records/r1/actions/send-invoice/invoke", r.URL.Path)
		raw, _ := io.ReadAll(r.Body)
		require.JSONEq(t, `{"recipientEmail":"a@b.com"}`, string(raw))
		_, _ = w.Write([]byte(`{"result":{"messageId":"msg-1"}}`))
	})
	got, err := f.Dispatch.InvokeEntityAction(context.Background(), "invoice", "r1", "send-invoice",
		json.RawMessage(`{"recipientEmail":"a@b.com"}`),
	)
	require.NoError(t, err)
	require.JSONEq(t, `{"messageId":"msg-1"}`, string(got))
}

func TestDispatch_InvokeGlobalAction_PathAndResultUnwrap(t *testing.T) {
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/functions/v1/actions/rebuild-search-index/invoke", r.URL.Path)
		raw, _ := io.ReadAll(r.Body)
		require.JSONEq(t, `{}`, string(raw))
		_, _ = w.Write([]byte(`{"result":{"rebuiltCount":42}}`))
	})
	got, err := f.Dispatch.InvokeGlobalAction(context.Background(), "rebuild-search-index", json.RawMessage(`{}`))
	require.NoError(t, err)
	require.JSONEq(t, `{"rebuiltCount":42}`, string(got))
}
