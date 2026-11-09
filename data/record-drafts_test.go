package data_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	datamodel "go.proteos.ai/model/data"
	dataapi "go.proteos.ai/model/data/api"
	sdk "go.proteos.ai/sdk"
	"go.proteos.ai/sdk/data"
)

const proposedDraftJSON = `{"id":"d1","org_id":"o1","entity_slug":"customer","intent":"update","status":"proposed","summary":"Rename Alice","items":[{"transaction_id":"1","id":"r1","data":{"name":"Alicia"}}],"created_at":"2026-11-06T00:00:00Z","created_by":{"type":"user","id":"u1"},"updated_at":"2026-11-06T00:00:00Z","updated_by":{"type":"user","id":"u1"}}`

func TestRecordDraftService_ListPage_Query(t *testing.T) {
	var seenPath string
	var seenQuery url.Values
	_, d := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		seenPath = r.URL.Path
		seenQuery = r.URL.Query()
		_, _ = w.Write([]byte(`{"meta":{"page":0,"page_size":20,"items_total":1,"pages_total":1},"data":[` + proposedDraftJSON + `]}`))
	})
	got, err := d.RecordDrafts.ListPage(context.Background(), &data.ListRecordDraftsOptions{
		PageSize:   20,
		EntitySlug: "customer",
		Status:     "proposed",
		Intent:     "update",
		RecordId:   "r1",
	})
	require.NoError(t, err)
	require.Equal(t, "/data/v1/record-drafts", seenPath)
	require.Equal(t, "20", seenQuery.Get("page_size"))
	require.Equal(t, "customer", seenQuery.Get("entity_slug"))
	require.Equal(t, "proposed", seenQuery.Get("status"))
	require.Equal(t, "update", seenQuery.Get("intent"))
	require.Equal(t, "r1", seenQuery.Get("record_id"))
	require.Len(t, got.Data, 1)
	require.Equal(t, datamodel.RecordDraftStatusProposed, got.Data[0].Status)
	require.Equal(t, datamodel.RecordDraftIntentUpdate, got.Data[0].Intent)
	require.Equal(t, "Alicia", got.Data[0].Items[0].Data["name"])
}

func TestRecordDraftService_ListPage_OmitsEmptyFilters(t *testing.T) {
	var seenQuery url.Values
	_, d := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		seenQuery = r.URL.Query()
		_, _ = w.Write([]byte(`{"meta":{"page":0,"page_size":100,"items_total":0,"pages_total":0},"data":[]}`))
	})
	_, err := d.RecordDrafts.ListPage(context.Background(), &data.ListRecordDraftsOptions{PageSize: 100})
	require.NoError(t, err)
	require.Equal(t, "100", seenQuery.Get("page_size"))
	for _, key := range []string{"entity_slug", "status", "intent", "record_id"} {
		_, present := seenQuery[key]
		require.False(t, present, "query param %q must be omitted when empty", key)
	}
}

func TestRecordDraftService_Get(t *testing.T) {
	_, d := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/data/v1/record-drafts/d1", r.URL.Path)
		_, _ = w.Write([]byte(proposedDraftJSON))
	})
	got, err := d.RecordDrafts.Get(context.Background(), "d1")
	require.NoError(t, err)
	require.Equal(t, "d1", got.Id)
	require.Equal(t, "customer", got.EntitySlug)
	require.Equal(t, "Rename Alice", got.Summary)
}

func TestRecordDraftService_Get_NotFound(t *testing.T) {
	_, d := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":"record_draft_not_found","message":"missing"}`))
	})
	_, err := d.RecordDrafts.Get(context.Background(), "missing")
	require.True(t, sdk.IsNotFound(err))
}

