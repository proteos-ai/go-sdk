package functions_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	sdk "go.proteos.ai/sdk"
	"go.proteos.ai/sdk/functions"
)

type sendInvoiceParams struct {
	RecipientEmail string `json:"recipient_email"`
}

type sendInvoiceResult struct {
	MessageId string `json:"message_id"`
}

func TestInvokeGlobalActionTyped_MarshalsParamsAndDecodesResult(t *testing.T) {
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/functions/v1/actions/send-invoice/invoke", r.URL.Path)
		raw, _ := io.ReadAll(r.Body)
		// Params are sent verbatim as the body — no envelope.
		require.JSONEq(t, `{"recipient_email":"a@b.com"}`, string(raw))
		_, _ = w.Write([]byte(`{"result":{"message_id":"msg-1"}}`))
	})

	got, err := functions.InvokeGlobalActionTyped[sendInvoiceParams, sendInvoiceResult](
		context.Background(), f.Dispatch, "send-invoice",
		sendInvoiceParams{RecipientEmail: "a@b.com"},
	)
	require.NoError(t, err)
	require.Equal(t, "msg-1", got.MessageId)
}

func TestInvokeEntityActionTyped_HitsEntityPathAndDecodesResult(t *testing.T) {
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/functions/v1/entities/invoice/records/r1/actions/send-invoice/invoke", r.URL.Path)
		raw, _ := io.ReadAll(r.Body)
		require.JSONEq(t, `{"recipient_email":"a@b.com"}`, string(raw))
		_, _ = w.Write([]byte(`{"result":{"message_id":"msg-2"}}`))
	})

	got, err := functions.InvokeEntityActionTyped[sendInvoiceParams, sendInvoiceResult](
		context.Background(), f.Dispatch, "invoice", "r1", "send-invoice",
		sendInvoiceParams{RecipientEmail: "a@b.com"},
	)
	require.NoError(t, err)
	require.Equal(t, "msg-2", got.MessageId)
}

// A nil pointer param must be sent as `{}`, not `null` — the invoke endpoint
// only normalizes an empty body, so `null` would reach the guest.
func TestInvokeGlobalActionTyped_NilPointerParamsSentAsEmptyObject(t *testing.T) {
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		require.JSONEq(t, `{}`, string(raw))
		_, _ = w.Write([]byte(`{"result":null}`))
	})

	_, err := functions.InvokeGlobalActionTyped[*sendInvoiceParams, sendInvoiceResult](
		context.Background(), f.Dispatch, "send-invoice", nil,
	)
	require.NoError(t, err)
}

// A guest that emits nothing yields a JSON null result; the typed invoke must
// decode that to the zero value, not error.
func TestInvokeGlobalActionTyped_NullResultYieldsZero(t *testing.T) {
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"result":null}`))
	})

	got, err := functions.InvokeGlobalActionTyped[sendInvoiceParams, sendInvoiceResult](
		context.Background(), f.Dispatch, "send-invoice", sendInvoiceParams{},
	)
	require.NoError(t, err)
	require.Equal(t, sendInvoiceResult{}, got)
}

func TestInvokeGlobalActionTyped_PropagatesNotFound(t *testing.T) {
	_, f := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":"not_found","message":"no such action"}`))
	})

	_, err := functions.InvokeGlobalActionTyped[sendInvoiceParams, sendInvoiceResult](
		context.Background(), f.Dispatch, "missing", sendInvoiceParams{},
	)
	require.Error(t, err)
	require.True(t, sdk.IsNotFound(err))
}
