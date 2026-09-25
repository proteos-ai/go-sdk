// Package data provides the records and queries services backed by the
// data-service.
package data

import (
	"go.proteos.ai/model/common"
	datamodel "go.proteos.ai/model/data"
	dataapi "go.proteos.ai/model/data/api"
)

// ListRecordsOptions filters and paginates GET /data/v1/records/{entitySlug}.
//
// Filters carries arbitrary attribute filters as flat query-string params.
// Operators use bracket syntax: [eq], [ne], [gt], [gte], [lt], [lte], [in]
// (pipe-separated), [not_in], [contains], [not_contains], [contains_all]
// (pipe-separated, arrays only), [starts_with], [ends_with], [empty],
// [not_empty]. The default operator is [eq]. On array attributes
// contains / not_contains / in / not_in / contains_all are set-membership
// tests (has / does not have / has any of / has none of / has all of). Flat filters are
// AND-combined; an attribute may reach one hop through a relation with a
// dotted path ("company_id.name[contains]").
//
// Filter carries a nested filter-group tree (AND/OR groups, the same wire
// dialect as visible_when and List.filters). It is sent as the data-service's
// `_filter` JSON query param and composes with Filters: the server ANDs the
// tree with any flat params. Element values are strings (pipe-joined for
// in / not_in / contains_all), and fields may use the same dotted relation-hop paths.
//
// Example:
//
//	ListRecordsOptions{
//	    Page: 0, PageSize: 50, Sort: "created_at:desc",
//	    Filters: map[string]any{
//	        "name[contains]": "alice",
//	        "age[gte]":       21,
//	    },
//	    Filter: &common.FilterGroup{
//	        LogicalOperator: common.LogicalOperatorOr,
//	        Elements: []common.FilterElement{
//	            {Field: "stage", Operator: common.ComparisonOperatorEquals, Value: "won"},
//	            {Field: "company_id.name", Operator: common.ComparisonOperatorContains, Value: "acme"},
//	        },
//	    },
//	}
type ListRecordsOptions struct {
	Page     int            `query:"page"               json:"page"`
	PageSize int            `query:"page_size"           json:"page_size"`
	Sort     string         `query:"sort,omitempty"     json:"sort,omitempty"`
	Filters  map[string]any `query:",flatten"           json:"filters,omitempty"`
	// Filter is serialized into the `_filter` query param by ListPage — the
	// query encoder itself skips struct fields, hence the query:"-".
	Filter *common.FilterGroup `query:"-"                json:"filter,omitempty"`
}

// BatchTransactionStatus is the per-transaction status returned by batch
// operations.
type BatchTransactionStatus string

const (
	BatchTransactionSuccess BatchTransactionStatus = "success"
	BatchTransactionError   BatchTransactionStatus = "error"
)

// BatchTransactionErr describes a per-transaction error in a batch result.
type BatchTransactionErr struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// BatchUpsertTransaction is one transaction within a batch upsert. The
// client supplies transactionId so it can correlate result rows back to its
// source records.
type BatchUpsertTransaction struct {
	TransactionID string         `json:"transaction_id"`
	Data          map[string]any `json:"data"`
}

// BatchUpsertTransactionResult is one entry in the batch upsert response.
type BatchUpsertTransactionResult struct {
	TransactionID string                 `json:"transaction_id"`
	Status        BatchTransactionStatus `json:"status"`
	Record        datamodel.Record       `json:"record,omitempty"`
	Error         *BatchTransactionErr   `json:"error,omitempty"`
}

// BatchUpsertRecordsResponse is the response shape from batch upsert.
type BatchUpsertRecordsResponse struct {
	Results []BatchUpsertTransactionResult `json:"results"`
}

// Query types are re-exported from go.proteos.ai/model/data/api so the
// wasip1 guest SDK can share the exact JSON shapes without dragging the
// SDK's net/http dep into the wasm build.

type QueryRow = dataapi.QueryRow
type QueryExecuteMeta = dataapi.QueryExecuteMeta
type QueryExecuteResponse = dataapi.QueryExecuteResponse
type QueryValidateMeta = dataapi.QueryValidateMeta
type QueryValidateResponse = dataapi.QueryValidateResponse

// Record duplicates + contact-observation replay are re-exported from
// go.proteos.ai/model/data so callers share the exact wire shapes.

type RecordDuplicate = datamodel.RecordDuplicate
type RecordDuplicateStatus = datamodel.RecordDuplicateStatus
type PublishContactObservationsResponse = dataapi.PublishContactObservationsResponse

// ListRecordDuplicatesOptions filters GET /data/v1/records/{entitySlug}/duplicates:
// RecordId narrows to the pairs one record takes part in (either side);
// Status defaults to open on the server.
type ListRecordDuplicatesOptions struct {
	Page     int    `query:"page"`
	PageSize int    `query:"page_size"`
	RecordId string `query:"record_id,omitempty"`
	Status   string `query:"status,omitempty"`
}

// PublishContactObservationsOptions pages POST /data/v1/records/{entitySlug}/contact-observations.
type PublishContactObservationsOptions struct {
	Page     int `query:"page"`
	PageSize int `query:"page_size"`
}