func TestRecordDraftService_Create_PostsBody(t *testing.T) {
	var seen dataapi.CreateRecordDraftRequest
	_, d := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/data/v1/record-drafts", r.URL.Path)
		raw, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(raw, &seen))
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(proposedDraftJSON))
	})
	got, err := d.RecordDrafts.Create(context.Background(), data.CreateRecordDraftRequest{
		EntitySlug: "customer",
		Intent:     datamodel.RecordDraftIntentUpdate,
		Items:      []data.RecordDraftItem{{Id: "r1", Data: map[string]any{"name": "Alicia"}}},
		Summary:    "Rename Alice",
		Source:     map[string]any{"agent_key": "crm-assistant"},
	})
	require.NoError(t, err)
	require.Equal(t, "d1", got.Id)
	require.Equal(t, "customer", seen.EntitySlug)
	require.Equal(t, datamodel.RecordDraftIntentUpdate, seen.Intent)
	require.Len(t, seen.Items, 1)
	require.Equal(t, "r1", seen.Items[0].Id)
	require.Equal(t, "Alicia", seen.Items[0].Data["name"])
	require.Equal(t, "Rename Alice", seen.Summary)
	require.Equal(t, "crm-assistant", seen.Source["agent_key"])
}

func TestRecordDraftService_Approve_AsProposed(t *testing.T) {
	var seenBody []byte
	_, d := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/data/v1/record-drafts/d1/approve", r.URL.Path)
		seenBody, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(`{"id":"d1","status":"approved","items":[],"results":[{"transaction_id":"1","status":"success","record_id":"r1"}]}`))
	})
	got, err := d.RecordDrafts.Approve(context.Background(), "d1", data.ApproveRecordDraftRequest{})
	require.NoError(t, err)
	require.JSONEq(t, `{}`, string(seenBody))
	require.Equal(t, datamodel.RecordDraftStatusApproved, got.Status)
	require.Equal(t, datamodel.RecordDraftItemResultStatusSuccess, got.Results[0].Status)
	require.Equal(t, "r1", got.Results[0].RecordId)
}

func TestRecordDraftService_Approve_WithItems(t *testing.T) {
	var seen dataapi.ApproveRecordDraftRequest
	_, d := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/data/v1/record-drafts/d1/approve", r.URL.Path)
		raw, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(raw, &seen))
		_, _ = w.Write([]byte(`{"id":"d1","status":"approved_with_changes","items":[],"changes":[{"transaction_id":"1","kind":"attribute_changed","attribute":"name","current":"Alice","proposed":"Alicia","applied":"Alicia Smith"}]}`))
	})
	got, err := d.RecordDrafts.Approve(context.Background(), "d1", data.ApproveRecordDraftRequest{
		Items: []data.RecordDraftItem{{TransactionId: "1", Id: "r1", Data: map[string]any{"name": "Alicia Smith"}}},
	})
	require.NoError(t, err)
	require.Len(t, seen.Items, 1)
	require.Equal(t, "1", seen.Items[0].TransactionId)
	require.Equal(t, "Alicia Smith", seen.Items[0].Data["name"])
	require.Equal(t, datamodel.RecordDraftStatusApprovedWithChanges, got.Status)
	require.Len(t, got.Changes, 1)
	require.Equal(t, datamodel.RecordDraftChangeAttributeChanged, got.Changes[0].Kind)
	require.Equal(t, "Alice", got.Changes[0].Current)
	require.Equal(t, "Alicia Smith", got.Changes[0].Applied)
}

func TestRecordDraftService_Approve_Conflict(t *testing.T) {
	_, d := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"code":"record_draft_not_proposed","message":"already resolved"}`))
	})
	_, err := d.RecordDrafts.Approve(context.Background(), "d1", data.ApproveRecordDraftRequest{})
	require.True(t, sdk.IsConflict(err))
}

func TestRecordDraftService_Reject_PostsReason(t *testing.T) {
	var seen dataapi.RejectRecordDraftRequest
	_, d := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/data/v1/record-drafts/d1/reject", r.URL.Path)
		raw, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(raw, &seen))
		_, _ = w.Write([]byte(`{"id":"d1","status":"rejected","items":[],"reason":"wrong customer"}`))
	})
	got, err := d.RecordDrafts.Reject(context.Background(), "d1", data.RejectRecordDraftRequest{Reason: "wrong customer"})
	require.NoError(t, err)
	require.Equal(t, "wrong customer", seen.Reason)
	require.Equal(t, datamodel.RecordDraftStatusRejected, got.Status)
	require.NotNil(t, got.Reason)
	require.Equal(t, "wrong customer", *got.Reason)
}
