// Package data provides the records and queries services backed by the
// data-service.
package data

import (
	"go.proteos.ai/model/data"
	dataapi "go.proteos.ai/model/data/api"
)

// ListRecordsOptions filters and paginates GET /data/v1/records/{entitySlug}.
//
// Filters carries arbitrary attribute filters as flat query-string params.
// Operators use bracket syntax: [eq], [ne], [gt], [gte], [lt], [lte], [in]
// (pipe-separated), [not_in], [contains], [starts_with], [ends_with],
// [empty], [not_empty]. The default operator is [eq].
//
// Example:
//
//	ListRecordsOptions{
//	    Page: 0, PageSize: 50, Sort: "created_at:desc",
//	    Filters: map[string]any{
//	        "name[contains]": "alice",
//	        "age[gte]":       21,
//	    },
//	}
type ListRecordsOptions struct {
	Page     int            `query:"page"               json:"page"`
	PageSize int            `query:"page_size"           json:"page_size"`
	Sort     string         `query:"sort,omitempty"     json:"sort,omitempty"`
	Filters  map[string]any `query:",flatten"           json:"filters,omitempty"`
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
