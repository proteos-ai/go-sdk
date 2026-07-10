package data_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	sdk "go.proteos.ai/sdk"
	"go.proteos.ai/sdk/data"
	datamodel "go.proteos.ai/model/data"
)

func newClient(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *data.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := sdk.NewClient(sdk.WithBaseURL(srv.URL), sdk.WithToken("t"))
	require.NoError(t, err)
	return srv, data.New(c)
}

func queryPage(r *http.Request) int {
	v := r.URL.Query().Get("page")
	if v == "" {
		return 0
	}
	n, _ := strconv.Atoi(v)
	return n
}

func TestRecordService_Get_Success(t *testing.T) {
	want := datamodel.Record{"id": "r1", "name": "Alice"}
	_, d := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/data/v1/records/customer/r1", r.URL.Path)
		_ = json.NewEncoder(w).Encode(want)
	})
	got, err := d.Records.Get(context.Background(), "customer", "r1")
	require.NoError(t, err)
	require.Equal(t, "Alice", got["name"])
}

func TestRecordService_Get_NotFound(t *testing.T) {
	_, d := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":"not_found","message":"missing"}`))
	})
	_, err := d.Records.Get(context.Background(), "customer", "missing")
	require.True(t, sdk.IsNotFound(err))
}

func TestRecordService_ListPage_DefaultURL(t *testing.T) {
	var seenURL string
	_, d := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		seenURL = r.URL.String()
		_, _ = w.Write([]byte(`{"meta":{"page":0,"page_size":100,"items_total":0,"pages_total":0},"data":[]}`))
	})
	_, err := d.Records.ListPage(context.Background(), "customer", &data.ListRecordsOptions{
		Page: 0, PageSize: 100,
	})
	require.NoError(t, err)
	require.Contains(t, seenURL, "/data/v1/records/customer?")
	require.Contains(t, seenURL, "page=0")
	require.Contains(t, seenURL, "page_size=100")
}

func TestRecordService_ListPage_FiltersFlattened(t *testing.T) {
	var seenURL string
	_, d := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		seenURL = r.URL.String()
		_, _ = w.Write([]byte(`{"meta":{"page":0,"page_size":10,"items_total":0,"pages_total":0},"data":[]}`))
	})
	_, err := d.Records.ListPage(context.Background(), "customer", &data.ListRecordsOptions{
		PageSize: 10,
		Sort:     "name:asc",
		Filters: map[string]any{
			"name[contains]": "alice",
			"age[gt]":        21,
			"status":         "active",
		},
	})
	require.NoError(t, err)
	q, err := url_unescape(seenURL)
	require.NoError(t, err)
	require.Contains(t, q, "name[contains]=alice")
	require.Contains(t, q, "age[gt]=21")
	require.Contains(t, q, "status=active")
	require.Contains(t, q, "sort=name:asc")
}

// url_unescape returns the URL with query-string components unescaped, to
// make Contains assertions readable.
func url_unescape(rawURL string) (string, error) {
	u, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	values := u.URL.Query()
	out := u.URL.Path + "?"
	first := true
	for k, vs := range values {
		for _, v := range vs {
			if !first {
				out += "&"
			}
			out += k + "=" + v
			first = false
		}
	}
	return out, nil
}

func TestRecordService_List_Iterates(t *testing.T) {
	pages := [][]datamodel.Record{
		{{"id": "r1"}, {"id": "r2"}},
		{{"id": "r3"}},
	}
	_, d := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		page := queryPage(r)
		var rows []datamodel.Record
		if page < len(pages) {
			rows = pages[page]
		}
		body := map[string]any{
			"meta": map[string]any{"page": page, "page_size": 2, "items_total": 3, "pages_total": 2},
			"data": rows,
		}
		_ = json.NewEncoder(w).Encode(body)
	})
	all, err := d.Records.List("customer", &data.ListRecordsOptions{PageSize: 2}).All(context.Background())
	require.NoError(t, err)
	require.Len(t, all, 3)
	require.Equal(t, "r1", all[0]["id"])
	require.Equal(t, "r3", all[2]["id"])
}

func TestRecordService_Create_PostsBody(t *testing.T) {
	var seenBody []byte
	_, d := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/data/v1/records/customer", r.URL.Path)
		seenBody, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(`{"id":"r1","name":"Alice"}`))
	})
	got, err := d.Records.Create(context.Background(), "customer", datamodel.Record{"name": "Alice", "age": 30})
	require.NoError(t, err)
	require.Equal(t, "r1", got["id"])
	require.Contains(t, string(seenBody), `"name":"Alice"`)
	require.Contains(t, string(seenBody), `"age":30`)
}

func TestRecordService_Update_PatchesBody(t *testing.T) {
	var seen []byte
	_, d := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		require.Equal(t, "/data/v1/records/customer/r1", r.URL.Path)
		seen, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(`{"id":"r1","name":"Alicia"}`))
	})
	_, err := d.Records.Update(context.Background(), "customer", "r1", datamodel.Record{"name": "Alicia"})
	require.NoError(t, err)
	require.Contains(t, string(seen), `"name":"Alicia"`)
}

func TestRecordService_Delete(t *testing.T) {
	_, d := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		require.Equal(t, "/data/v1/records/customer/r1", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, d.Records.Delete(context.Background(), "customer", "r1"))
}

func TestRecordService_BatchUpsert_PostsTransactions(t *testing.T) {
	var seen []byte
	want := data.BatchUpsertRecordsResponse{
		Results: []data.BatchUpsertTransactionResult{
			{TransactionID: "t1", Status: data.BatchTransactionSuccess, Record: datamodel.Record{"id": "r1"}},
			{TransactionID: "t2", Status: data.BatchTransactionError, Error: &data.BatchTransactionErr{Code: "X", Message: "no"}},
		},
	}
	_, d := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/data/v1/batch/records/customer/upsert", r.URL.Path)
		seen, _ = io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(want)
	})
	got, err := d.Records.BatchUpsert(context.Background(), "customer", []data.BatchUpsertTransaction{
		{TransactionID: "t1", Data: map[string]any{"name": "Alice"}},
		{TransactionID: "t2", Data: map[string]any{"name": "Bob"}},
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(got.Results))
	require.Equal(t, data.BatchTransactionSuccess, got.Results[0].Status)
	require.Equal(t, "X", got.Results[1].Error.Code)
	require.Contains(t, string(seen), `"transaction_id":"t1"`)
	require.Contains(t, string(seen), `"transaction_id":"t2"`)
}

func TestRecordService_List_DefaultPageSize(t *testing.T) {
	var seen string
	_, d := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"meta":{"page":0,"page_size":100,"items_total":0,"pages_total":0},"data":[]}`))
	})
	_, err := d.Records.List("customer", nil).All(context.Background())
	require.NoError(t, err)
	require.Contains(t, seen, fmt.Sprintf("page_size=%d", sdk.DefaultPageSize))
}
